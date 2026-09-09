package outbound

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
	dto "wms-api/dto/outbound"
	authmodel "wms-api/models/authentication"
	inventorymodel "wms-api/models/inventory"
	mastermodel "wms-api/models/master"
	model "wms-api/models/outbound"
	authrepository "wms-api/repository/authentication"
	inventoryrepository "wms-api/repository/inventory"
	masterrepository "wms-api/repository/master"
	repository "wms-api/repository/outbound"
	stockrepository "wms-api/repository/stock_control"
	masterservice "wms-api/services/master"
)

func outboundOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func outboundPointer[T any](value T) *T { return &value }

func TestOutboundPostgreSQL(t *testing.T) {
	if os.Getenv("WMS_INTEGRATION_TEST") != "1" {
		t.Skip("set WMS_INTEGRATION_TEST=1")
	}
	for _, fresh := range []bool{false, true} {
		t.Run(fmt.Sprintf("fresh_%t", fresh), func(t *testing.T) { testOutboundPartOne(t, fresh) })
	}
}

func testOutboundPartOne(t *testing.T, fresh bool) {
	_ = godotenv.Load("../../.env")
	cfg, err := config.Load()
	outboundOK(t, err)
	cfg.Database.AutoCreate = false
	db, err := config.OpenDatabase(cfg.Database)
	outboundOK(t, err)
	sqlDB, err := db.DB()
	outboundOK(t, err)
	defer sqlDB.Close()
	tx := db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Begin()
	outboundOK(t, tx.Error)
	defer tx.Rollback()
	suffix := fmt.Sprint(time.Now().UnixNano())
	if fresh {
		schema := "outbound_test_" + suffix
		outboundOK(t, tx.Exec("CREATE SCHEMA "+schema).Error)
		outboundOK(t, tx.Exec("SET LOCAL search_path TO "+schema).Error)
	}
	outboundOK(t, authrepository.Migrate(tx))
	outboundOK(t, masterrepository.Migrate(tx))
	outboundOK(t, masterrepository.MigrateCatalog(tx))
	outboundOK(t, masterrepository.MigrateOperational(tx))
	outboundOK(t, inventoryrepository.Migrate(tx))
	outboundOK(t, stockrepository.Migrate(tx))
	outboundOK(t, repository.Migrate(tx))
	ctx := context.Background()
	outboundOK(t, masterservice.NewCatalogService(masterrepository.NewCatalogRepositories(tx)).SeedCatalog(ctx))
	operational, err := masterservice.NewOperationalService(masterrepository.NewOperationalRepositories(tx), "Asia/Jakarta")
	outboundOK(t, err)
	outboundOK(t, operational.SeedOperational(ctx))
	outboundOK(t, repository.SeedReferenceData(tx))
	accountStatus := authmodel.AccountStatus{Code: "OBS_" + suffix, Name: "Active", AllowsLogin: true}
	outboundOK(t, tx.Create(&accountStatus).Error)
	account := authmodel.AppAccount{Username: "outbound_" + suffix, DisplayName: "Outbound test", AccountStatusID: accountStatus.ID, ExternalSubject: outboundPointer("outbound_" + suffix)}
	outboundOK(t, tx.Create(&account).Error)
	owner := mastermodel.Organization{Code: "OBO_" + suffix, Name: "Outbound owner", TimezoneName: "Asia/Jakarta"}
	outboundOK(t, tx.Create(&owner).Error)
	warehouse := mastermodel.Warehouse{OperatorID: owner.ID, Code: "OBW_" + suffix, Name: "Outbound warehouse", TimezoneName: "Asia/Jakarta"}
	outboundOK(t, tx.Create(&warehouse).Error)
	outboundOK(t, tx.Create(&mastermodel.WarehouseOwner{WarehouseID: warehouse.ID, OwnerID: owner.ID}).Error)
	outboundOK(t, tx.Create(&mastermodel.AccountOwnerAccess{AccountID: account.ID, OwnerID: owner.ID}).Error)
	outboundOK(t, tx.Create(&mastermodel.AccountWarehouseAccess{AccountID: account.ID, WarehouseID: warehouse.ID}).Error)
	zone := mastermodel.WarehouseZone{WarehouseID: warehouse.ID, Code: "OBZ", Name: "Outbound zone"}
	outboundOK(t, tx.Create(&zone).Error)
	pickType := mastermodel.LocationType{Code: "OBP_" + suffix, Name: "Pick storage", AllowsStorage: true, AllowsPicking: true}
	outboundOK(t, tx.Create(&pickType).Error)
	packingType := mastermodel.LocationType{Code: "OBK_" + suffix, Name: "Packing"}
	outboundOK(t, tx.Create(&packingType).Error)
	var stagingType mastermodel.LocationType
	err = tx.Where("code='STAGING'").Take(&stagingType).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		stagingType = mastermodel.LocationType{Code: "STAGING", Name: "Staging"}
		outboundOK(t, tx.Create(&stagingType).Error)
	} else {
		outboundOK(t, err)
	}
	pickLocation := mastermodel.WarehouseLocation{WarehouseID: warehouse.ID, ZoneID: zone.ID, LocationTypeID: pickType.ID, Code: "PICK-01", PickSequence: 10}
	stagingLocation := mastermodel.WarehouseLocation{WarehouseID: warehouse.ID, ZoneID: zone.ID, LocationTypeID: stagingType.ID, Code: "STAGE-01"}
	packingLocation := mastermodel.WarehouseLocation{WarehouseID: warehouse.ID, ZoneID: zone.ID, LocationTypeID: packingType.ID, Code: "PACK-01"}
	outboundOK(t, tx.Create(&pickLocation).Error)
	outboundOK(t, tx.Create(&stagingLocation).Error)
	outboundOK(t, tx.Create(&packingLocation).Error)
	address := "Jl. Test 1"
	customer := mastermodel.BusinessPartner{OwnerID: owner.ID, Code: "CUS_" + suffix, Name: "Customer", AddressLine1: &address}
	shipTo := mastermodel.BusinessPartner{OwnerID: owner.ID, Code: "STORE_" + suffix, Name: "Store", AddressLine1: &address}
	outboundOK(t, tx.Create(&customer).Error)
	outboundOK(t, tx.Create(&shipTo).Error)
	var customerType mastermodel.PartnerType
	outboundOK(t, tx.Where("code='CUSTOMER'").Take(&customerType).Error)
	outboundOK(t, tx.Create(&mastermodel.BusinessPartnerType{PartnerID: customer.ID, PartnerTypeID: customerType.ID}).Error)
	outboundOK(t, tx.Create(&mastermodel.BusinessPartnerType{PartnerID: shipTo.ID, PartnerTypeID: customerType.ID}).Error)
	var each mastermodel.UOM
	outboundOK(t, tx.Where("code='EA'").Take(&each).Error)
	item := mastermodel.Item{OwnerID: owner.ID, Code: "OBI_" + suffix, Name: "Outbound item", BaseUOMID: each.ID}
	outboundOK(t, tx.Create(&item).Error)
	outboundOK(t, tx.Create(&mastermodel.ItemUOM{ItemID: item.ID, UOMID: each.ID, ConversionToBase: "1", IsPickingUOM: true}).Error)
	var available mastermodel.InventoryStatus
	outboundOK(t, tx.Where("code='AVAILABLE'").Take(&available).Error)
	var fefo mastermodel.PickingSortMethod
	outboundOK(t, tx.Where("code='FEFO'").Take(&fefo).Error)
	strategy := mastermodel.PickingStrategy{OwnerID: &owner.ID, WarehouseID: &warehouse.ID, Code: "OBP_" + suffix, Name: "Outbound FEFO", IsActive: true}
	outboundOK(t, tx.Create(&strategy).Error)
	outboundOK(t, tx.Create(&mastermodel.PickingStrategyRule{PickingStrategyID: strategy.ID, SequenceNo: 10, InventoryStatusID: &available.ID, ZoneID: &zone.ID, PickingSortMethodID: fefo.ID, IsActive: true}).Error)
	balance := inventorymodel.InventoryBalance{ID: "BAL-" + suffix, OwnerID: owner.ID, WarehouseID: warehouse.ID, LocationID: pickLocation.ID, ItemID: item.ID, InventoryStatusID: available.ID, OnHandQty: "10", ReservedQty: "0", UOMID: each.ID, VersionNo: 1}
	outboundOK(t, tx.Create(&balance).Error)
	service, err := NewService(repository.NewRepositories(tx), "Asia/Jakarta")
	outboundOK(t, err)
	allowed, err := service.CanAccess(ctx, account.ID, owner.ID, warehouse.ID)
	outboundOK(t, err)
	if !allowed {
		t.Fatal("configured account scope was denied")
	}
	deniedAccount := authmodel.AppAccount{Username: "outbound_denied_" + suffix, DisplayName: "Denied outbound test", AccountStatusID: accountStatus.ID, ExternalSubject: outboundPointer("outbound_denied_" + suffix)}
	outboundOK(t, tx.Create(&deniedAccount).Error)
	allowed, err = service.CanAccess(ctx, deniedAccount.ID, owner.ID, warehouse.ID)
	outboundOK(t, err)
	if allowed {
		t.Fatal("account without owner/warehouse grants was allowed")
	}
	businessDate := time.Now().Format("2006-01-02")
	order, err := service.CreateOutboundOrder(ctx, dto.CreateOutboundOrderRequest{OwnerID: owner.ID, CustomerID: customer.ID, ShipToPartnerID: shipTo.ID, WarehouseID: warehouse.ID, BusinessDate: businessDate, ClientDeliveryOrderNo: "CLIENT-DO-" + suffix, RequestedShipAt: businessDate + "T15:00:00+07:00", Lines: []dto.OutboundOrderLineRequest{{ItemID: item.ID, OrderedQty: "6"}}}, account.ID)
	outboundOK(t, err)
	scopeOwner, scopeWarehouse, err := service.ResourceScope(ctx, "orders", order.ID)
	outboundOK(t, err)
	if scopeOwner != owner.ID || scopeWarehouse != warehouse.ID {
		t.Fatalf("unexpected outbound resource scope: %s %s", scopeOwner, scopeWarehouse)
	}
	if order.StatusCode != "DRAFT" || len(order.Lines) != 1 {
		t.Fatalf("unexpected order: %+v", order)
	}
	validation, err := service.ValidateOutboundOrder(ctx, order.ID, dto.ValidateOutboundRequest{ExpectedVersion: order.VersionNo}, account.ID)
	outboundOK(t, err)
	if validation.StatusCode != "PASSED" || len(validation.Results) != 5 {
		t.Fatalf("unexpected validation: %+v", validation)
	}
	order, err = service.GetOutboundOrder(ctx, order.ID)
	outboundOK(t, err)
	order, err = service.ReleaseOutboundOrder(ctx, order.ID, dto.TransitionRequest{ExpectedVersion: order.VersionNo}, account.ID)
	outboundOK(t, err)
	allocation, err := service.AllocateOutboundOrder(ctx, order.ID, dto.AllocateOutboundRequest{ExpectedVersion: order.VersionNo, PickingStrategyID: &strategy.ID}, account.ID)
	outboundOK(t, err)
	if allocation.Order.StatusCode != "ALLOCATED" || allocation.AllocatedQty != "6.000000" || len(allocation.Reservations) != 1 {
		t.Fatalf("unexpected allocation: %+v", allocation)
	}
	outboundOK(t, tx.Where("balance_id=?", balance.ID).Take(&balance).Error)
	if balance.OnHandQty != "10.000000" || balance.ReservedQty != "6.000000" {
		t.Fatalf("allocation changed the wrong quantities: %+v", balance)
	}
	wave, err := service.CreateWave(ctx, dto.CreateWaveRequest{OwnerID: owner.ID, WarehouseID: warehouse.ID, BusinessDate: businessDate, WaveTypeCode: "SINGLE_ORDER", PickingStrategyID: &strategy.ID, OutboundIDs: []string{order.ID}}, account.ID)
	outboundOK(t, err)
	wave, err = service.ReleaseWave(ctx, wave.ID, dto.ReleaseWaveRequest{ExpectedVersion: wave.VersionNo, StagingLocationID: stagingLocation.ID, PriorityCode: "NORMAL"}, account.ID)
	outboundOK(t, err)
	if wave.StatusCode != "RELEASED" || wave.PickTaskCount != 1 {
		t.Fatalf("unexpected released wave: %+v", wave)
	}
	tasks, err := service.ListPickTasks(ctx, repository.ListFilter{OwnerID: owner.ID, WarehouseID: warehouse.ID, Page: 1, PageSize: 20}, "")
	outboundOK(t, err)
	if len(tasks.Items) != 1 {
		t.Fatalf("expected one pick task: %+v", tasks)
	}
	task, err := service.StartPickTask(ctx, tasks.Items[0].ID, account.ID)
	outboundOK(t, err)
	if task.StatusCode != "IN_PROGRESS" {
		t.Fatalf("task not started: %+v", task)
	}
	confirmationRequest := dto.ConfirmPickRequest{PickedQty: "6", ExpectedBalanceVersion: task.BalanceVersion, BusinessDate: businessDate, OperationKey: "pick-" + suffix}
	confirmation, err := service.ConfirmPick(ctx, task.ID, confirmationRequest, account.ID)
	outboundOK(t, err)
	if confirmation.Task.StatusCode != "COMPLETED" {
		t.Fatalf("task not completed: %+v", confirmation)
	}
	replay, err := service.ConfirmPick(ctx, task.ID, confirmationRequest, account.ID)
	outboundOK(t, err)
	if replay.MovementID != confirmation.MovementID {
		t.Fatal("pick operation-key replay created a different movement")
	}
	outboundOK(t, tx.Where("balance_id=?", balance.ID).Take(&balance).Error)
	if balance.OnHandQty != "4.000000" || balance.ReservedQty != "0.000000" {
		t.Fatalf("unexpected source balance: %+v", balance)
	}
	var stagingBalance inventorymodel.InventoryBalance
	outboundOK(t, tx.Where("balance_id=?", confirmation.StagingBalanceID).Take(&stagingBalance).Error)
	if stagingBalance.OnHandQty != "6.000000" || stagingBalance.LocationID != stagingLocation.ID {
		t.Fatalf("unexpected staging balance: %+v", stagingBalance)
	}
	var reservation model.InventoryReservation
	outboundOK(t, tx.Where("reservation_id=?", allocation.Reservations[0].ID).Take(&reservation).Error)
	var consumed mastermodel.DocumentStatus
	outboundOK(t, tx.Where("document_type_id=? AND code='CONSUMED'", reservation.DocumentTypeID).Take(&consumed).Error)
	if reservation.StatusID != consumed.ID || reservation.PickedQty != "6.000000" {
		t.Fatalf("reservation was not consumed: %+v", reservation)
	}
	wave, err = service.GetWave(ctx, wave.ID)
	outboundOK(t, err)
	if wave.StatusCode != "COMPLETED" {
		t.Fatalf("wave not completed: %+v", wave)
	}
	staging, err := service.CreateStaging(ctx, dto.CreateStagingRequest{OutboundID: order.ID, WaveID: wave.ID}, account.ID)
	outboundOK(t, err)
	if staging.StatusCode != "OPEN" || len(staging.Lines) != 1 {
		t.Fatalf("unexpected staging document: %+v", staging)
	}
	staging, err = service.CompleteStaging(ctx, staging.ID, account.ID)
	outboundOK(t, err)
	if staging.StatusCode != "COMPLETED" {
		t.Fatalf("staging not completed: %+v", staging)
	}
	order, err = service.GetOutboundOrder(ctx, order.ID)
	outboundOK(t, err)
	if order.StatusCode != "STAGED" || order.Lines[0].PickedQty != "6.000000" {
		t.Fatalf("order did not reach staging: %+v", order)
	}

	check, err := service.CreateCheck(ctx, dto.CreateCheckRequest{StagingID: staging.ID}, account.ID)
	outboundOK(t, err)
	if check.StatusCode != "OPEN" || len(check.Lines) != 1 {
		t.Fatalf("unexpected outbound check: %+v", check)
	}
	check, err = service.RecordCheckLine(ctx, check.ID, check.Lines[0].ID, dto.RecordCheckLineRequest{CheckedQty: "6", ResultCode: "DAMAGED", Notes: outboundPointer("Cartons damaged during staging")}, account.ID)
	outboundOK(t, err)
	check, err = service.CompleteCheck(ctx, check.ID, account.ID)
	outboundOK(t, err)
	if check.StatusCode != "FAILED" {
		t.Fatalf("outbound check did not fail: %+v", check)
	}
	exceptions, err := service.ListCheckExceptions(ctx, check.ID)
	outboundOK(t, err)
	if len(exceptions) != 1 {
		t.Fatalf("expected one check exception: %+v", exceptions)
	}
	var correctionType inventorymodel.MovementType
	outboundOK(t, tx.Where("code='OUTBOUND_CHECK_CORRECTION'").Take(&correctionType).Error)
	correctionMovement := inventorymodel.InventoryMovement{ID: "CORR-" + suffix, MovementTypeID: correctionType.ID, OwnerID: owner.ID, WarehouseID: warehouse.ID, BusinessDate: time.Now(), OccurredAt: time.Now(), ItemID: item.ID, FromLocationID: &stagingLocation.ID, FromStatusID: &available.ID, Quantity: "6", UOMID: each.ID, SourceDocumentID: check.ID, CreatedBy: account.ID}
	outboundOK(t, tx.Create(&correctionMovement).Error)
	outboundOK(t, tx.Model(&inventorymodel.InventoryBalance{}).Where("balance_id=?", confirmation.StagingBalanceID).Update("on_hand_qty", "0").Error)
	resolvedException, err := service.ResolveCheckException(ctx, exceptions[0].ID, dto.ResolveCheckExceptionRequest{ResolutionTypeCode: "STOCK_CORRECTION", ResolvedQty: "6", MovementID: &correctionMovement.ID, Notes: outboundPointer("Damaged stock quarantined")}, account.ID)
	outboundOK(t, err)
	if resolvedException.StatusCode == "RESOLVED" {
		t.Fatal("damaged exception resolved before replacement")
	}
	outboundOK(t, tx.Model(&inventorymodel.InventoryBalance{}).Where("balance_id=?", balance.ID).Updates(map[string]interface{}{"on_hand_qty": "10", "version_no": gorm.Expr("version_no+1")}).Error)
	outboundOK(t, tx.Where("balance_id=?", balance.ID).Take(&balance).Error)
	replacement, err := service.CreateReplacementPick(ctx, exceptions[0].ID, dto.CreateReplacementPickRequest{BalanceID: balance.ID, ReplacementQty: "6", ExpectedBalanceVersion: balance.VersionNo, PickingStrategyID: &strategy.ID, PriorityCode: "HIGH", Notes: outboundPointer("Supplemental quality replacement")}, account.ID)
	outboundOK(t, err)
	replacementTask, err := service.StartPickTask(ctx, replacement.Task.ID, account.ID)
	outboundOK(t, err)
	_, err = service.ConfirmPick(ctx, replacementTask.ID, dto.ConfirmPickRequest{PickedQty: "6", ExpectedBalanceVersion: replacementTask.BalanceVersion, BusinessDate: businessDate, OperationKey: "pick-replacement-" + suffix}, account.ID)
	outboundOK(t, err)
	exceptions, err = service.ListCheckExceptions(ctx, check.ID)
	outboundOK(t, err)
	if exceptions[0].StatusCode != "RESOLVED" {
		t.Fatalf("replacement did not resolve exception: %+v", exceptions[0])
	}
	recheck, err := service.CreateCheck(ctx, dto.CreateCheckRequest{StagingID: staging.ID, ParentCheckID: &check.ID, Notes: outboundPointer("Replacement verification")}, account.ID)
	outboundOK(t, err)
	if len(recheck.Lines) != 1 || recheck.Lines[0].ExpectedQty != "6.000000" {
		t.Fatalf("unexpected replacement recheck: %+v", recheck)
	}
	recheck, err = service.RecordCheckLine(ctx, recheck.ID, recheck.Lines[0].ID, dto.RecordCheckLineRequest{CheckedQty: "6", ResultCode: "PASS"}, account.ID)
	outboundOK(t, err)
	check, err = service.CompleteCheck(ctx, recheck.ID, account.ID)
	outboundOK(t, err)
	if check.StatusCode != "PASSED" {
		t.Fatalf("replacement recheck did not pass: %+v", check)
	}

	packing, err := service.CreatePacking(ctx, dto.CreatePackingRequest{OutboundCheckID: recheck.ID, PackingLocationID: packingLocation.ID}, account.ID)
	outboundOK(t, err)
	if packing.StatusCode != "OPEN" || len(packing.Lines) != 1 {
		t.Fatalf("unexpected packing: %+v", packing)
	}
	packing, err = service.PackLine(ctx, packing.ID, packing.Lines[0].OutboundCheckLineID, dto.PackLineRequest{ExpectedBalanceVersion: packing.Lines[0].SourceBalanceVersion, BusinessDate: businessDate, OperationKey: "pack-" + suffix}, account.ID)
	outboundOK(t, err)
	packing, err = service.CompletePacking(ctx, packing.ID, account.ID)
	outboundOK(t, err)
	if packing.StatusCode != "COMPLETED" || packing.Lines[0].PackedQty != "6.000000" {
		t.Fatalf("packing did not complete: %+v", packing)
	}

	carrier, err := service.CreateCarrier(ctx, dto.CreateCarrierRequest{Code: "CAR_" + suffix, Name: "Test carrier"})
	outboundOK(t, err)
	carrierService, err := service.CreateCarrierService(ctx, carrier.ID, dto.CreateCarrierServiceRequest{Code: "REGULAR", Name: "Regular delivery"})
	outboundOK(t, err)
	driver, err := service.CreateCarrierDriver(ctx, carrier.ID, dto.CreateCarrierDriverRequest{Code: "DRV_" + suffix, Name: "Test driver"}, account.ID)
	outboundOK(t, err)
	shipment, err := service.CreateShipment(ctx, dto.CreateShipmentRequest{OwnerID: owner.ID, WarehouseID: warehouse.ID, BusinessDate: businessDate, PackingIDs: []string{packing.ID}, CarrierServiceID: &carrierService.ID, TrackingNumber: outboundPointer("TRACK-" + suffix)}, account.ID)
	outboundOK(t, err)
	if shipment.StatusCode != "PLANNED" || len(shipment.Lines) != 1 {
		t.Fatalf("unexpected shipment: %+v", shipment)
	}
	shipment, err = service.DispatchShipmentLine(ctx, shipment.ID, shipment.Lines[0].PackingLineID, dto.DispatchShipmentLineRequest{ExpectedBalanceVersion: shipment.Lines[0].SourceBalanceVersion, BusinessDate: businessDate, OperationKey: "ship-" + suffix}, account.ID)
	outboundOK(t, err)
	_, err = service.CompleteShipment(ctx, shipment.ID, account.ID)
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("shipment without a primary driver got %v", err)
	}
	drivers, err := service.AssignShipmentDriver(ctx, shipment.ID, dto.AssignShipmentDriverRequest{DriverID: driver.ID, IsPrimary: true}, account.ID)
	outboundOK(t, err)
	if len(drivers) != 1 || !drivers[0].IsPrimary {
		t.Fatalf("unexpected shipment driver assignment: %+v", drivers)
	}
	shipment, err = service.CompleteShipment(ctx, shipment.ID, account.ID)
	outboundOK(t, err)
	if shipment.StatusCode != "SHIPPED" || shipment.Lines[0].ShippedQty != "6.000000" {
		t.Fatalf("shipment did not complete: %+v", shipment)
	}

	now := time.Now()
	delivery, err := service.CreateDelivery(ctx, dto.CreateDeliveryRequest{ShipmentID: shipment.ID, OutboundID: order.ID, BusinessDate: businessDate, PlannedDeliveryAt: outboundPointer(now.Add(time.Hour).Format(time.RFC3339))}, account.ID)
	outboundOK(t, err)
	if delivery.StatusCode != "PLANNED" || len(delivery.Lines) != 1 {
		t.Fatalf("unexpected delivery: %+v", delivery)
	}
	delivery, err = service.DepartDelivery(ctx, delivery.ID, dto.DeliveryEventRequest{EventAt: now.Format(time.RFC3339)}, account.ID)
	outboundOK(t, err)
	delivery, err = service.ArriveDelivery(ctx, delivery.ID, dto.DeliveryEventRequest{EventAt: now.Add(15 * time.Minute).Format(time.RFC3339)}, account.ID)
	outboundOK(t, err)
	delivery, err = service.DeliverLine(ctx, delivery.ID, delivery.Lines[0].ID, dto.DeliverLineRequest{DeliveredQty: "6", EventAt: now.Add(20 * time.Minute).Format(time.RFC3339), RecipientName: "Store receiver", ProofReference: "POD-" + suffix}, account.ID)
	outboundOK(t, err)
	delivery, err = service.CompleteDelivery(ctx, delivery.ID, dto.CompleteDeliveryRequest{DeliveredAt: now.Add(20 * time.Minute).Format(time.RFC3339)}, account.ID)
	outboundOK(t, err)
	if delivery.StatusCode != "DELIVERED" || delivery.Lines[0].DeliveredQty != "6.000000" {
		t.Fatalf("delivery did not complete: %+v", delivery)
	}
	order, err = service.GetOutboundOrder(ctx, order.ID)
	outboundOK(t, err)
	if order.StatusCode != "DELIVERED" || order.Lines[0].CheckedQty != "6.000000" || order.Lines[0].PackedQty != "6.000000" || order.Lines[0].ShippedQty != "6.000000" || order.Lines[0].DeliveredQty != "6.000000" {
		t.Fatalf("phase-two quantities did not reach the order: %+v", order)
	}

	shortageOrder, err := service.CreateOutboundOrder(ctx, dto.CreateOutboundOrderRequest{OwnerID: owner.ID, CustomerID: customer.ID, ShipToPartnerID: shipTo.ID, WarehouseID: warehouse.ID, BusinessDate: businessDate, ClientDeliveryOrderNo: "CLIENT-DO-SHORT-" + suffix, RequestedShipAt: businessDate + "T16:00:00+07:00", Lines: []dto.OutboundOrderLineRequest{{ItemID: item.ID, OrderedQty: "6"}}}, account.ID)
	outboundOK(t, err)
	_, err = service.ValidateOutboundOrder(ctx, shortageOrder.ID, dto.ValidateOutboundRequest{ExpectedVersion: shortageOrder.VersionNo}, account.ID)
	outboundOK(t, err)
	shortageOrder, err = service.GetOutboundOrder(ctx, shortageOrder.ID)
	outboundOK(t, err)
	shortageOrder, err = service.ReleaseOutboundOrder(ctx, shortageOrder.ID, dto.TransitionRequest{ExpectedVersion: shortageOrder.VersionNo}, account.ID)
	outboundOK(t, err)
	_, err = service.AllocateOutboundOrder(ctx, shortageOrder.ID, dto.AllocateOutboundRequest{ExpectedVersion: shortageOrder.VersionNo, PickingStrategyID: &strategy.ID}, account.ID)
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("full allocation shortage got %v", err)
	}
	outboundOK(t, tx.Where("balance_id=?", balance.ID).Take(&balance).Error)
	if balance.ReservedQty != "0.000000" {
		t.Fatalf("failed allocation was not rolled back: %+v", balance)
	}
	partialAllocation, err := service.AllocateOutboundOrder(ctx, shortageOrder.ID, dto.AllocateOutboundRequest{ExpectedVersion: shortageOrder.VersionNo, PickingStrategyID: &strategy.ID, AllowPartial: true}, account.ID)
	outboundOK(t, err)
	if partialAllocation.Order.StatusCode != "PARTIALLY_ALLOCATED" || partialAllocation.AllocatedQty != "4.000000" || partialAllocation.ShortageQty != "2.000000" {
		t.Fatalf("unexpected partial allocation: %+v", partialAllocation)
	}
	shortageOrder, err = service.ReleaseReservation(ctx, partialAllocation.Reservations[0].ID, dto.ReleaseReservationRequest{Reason: "Release shortage study reservation"}, account.ID)
	outboundOK(t, err)
	if shortageOrder.StatusCode != "RELEASED" || shortageOrder.Lines[0].AllocatedQty != "0.000000" {
		t.Fatalf("reservation release did not restore order: %+v", shortageOrder)
	}

	returnOrder, err := service.CreateOutboundOrder(ctx, dto.CreateOutboundOrderRequest{OwnerID: owner.ID, CustomerID: customer.ID, ShipToPartnerID: shipTo.ID, WarehouseID: warehouse.ID, BusinessDate: businessDate, ClientDeliveryOrderNo: "CLIENT-DO-RETURN-" + suffix, RequestedShipAt: businessDate + "T17:00:00+07:00", Lines: []dto.OutboundOrderLineRequest{{ItemID: item.ID, OrderedQty: "4"}}}, account.ID)
	outboundOK(t, err)
	_, err = service.ValidateOutboundOrder(ctx, returnOrder.ID, dto.ValidateOutboundRequest{ExpectedVersion: returnOrder.VersionNo}, account.ID)
	outboundOK(t, err)
	returnOrder, err = service.GetOutboundOrder(ctx, returnOrder.ID)
	outboundOK(t, err)
	returnOrder, err = service.ReleaseOutboundOrder(ctx, returnOrder.ID, dto.TransitionRequest{ExpectedVersion: returnOrder.VersionNo}, account.ID)
	outboundOK(t, err)
	returnAllocation, err := service.AllocateOutboundOrder(ctx, returnOrder.ID, dto.AllocateOutboundRequest{ExpectedVersion: returnOrder.VersionNo, PickingStrategyID: &strategy.ID}, account.ID)
	outboundOK(t, err)
	if returnAllocation.Order.StatusCode != "ALLOCATED" {
		t.Fatalf("return-path order did not allocate: %+v", returnAllocation)
	}
	returnWave, err := service.CreateWave(ctx, dto.CreateWaveRequest{OwnerID: owner.ID, WarehouseID: warehouse.ID, BusinessDate: businessDate, WaveTypeCode: "SINGLE_ORDER", PickingStrategyID: &strategy.ID, OutboundIDs: []string{returnOrder.ID}}, account.ID)
	outboundOK(t, err)
	returnWave, err = service.ReleaseWave(ctx, returnWave.ID, dto.ReleaseWaveRequest{ExpectedVersion: returnWave.VersionNo, StagingLocationID: stagingLocation.ID, PriorityCode: "NORMAL"}, account.ID)
	outboundOK(t, err)
	tasks, err = service.ListPickTasks(ctx, repository.ListFilter{OwnerID: owner.ID, WarehouseID: warehouse.ID, Page: 1, PageSize: 20}, "")
	outboundOK(t, err)
	var returnTask dto.PickTaskResponse
	for _, candidate := range tasks.Items {
		if candidate.OutboundID == returnOrder.ID {
			returnTask = candidate
			break
		}
	}
	if returnTask.ID == "" {
		t.Fatal("return-path pick task was not found")
	}
	returnTask, err = service.StartPickTask(ctx, returnTask.ID, account.ID)
	outboundOK(t, err)
	_, err = service.ConfirmPick(ctx, returnTask.ID, dto.ConfirmPickRequest{PickedQty: "4", ExpectedBalanceVersion: returnTask.BalanceVersion, BusinessDate: businessDate, OperationKey: "pick-return-" + suffix}, account.ID)
	outboundOK(t, err)
	returnStaging, err := service.CreateStaging(ctx, dto.CreateStagingRequest{OutboundID: returnOrder.ID, WaveID: returnWave.ID}, account.ID)
	outboundOK(t, err)
	returnStaging, err = service.CompleteStaging(ctx, returnStaging.ID, account.ID)
	outboundOK(t, err)
	returnCheck, err := service.CreateCheck(ctx, dto.CreateCheckRequest{StagingID: returnStaging.ID}, account.ID)
	outboundOK(t, err)
	returnCheck, err = service.RecordCheckLine(ctx, returnCheck.ID, returnCheck.Lines[0].ID, dto.RecordCheckLineRequest{CheckedQty: "4", ResultCode: "PASS"}, account.ID)
	outboundOK(t, err)
	returnCheck, err = service.CompleteCheck(ctx, returnCheck.ID, account.ID)
	outboundOK(t, err)
	returnPacking, err := service.CreatePacking(ctx, dto.CreatePackingRequest{OutboundCheckID: returnCheck.ID, PackingLocationID: packingLocation.ID}, account.ID)
	outboundOK(t, err)
	returnPacking, err = service.PackLine(ctx, returnPacking.ID, returnPacking.Lines[0].OutboundCheckLineID, dto.PackLineRequest{ExpectedBalanceVersion: returnPacking.Lines[0].SourceBalanceVersion, BusinessDate: businessDate, OperationKey: "pack-return-" + suffix}, account.ID)
	outboundOK(t, err)
	returnPacking, err = service.CompletePacking(ctx, returnPacking.ID, account.ID)
	outboundOK(t, err)
	returnShipment, err := service.CreateShipment(ctx, dto.CreateShipmentRequest{OwnerID: owner.ID, WarehouseID: warehouse.ID, BusinessDate: businessDate, PackingIDs: []string{returnPacking.ID}}, account.ID)
	outboundOK(t, err)
	returnShipment, err = service.DispatchShipmentLine(ctx, returnShipment.ID, returnShipment.Lines[0].PackingLineID, dto.DispatchShipmentLineRequest{ExpectedBalanceVersion: returnShipment.Lines[0].SourceBalanceVersion, BusinessDate: businessDate, OperationKey: "ship-return-" + suffix}, account.ID)
	outboundOK(t, err)
	returnShipment, err = service.CompleteShipment(ctx, returnShipment.ID, account.ID)
	outboundOK(t, err)
	returnDelivery, err := service.CreateDelivery(ctx, dto.CreateDeliveryRequest{ShipmentID: returnShipment.ID, OutboundID: returnOrder.ID, BusinessDate: businessDate}, account.ID)
	outboundOK(t, err)
	returnDelivery, err = service.DepartDelivery(ctx, returnDelivery.ID, dto.DeliveryEventRequest{EventAt: now.Add(30 * time.Minute).Format(time.RFC3339)}, account.ID)
	outboundOK(t, err)
	returnDelivery, err = service.ArriveDelivery(ctx, returnDelivery.ID, dto.DeliveryEventRequest{EventAt: now.Add(45 * time.Minute).Format(time.RFC3339)}, account.ID)
	outboundOK(t, err)
	returnDelivery, err = service.DeliverLine(ctx, returnDelivery.ID, returnDelivery.Lines[0].ID, dto.DeliverLineRequest{DeliveredQty: "1", EventAt: now.Add(50 * time.Minute).Format(time.RFC3339), RecipientName: "Store receiver", ProofReference: "PARTIAL-POD-" + suffix}, account.ID)
	outboundOK(t, err)
	returnDelivery, err = service.FailDelivery(ctx, returnDelivery.ID, dto.FailDeliveryRequest{ReasonCode: "RECIPIENT_REJECTED", EventAt: now.Add(51 * time.Minute).Format(time.RFC3339), Notes: outboundPointer("Store rejected the remaining cartons")}, account.ID)
	outboundOK(t, err)
	var hold mastermodel.InventoryStatus
	outboundOK(t, tx.Where("code='HOLD'").Take(&hold).Error)
	_, err = service.UpsertReturnPolicy(ctx, dto.UpsertReturnPolicyRequest{OwnerID: owner.ID, WarehouseID: warehouse.ID, ReturnLocationID: stagingLocation.ID, ReturnInventoryStatusID: hold.ID, IsActive: true}, account.ID)
	outboundOK(t, err)
	returned, err := service.ReturnDeliveryLine(ctx, returnDelivery.ID, returnDelivery.Lines[0].ID, dto.ReturnDeliveryLineRequest{ReturnedQty: "3", EventAt: now.Add(time.Hour).Format(time.RFC3339), OperationKey: "delivery-return-" + suffix, Notes: outboundPointer("Returned after recipient rejection")}, account.ID)
	outboundOK(t, err)
	if returned.ReturnedQty != "3.000000" {
		t.Fatalf("unexpected returned stock: %+v", returned)
	}
	returnDelivery, err = service.CloseReturnedDelivery(ctx, returnDelivery.ID, dto.DeliveryEventRequest{EventAt: now.Add(time.Hour).Format(time.RFC3339), Notes: outboundPointer("Vehicle returned to depot")}, account.ID)
	outboundOK(t, err)
	if returnDelivery.StatusCode != "PARTIALLY_DELIVERED" || returnDelivery.Lines[0].DeliveredQty != "1.000000" || returnDelivery.Lines[0].ReturnedQty != "3.000000" {
		t.Fatalf("failed delivery did not close correctly: %+v", returnDelivery)
	}
	var returnedBalance inventorymodel.InventoryBalance
	outboundOK(t, tx.Where("balance_id=?", returned.ReturnedBalanceID).Take(&returnedBalance).Error)
	if returnedBalance.LocationID != stagingLocation.ID || returnedBalance.InventoryStatusID != hold.ID || returnedBalance.OnHandQty != "3.000000" {
		t.Fatalf("returned stock was not quarantined by policy: %+v", returnedBalance)
	}
}
