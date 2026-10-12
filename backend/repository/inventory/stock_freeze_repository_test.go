package inventory

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"wms-api/config"
	model "wms-api/models/inventory"
)

func TestGrandFreezeSerializesStockWriters(t *testing.T) {
	if os.Getenv("WMS_INTEGRATION_TEST") != "1" {
		t.Skip("set WMS_INTEGRATION_TEST=1")
	}
	check := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	_ = godotenv.Load("../../.env")
	cfg, err := config.Load()
	check(err)
	cfg.Database.AutoCreate = false
	db, err := config.OpenDatabase(cfg.Database)
	check(err)
	sqlDB, err := db.DB()
	check(err)
	db = db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
	schema := fmt.Sprintf("freeze_concurrency_%d", time.Now().UnixNano())
	check(db.Exec("CREATE SCHEMA " + schema).Error)
	t.Cleanup(func() {
		if err := db.Exec("DROP SCHEMA " + schema + " CASCADE").Error; err != nil {
			t.Errorf("cleanup generated test schema: %v", err)
		}
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	begin := func() *gorm.DB {
		t.Helper()
		tx := db.WithContext(ctx).Begin()
		check(tx.Error)
		check(tx.Exec("SET LOCAL search_path TO " + schema).Error)
		t.Cleanup(func() { _ = tx.Rollback().Error })
		return tx
	}
	const owner = "10000000-0000-4000-8000-000000000001"
	const otherOwner = "10000000-0000-4000-8000-000000000002"
	const warehouse = "20000000-0000-4000-8000-000000000001"
	const location = "30000000-0000-4000-8000-000000000001"
	const item = "40000000-0000-4000-8000-000000000001"
	const status = "50000000-0000-4000-8000-000000000001"
	const uom = "60000000-0000-4000-8000-000000000001"
	setup := begin()
	check(setup.AutoMigrate(&model.InventoryBalance{}, &model.StockFreeze{}))
	for id, ownerID := range map[string]string{"FROZEN-BALANCE": owner, "OTHER-BALANCE": otherOwner} {
		check(setup.Create(&model.InventoryBalance{ID: id, OwnerID: ownerID, WarehouseID: warehouse, LocationID: location, ItemID: item, InventoryStatusID: status, UOMID: uom, OnHandQty: "10", ReservedQty: "0", VersionNo: 1}).Error)
	}
	check(setup.Commit().Error)
	pid := func(tx *gorm.DB) int {
		t.Helper()
		var id int
		check(tx.Raw("SELECT pg_backend_pid()").Scan(&id).Error)
		return id
	}
	waitForLock := func(backendPID int) {
		t.Helper()
		deadline := time.NewTimer(5 * time.Second)
		defer deadline.Stop()
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			var blocked bool
			check(db.WithContext(ctx).Raw("SELECT EXISTS (SELECT 1 FROM pg_locks WHERE pid=? AND locktype='advisory' AND NOT granted)", backendPID).Scan(&blocked).Error)
			if blocked {
				return
			}
			select {
			case <-ticker.C:
			case <-deadline.C:
				t.Fatalf("backend %d did not wait for the opposing scope lock", backendPID)
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
		}
	}
	await := func(result <-chan error) error {
		t.Helper()
		select {
		case err := <-result:
			return err
		case <-ctx.Done():
			t.Fatal(ctx.Err())
			return ctx.Err()
		}
	}

	// A write already in progress finishes before grand opname takes its snapshot.
	writer := begin()
	check(NewInventoryBalanceRepository(writer).SetOnHand(ctx, "FROZEN-BALANCE", "11"))
	snapshot := begin()
	snapshotPID := pid(snapshot)
	snapshotDone := make(chan error, 1)
	go func() {
		r := NewInventoryBalanceRepository(snapshot)
		if err := r.LockCountScope(ctx, owner, warehouse); err != nil {
			snapshotDone <- err
			return
		}
		var balance model.InventoryBalance
		if err := snapshot.Where("balance_id=?", "FROZEN-BALANCE").Take(&balance).Error; err != nil {
			snapshotDone <- err
			return
		}
		if balance.OnHandQty != "11.000000" || balance.VersionNo != 2 {
			snapshotDone <- fmt.Errorf("grand snapshot missed committed writer: quantity=%s version=%d", balance.OnHandQty, balance.VersionNo)
			return
		}
		if err := r.BeginCountFreeze(ctx, owner, warehouse, "GSO-CONCURRENT-ONE"); err != nil {
			snapshotDone <- err
			return
		}
		snapshotDone <- snapshot.Commit().Error
	}()
	waitForLock(snapshotPID)
	check(writer.Commit().Error)
	check(await(snapshotDone))
	release := begin()
	check(NewInventoryBalanceRepository(release).ReleaseCountFreeze(ctx, "GSO-CONCURRENT-ONE"))
	check(release.Commit().Error)

	// A freeze being created blocks later writes until commit, then rejects them.
	freezer := begin()
	freezerRepo := NewInventoryBalanceRepository(freezer)
	check(freezerRepo.LockCountScope(ctx, owner, warehouse))
	check(freezerRepo.BeginCountFreeze(ctx, owner, warehouse, "GSO-CONCURRENT-TWO"))
	lateWriter := begin()
	lateWriterPID := pid(lateWriter)
	lateDone := make(chan error, 1)
	go func() {
		err := NewInventoryBalanceRepository(lateWriter).SetQuantities(ctx, "FROZEN-BALANCE", "11", "1")
		if err == nil {
			err = lateWriter.Commit().Error
		} else {
			_ = lateWriter.Rollback().Error
		}
		lateDone <- err
	}()
	waitForLock(lateWriterPID)
	unaffected := begin()
	check(NewInventoryBalanceRepository(unaffected).SetOnHand(ctx, "OTHER-BALANCE", "12"))
	check(unaffected.Commit().Error)
	check(freezer.Commit().Error)
	if err := await(lateDone); !errors.Is(err, ErrStockFrozen) {
		t.Fatalf("writer arriving during freeze got %v, want ErrStockFrozen", err)
	}
	verify := begin()
	var frozen model.InventoryBalance
	check(verify.Where("balance_id=?", "FROZEN-BALANCE").Take(&frozen).Error)
	if frozen.OnHandQty != "11.000000" || frozen.ReservedQty != "0.000000" || frozen.VersionNo != 2 {
		t.Fatalf("blocked concurrent reservation changed stock: %+v", frozen)
	}
	var unaffectedBalance model.InventoryBalance
	check(verify.Where("balance_id=?", "OTHER-BALANCE").Take(&unaffectedBalance).Error)
	if unaffectedBalance.OnHandQty != "12.000000" {
		t.Fatalf("unrelated customer was blocked: %+v", unaffectedBalance)
	}
	check(verify.Commit().Error)
}
