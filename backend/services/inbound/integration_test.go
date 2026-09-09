package inbound

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
	dto "wms-api/dto/inbound"
	authmodel "wms-api/models/authentication"
	inventorymodel "wms-api/models/inventory"
	mastermodel "wms-api/models/master"
	authrepository "wms-api/repository/authentication"
	repository "wms-api/repository/inbound"
	inventoryrepository "wms-api/repository/inventory"
	masterrepository "wms-api/repository/master"
	masterservice "wms-api/services/master"
)

func inboundOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func inboundWant(t *testing.T, err, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("got %v, want %v", err, target)
	}
}

func inboundPointer[T any](value T) *T { return &value }

// Opt-in integration tests use a transaction and leave no business data behind.
func TestInboundWorkflowPostgreSQL(t *testing.T) {
	if os.Getenv("WMS_INTEGRATION_TEST") != "1" {
		t.Skip("set WMS_INTEGRATION_TEST=1")
	}
	for _, fresh := range []bool{false, true} {
		t.Run(fmt.Sprintf("fresh_%t", fresh), func(t *testing.T) { testInboundWorkflow(t, fresh) })
	}
}

func testInboundWorkflow(t *testing.T, fresh bool) {
	_ = godotenv.Load("../../.env")
	cfg, err := config.Load()
	inboundOK(t, err)
	cfg.Database.AutoCreate = false
	db, err := config.OpenDatabase(cfg.Database)
	inboundOK(t, err)
	sqlDB, err := db.DB()
	inboundOK(t, err)
	defer sqlDB.Close()
	tx := db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Begin()
	inboundOK(t, tx.Error)
	defer tx.Rollback()
	suffix := fmt.Sprint(time.Now().UnixNano())
	if fresh {
		schema := "inbound_test_" + suffix
		inboundOK(t, tx.Exec("CREATE SCHEMA "+schema).Error)
		inboundOK(t, tx.Exec("SET LOCAL search_path TO "+schema).Error)
	}
	inboundOK(t, authrepository.Migrate(tx))
	inboundOK(t, masterrepository.Migrate(tx))
	inboundOK(t, masterrepository.MigrateCatalog(tx))
	inboundOK(t, masterrepository.MigrateOperational(tx))
	inboundOK(t, inventoryrepository.Migrate(tx))
	inboundOK(t, repository.Migrate(tx))
	inboundOK(t, repository.SeedReferenceData(tx))

	ctx := context.Background()
	inboundOK(t, masterservice.NewCatalogService(masterrepository.NewCatalogRepositories(tx)).SeedCatalog(ctx))
	operational, err := masterservice.NewOperationalService(masterrepository.NewOperationalRepositories(tx), "Asia/Jakarta")
	inboundOK(t, err)
	inboundOK(t, operational.SeedOperational(ctx))

	accountStatus := authmodel.AccountStatus{Code: "IBS_" + suffix, Name: "Active", AllowsLogin: true}
	inboundOK(t, tx.Create(&accountStatus).Error)
	account := authmodel.AppAccount{Username: "inbound_" + suffix, DisplayName: "Inbound test", AccountStatusID: accountStatus.ID, ExternalSubject: inboundPointer("inbound_" + suffix)}
	inboundOK(t, tx.Create(&account).Error)
	deniedAccount := authmodel.AppAccount{Username: "inbound_denied_" + suffix, DisplayName: "Inbound denied", AccountStatusID: accountStatus.ID, ExternalSubject: inboundPointer("inbound_denied_" + suffix)}
	inboundOK(t, tx.Create(&deniedAccount).Error)
	owner := mastermodel.Organization{Code: "IBO_" + suffix, Name: "Inbound owner", TimezoneName: "Asia/Jakarta"}
	inboundOK(t, tx.Create(&owner).Error)
	warehouse := mastermodel.Warehouse{OperatorID: owner.ID, Code: "IBW_" + suffix, Name: "Inbound warehouse", TimezoneName: "Asia/Jakarta"}
	inboundOK(t, tx.Create(&warehouse).Error)
	inboundOK(t, tx.Create(&mastermodel.WarehouseOwner{WarehouseID: warehouse.ID, OwnerID: owner.ID}).Error)
	inboundOK(t, tx.Create(&mastermodel.AccountOwnerAccess{AccountID: account.ID, OwnerID: owner.ID}).Error)
	inboundOK(t, tx.Create(&mastermodel.AccountWarehouseAccess{AccountID: account.ID, WarehouseID: warehouse.ID}).Error)
	zone := mastermodel.WarehouseZone{WarehouseID: warehouse.ID, Code: "IBZ", Name: "Receiving zone"}
	inboundOK(t, tx.Create(&zone).Error)
	receivingType := mastermodel.LocationType{Code: "IBL_" + suffix, Name: "Receiving", AllowsReceiving: true}
	storageType := mastermodel.LocationType{Code: "IBS_" + suffix, Name: "Storage", AllowsStorage: true}
	inboundOK(t, tx.Create(&receivingType).Error)
	inboundOK(t, tx.Create(&storageType).Error)
	dock := mastermodel.WarehouseLocation{WarehouseID: warehouse.ID, ZoneID: zone.ID, LocationTypeID: receivingType.ID, Code: "DOCK-01"}
	qc := mastermodel.WarehouseLocation{WarehouseID: warehouse.ID, ZoneID: zone.ID, LocationTypeID: receivingType.ID, Code: "QC-01"}
	storage := mastermodel.WarehouseLocation{WarehouseID: warehouse.ID, ZoneID: zone.ID, LocationTypeID: storageType.ID, Code: "STORAGE-01"}
	storageTwo := mastermodel.WarehouseLocation{WarehouseID: warehouse.ID, ZoneID: zone.ID, LocationTypeID: storageType.ID, Code: "STORAGE-02"}
	inboundOK(t, tx.Create(&dock).Error)
	inboundOK(t, tx.Create(&qc).Error)
	inboundOK(t, tx.Create(&storage).Error)
	inboundOK(t, tx.Create(&storageTwo).Error)

	var supplierType mastermodel.PartnerType
	inboundOK(t, tx.Where("code = ?", "SUPPLIER").Take(&supplierType).Error)
	vendor := mastermodel.BusinessPartner{OwnerID: owner.ID, Code: "SUP_" + suffix, Name: "Supplier"}
	inboundOK(t, tx.Create(&vendor).Error)
	inboundOK(t, tx.Create(&mastermodel.BusinessPartnerType{PartnerID: vendor.ID, PartnerTypeID: supplierType.ID}).Error)
	var each mastermodel.UOM
	inboundOK(t, tx.Where("code = ?", "EA").Take(&each).Error)
	minimumReceiveDays := 10
	item := mastermodel.Item{OwnerID: owner.ID, Code: "ITEM_" + suffix, Name: "Lot item", BaseUOMID: each.ID, LotControlled: true, MinimumReceiveDays: &minimumReceiveDays}
	inboundOK(t, tx.Create(&item).Error)
	inboundOK(t, tx.Create(&mastermodel.ItemUOM{ItemID: item.ID, UOMID: each.ID, ConversionToBase: "1", IsReceivingUOM: true, IsPickingUOM: true}).Error)
	strategy := mastermodel.PutawayStrategy{OwnerID: &owner.ID, WarehouseID: &warehouse.ID, Code: "IBP_" + suffix, Name: "Inbound storage"}
	inboundOK(t, tx.Create(&strategy).Error)
	inboundOK(t, tx.Create(&mastermodel.PutawayStrategyRule{PutawayStrategyID: strategy.ID, SequenceNo: 10, LocationTypeID: &storageType.ID, ZoneID: &zone.ID, IsActive: true}).Error)

	location, _ := time.LoadLocation("Asia/Jakarta")
	businessDate := time.Now().In(location).Format("2006-01-02")
	orderedAt := businessDate + "T08:00:00+07:00"
	expiry, _ := time.Parse("2006-01-02", businessDate)
	expiryText := expiry.AddDate(1, 0, 0).Format("2006-01-02")
	service, err := NewService(repository.NewRepositories(tx), "Asia/Jakarta")
	inboundOK(t, err)
	allowed, err := service.CanAccess(ctx, account.ID, owner.ID, warehouse.ID)
	inboundOK(t, err)
	if !allowed {
		t.Fatal("granted inbound scope was denied")
	}
	allowed, err = service.CanAccess(ctx, deniedAccount.ID, owner.ID, warehouse.ID)
	inboundOK(t, err)
	if allowed {
		t.Fatal("ungranted inbound scope was allowed")
	}
	po, err := service.CreatePurchaseOrder(ctx, dto.CreatePurchaseOrderRequest{
		OwnerID: owner.ID, VendorID: vendor.ID, WarehouseID: warehouse.ID, BusinessDate: businessDate,
		PurchaseOrderNo: "CLIENT-PO-" + suffix, OrderedAt: orderedAt,
		Lines: []dto.PurchaseOrderLineRequest{
			{ItemID: item.ID, OrderedQty: "10", UOMID: each.ID},
			{ItemID: item.ID, OrderedQty: "2", UOMID: each.ID},
		},
	}, account.ID)
	inboundOK(t, err)
	if po.StatusCode != "DRAFT" || po.VersionNo != 1 || len(po.Lines) != 2 {
		t.Fatalf("unexpected purchase order: %+v", po)
	}
	scopeOwner, scopeWarehouse, err := service.ResourceScope(ctx, "purchase-orders", po.ID)
	inboundOK(t, err)
	if scopeOwner != owner.ID || scopeWarehouse != warehouse.ID {
		t.Fatalf("wrong purchase-order scope: %s %s", scopeOwner, scopeWarehouse)
	}
	po, err = service.UpdatePurchaseOrder(ctx, po.ID, dto.UpdatePurchaseOrderRequest{ExpectedVersion: po.VersionNo, PurchaseOrderNo: "CLIENT-PO-EDITED-" + suffix, OrderedAt: orderedAt, Notes: inboundPointer("draft edited")}, account.ID)
	inboundOK(t, err)
	po, err = service.AddPurchaseOrderLine(ctx, po.ID, dto.AddPurchaseOrderLineRequest{ExpectedVersion: po.VersionNo, PurchaseOrderLineRequest: dto.PurchaseOrderLineRequest{ItemID: item.ID, OrderedQty: "1", UOMID: each.ID}}, account.ID)
	inboundOK(t, err)
	addedPOLine := po.Lines[len(po.Lines)-1]
	po, err = service.UpdatePurchaseOrderLine(ctx, po.ID, addedPOLine.ID, dto.UpdatePurchaseOrderLineRequest{ExpectedVersion: po.VersionNo, OrderedQty: "2"}, account.ID)
	inboundOK(t, err)
	po, err = service.DeletePurchaseOrderLine(ctx, po.ID, addedPOLine.ID, po.VersionNo, account.ID)
	inboundOK(t, err)
	_, err = service.UpdatePurchaseOrder(ctx, po.ID, dto.UpdatePurchaseOrderRequest{ExpectedVersion: 1, PurchaseOrderNo: "STALE-" + suffix, OrderedAt: orderedAt}, account.ID)
	inboundWant(t, err, repository.ErrConcurrentWrite)
	po, err = service.ApprovePurchaseOrder(ctx, po.ID, dto.TransitionRequest{ExpectedVersion: po.VersionNo}, account.ID)
	inboundOK(t, err)
	if po.StatusCode != "APPROVED" || po.VersionNo != 6 {
		t.Fatalf("purchase order was not approved: %+v", po)
	}

	inboundOrder, err := service.CreateInboundOrder(ctx, dto.CreateInboundOrderRequest{
		PurchaseOrderID: po.ID, BusinessDate: businessDate,
		Lines: []dto.InboundOrderLineRequest{{PurchaseOrderLineID: po.Lines[0].ID, ExpectedQty: "10"}},
	}, account.ID)
	inboundOK(t, err)
	inboundOrder, err = service.UpdateInboundOrder(ctx, inboundOrder.ID, dto.UpdateInboundOrderRequest{ExpectedVersion: inboundOrder.VersionNo, ExternalReference: inboundPointer("EDITED-" + suffix), Notes: inboundPointer("draft edited")}, account.ID)
	inboundOK(t, err)
	inboundOrder, err = service.AddInboundOrderLine(ctx, inboundOrder.ID, dto.AddInboundOrderLineRequest{ExpectedVersion: inboundOrder.VersionNo, InboundOrderLineRequest: dto.InboundOrderLineRequest{PurchaseOrderLineID: po.Lines[1].ID, ExpectedQty: "2"}}, account.ID)
	inboundOK(t, err)
	inboundOrder, err = service.UpdateInboundOrderLine(ctx, inboundOrder.ID, inboundOrder.Lines[0].ID, dto.UpdateInboundOrderLineRequest{ExpectedVersion: inboundOrder.VersionNo, ExpectedQty: "10", Notes: inboundPointer("line edited")}, account.ID)
	inboundOK(t, err)
	inboundOrder, err = service.ReleaseInboundOrder(ctx, inboundOrder.ID, dto.TransitionRequest{ExpectedVersion: inboundOrder.VersionNo}, account.ID)
	inboundOK(t, err)
	if inboundOrder.StatusCode != "RELEASED" || inboundOrder.VersionNo != 5 {
		t.Fatalf("inbound order was not released: %+v", inboundOrder)
	}

	tooSoon := expiry.AddDate(0, 0, 5).Format("2006-01-02")
	_, err = service.CreateReceipt(ctx, dto.CreateReceiptRequest{
		InboundID: inboundOrder.ID, BusinessDate: businessDate, ReceivedAt: businessDate + "T08:30:00+07:00", DockLocationID: dock.ID,
		Lines: []dto.ReceiptLineRequest{{InboundLineID: inboundOrder.Lines[0].ID, ReceivedQty: "1", RejectedQty: "0", Batches: []dto.ReceiptBatchRequest{{SourceQty: "1", ReceivedLocationID: qc.ID, Lot: &dto.ReceiptLotRequest{LotNumber: "SHORT-" + suffix, ExpiryDate: &tooSoon}}}}},
	}, account.ID)
	inboundWant(t, err, ErrInvalidInput)
	var rolledBackLots int64
	inboundOK(t, tx.Table("inventory_lot").Where("lot_number = ?", "SHORT-"+suffix).Count(&rolledBackLots).Error)
	if rolledBackLots != 0 {
		t.Fatal("invalid receipt did not roll back its new lot")
	}

	receipt, err := service.CreateReceipt(ctx, dto.CreateReceiptRequest{
		InboundID: inboundOrder.ID, BusinessDate: businessDate, ReceivedAt: businessDate + "T09:00:00+07:00", DockLocationID: dock.ID,
		Lines: []dto.ReceiptLineRequest{
			{InboundLineID: inboundOrder.Lines[0].ID, ReceivedQty: "10", RejectedQty: "2", ExceptionNotes: inboundPointer("Two damaged units rejected at dock"), ExceptionTypeCode: inboundPointer("DAMAGED"), Batches: []dto.ReceiptBatchRequest{{SourceQty: "8", ReceivedLocationID: qc.ID, Lot: &dto.ReceiptLotRequest{LotNumber: "LOT-" + suffix, ExpiryDate: &expiryText}}}},
			{InboundLineID: inboundOrder.Lines[1].ID, ReceivedQty: "2", RejectedQty: "2", ExceptionNotes: inboundPointer("Entire line is the wrong supplied item"), ExceptionTypeCode: inboundPointer("WRONG_ITEM")},
		},
	}, account.ID)
	inboundOK(t, err)
	if receipt.StatusCode != "OPEN" || len(receipt.Lines) != 2 || receipt.Lines[0].AcceptedQty != "8.000000" || receipt.Lines[1].AcceptedQty != "0.000000" {
		t.Fatalf("unexpected open receipt: %+v", receipt)
	}
	receipt, err = service.UpdateReceipt(ctx, receipt.ID, dto.UpdateReceiptRequest{ExpectedVersion: receipt.VersionNo, ReceivedAt: businessDate + "T09:05:00+07:00", DockLocationID: dock.ID, VehicleNumber: inboundPointer("EDITED-TRUCK"), Lines: []dto.ReceiptLineRequest{
		{InboundLineID: inboundOrder.Lines[0].ID, ReceivedQty: "10", RejectedQty: "2", ExceptionNotes: inboundPointer("Two damaged units rejected at dock"), ExceptionTypeCode: inboundPointer("DAMAGED"), Batches: []dto.ReceiptBatchRequest{{SourceQty: "8", ReceivedLocationID: qc.ID, Lot: &dto.ReceiptLotRequest{LotNumber: "LOT-" + suffix, ExpiryDate: &expiryText}}}},
		{InboundLineID: inboundOrder.Lines[1].ID, ReceivedQty: "2", RejectedQty: "2", ExceptionNotes: inboundPointer("Entire line is the wrong supplied item"), ExceptionTypeCode: inboundPointer("WRONG_ITEM")},
	}}, account.ID)
	inboundOK(t, err)
	if receipt.VersionNo != 2 || receipt.VehicleNumber == nil || *receipt.VehicleNumber != "EDITED-TRUCK" {
		t.Fatalf("receipt draft edit failed: %+v", receipt)
	}
	var before int64
	inboundOK(t, tx.Model(&inventorymodel.InventoryMovement{}).Where("source_document_id = ?", receipt.ID).Count(&before).Error)
	if before != 0 || receipt.Lines[0].Batches[0].InitialBalanceID != nil {
		t.Fatal("opening a receipt must not post inventory")
	}

	receipt, err = service.CompleteReceipt(ctx, receipt.ID, dto.TransitionRequest{ExpectedVersion: receipt.VersionNo}, account.ID)
	inboundOK(t, err)
	if receipt.StatusCode != "COMPLETED" || receipt.VersionNo != 3 || receipt.Lines[0].Batches[0].InitialBalanceID == nil {
		t.Fatalf("receipt was not completed: %+v", receipt)
	}
	var movements []inventorymodel.InventoryMovement
	inboundOK(t, tx.Where("source_document_id = ?", receipt.ID).Find(&movements).Error)
	if len(movements) != 1 || movements[0].Quantity != "8.000000" {
		t.Fatalf("unexpected receipt movements: %+v", movements)
	}
	var balance inventorymodel.InventoryBalance
	inboundOK(t, tx.Where("balance_id = ?", *receipt.Lines[0].Batches[0].InitialBalanceID).Take(&balance).Error)
	var pending mastermodel.InventoryStatus
	inboundOK(t, tx.Where("code = ?", "QC_PENDING").Take(&pending).Error)
	if balance.OnHandQty != "8.000000" || balance.InventoryStatusID != pending.ID || balance.LocationID != qc.ID {
		t.Fatalf("unexpected QC-pending balance: %+v", balance)
	}
	po, err = service.GetPurchaseOrder(ctx, po.ID)
	inboundOK(t, err)
	inboundOrder, err = service.GetInboundOrder(ctx, inboundOrder.ID)
	inboundOK(t, err)
	if po.StatusCode != "RECEIVED" || inboundOrder.StatusCode != "RECEIVED" || po.Lines[0].CompletedReceiptQty != "10.000000" {
		t.Fatalf("document progress was not completed: po=%+v inbound=%+v", po, inboundOrder)
	}

	// Completion is idempotent and cannot create a second stock movement.
	_, err = service.CompleteReceipt(ctx, receipt.ID, dto.TransitionRequest{ExpectedVersion: 1}, account.ID)
	inboundOK(t, err)
	var after int64
	inboundOK(t, tx.Model(&inventorymodel.InventoryMovement{}).Where("source_document_id = ?", receipt.ID).Count(&after).Error)
	if after != 1 {
		t.Fatalf("idempotent completion created %d movements", after)
	}

	inspection, err := service.CreateQualityInspection(ctx, dto.CreateQualityInspectionRequest{ReceiptInventoryID: receipt.Lines[0].Batches[0].ID}, account.ID)
	inboundOK(t, err)
	if inspection.QualityStatusCode != "PENDING" || inspection.VersionNo != 1 {
		t.Fatalf("unexpected open inspection: %+v", inspection)
	}
	var qcBalance inventorymodel.InventoryBalance
	inboundOK(t, tx.Where("balance_id=?", *receipt.Lines[0].Batches[0].InitialBalanceID).Take(&qcBalance).Error)
	inspectionCompletion := dto.CompleteQualityInspectionRequest{ExpectedVersion: inspection.VersionNo, ExpectedBalanceVersion: qcBalance.VersionNo, PassedQty: "6", FailedQty: "2", PutawayTargetLocationID: &storage.ID}
	inspection, err = service.CompleteQualityInspection(ctx, inspection.ID, inspectionCompletion, account.ID)
	inboundOK(t, err)
	if inspection.InspectionResultCode != "PARTIAL" || inspection.PutawayTask == nil || inspection.QuarantineCase == nil {
		t.Fatalf("inspection did not create both outcomes: %+v", inspection)
	}
	var passBalance inventorymodel.InventoryBalance
	inboundOK(t, tx.Where("balance_id=?", inspection.PutawayTask.SourceBalanceID).Take(&passBalance).Error)
	var putawayPending mastermodel.InventoryStatus
	inboundOK(t, tx.Where("code=?", "PUTAWAY_PENDING").Take(&putawayPending).Error)
	if passBalance.OnHandQty != "6.000000" || passBalance.InventoryStatusID != putawayPending.ID {
		t.Fatalf("unexpected putaway-pending balance: %+v", passBalance)
	}
	inspectionReplay, err := service.CompleteQualityInspection(ctx, inspection.ID, inspectionCompletion, account.ID)
	inboundOK(t, err)
	if inspectionReplay.VersionNo != inspection.VersionNo || inspectionReplay.PutawayTask == nil || inspectionReplay.QuarantineCase == nil {
		t.Fatal("quality completion retry was not idempotent")
	}

	task, err := service.AssignPutawayTask(ctx, inspection.PutawayTask.ID, dto.AssignPutawayRequest{ExpectedVersion: inspection.PutawayTask.VersionNo, AccountID: account.ID}, account.ID)
	inboundOK(t, err)
	if task.TaskStatusCode != "ASSIGNED" {
		t.Fatalf("putaway task was not assigned: %+v", task)
	}
	task, err = service.RetargetPutawayTask(ctx, task.ID, dto.RetargetPutawayRequest{ExpectedVersion: task.VersionNo, TargetLocationID: storageTwo.ID}, account.ID)
	inboundOK(t, err)
	if task.TargetLocationID != storageTwo.ID {
		t.Fatalf("putaway task was not retargeted: %+v", task)
	}
	task, err = service.StartPutawayTask(ctx, task.ID, dto.PutawayTransitionRequest{ExpectedVersion: task.VersionNo}, account.ID)
	inboundOK(t, err)
	if task.TaskStatusCode != "IN_PROGRESS" || task.VersionNo != 4 {
		t.Fatalf("putaway task was not started: %+v", task)
	}
	inboundOK(t, tx.Where("balance_id=?", task.SourceBalanceID).Take(&passBalance).Error)
	putawayCompletion := dto.CompletePutawayRequest{ExpectedVersion: task.VersionNo, ExpectedBalanceVersion: passBalance.VersionNo, BusinessDate: businessDate}
	task, err = service.CompletePutawayTask(ctx, task.ID, putawayCompletion, account.ID)
	inboundOK(t, err)
	if task.TaskStatusCode != "COMPLETED" || task.ResultingBalanceID == nil {
		t.Fatalf("putaway task was not completed: %+v", task)
	}
	taskReplay, err := service.CompletePutawayTask(ctx, task.ID, putawayCompletion, account.ID)
	inboundOK(t, err)
	if taskReplay.VersionNo != task.VersionNo || taskReplay.InventoryMovementID == nil {
		t.Fatal("putaway completion retry was not idempotent")
	}
	var completedPutawayBalance inventorymodel.InventoryBalance
	inboundOK(t, tx.Where("balance_id=?", *task.ResultingBalanceID).Take(&completedPutawayBalance).Error)
	task, err = service.ReversePutawayTask(ctx, task.ID, dto.CancelPutawayRequest{ExpectedVersion: task.VersionNo, ExpectedBalanceVersion: completedPutawayBalance.VersionNo, BusinessDate: businessDate, Reason: "Completed into the wrong storage slot"}, account.ID)
	inboundOK(t, err)
	if task.TaskStatusCode != "REVERSED" || task.ReplacementInspectionID == nil {
		t.Fatalf("putaway reversal failed: %+v", task)
	}
	putawayRecoveryInspection, err := service.GetQualityInspection(ctx, *task.ReplacementInspectionID)
	inboundOK(t, err)
	var putawayRecoveryBalance inventorymodel.InventoryBalance
	inboundOK(t, tx.Where("balance_id=?", *putawayRecoveryInspection.SourceBalanceID).Take(&putawayRecoveryBalance).Error)
	putawayRecoveryInspection, err = service.CompleteQualityInspection(ctx, putawayRecoveryInspection.ID, dto.CompleteQualityInspectionRequest{ExpectedVersion: putawayRecoveryInspection.VersionNo, ExpectedBalanceVersion: putawayRecoveryBalance.VersionNo, PassedQty: "6", FailedQty: "0", PutawayTargetLocationID: &storageTwo.ID}, account.ID)
	inboundOK(t, err)
	recoveryPutaway, err := service.StartPutawayTask(ctx, putawayRecoveryInspection.PutawayTask.ID, dto.PutawayTransitionRequest{ExpectedVersion: putawayRecoveryInspection.PutawayTask.VersionNo}, account.ID)
	inboundOK(t, err)
	var recoveryPutawayBalance inventorymodel.InventoryBalance
	inboundOK(t, tx.Where("balance_id=?", recoveryPutaway.SourceBalanceID).Take(&recoveryPutawayBalance).Error)
	_, err = service.CompletePutawayTask(ctx, recoveryPutaway.ID, dto.CompletePutawayRequest{ExpectedVersion: recoveryPutaway.VersionNo, ExpectedBalanceVersion: recoveryPutawayBalance.VersionNo, BusinessDate: businessDate}, account.ID)
	inboundOK(t, err)

	caseResult := *inspection.QuarantineCase
	var quarantineBalance inventorymodel.InventoryBalance
	inboundOK(t, tx.Where("balance_id=?", caseResult.QuarantineBalanceID).Take(&quarantineBalance).Error)
	caseResult, err = service.CreateQuarantineDisposition(ctx, caseResult.ID, dto.CreateQuarantineDispositionRequest{ExpectedCaseVersion: caseResult.VersionNo, ExpectedBalanceVersion: quarantineBalance.VersionNo, DispositionTypeCode: "RETURN", DispositionQty: "1", BusinessDate: businessDate, DecidedAt: businessDate + "T11:00:00+07:00"}, account.ID)
	inboundOK(t, err)
	if caseResult.StatusCode != "PARTIALLY_DECIDED" || caseResult.DisposedQty != "1.000000" {
		t.Fatalf("quarantine case was not partially decided: %+v", caseResult)
	}
	inboundOK(t, tx.Where("balance_id=?", caseResult.QuarantineBalanceID).Take(&quarantineBalance).Error)
	caseResult, err = service.CreateQuarantineDisposition(ctx, caseResult.ID, dto.CreateQuarantineDispositionRequest{ExpectedCaseVersion: caseResult.VersionNo, ExpectedBalanceVersion: quarantineBalance.VersionNo, DispositionTypeCode: "REWORK", DispositionQty: "1", BusinessDate: businessDate, DecidedAt: businessDate + "T12:00:00+07:00", WorkInstructions: inboundPointer("Replace damaged seal and clean package")}, account.ID)
	inboundOK(t, err)
	if caseResult.StatusCode != "CLOSED" || caseResult.DisposedQty != "2.000000" || len(caseResult.Dispositions) != 2 {
		t.Fatalf("quarantine case was not closed: %+v", caseResult)
	}
	rework := caseResult.Dispositions[1].ReworkTask
	if rework == nil {
		t.Fatal("REWORK disposition did not create a rework task")
	}
	reworkResult, err := service.StartReworkTask(ctx, rework.ID, dto.ReworkTransitionRequest{ExpectedVersion: rework.VersionNo}, account.ID)
	inboundOK(t, err)
	reworkResult, err = service.CompleteReworkTask(ctx, reworkResult.ID, dto.CompleteReworkRequest{ExpectedVersion: reworkResult.VersionNo, ResultNotes: inboundPointer("Seal replaced")}, account.ID)
	inboundOK(t, err)
	if reworkResult.ReinspectionID == nil {
		t.Fatalf("rework did not open reinspection: %+v", reworkResult)
	}
	reinspection, err := service.GetQualityInspection(ctx, *reworkResult.ReinspectionID)
	inboundOK(t, err)
	var reworkBalance inventorymodel.InventoryBalance
	inboundOK(t, tx.Where("balance_id=?", reworkResult.SourceBalanceID).Take(&reworkBalance).Error)
	reinspection, err = service.CompleteQualityInspection(ctx, reinspection.ID, dto.CompleteQualityInspectionRequest{ExpectedVersion: reinspection.VersionNo, ExpectedBalanceVersion: reworkBalance.VersionNo, PassedQty: "1", FailedQty: "0", PutawayTargetLocationID: &storage.ID}, account.ID)
	inboundOK(t, err)
	if reinspection.PutawayTask == nil {
		t.Fatalf("reinspection did not create putaway: %+v", reinspection)
	}
	reworkPutaway, err := service.StartPutawayTask(ctx, reinspection.PutawayTask.ID, dto.PutawayTransitionRequest{ExpectedVersion: reinspection.PutawayTask.VersionNo}, account.ID)
	inboundOK(t, err)
	var reworkPutawayBalance inventorymodel.InventoryBalance
	inboundOK(t, tx.Where("balance_id=?", reworkPutaway.SourceBalanceID).Take(&reworkPutawayBalance).Error)
	_, err = service.CompletePutawayTask(ctx, reworkPutaway.ID, dto.CompletePutawayRequest{ExpectedVersion: reworkPutaway.VersionNo, ExpectedBalanceVersion: reworkPutawayBalance.VersionNo, BusinessDate: businessDate}, account.ID)
	inboundOK(t, err)
	var availableTotal string
	inboundOK(t, tx.Table("inventory_balance balance").Select("COALESCE(sum(balance.on_hand_qty),0)::text").Joins("JOIN inventory_status status ON status.inventory_status_id=balance.inventory_status_id").Where("balance.item_id=? AND status.code='AVAILABLE'", item.ID).Scan(&availableTotal).Error)
	if availableTotal != "7.000000" {
		t.Fatalf("available stock after putaway and quarantine disposition=%s", availableTotal)
	}

	// Multiple receipts, tolerance, reversal, short closure, and recovery paths.
	overTolerance, underTolerance := "10", "5"
	variancePO, err := service.CreatePurchaseOrder(ctx, dto.CreatePurchaseOrderRequest{OwnerID: owner.ID, VendorID: vendor.ID, WarehouseID: warehouse.ID, BusinessDate: businessDate, PurchaseOrderNo: "VARIANCE-PO-" + suffix, OrderedAt: orderedAt, Lines: []dto.PurchaseOrderLineRequest{{ItemID: item.ID, OrderedQty: "10", UOMID: each.ID, OverReceiptTolerancePct: &overTolerance, UnderReceiptTolerancePct: &underTolerance}}}, account.ID)
	inboundOK(t, err)
	variancePO, err = service.ApprovePurchaseOrder(ctx, variancePO.ID, dto.TransitionRequest{ExpectedVersion: variancePO.VersionNo}, account.ID)
	inboundOK(t, err)
	varianceInbound, err := service.CreateInboundOrder(ctx, dto.CreateInboundOrderRequest{PurchaseOrderID: variancePO.ID, BusinessDate: businessDate, Lines: []dto.InboundOrderLineRequest{{PurchaseOrderLineID: variancePO.Lines[0].ID, ExpectedQty: "10"}}}, account.ID)
	inboundOK(t, err)
	varianceInbound, err = service.ReleaseInboundOrder(ctx, varianceInbound.ID, dto.TransitionRequest{ExpectedVersion: varianceInbound.VersionNo}, account.ID)
	inboundOK(t, err)
	partialReceipt, err := service.CreateReceipt(ctx, dto.CreateReceiptRequest{InboundID: varianceInbound.ID, BusinessDate: businessDate, ReceivedAt: businessDate + "T13:00:00+07:00", DockLocationID: dock.ID, Lines: []dto.ReceiptLineRequest{{InboundLineID: varianceInbound.Lines[0].ID, ReceivedQty: "4", RejectedQty: "0", Batches: []dto.ReceiptBatchRequest{{SourceQty: "4", ReceivedLocationID: qc.ID, Lot: &dto.ReceiptLotRequest{LotNumber: "VAR-A-" + suffix, ExpiryDate: &expiryText}}}}}}, account.ID)
	inboundOK(t, err)
	partialReceipt, err = service.CompleteReceipt(ctx, partialReceipt.ID, dto.TransitionRequest{ExpectedVersion: partialReceipt.VersionNo}, account.ID)
	inboundOK(t, err)
	varianceInbound, err = service.GetInboundOrder(ctx, varianceInbound.ID)
	inboundOK(t, err)
	if varianceInbound.StatusCode != "PARTIALLY_RECEIVED" {
		t.Fatalf("partial receipt did not preserve open inbound: %+v", varianceInbound)
	}
	overReceipt, err := service.CreateReceipt(ctx, dto.CreateReceiptRequest{InboundID: varianceInbound.ID, BusinessDate: businessDate, ReceivedAt: businessDate + "T14:00:00+07:00", DockLocationID: dock.ID, Lines: []dto.ReceiptLineRequest{{InboundLineID: varianceInbound.Lines[0].ID, ReceivedQty: "7", RejectedQty: "0", ExceptionNotes: inboundPointer("Supplier shipped one approved extra unit"), Batches: []dto.ReceiptBatchRequest{{SourceQty: "7", ReceivedLocationID: qc.ID, Lot: &dto.ReceiptLotRequest{LotNumber: "VAR-B-" + suffix, ExpiryDate: &expiryText}}}}}}, account.ID)
	inboundOK(t, err)
	overReceipt, err = service.CompleteReceipt(ctx, overReceipt.ID, dto.TransitionRequest{ExpectedVersion: overReceipt.VersionNo}, account.ID)
	inboundOK(t, err)
	if overReceipt.StatusCode != "COMPLETED" {
		t.Fatalf("over receipt was not completed: %+v", overReceipt)
	}
	var overBalance inventorymodel.InventoryBalance
	inboundOK(t, tx.Where("balance_id=?", *overReceipt.Lines[0].Batches[0].InitialBalanceID).Take(&overBalance).Error)
	overReceipt, err = service.ReverseReceipt(ctx, overReceipt.ID, dto.ReverseReceiptRequest{ExpectedVersion: overReceipt.VersionNo, BusinessDate: businessDate, Reason: "Duplicate supplier delivery", Balances: []dto.BalanceVersionRequest{{BalanceID: overBalance.ID, ExpectedVersion: overBalance.VersionNo}}}, account.ID)
	inboundOK(t, err)
	if overReceipt.StatusCode != "REVERSED" {
		t.Fatalf("receipt was not reversed: %+v", overReceipt)
	}
	varianceInbound, err = service.GetInboundOrder(ctx, varianceInbound.ID)
	inboundOK(t, err)
	if varianceInbound.StatusCode != "PARTIALLY_RECEIVED" {
		t.Fatalf("reversal did not reopen inbound progress: %+v", varianceInbound)
	}
	varianceInbound, err = service.CloseInboundOrder(ctx, varianceInbound.ID, dto.ExceptionTransitionRequest{ExpectedVersion: varianceInbound.VersionNo, Reason: "Supplier confirmed remaining quantity will not ship"}, account.ID)
	inboundOK(t, err)
	if varianceInbound.StatusCode != "CLOSED" {
		t.Fatalf("inbound was not closed short: %+v", varianceInbound)
	}
	variancePO, err = service.GetPurchaseOrder(ctx, variancePO.ID)
	inboundOK(t, err)
	variancePO, err = service.ClosePurchaseOrder(ctx, variancePO.ID, dto.ExceptionTransitionRequest{ExpectedVersion: variancePO.VersionNo, Reason: "Owner accepted short closure"}, account.ID)
	inboundOK(t, err)
	if variancePO.StatusCode != "CLOSED" {
		t.Fatalf("purchase order was not closed short: %+v", variancePO)
	}

	cancelledInspection, err := service.CreateQualityInspection(ctx, dto.CreateQualityInspectionRequest{ReceiptInventoryID: partialReceipt.Lines[0].Batches[0].ID}, account.ID)
	inboundOK(t, err)
	cancelledInspection, err = service.CancelQualityInspection(ctx, cancelledInspection.ID, dto.ExceptionTransitionRequest{ExpectedVersion: cancelledInspection.VersionNo, Reason: "Wrong QC sampling plan"}, account.ID)
	inboundOK(t, err)
	if cancelledInspection.QualityStatusCode != "WAIVED" || cancelledInspection.ReplacementInspectionID == nil {
		t.Fatalf("inspection cancellation did not create replacement: %+v", cancelledInspection)
	}
	replacement, err := service.GetQualityInspection(ctx, *cancelledInspection.ReplacementInspectionID)
	inboundOK(t, err)
	var partialBalance inventorymodel.InventoryBalance
	inboundOK(t, tx.Where("balance_id=?", *replacement.SourceBalanceID).Take(&partialBalance).Error)
	replacement, err = service.CompleteQualityInspection(ctx, replacement.ID, dto.CompleteQualityInspectionRequest{ExpectedVersion: replacement.VersionNo, ExpectedBalanceVersion: partialBalance.VersionNo, PassedQty: "4", FailedQty: "0", PutawayTargetLocationID: &storage.ID}, account.ID)
	inboundOK(t, err)
	var pendingPutawayBalance inventorymodel.InventoryBalance
	inboundOK(t, tx.Where("balance_id=?", replacement.PutawayTask.SourceBalanceID).Take(&pendingPutawayBalance).Error)
	cancelledTask, err := service.CancelPutawayTask(ctx, replacement.PutawayTask.ID, dto.CancelPutawayRequest{ExpectedVersion: replacement.PutawayTask.VersionNo, ExpectedBalanceVersion: pendingPutawayBalance.VersionNo, BusinessDate: businessDate, Reason: "Putaway plan changed; return to QC queue"}, account.ID)
	inboundOK(t, err)
	if cancelledTask.TaskStatusCode != "CANCELLED" {
		t.Fatalf("putaway cancellation failed: %+v", cancelledTask)
	}

	draftPO, err := service.CreatePurchaseOrder(ctx, dto.CreatePurchaseOrderRequest{OwnerID: owner.ID, VendorID: vendor.ID, WarehouseID: warehouse.ID, BusinessDate: businessDate, PurchaseOrderNo: "CANCEL-PO-" + suffix, OrderedAt: orderedAt, Lines: []dto.PurchaseOrderLineRequest{{ItemID: item.ID, OrderedQty: "1", UOMID: each.ID}}}, account.ID)
	inboundOK(t, err)
	draftPO, err = service.CancelPurchaseOrder(ctx, draftPO.ID, dto.ExceptionTransitionRequest{ExpectedVersion: draftPO.VersionNo, Reason: "Owner cancelled before approval"}, account.ID)
	inboundOK(t, err)
	if draftPO.StatusCode != "CANCELLED" {
		t.Fatalf("purchase order cancellation failed: %+v", draftPO)
	}
	_, err = service.UpdatePurchaseOrder(ctx, draftPO.ID, dto.UpdatePurchaseOrderRequest{ExpectedVersion: draftPO.VersionNo, PurchaseOrderNo: "CANNOT-EDIT-" + suffix, OrderedAt: orderedAt}, account.ID)
	inboundWant(t, err, ErrInvalidState)
	draftPOSuccessor, err := service.CreatePurchaseOrder(ctx, dto.CreatePurchaseOrderRequest{OwnerID: owner.ID, VendorID: vendor.ID, WarehouseID: warehouse.ID, BusinessDate: businessDate, PurchaseOrderNo: "REPLACEMENT-PO-" + suffix, OrderedAt: orderedAt, SupersedesPurchaseOrderID: &draftPO.ID, Lines: []dto.PurchaseOrderLineRequest{{ItemID: item.ID, OrderedQty: "1", UOMID: each.ID}}}, account.ID)
	inboundOK(t, err)
	if draftPOSuccessor.SupersedesPurchaseOrderID == nil || *draftPOSuccessor.SupersedesPurchaseOrderID != draftPO.ID {
		t.Fatalf("purchase-order replacement link missing: %+v", draftPOSuccessor)
	}
	_, err = service.CreatePurchaseOrder(ctx, dto.CreatePurchaseOrderRequest{OwnerID: owner.ID, VendorID: vendor.ID, WarehouseID: warehouse.ID, BusinessDate: businessDate, PurchaseOrderNo: "DUPLICATE-REPLACEMENT-PO-" + suffix, OrderedAt: orderedAt, SupersedesPurchaseOrderID: &draftPO.ID, Lines: []dto.PurchaseOrderLineRequest{{ItemID: item.ID, OrderedQty: "1", UOMID: each.ID}}}, account.ID)
	inboundWant(t, err, ErrInvalidState)
	cancelFlowPO, err := service.CreatePurchaseOrder(ctx, dto.CreatePurchaseOrderRequest{OwnerID: owner.ID, VendorID: vendor.ID, WarehouseID: warehouse.ID, BusinessDate: businessDate, PurchaseOrderNo: "CANCEL-FLOW-PO-" + suffix, OrderedAt: orderedAt, Lines: []dto.PurchaseOrderLineRequest{{ItemID: item.ID, OrderedQty: "1", UOMID: each.ID}}}, account.ID)
	inboundOK(t, err)
	cancelFlowPO, err = service.ApprovePurchaseOrder(ctx, cancelFlowPO.ID, dto.TransitionRequest{ExpectedVersion: cancelFlowPO.VersionNo}, account.ID)
	inboundOK(t, err)
	cancelFlowInbound, err := service.CreateInboundOrder(ctx, dto.CreateInboundOrderRequest{PurchaseOrderID: cancelFlowPO.ID, BusinessDate: businessDate, Lines: []dto.InboundOrderLineRequest{{PurchaseOrderLineID: cancelFlowPO.Lines[0].ID, ExpectedQty: "1"}}}, account.ID)
	inboundOK(t, err)
	cancelFlowInbound, err = service.ReleaseInboundOrder(ctx, cancelFlowInbound.ID, dto.TransitionRequest{ExpectedVersion: cancelFlowInbound.VersionNo}, account.ID)
	inboundOK(t, err)
	cancelFlowReceipt, err := service.CreateReceipt(ctx, dto.CreateReceiptRequest{InboundID: cancelFlowInbound.ID, BusinessDate: businessDate, ReceivedAt: businessDate + "T15:00:00+07:00", DockLocationID: dock.ID, Lines: []dto.ReceiptLineRequest{{InboundLineID: cancelFlowInbound.Lines[0].ID, ReceivedQty: "1", RejectedQty: "0", Batches: []dto.ReceiptBatchRequest{{SourceQty: "1", ReceivedLocationID: qc.ID, Lot: &dto.ReceiptLotRequest{LotNumber: "CANCEL-LOT-" + suffix, ExpiryDate: &expiryText}}}}}}, account.ID)
	inboundOK(t, err)
	cancelFlowReceipt, err = service.CancelReceipt(ctx, cancelFlowReceipt.ID, dto.ExceptionTransitionRequest{ExpectedVersion: cancelFlowReceipt.VersionNo, Reason: "Truck manifest was duplicated"}, account.ID)
	inboundOK(t, err)
	if cancelFlowReceipt.StatusCode != "CANCELLED" {
		t.Fatalf("receipt cancellation failed: %+v", cancelFlowReceipt)
	}
	replacementReceipt, err := service.CreateReceipt(ctx, dto.CreateReceiptRequest{InboundID: cancelFlowInbound.ID, BusinessDate: businessDate, ReceivedAt: businessDate + "T15:05:00+07:00", DockLocationID: dock.ID, SupersedesReceiptID: &cancelFlowReceipt.ID, Lines: []dto.ReceiptLineRequest{{InboundLineID: cancelFlowInbound.Lines[0].ID, ReceivedQty: "1", RejectedQty: "0", Batches: []dto.ReceiptBatchRequest{{SourceQty: "1", ReceivedLocationID: qc.ID, Lot: &dto.ReceiptLotRequest{LotNumber: "CANCEL-LOT-" + suffix, ExpiryDate: &expiryText}}}}}}, account.ID)
	inboundOK(t, err)
	if replacementReceipt.SupersedesReceiptID == nil || *replacementReceipt.SupersedesReceiptID != cancelFlowReceipt.ID {
		t.Fatalf("receipt replacement link missing: %+v", replacementReceipt)
	}
	replacementReceipt, err = service.CancelReceipt(ctx, replacementReceipt.ID, dto.ExceptionTransitionRequest{ExpectedVersion: replacementReceipt.VersionNo, Reason: "Replacement receipt no longer required"}, account.ID)
	inboundOK(t, err)
	cancelFlowInbound, err = service.CancelInboundOrder(ctx, cancelFlowInbound.ID, dto.ExceptionTransitionRequest{ExpectedVersion: cancelFlowInbound.VersionNo, Reason: "No valid delivery remains"}, account.ID)
	inboundOK(t, err)
	if cancelFlowInbound.StatusCode != "CANCELLED" {
		t.Fatalf("inbound cancellation failed: %+v", cancelFlowInbound)
	}
	replacementInbound, err := service.CreateInboundOrder(ctx, dto.CreateInboundOrderRequest{PurchaseOrderID: cancelFlowPO.ID, BusinessDate: businessDate, SupersedesInboundID: &cancelFlowInbound.ID, Lines: []dto.InboundOrderLineRequest{{PurchaseOrderLineID: cancelFlowPO.Lines[0].ID, ExpectedQty: "1"}}}, account.ID)
	inboundOK(t, err)
	if replacementInbound.SupersedesInboundID == nil || *replacementInbound.SupersedesInboundID != cancelFlowInbound.ID {
		t.Fatalf("inbound replacement link missing: %+v", replacementInbound)
	}
	replacementInbound, err = service.CancelInboundOrder(ctx, replacementInbound.ID, dto.ExceptionTransitionRequest{ExpectedVersion: replacementInbound.VersionNo, Reason: "Replacement inbound no longer required"}, account.ID)
	inboundOK(t, err)
	cancelFlowPO, err = service.CancelPurchaseOrder(ctx, cancelFlowPO.ID, dto.ExceptionTransitionRequest{ExpectedVersion: cancelFlowPO.VersionNo, Reason: "Owner cancelled order"}, account.ID)
	inboundOK(t, err)
	if cancelFlowPO.StatusCode != "CANCELLED" {
		t.Fatalf("approved purchase order cancellation failed: %+v", cancelFlowPO)
	}

	exceptions, err := service.ListInboundExceptions(ctx, repository.ListFilter{OwnerID: owner.ID, Page: 1, PageSize: 100})
	inboundOK(t, err)
	if exceptions.TotalItems < 11 {
		t.Fatalf("expected lifecycle exception audit records, got %d", exceptions.TotalItems)
	}

	_, err = service.CreateInboundOrder(ctx, dto.CreateInboundOrderRequest{PurchaseOrderID: po.ID, BusinessDate: businessDate, Lines: []dto.InboundOrderLineRequest{{PurchaseOrderLineID: po.Lines[0].ID, ExpectedQty: "1"}}}, account.ID)
	inboundWant(t, err, ErrInvalidState)
}
