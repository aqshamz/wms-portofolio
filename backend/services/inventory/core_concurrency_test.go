package inventory

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"sync"
	"testing"
	"time"
	"wms-api/config"
	dto "wms-api/dto/inventory"
	authmodel "wms-api/models/authentication"
	master "wms-api/models/master"
	authrepo "wms-api/repository/authentication"
	repository "wms-api/repository/inventory"
	masterrepo "wms-api/repository/master"
)

func TestConcurrentPosting(t *testing.T) {
	if os.Getenv("WMS_INTEGRATION_TEST") != "1" {
		t.Skip("set WMS_INTEGRATION_TEST=1")
	}
	_ = godotenv.Load("../../.env")
	cfg, err := config.Load()
	ok(t, err)
	cfg.Database.AutoCreate = false
	db, err := config.OpenDatabase(cfg.Database)
	ok(t, err)
	sqlDB, err := db.DB()
	ok(t, err)
	db = db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
	schema := fmt.Sprintf("core_concurrency_%d", time.Now().UnixNano())
	ok(t, db.Exec("CREATE SCHEMA "+schema).Error)
	t.Cleanup(func() {
		if err := db.Exec("DROP SCHEMA " + schema + " CASCADE").Error; err != nil {
			t.Errorf("cleanup: %v", err)
		}
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close DB: %v", err)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	withSchema := func(work func(*Service, *gorm.DB) error) error {
		return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("SET LOCAL search_path TO " + schema).Error; err != nil {
				return err
			}
			return work(NewService(repository.NewRepositories(tx)), tx)
		})
	}
	var actor, owner, warehouse, item, location, status string
	ok(t, withSchema(func(_ *Service, tx *gorm.DB) error {
		if err := authrepo.Migrate(tx); err != nil {
			return err
		}
		if err := masterrepo.Migrate(tx); err != nil {
			return err
		}
		if err := masterrepo.MigrateCatalog(tx); err != nil {
			return err
		}
		if err := repository.Migrate(tx); err != nil {
			return err
		}
		accountStatus := authmodel.AccountStatus{Code: "ACTIVE", Name: "Active", AllowsLogin: true}
		if err := tx.Create(&accountStatus).Error; err != nil {
			return err
		}
		account := authmodel.AppAccount{Username: "concurrent", DisplayName: "Concurrent", AccountStatusID: accountStatus.ID, ExternalSubject: ptr("concurrent")}
		if err := tx.Create(&account).Error; err != nil {
			return err
		}
		actor = account.ID
		org := master.Organization{Code: "OWNER", Name: "Owner", TimezoneName: "Asia/Jakarta"}
		if err := tx.Create(&org).Error; err != nil {
			return err
		}
		owner = org.ID
		unit := master.UOM{Code: "EA", Name: "Each"}
		if err := tx.Create(&unit).Error; err != nil {
			return err
		}
		product := master.Item{OwnerID: org.ID, Code: "ITEM", Name: "Item", BaseUOMID: unit.ID}
		if err := tx.Create(&product).Error; err != nil {
			return err
		}
		item = product.ID
		wh := master.Warehouse{OperatorID: org.ID, Code: "WH", Name: "Warehouse", TimezoneName: "Asia/Jakarta"}
		if err := tx.Create(&wh).Error; err != nil {
			return err
		}
		warehouse = wh.ID
		if err := tx.Create(&master.WarehouseOwner{OwnerID: org.ID, WarehouseID: wh.ID}).Error; err != nil {
			return err
		}
		lt := master.LocationType{Code: "STORAGE", Name: "Storage", AllowsStorage: true}
		if err := tx.Create(&lt).Error; err != nil {
			return err
		}
		zone := master.WarehouseZone{WarehouseID: wh.ID, Code: "ZONE", Name: "Zone"}
		if err := tx.Create(&zone).Error; err != nil {
			return err
		}
		loc := master.WarehouseLocation{WarehouseID: wh.ID, ZoneID: zone.ID, LocationTypeID: lt.ID, Code: "A"}
		if err := tx.Create(&loc).Error; err != nil {
			return err
		}
		location = loc.ID
		available := master.InventoryStatus{Code: "AVAILABLE", Name: "Available", IsAllocatable: true, IsPickable: true}
		if err := tx.Create(&available).Error; err != nil {
			return err
		}
		status = available.ID
		return nil
	}))
	post := func(key string) (dto.PostingResult, error) {
		var result dto.PostingResult
		err := withSchema(func(s *Service, _ *gorm.DB) error {
			var e error
			result, e = s.PostMovement(ctx, dto.PostingRequest{OperationKey: key, MovementTypeCode: "RECEIVE", OwnerID: owner, WarehouseID: warehouse, BusinessDate: "2026-09-07", ItemID: item, To: &dto.BalanceDimension{LocationID: location, InventoryStatusID: status}, Quantity: "1", SourceDocumentID: "CONCURRENT-RECEIPT"}, actor)
			return e
		})
		return result, err
	}
	run := func(keys []string) []dto.PostingResult {
		results := make(chan dto.PostingResult, len(keys))
		failures := make(chan error, len(keys))
		var wg sync.WaitGroup
		for _, key := range keys {
			wg.Add(1)
			go func(k string) {
				defer wg.Done()
				result, err := post(k)
				if err != nil {
					failures <- err
				} else {
					results <- result
				}
			}(key)
		}
		wg.Wait()
		close(results)
		close(failures)
		for err := range failures {
			t.Error(err)
		}
		output := make([]dto.PostingResult, 0, len(keys))
		for result := range results {
			output = append(output, result)
		}
		return output
	}
	keys := make([]string, 12)
	for i := range keys {
		keys[i] = fmt.Sprintf("parallel-%02d", i)
	}
	if got := run(keys); len(got) != 12 {
		t.Fatalf("distinct postings=%d", len(got))
	}
	replayKeys := make([]string, 12)
	for i := range replayKeys {
		replayKeys[i] = "one-operation"
	}
	got := run(replayKeys)
	if len(got) != 12 {
		t.Fatalf("replays=%d", len(got))
	}
	seen := map[string]bool{}
	for _, v := range got {
		seen[v.Movement.ID] = true
	}
	if len(seen) != 1 {
		t.Fatalf("idempotency produced %d movements", len(seen))
	}
	ok(t, withSchema(func(s *Service, _ *gorm.DB) error {
		balances, err := s.ListBalances(ctx, repository.BalanceFilter{OwnerID: owner, WarehouseID: warehouse, ItemID: item, Page: 1, PageSize: 20})
		if err != nil {
			return err
		}
		if balances.TotalItems != 1 || balances.Items[0].OnHandQty != "13.000000" {
			return fmt.Errorf("bad concurrent balance: %+v", balances)
		}
		movements, err := s.ListMovements(ctx, repository.MovementFilter{OwnerID: owner, WarehouseID: warehouse, Page: 1, PageSize: 20})
		if err != nil {
			return err
		}
		if movements.TotalItems != 13 {
			return fmt.Errorf("movements=%d", movements.TotalItems)
		}
		return nil
	}))
}
