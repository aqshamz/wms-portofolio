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
	inventorymodel "wms-api/models/inventory"
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
	okay(t, masterrepo.MigrateOperational(tx))
	okay(t, inventoryrepo.Migrate(tx))
	okay(t, repository.Migrate(tx))
	okay(t, repository.MigrateReplenishments(tx))
	okay(t, repository.MigrateInventoryAdjustments(tx))
	okay(t, repository.MigrateInventoryAdjustmentLines(tx))
	ctx := context.Background()
	inventoryModule := master.AppModule{Code: "INVENTORY", Name: "Inventory", IsActive: true}
	okay(t, tx.Where("code=?", inventoryModule.Code).FirstOrCreate(&inventoryModule).Error)
	adjustmentType := master.DocumentType{Code: "INVENTORY_ADJUSTMENT", Name: "Inventory adjustment", ModuleCode: "INVENTORY", IsActive: true}
	okay(t, tx.Where("code=?", adjustmentType.Code).FirstOrCreate(&adjustmentType).Error)
	for _, status := range []master.DocumentStatus{
		{DocumentTypeID: adjustmentType.ID, Code: "DRAFT", Name: "Draft", IsInitial: true, IsActive: true},
		{DocumentTypeID: adjustmentType.ID, Code: "PARTIALLY_POSTED", Name: "Partially posted", IsActive: true},
		{DocumentTypeID: adjustmentType.ID, Code: "POSTED", Name: "Posted", IsFinal: true, IsActive: true},
		{DocumentTypeID: adjustmentType.ID, Code: "CANCELLED", Name: "Cancelled", IsFinal: true, IsCancelled: true, IsActive: true},
	} {
		okay(t, tx.Where("document_type_id=? AND code=?", status.DocumentTypeID, status.Code).FirstOrCreate(&status).Error)
	}
	for _, status := range []master.TaskStatus{
		{Code: "OPEN", Name: "Open", IsInitial: true, IsActive: true},
		{Code: "ASSIGNED", Name: "Assigned", IsActive: true},
		{Code: "IN_PROGRESS", Name: "In progress", IsActive: true},
		{Code: "COMPLETED", Name: "Completed", IsFinal: true, IsActive: true},
		{Code: "CANCELLED", Name: "Cancelled", IsFinal: true, IsCancelled: true, IsActive: true},
	} {
		okay(t, tx.Where("code=?", status.Code).FirstOrCreate(&status).Error)
	}
	taskType := master.TaskType{Code: "REPLENISHMENT", Name: "Replenishment", IsActive: true}
	okay(t, tx.Where("code=?", taskType.Code).FirstOrCreate(&taskType).Error)
	priority := master.TaskPriority{Code: "NORMAL", Name: "Normal", PriorityValue: 200, IsActive: true}
	okay(t, tx.Where("code=?", priority.Code).FirstOrCreate(&priority).Error)
	accountStatus := authmodel.AccountStatus{Code: "S" + suffix, Name: "Active", AllowsLogin: true}
	okay(t, tx.Create(&accountStatus).Error)
	account := authmodel.AppAccount{Username: "stock_" + suffix, DisplayName: "Stock", AccountStatusID: accountStatus.ID, ExternalSubject: pointer("stock_" + suffix)}
	okay(t, tx.Create(&account).Error)
	approver := authmodel.AppAccount{Username: "stock_approver_" + suffix, DisplayName: "Approver", AccountStatusID: accountStatus.ID, ExternalSubject: pointer("stock_approver_" + suffix)}
	okay(t, tx.Create(&approver).Error)
	deniedAccount := authmodel.AppAccount{Username: "stock_denied_" + suffix, DisplayName: "Denied", AccountStatusID: accountStatus.ID, ExternalSubject: pointer("stock_denied_" + suffix)}
	okay(t, tx.Create(&deniedAccount).Error)
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
	okay(t, tx.Create(&master.AccountOwnerAccess{AccountID: account.ID, OwnerID: owner.ID}).Error)
	okay(t, tx.Create(&master.AccountWarehouseAccess{AccountID: account.ID, WarehouseID: wh1.ID}).Error)
	okay(t, tx.Create(&master.AccountWarehouseAccess{AccountID: account.ID, WarehouseID: wh2.ID}).Error)
	okay(t, tx.Create(&master.AccountOwnerAccess{AccountID: approver.ID, OwnerID: owner.ID}).Error)
	okay(t, tx.Create(&master.AccountWarehouseAccess{AccountID: approver.ID, WarehouseID: wh1.ID}).Error)
	storageType := master.LocationType{Code: "STORAGE", Name: "Storage", AllowsStorage: true, IsActive: true}
	okay(t, tx.Where("code=?", storageType.Code).FirstOrCreate(&storageType).Error)
	pickType := master.LocationType{Code: "PICK_FACE", Name: "Pick face", AllowsStorage: true, AllowsPicking: true, IsActive: true}
	okay(t, tx.Where("code=?", pickType.Code).FirstOrCreate(&pickType).Error)
	receivingType := master.LocationType{Code: "RECEIVING", Name: "Receiving", AllowsReceiving: true, IsActive: true}
	okay(t, tx.Where("code=?", receivingType.Code).FirstOrCreate(&receivingType).Error)
	z1 := master.WarehouseZone{WarehouseID: wh1.ID, Code: "Z1", Name: "Zone 1"}
	z2 := master.WarehouseZone{WarehouseID: wh2.ID, Code: "Z2", Name: "Zone 2"}
	okay(t, tx.Create(&z1).Error)
	okay(t, tx.Create(&z2).Error)
	loc1 := master.WarehouseLocation{WarehouseID: wh1.ID, ZoneID: z1.ID, LocationTypeID: storageType.ID, Code: "A"}
	loc2 := master.WarehouseLocation{WarehouseID: wh1.ID, ZoneID: z1.ID, LocationTypeID: storageType.ID, Code: "B"}
	pickFace := master.WarehouseLocation{WarehouseID: wh1.ID, ZoneID: z1.ID, LocationTypeID: pickType.ID, Code: "P", IsPickFace: true}
	receiving := master.WarehouseLocation{WarehouseID: wh1.ID, ZoneID: z1.ID, LocationTypeID: receivingType.ID, Code: "R"}
	target := master.WarehouseLocation{WarehouseID: wh2.ID, ZoneID: z2.ID, LocationTypeID: storageType.ID, Code: "C"}
	okay(t, tx.Create(&loc1).Error)
	okay(t, tx.Create(&loc2).Error)
	okay(t, tx.Create(&pickFace).Error)
	okay(t, tx.Create(&receiving).Error)
	okay(t, tx.Create(&target).Error)
	available := master.InventoryStatus{Code: "AVAILABLE", Name: "Available", IsAllocatable: true, IsPickable: true, IsActive: true}
	hold := master.InventoryStatus{Code: "H" + suffix, Name: "Hold"}
	okay(t, tx.Where("code=?", available.Code).FirstOrCreate(&available).Error)
	okay(t, tx.Create(&hold).Error)
	inv := inventory.NewService(inventoryrepo.NewRepositories(tx))
	initial, err := inv.PostMovement(ctx, inventorydto.PostingRequest{OperationKey: "seed-" + suffix, MovementTypeCode: "RECEIVE", OwnerID: owner.ID, WarehouseID: wh1.ID, BusinessDate: "2026-09-07", ItemID: item.ID, To: &inventorydto.BalanceDimension{LocationID: loc1.ID, InventoryStatusID: available.ID}, Quantity: "10", SourceDocumentID: "TEST-SEED"}, account.ID)
	okay(t, err)
	service := NewService(repository.NewRepositories(tx))
	command := dto.CommandBase{OperationKey: "move-" + suffix, BusinessDate: "2026-09-07", SourceDocumentID: "MOVE-" + suffix}
	invalidTarget := command
	invalidTarget.OperationKey = "move-receiving-" + suffix
	_, err = service.InternalMove(ctx, dto.InternalMoveRequest{CommandBase: invalidTarget, SourceBalanceID: initial.ToBalance.ID, TargetLocationID: receiving.ID, Quantity: "1", ExpectedVersion: initial.ToBalance.VersionNo}, account.ID)
	wants(t, err, ErrInvalidInput)
	moved, err := service.InternalMove(ctx, dto.InternalMoveRequest{CommandBase: command, SourceBalanceID: initial.ToBalance.ID, TargetLocationID: loc2.ID, Quantity: "4", ExpectedVersion: initial.ToBalance.VersionNo}, account.ID)
	okay(t, err)
	if moved.SourceBalance.OnHandQty != "6.000000" || moved.DestinationBalance.OnHandQty != "4.000000" {
		t.Fatal("internal move quantities")
	}
	deniedCommand := command
	deniedCommand.OperationKey = "move-denied-" + suffix
	_, err = service.InternalMove(ctx, dto.InternalMoveRequest{CommandBase: deniedCommand, SourceBalanceID: moved.SourceBalance.ID, TargetLocationID: loc2.ID, Quantity: "1", ExpectedVersion: moved.SourceBalance.VersionNo}, deniedAccount.ID)
	wants(t, err, ErrForbidden)
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
	receivingBalance := inventorymodel.InventoryBalance{ID: "BAL-RECEIVING-" + suffix, OwnerID: owner.ID, WarehouseID: wh1.ID, LocationID: receiving.ID, ItemID: item.ID, InventoryStatusID: available.ID, OnHandQty: "1", ReservedQty: "0", UOMID: unit.ID, VersionNo: 1}
	okay(t, tx.Create(&receivingBalance).Error)
	_, err = service.CreateAdjustment(ctx, dto.CreateAdjustmentRequest{BusinessDate: "2026-09-07", Direction: "DECREASE", ReasonCode: "EXPIRY", Lines: []dto.CreateAdjustmentLineRequest{{BalanceID: receivingBalance.ID, Quantity: "1", ExpectedVersion: 1}}}, account.ID)
	wants(t, err, ErrInvalidInput)
	increaseRequest, err := service.CreateAdjustment(ctx, dto.CreateAdjustmentRequest{BusinessDate: "2026-09-07", Direction: "INCREASE", ReasonCode: "EXPIRY", Lines: []dto.CreateAdjustmentLineRequest{
		{BalanceID: statusResult.DestinationBalance.ID, Quantity: "2", ExpectedVersion: statusResult.DestinationBalance.VersionNo},
		{BalanceID: statusResult.SourceBalance.ID, Quantity: "1", ExpectedVersion: statusResult.SourceBalance.VersionNo},
	}}, account.ID)
	okay(t, err)
	staleRequest, err := service.CreateAdjustment(ctx, dto.CreateAdjustmentRequest{BusinessDate: "2026-09-07", Direction: "DECREASE", ReasonCode: "EXPIRY", Lines: []dto.CreateAdjustmentLineRequest{{BalanceID: statusResult.DestinationBalance.ID, Quantity: "1", ExpectedVersion: statusResult.DestinationBalance.VersionNo}}}, account.ID)
	okay(t, err)
	unchanged, err := inv.GetBalance(ctx, statusResult.DestinationBalance.ID)
	okay(t, err)
	if unchanged.OnHandQty != statusResult.DestinationBalance.OnHandQty || unchanged.VersionNo != statusResult.DestinationBalance.VersionNo {
		t.Fatal("draft adjustment changed stock")
	}
	_, err = service.ApproveAdjustment(ctx, increaseRequest.ID, dto.AdjustmentLineSelectionRequest{ExpectedVersion: increaseRequest.VersionNo, LineIDs: []string{increaseRequest.Lines[0].ID}}, account.ID)
	wants(t, err, ErrInvalidInput)
	increaseRequest, err = service.ApproveAdjustment(ctx, increaseRequest.ID, dto.AdjustmentLineSelectionRequest{ExpectedVersion: increaseRequest.VersionNo, LineIDs: []string{increaseRequest.Lines[0].ID}}, approver.ID)
	okay(t, err)
	if increaseRequest.StatusCode != "PARTIALLY_POSTED" || increaseRequest.PostedLines != 1 || increaseRequest.PendingLines != 1 || increaseRequest.Lines[0].InventoryMovementID == nil || increaseRequest.Lines[0].ApprovedBy == nil || *increaseRequest.Lines[0].ApprovedBy != approver.ID {
		t.Fatal("adjustment approval audit")
	}
	increaseRequest, err = service.ApproveAdjustment(ctx, increaseRequest.ID, dto.AdjustmentLineSelectionRequest{ExpectedVersion: increaseRequest.VersionNo, LineIDs: []string{increaseRequest.Lines[1].ID}}, approver.ID)
	okay(t, err)
	if increaseRequest.StatusCode != "POSTED" || increaseRequest.PostedLines != 2 {
		t.Fatal("partial adjustment completion")
	}
	_, err = service.ApproveAdjustment(ctx, staleRequest.ID, dto.AdjustmentLineSelectionRequest{ExpectedVersion: staleRequest.VersionNo, LineIDs: []string{staleRequest.Lines[0].ID}}, approver.ID)
	wants(t, err, ErrInvalidInput)
	staleRequest, err = service.CancelAdjustment(ctx, staleRequest.ID, dto.CancelAdjustmentRequest{ExpectedVersion: staleRequest.VersionNo, Reason: "Balance changed"}, account.ID)
	okay(t, err)
	if staleRequest.StatusCode != "CANCELLED" {
		t.Fatal("stale adjustment cancellation")
	}
	increased, err := inv.GetBalance(ctx, statusResult.DestinationBalance.ID)
	okay(t, err)
	decreaseRequest, err := service.CreateAdjustment(ctx, dto.CreateAdjustmentRequest{BusinessDate: "2026-09-07", Direction: "DECREASE", ReasonCode: "EXPIRY", Lines: []dto.CreateAdjustmentLineRequest{{BalanceID: increased.ID, Quantity: "1", ExpectedVersion: increased.VersionNo}}}, account.ID)
	okay(t, err)
	decreaseRequest, err = service.ApproveAdjustment(ctx, decreaseRequest.ID, dto.AdjustmentLineSelectionRequest{ExpectedVersion: decreaseRequest.VersionNo, LineIDs: []string{decreaseRequest.Lines[0].ID}}, approver.ID)
	okay(t, err)
	decreased, err := inv.GetBalance(ctx, increased.ID)
	okay(t, err)
	countBase := dto.CommandBase{OperationKey: "count-" + suffix, BusinessDate: "2026-09-07", SourceDocumentID: "COUNT-" + suffix, ReasonCode: pointer("EXPIRY")}
	counted, err := service.ReconcileCount(ctx, dto.StockCountReconcileRequest{CommandBase: countBase, BalanceID: decreased.ID, CountedQty: "3", ExpectedVersion: decreased.VersionNo}, account.ID)
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
	if total.Cmp(big.NewRat(11, 1)) != 0 {
		t.Fatalf("adjusted warehouse total=%s", total.RatString())
	}
	movements, err := inv.ListMovements(ctx, inventoryrepo.MovementFilter{OwnerID: owner.ID, Page: 1, PageSize: 100})
	okay(t, err)
	if movements.TotalItems != 9 {
		t.Fatalf("movements=%d", movements.TotalItems)
	}

	huType := master.HandlingUnitType{Code: "HU_" + suffix, Name: "Pallet"}
	okay(t, tx.Create(&huType).Error)
	hu := inventorymodel.HandlingUnit{ID: "HU-" + suffix, WarehouseID: wh1.ID, OwnerID: owner.ID, HandlingUnitTypeID: huType.ID, CurrentLocationID: pointer(loc1.ID), Barcode: "HU-BC-" + suffix}
	okay(t, tx.Create(&hu).Error)
	huStock, err := inv.PostMovement(ctx, inventorydto.PostingRequest{OperationKey: "hu-seed-" + suffix, MovementTypeCode: "RECEIVE", OwnerID: owner.ID, WarehouseID: wh1.ID, BusinessDate: "2026-09-07", ItemID: item.ID, HandlingUnitID: pointer(hu.ID), To: &inventorydto.BalanceDimension{LocationID: loc1.ID, InventoryStatusID: available.ID}, Quantity: "2", SourceDocumentID: "HU-SEED"}, account.ID)
	okay(t, err)
	huCommand := dto.CommandBase{OperationKey: "hu-move-" + suffix, BusinessDate: "2026-09-07", SourceDocumentID: "HU-MOVE-" + suffix}
	huMoved, err := service.InternalMove(ctx, dto.InternalMoveRequest{CommandBase: huCommand, SourceBalanceID: huStock.ToBalance.ID, TargetLocationID: loc2.ID, Quantity: "2", ExpectedVersion: huStock.ToBalance.VersionNo}, account.ID)
	okay(t, err)
	if huMoved.DestinationBalance == nil || huMoved.DestinationBalance.HandlingUnitID == nil || *huMoved.DestinationBalance.HandlingUnitID != hu.ID {
		t.Fatal("handling-unit stock was not moved")
	}
	var relocated inventorymodel.HandlingUnit
	okay(t, tx.Where("handling_unit_id=?", hu.ID).Take(&relocated).Error)
	if relocated.CurrentLocationID == nil || *relocated.CurrentLocationID != loc2.ID {
		t.Fatal("handling-unit location was not relocated")
	}

	replenishmentStock, err := inv.PostMovement(ctx, inventorydto.PostingRequest{OperationKey: "replenishment-seed-" + suffix, MovementTypeCode: "RECEIVE", OwnerID: owner.ID, WarehouseID: wh1.ID, BusinessDate: "2026-09-07", ItemID: item.ID, To: &inventorydto.BalanceDimension{LocationID: loc1.ID, InventoryStatusID: available.ID}, Quantity: "6", SourceDocumentID: "REPLENISHMENT-SEED"}, account.ID)
	okay(t, err)
	task, err := service.CreateReplenishment(ctx, dto.CreateReplenishmentRequest{SourceBalanceID: replenishmentStock.ToBalance.ID, TargetLocationID: pickFace.ID, Quantity: "2", PriorityCode: "NORMAL", ExpectedVersion: replenishmentStock.ToBalance.VersionNo}, account.ID)
	okay(t, err)
	if task.TaskStatusCode != "OPEN" || task.SourceBalanceVersionNo != replenishmentStock.ToBalance.VersionNo+1 {
		t.Fatalf("unexpected replenishment creation: %+v", task)
	}
	var reserved inventorymodel.InventoryBalance
	okay(t, tx.Where("balance_id=?", task.SourceBalanceID).Take(&reserved).Error)
	if reserved.ReservedQty != "2.000000" {
		t.Fatalf("replenishment reservation=%s", reserved.ReservedQty)
	}
	task, err = service.StartReplenishment(ctx, task.ID, dto.ReplenishmentTransitionRequest{ExpectedVersion: task.VersionNo}, account.ID)
	okay(t, err)
	if task.TaskStatusCode != "IN_PROGRESS" || task.AssignedTo == nil || *task.AssignedTo != account.ID {
		t.Fatalf("unexpected replenishment start: %+v", task)
	}
	task, err = service.CompleteReplenishment(ctx, task.ID, dto.CompleteReplenishmentRequest{ExpectedVersion: task.VersionNo, ExpectedBalanceVersion: task.SourceBalanceVersionNo, BusinessDate: "2026-09-07"}, account.ID)
	okay(t, err)
	if task.TaskStatusCode != "COMPLETED" || task.InventoryMovementID == nil || task.ResultingBalanceID == nil {
		t.Fatalf("unexpected replenishment completion: %+v", task)
	}
	okay(t, tx.Where("balance_id=?", task.SourceBalanceID).Take(&reserved).Error)
	if reserved.ReservedQty != "0.000000" {
		t.Fatalf("completed reservation=%s", reserved.ReservedQty)
	}

	cancelTask, err := service.CreateReplenishment(ctx, dto.CreateReplenishmentRequest{SourceBalanceID: task.SourceBalanceID, TargetLocationID: pickFace.ID, Quantity: "1", PriorityCode: "NORMAL", ExpectedVersion: reserved.VersionNo}, account.ID)
	okay(t, err)
	cancelTask, err = service.CancelReplenishment(ctx, cancelTask.ID, dto.CancelReplenishmentRequest{ExpectedVersion: cancelTask.VersionNo, ExpectedBalanceVersion: cancelTask.SourceBalanceVersionNo, Reason: "Demand changed"}, account.ID)
	okay(t, err)
	if cancelTask.TaskStatusCode != "CANCELLED" || cancelTask.CancellationReason == nil {
		t.Fatalf("unexpected replenishment cancellation: %+v", cancelTask)
	}
	okay(t, tx.Where("balance_id=?", cancelTask.SourceBalanceID).Take(&reserved).Error)
	if reserved.ReservedQty != "0.000000" {
		t.Fatalf("cancelled reservation=%s", reserved.ReservedQty)
	}
}
