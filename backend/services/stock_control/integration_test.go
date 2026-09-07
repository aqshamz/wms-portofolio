package stockcontrol

import (
	"context"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"math/big"
	"os"
	"testing"
	"time"
	"wms-api/config"
	inventorydto "wms-api/dto/inventory"
	dto "wms-api/dto/stock_control"
	authmodel "wms-api/models/authentication"
	master "wms-api/models/master"
	authrepo "wms-api/repository/authentication"
	inventoryrepo "wms-api/repository/inventory"
	masterrepo "wms-api/repository/master"
	repository "wms-api/repository/stock_control"
	inventory "wms-api/services/inventory"
)

func okay(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func wants(t *testing.T, err, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("got %v want %v", err, target)
	}
}
func pointer[T any](v T) *T { return &v }
func TestStockControlPostgreSQL(t *testing.T) {
	if os.Getenv("WMS_INTEGRATION_TEST") != "1" {
		t.Skip("set WMS_INTEGRATION_TEST=1")
	}
	for _, fresh := range []bool{false, true} {
		t.Run(fmt.Sprintf("fresh_%t", fresh), func(t *testing.T) { testStockControl(t, fresh) })
	}
}
func testStockControl(t *testing.T, fresh bool) {
	_ = godotenv.Load("../../.env")
	cfg, err := config.Load()
	okay(t, err)
	cfg.Database.AutoCreate = false
	db, err := config.OpenDatabase(cfg.Database)
	okay(t, err)
	sqlDB, err := db.DB()
	okay(t, err)
	defer sqlDB.Close()
	tx := db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Begin()
	okay(t, tx.Error)
	defer tx.Rollback()
	suffix := fmt.Sprint(time.Now().UnixNano())
	if fresh {
		schema := "stock_test_" + suffix
		okay(t, tx.Exec("CREATE SCHEMA "+schema).Error)
		okay(t, tx.Exec("SET LOCAL search_path TO "+schema).Error)
	}
	okay(t, authrepo.Migrate(tx))
	okay(t, masterrepo.Migrate(tx))
	okay(t, masterrepo.MigrateCatalog(tx))
	okay(t, inventoryrepo.Migrate(tx))
	okay(t, repository.Migrate(tx))
	ctx := context.Background()
	accountStatus := authmodel.AccountStatus{Code: "S" + suffix, Name: "Active", AllowsLogin: true}
	okay(t, tx.Create(&accountStatus).Error)
	account := authmodel.AppAccount{Username: "stock_" + suffix, DisplayName: "Stock", AccountStatusID: accountStatus.ID, ExternalSubject: pointer("stock_" + suffix)}
	okay(t, tx.Create(&account).Error)
	owner := master.Organization{Code: "O" + suffix, Name: "Owner", TimezoneName: "Asia/Jakarta"}
	okay(t, tx.Create(&owner).Error)
	unit := master.UOM{Code: "U" + suffix, Name: "Each"}
	okay(t, tx.Create(&unit).Error)
	item := master.Item{OwnerID: owner.ID, Code: "I" + suffix, Name: "Item", BaseUOMID: unit.ID}
	okay(t, tx.Create(&item).Error)
	wh1 := master.Warehouse{OperatorID: owner.ID, Code: "W1_" + suffix, Name: "Source", TimezoneName: "Asia/Jakarta"}
	wh2 := master.Warehouse{OperatorID: owner.ID, Code: "W2_" + suffix, Name: "Target", TimezoneName: "Asia/Jakarta"}
	okay(t, tx.Create(&wh1).Error)
	okay(t, tx.Create(&wh2).Error)
	okay(t, tx.Create(&master.WarehouseOwner{OwnerID: owner.ID, WarehouseID: wh1.ID}).Error)
	okay(t, tx.Create(&master.WarehouseOwner{OwnerID: owner.ID, WarehouseID: wh2.ID}).Error)
	lt := master.LocationType{Code: "L" + suffix, Name: "Storage", AllowsStorage: true}
	okay(t, tx.Create(&lt).Error)
	z1 := master.WarehouseZone{WarehouseID: wh1.ID, Code: "Z1", Name: "Zone 1"}
	z2 := master.WarehouseZone{WarehouseID: wh2.ID, Code: "Z2", Name: "Zone 2"}
	okay(t, tx.Create(&z1).Error)
	okay(t, tx.Create(&z2).Error)
	loc1 := master.WarehouseLocation{WarehouseID: wh1.ID, ZoneID: z1.ID, LocationTypeID: lt.ID, Code: "A"}
	loc2 := master.WarehouseLocation{WarehouseID: wh1.ID, ZoneID: z1.ID, LocationTypeID: lt.ID, Code: "B"}
	target := master.WarehouseLocation{WarehouseID: wh2.ID, ZoneID: z2.ID, LocationTypeID: lt.ID, Code: "C"}
	okay(t, tx.Create(&loc1).Error)
	okay(t, tx.Create(&loc2).Error)
	okay(t, tx.Create(&target).Error)
	available := master.InventoryStatus{Code: "A" + suffix, Name: "Available", IsAllocatable: true, IsPickable: true}
	hold := master.InventoryStatus{Code: "H" + suffix, Name: "Hold"}
	okay(t, tx.Create(&available).Error)
	okay(t, tx.Create(&hold).Error)
	inv := inventory.NewService(inventoryrepo.NewRepositories(tx))
	initial, err := inv.PostMovement(ctx, inventorydto.PostingRequest{OperationKey: "seed-" + suffix, MovementTypeCode: "RECEIVE", OwnerID: owner.ID, WarehouseID: wh1.ID, BusinessDate: "2026-09-07", ItemID: item.ID, To: &inventorydto.BalanceDimension{LocationID: loc1.ID, InventoryStatusID: available.ID}, Quantity: "10", SourceDocumentID: "TEST-SEED"}, account.ID)
	okay(t, err)
	service := NewService(repository.NewRepositories(tx))
	command := dto.CommandBase{OperationKey: "move-" + suffix, BusinessDate: "2026-09-07", SourceDocumentID: "MOVE-" + suffix}
	moved, err := service.InternalMove(ctx, dto.InternalMoveRequest{CommandBase: command, SourceBalanceID: initial.ToBalance.ID, TargetLocationID: loc2.ID, Quantity: "4", ExpectedVersion: initial.ToBalance.VersionNo}, account.ID)
	okay(t, err)
	if moved.SourceBalance.OnHandQty != "6.000000" || moved.DestinationBalance.OnHandQty != "4.000000" {
		t.Fatal("internal move quantities")
	}
	replay, err := service.InternalMove(ctx, dto.InternalMoveRequest{CommandBase: command, SourceBalanceID: initial.ToBalance.ID, TargetLocationID: loc2.ID, Quantity: "4", ExpectedVersion: initial.ToBalance.VersionNo}, account.ID)
	okay(t, err)
	if !replay.IdempotentReplay {
		t.Fatal("move replay")
	}
	stale := command
	stale.OperationKey = "stale-" + suffix
	_, err = service.InternalMove(ctx, dto.InternalMoveRequest{CommandBase: stale, SourceBalanceID: initial.ToBalance.ID, TargetLocationID: loc2.ID, Quantity: "1", ExpectedVersion: initial.ToBalance.VersionNo}, account.ID)
	wants(t, err, inventory.ErrInvalidInput)
	missingReason := dto.CommandBase{OperationKey: "missing-reason-" + suffix, BusinessDate: "2026-09-07", SourceDocumentID: "STATUS-MISSING-" + suffix}
	_, err = service.StatusChange(ctx, dto.StatusChangeRequest{CommandBase: missingReason, SourceBalanceID: moved.DestinationBalance.ID, TargetInventoryStatusID: hold.ID, Quantity: "1", ExpectedVersion: moved.DestinationBalance.VersionNo}, account.ID)
	wants(t, err, ErrInvalidInput)
	requiredNote := dto.CommandBase{OperationKey: "required-note-" + suffix, BusinessDate: "2026-09-07", SourceDocumentID: "ADJ-NOTE-" + suffix, ReasonCode: pointer("MANUAL_ADJUSTMENT")}
	_, err = service.Adjustment(ctx, dto.AdjustmentRequest{CommandBase: requiredNote, BalanceID: moved.DestinationBalance.ID, Direction: "INCREASE", Quantity: "1", ExpectedVersion: moved.DestinationBalance.VersionNo}, account.ID)
	wants(t, err, ErrInvalidInput)
	statusCommand := dto.CommandBase{OperationKey: "status-" + suffix, BusinessDate: "2026-09-07", SourceDocumentID: "STATUS-" + suffix, ReasonCode: pointer("EXPIRY")}
	statusResult, err := service.StatusChange(ctx, dto.StatusChangeRequest{CommandBase: statusCommand, SourceBalanceID: moved.DestinationBalance.ID, TargetInventoryStatusID: hold.ID, Quantity: "4", ExpectedVersion: moved.DestinationBalance.VersionNo}, account.ID)
	okay(t, err)
	adjustBase := dto.CommandBase{OperationKey: "inc-" + suffix, BusinessDate: "2026-09-07", SourceDocumentID: "ADJ-" + suffix, ReasonCode: pointer("EXPIRY")}
	increased, err := service.Adjustment(ctx, dto.AdjustmentRequest{CommandBase: adjustBase, BalanceID: statusResult.DestinationBalance.ID, Direction: "INCREASE", Quantity: "2", ExpectedVersion: statusResult.DestinationBalance.VersionNo}, account.ID)
	okay(t, err)
	adjustBase.OperationKey = "dec-" + suffix
	decreased, err := service.Adjustment(ctx, dto.AdjustmentRequest{CommandBase: adjustBase, BalanceID: increased.DestinationBalance.ID, Direction: "DECREASE", Quantity: "1", ExpectedVersion: increased.DestinationBalance.VersionNo}, account.ID)
	okay(t, err)
	countBase := dto.CommandBase{OperationKey: "count-" + suffix, BusinessDate: "2026-09-07", SourceDocumentID: "COUNT-" + suffix, ReasonCode: pointer("EXPIRY")}
	counted, err := service.ReconcileCount(ctx, dto.StockCountReconcileRequest{CommandBase: countBase, BalanceID: decreased.SourceBalance.ID, CountedQty: "3", ExpectedVersion: decreased.SourceBalance.VersionNo}, account.ID)
	okay(t, err)
	if counted.SourceBalance.OnHandQty != "3.000000" {
		t.Fatal("count correction")
	}
	countBase.OperationKey = "novariance-" + suffix
	noVariance, err := service.ReconcileCount(ctx, dto.StockCountReconcileRequest{CommandBase: countBase, BalanceID: counted.SourceBalance.ID, CountedQty: "3", ExpectedVersion: counted.SourceBalance.VersionNo}, account.ID)
	okay(t, err)
	if !noVariance.NoVariance || len(noVariance.Movements) != 0 {
		t.Fatal("no variance")
	}
	transferBase := dto.CommandBase{OperationKey: "transfer-" + suffix, BusinessDate: "2026-09-07", SourceDocumentID: "TRANSFER-" + suffix}
	transferred, err := service.WarehouseTransfer(ctx, dto.WarehouseTransferRequest{CommandBase: transferBase, SourceBalanceID: counted.SourceBalance.ID, TargetWarehouseID: wh2.ID, TargetLocationID: target.ID, TargetInventoryStatusID: available.ID, Quantity: "2", ExpectedVersion: counted.SourceBalance.VersionNo}, account.ID)
	okay(t, err)
	if len(transferred.Movements) != 2 || transferred.SourceBalance.OnHandQty != "1.000000" || transferred.DestinationBalance.OnHandQty != "2.000000" {
		t.Fatal("warehouse transfer")
	}
	transferReplay, err := service.WarehouseTransfer(ctx, dto.WarehouseTransferRequest{CommandBase: transferBase, SourceBalanceID: counted.SourceBalance.ID, TargetWarehouseID: wh2.ID, TargetLocationID: target.ID, TargetInventoryStatusID: available.ID, Quantity: "2", ExpectedVersion: counted.SourceBalance.VersionNo}, account.ID)
	okay(t, err)
	if !transferReplay.IdempotentReplay {
		t.Fatal("transfer replay")
	}
	balances, err := inv.ListBalances(ctx, inventoryrepo.BalanceFilter{OwnerID: owner.ID, Page: 1, PageSize: 100, IncludeZero: true})
	okay(t, err)
	var total big.Rat
	for _, balance := range balances.Items {
		value, ok := new(big.Rat).SetString(balance.OnHandQty)
		if !ok {
			t.Fatal("quantity")
		}
		total.Add(&total, value)
	}
	if total.Cmp(big.NewRat(9, 1)) != 0 {
		t.Fatalf("adjusted warehouse total=%s", total.RatString())
	}
	movements, err := inv.ListMovements(ctx, inventoryrepo.MovementFilter{OwnerID: owner.ID, Page: 1, PageSize: 100})
	okay(t, err)
	if movements.TotalItems != 8 {
		t.Fatalf("movements=%d", movements.TotalItems)
	}
}
