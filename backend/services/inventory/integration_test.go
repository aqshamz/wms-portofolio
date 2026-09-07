package inventory

import (
	"context"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"strings"
	"testing"
	"time"
	"wms-api/config"
	dto "wms-api/dto/inventory"
	masterdto "wms-api/dto/master"
	authmodel "wms-api/models/authentication"
	model "wms-api/models/inventory"
	master "wms-api/models/master"
	authrepo "wms-api/repository/authentication"
	repository "wms-api/repository/inventory"
	masterrepo "wms-api/repository/master"
	masterservice "wms-api/services/master"
)

func ok(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func want(t *testing.T, err, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("got %v, want %v", err, target)
	}
}
func ptr[T any](v T) *T { return &v }

// Opt-in integration tests: migrations, schemas and all fixtures are rolled back.
func TestInventoryIdentityPostgreSQL(t *testing.T) {
	if os.Getenv("WMS_INTEGRATION_TEST") != "1" {
		t.Skip("set WMS_INTEGRATION_TEST=1")
	}
	for _, fresh := range []bool{false, true} {
		t.Run(fmt.Sprintf("fresh_%t", fresh), func(t *testing.T) { testIdentity(t, fresh) })
	}
}
func testIdentity(t *testing.T, fresh bool) {
	_ = godotenv.Load("../../.env")
	cfg, err := config.Load()
	ok(t, err)
	cfg.Database.AutoCreate = false
	db, err := config.OpenDatabase(cfg.Database)
	ok(t, err)
	sqlDB, err := db.DB()
	ok(t, err)
	defer sqlDB.Close()
	tx := db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Begin()
	ok(t, tx.Error)
	defer tx.Rollback()
	suffix := fmt.Sprint(time.Now().UnixNano())
	if fresh {
		schema := "identity_test_" + suffix
		ok(t, tx.Exec("CREATE SCHEMA "+schema).Error)
		ok(t, tx.Exec("SET LOCAL search_path TO "+schema).Error)
	}
	ok(t, authrepo.Migrate(tx))
	ok(t, masterrepo.Migrate(tx))
	ok(t, masterrepo.MigrateCatalog(tx))
	ok(t, repository.Migrate(tx))
	ok(t, repository.Migrate(tx))
	ctx := context.Background()
	repos := repository.NewRepositories(tx)
	s := NewService(repos)
	ok(t, repos.Catalog.HandlingUnitType.SeedDefaults(ctx))
	ok(t, repos.Catalog.HandlingUnitType.SeedDefaults(ctx))
	kinds, _, err := repos.Catalog.HandlingUnitType.List(ctx, masterrepo.CatalogFilter{Page: 1, PageSize: 100})
	ok(t, err)
	var pallet, carton string
	for _, v := range kinds {
		if v.Code == "PALLET" {
			pallet = v.ID
		}
		if v.Code == "CARTON" {
			carton = v.ID
		}
	}
	if pallet == "" || carton == "" {
		t.Fatal("missing HU seeds")
	}
	status := authmodel.AccountStatus{Code: "IT_" + suffix, Name: "Identity test", AllowsLogin: true}
	ok(t, tx.Create(&status).Error)
	account := authmodel.AppAccount{Username: "identity_" + suffix, DisplayName: "Identity test", AccountStatusID: status.ID, ExternalSubject: ptr("identity_" + suffix)}
	ok(t, tx.Create(&account).Error)
	owner := master.Organization{Code: "IO_" + suffix, Name: "Owner", TimezoneName: "Asia/Jakarta"}
	other := master.Organization{Code: "IX_" + suffix, Name: "Other owner", TimezoneName: "Asia/Jakarta"}
	ok(t, tx.Create(&owner).Error)
	ok(t, tx.Create(&other).Error)
	unit := master.UOM{Code: "U" + suffix, Name: "Each"}
	ok(t, tx.Create(&unit).Error)
	item := master.Item{OwnerID: owner.ID, Code: "II_" + suffix, Name: "Tracked item", BaseUOMID: unit.ID, LotControlled: true, SerialControlled: true}
	ok(t, tx.Create(&item).Error)
	warehouse := master.Warehouse{OperatorID: owner.ID, Code: "IW_" + suffix, Name: "Warehouse", TimezoneName: "Asia/Jakarta"}
	elsewhere := master.Warehouse{OperatorID: owner.ID, Code: "IE_" + suffix, Name: "Elsewhere", TimezoneName: "Asia/Jakarta"}
	ok(t, tx.Create(&warehouse).Error)
	ok(t, tx.Create(&elsewhere).Error)
	assignment := master.WarehouseOwner{WarehouseID: warehouse.ID, OwnerID: owner.ID}
	ok(t, tx.Create(&assignment).Error)
	zone := master.WarehouseZone{WarehouseID: warehouse.ID, Code: "IZ_" + suffix, Name: "Zone"}
	otherZone := master.WarehouseZone{WarehouseID: elsewhere.ID, Code: "IZ_" + suffix, Name: "Other zone"}
	ok(t, tx.Create(&zone).Error)
	ok(t, tx.Create(&otherZone).Error)
	locationType := master.LocationType{Code: "IL_" + suffix, Name: "Storage", AllowsStorage: true}
	ok(t, tx.Create(&locationType).Error)
	location := master.WarehouseLocation{WarehouseID: warehouse.ID, ZoneID: zone.ID, LocationTypeID: locationType.ID, Code: "A"}
	otherLocation := master.WarehouseLocation{WarehouseID: elsewhere.ID, ZoneID: otherZone.ID, LocationTypeID: locationType.ID, Code: "B"}
	ok(t, tx.Create(&location).Error)
	ok(t, tx.Create(&otherLocation).Error)
	quality := master.QualityStatus{Code: "IQ_" + suffix, Name: "Pending"}
	ok(t, tx.Create(&quality).Error)
	available := master.InventoryStatus{Code: "IA_" + suffix, Name: "Available", IsAllocatable: true, IsPickable: true}
	hold := master.InventoryStatus{Code: "IH_" + suffix, Name: "Hold"}
	ok(t, tx.Create(&available).Error)
	ok(t, tx.Create(&hold).Error)

	q := dto.CreateLotRequest{OwnerID: strings.ToUpper(owner.ID), ItemID: item.ID, LotNumber: " Batch-a ", ManufactureDate: ptr("2026-09-01"), ExpiryDate: ptr("2027-09-01"), QualityStatusID: &quality.ID}
	lot, err := s.CreateLot(ctx, q, account.ID)
	ok(t, err)
	if lot.LotNumber != "Batch-a" || lot.OwnerID != owner.ID || *lot.ManufactureDate != "2026-09-01" || lot.CreatedBy == nil || *lot.CreatedBy != account.ID {
		t.Fatalf("bad lot %+v", lot)
	}
	fetched, err := s.GetLot(ctx, lot.ID)
	ok(t, err)
	if fetched.ID != lot.ID {
		t.Fatal("get lot")
	}
	_, err = s.CreateLot(ctx, q, account.ID)
	want(t, err, repository.ErrConflict)
	q.LotNumber = "bad-date"
	q.ExpiryDate = ptr("2026-08-31")
	_, err = s.CreateLot(ctx, q, account.ID)
	want(t, err, ErrInvalidInput)
	q.ExpiryDate = nil
	q.OwnerID = other.ID
	_, err = s.CreateLot(ctx, q, account.ID)
	want(t, err, ErrInvalidInput)
	q.OwnerID = owner.ID
	for _, target := range []struct{ table, key, id, field string }{
		{"organization", "organization_id", owner.ID, "is_active"}, {"item", "item_id", item.ID, "is_active"},
		{"item", "item_id", item.ID, "lot_controlled"}, {"quality_status", "quality_status_id", quality.ID, "is_active"},
	} {
		ok(t, tx.Table(target.table).Where(target.key+" = ?", target.id).Update(target.field, false).Error)
		_, err = s.CreateLot(ctx, q, account.ID)
		want(t, err, ErrInvalidInput)
		ok(t, tx.Table(target.table).Where(target.key+" = ?", target.id).Update(target.field, true).Error)
	}
	listed, err := s.ListLot(ctx, repository.Filter{OwnerID: owner.ID, ItemID: item.ID, Number: "Batch-a", Page: 1, PageSize: 1})
	ok(t, err)
	if listed.TotalItems != 1 || len(listed.Items) != 1 {
		t.Fatal("lot exact filter")
	}
	_, err = s.GetLot(ctx, "LOT-missing")
	want(t, err, repository.ErrNotFound)

	serial, err := s.CreateSerial(ctx, dto.CreateSerialRequest{OwnerID: owner.ID, ItemID: item.ID, SerialNo: " SN-001 "}, account.ID)
	ok(t, err)
	if serial.SerialNo != "SN-001" {
		t.Fatal("serial normalization")
	}
	serialRead, err := s.GetSerial(ctx, serial.ID)
	ok(t, err)
	if serialRead.ID != serial.ID {
		t.Fatal("serial get")
	}
	serials, err := s.ListSerial(ctx, repository.Filter{OwnerID: owner.ID, ItemID: item.ID, Number: "SN-001", Page: 1, PageSize: 20})
	ok(t, err)
	if serials.TotalItems != 1 {
		t.Fatal("serial exact filter")
	}
	_, err = s.CreateSerial(ctx, dto.CreateSerialRequest{OwnerID: owner.ID, ItemID: item.ID, SerialNo: "SN-001"}, account.ID)
	want(t, err, repository.ErrConflict)
	ok(t, tx.Model(&item).Update("serial_controlled", false).Error)
	_, err = s.CreateSerial(ctx, dto.CreateSerialRequest{OwnerID: owner.ID, ItemID: item.ID, SerialNo: "SN-002"}, account.ID)
	want(t, err, ErrInvalidInput)
	ok(t, tx.Model(&item).Update("serial_controlled", true).Error)
	_, err = s.CreateSerial(ctx, dto.CreateSerialRequest{OwnerID: other.ID, ItemID: item.ID, SerialNo: "SN-002"}, account.ID)
	want(t, err, ErrInvalidInput)

	// The core posting service is internal, atomic and idempotent. A serialized
	// posting represents exactly one unit and maintains its current-state pointer.
	receive := dto.PostingRequest{OperationKey: "receive-" + suffix, MovementTypeCode: "RECEIVE", OwnerID: owner.ID, WarehouseID: warehouse.ID,
		BusinessDate: "2026-09-07", ItemID: item.ID, LotID: &lot.ID, To: &dto.BalanceDimension{LocationID: location.ID, InventoryStatusID: available.ID},
		Quantity: "1", SerialIDs: []string{serial.ID}, SourceDocumentID: "TEST-RECEIPT-" + suffix}
	posted, err := s.PostMovement(ctx, receive, account.ID)
	ok(t, err)
	if posted.IdempotentReplay || posted.ToBalance == nil || posted.ToBalance.OnHandQty != "1.000000" || posted.Movement.UOMID != unit.ID {
		t.Fatalf("bad receive posting: %+v", posted)
	}
	replay, err := s.PostMovement(ctx, receive, account.ID)
	ok(t, err)
	if !replay.IdempotentReplay || replay.Movement.ID != posted.Movement.ID {
		t.Fatal("idempotent replay created a different movement")
	}
	changed := receive
	changed.Quantity = "2"
	_, err = s.PostMovement(ctx, changed, account.ID)
	want(t, err, ErrInvalidInput)
	state, err := s.GetSerialState(ctx, serial.ID)
	ok(t, err)
	if state.Balance.ID != posted.ToBalance.ID || state.SerialNo != "SN-001" {
		t.Fatalf("bad serial state: %+v", state)
	}

	statusChange := dto.PostingRequest{OperationKey: "status-" + suffix, MovementTypeCode: "STATUS_CHANGE", OwnerID: owner.ID, WarehouseID: warehouse.ID,
		BusinessDate: "2026-09-07", ItemID: item.ID, LotID: &lot.ID, From: &dto.BalanceDimension{LocationID: location.ID, InventoryStatusID: available.ID},
		To: &dto.BalanceDimension{LocationID: location.ID, InventoryStatusID: hold.ID}, Quantity: "1.000000", SerialIDs: []string{serial.ID}, SourceDocumentID: "TEST-QC-" + suffix}
	changedStatus, err := s.PostMovement(ctx, statusChange, account.ID)
	ok(t, err)
	if changedStatus.FromBalance.OnHandQty != "0.000000" || changedStatus.ToBalance.OnHandQty != "1.000000" {
		t.Fatal("status balances not transferred")
	}
	state, err = s.GetSerialState(ctx, serial.ID)
	ok(t, err)
	if state.Balance.InventoryStatusID != hold.ID || state.VersionNo != 2 {
		t.Fatal("serial state did not move")
	}

	ship := dto.PostingRequest{OperationKey: "ship-" + suffix, MovementTypeCode: "SHIP", OwnerID: owner.ID, WarehouseID: warehouse.ID,
		BusinessDate: "2026-09-07", ItemID: item.ID, LotID: &lot.ID, From: &dto.BalanceDimension{LocationID: location.ID, InventoryStatusID: hold.ID},
		Quantity: "1", SerialIDs: []string{serial.ID}, SourceDocumentID: "TEST-SHIP-" + suffix}
	_, err = s.PostMovement(ctx, ship, account.ID)
	ok(t, err)
	_, err = s.GetSerialState(ctx, serial.ID)
	want(t, err, repository.ErrNotFound)

	// Non-serialized base-UOM decimals aggregate in balances and cannot reduce
	// on-hand below reserved quantity.
	commodity := master.Item{OwnerID: owner.ID, Code: "IC_" + suffix, Name: "Commodity", BaseUOMID: unit.ID}
	ok(t, tx.Create(&commodity).Error)
	receiveBulk := dto.PostingRequest{OperationKey: "bulk-" + suffix, MovementTypeCode: "RECEIVE", OwnerID: owner.ID, WarehouseID: warehouse.ID,
		BusinessDate: "2026-09-07", ItemID: commodity.ID, To: &dto.BalanceDimension{LocationID: location.ID, InventoryStatusID: available.ID}, Quantity: "10.5", SourceDocumentID: "TEST-BULK-" + suffix}
	bulk, err := s.PostMovement(ctx, receiveBulk, account.ID)
	ok(t, err)
	if bulk.ToBalance.OnHandQty != "10.500000" {
		t.Fatal("bulk quantity not normalized")
	}
	ok(t, tx.Model(&model.InventoryBalance{}).Where("balance_id = ?", bulk.ToBalance.ID).Update("reserved_qty", "6.000000").Error)
	issue := dto.PostingRequest{OperationKey: "issue-" + suffix, MovementTypeCode: "SHIP", OwnerID: owner.ID, WarehouseID: warehouse.ID,
		BusinessDate: "2026-09-07", ItemID: commodity.ID, From: &dto.BalanceDimension{LocationID: location.ID, InventoryStatusID: available.ID}, Quantity: "5", SourceDocumentID: "TEST-ISSUE-" + suffix}
	_, err = s.PostMovement(ctx, issue, account.ID)
	want(t, err, ErrInvalidInput)
	issue.Quantity = "4.5"
	issued, err := s.PostMovement(ctx, issue, account.ID)
	ok(t, err)
	if issued.FromBalance.OnHandQty != "6.000000" || issued.FromBalance.AvailableQty != "0.000000" {
		t.Fatal("reserved invariant broken")
	}
	balances, err := s.ListBalances(ctx, repository.BalanceFilter{OwnerID: owner.ID, WarehouseID: warehouse.ID, ItemID: commodity.ID, Page: 1, PageSize: 20})
	ok(t, err)
	if balances.TotalItems != 1 {
		t.Fatal("balance inquiry")
	}
	movements, err := s.ListMovements(ctx, repository.MovementFilter{OwnerID: owner.ID, WarehouseID: warehouse.ID, Page: 1, PageSize: 20})
	ok(t, err)
	if movements.TotalItems != 5 {
		t.Fatalf("expected 5 movements, got %d", movements.TotalItems)
	}
	states, err := s.ListSerialStates(ctx, repository.SerialStateFilter{OwnerID: owner.ID, WarehouseID: warehouse.ID, Page: 1, PageSize: 20})
	ok(t, err)
	if states.TotalItems != 0 {
		t.Fatal("shipped serial still appears in inventory")
	}

	huRequest := dto.CreateHandlingUnitRequest{OwnerID: owner.ID, WarehouseID: warehouse.ID, HandlingUnitTypeID: pallet, CurrentLocationID: &location.ID, Barcode: "PLT-" + suffix}
	hu, err := s.CreateHandlingUnit(ctx, huRequest, account.ID)
	ok(t, err)
	if hu.IsClosed || hu.CurrentLocationID == nil || *hu.CurrentLocationID != location.ID {
		t.Fatal("HU defaults")
	}
	huRead, err := s.GetHandlingUnit(ctx, hu.ID)
	ok(t, err)
	if huRead.Barcode != hu.Barcode {
		t.Fatal("HU get")
	}
	_, err = s.CreateHandlingUnit(ctx, huRequest, account.ID)
	want(t, err, repository.ErrConflict)
	childRequest := dto.CreateHandlingUnitRequest{OwnerID: owner.ID, WarehouseID: warehouse.ID, HandlingUnitTypeID: carton, ParentHandlingUnitID: &hu.ID, Barcode: "CTN-" + suffix}
	child, err := s.CreateHandlingUnit(ctx, childRequest, account.ID)
	ok(t, err)
	if child.CurrentLocationID == nil || *child.CurrentLocationID != location.ID {
		t.Fatal("child must inherit parent location")
	}
	childRequest.Barcode = "CTN2-" + suffix
	childRequest.CurrentLocationID = &otherLocation.ID
	_, err = s.CreateHandlingUnit(ctx, childRequest, account.ID)
	want(t, err, ErrInvalidInput)
	childRequest.CurrentLocationID = nil
	ok(t, tx.Model(&model.HandlingUnit{}).Where("handling_unit_id = ?", hu.ID).Update("is_closed", true).Error)
	_, err = s.CreateHandlingUnit(ctx, childRequest, account.ID)
	want(t, err, ErrInvalidInput)
	ok(t, tx.Model(&model.HandlingUnit{}).Where("handling_unit_id = ?", hu.ID).Update("is_closed", false).Error)
	// A closed ancestor (not just the direct parent) also rejects nesting.
	childRequest.ParentHandlingUnitID = &child.ID
	ok(t, tx.Model(&model.HandlingUnit{}).Where("handling_unit_id = ?", hu.ID).Update("is_closed", true).Error)
	_, err = s.CreateHandlingUnit(ctx, childRequest, account.ID)
	want(t, err, ErrInvalidInput)
	ok(t, tx.Model(&model.HandlingUnit{}).Where("handling_unit_id = ?", hu.ID).Update("is_closed", false).Error)
	huRequest.Barcode = "OTHER-" + suffix
	huRequest.CurrentLocationID = &otherLocation.ID
	_, err = s.CreateHandlingUnit(ctx, huRequest, account.ID)
	want(t, err, ErrInvalidInput)
	huRequest.CurrentLocationID = &location.ID
	for _, target := range []struct {
		table, key, id, field string
		bad, good             bool
	}{
		{"organization", "organization_id", owner.ID, "is_active", false, true},
		{"warehouse", "warehouse_id", warehouse.ID, "is_active", false, true},
		{"handling_unit_type", "handling_unit_type_id", pallet, "is_active", false, true},
		{"warehouse_location", "location_id", location.ID, "is_locked", true, false},
		{"warehouse_location", "location_id", location.ID, "is_active", false, true},
		{"warehouse_zone", "zone_id", zone.ID, "is_active", false, true},
		{"location_type", "location_type_id", locationType.ID, "is_active", false, true},
	} {
		ok(t, tx.Table(target.table).Where(target.key+" = ?", target.id).Update(target.field, target.bad).Error)
		_, err = s.CreateHandlingUnit(ctx, huRequest, account.ID)
		want(t, err, ErrInvalidInput)
		ok(t, tx.Table(target.table).Where(target.key+" = ?", target.id).Update(target.field, target.good).Error)
	}
	ok(t, tx.Model(&assignment).Update("is_active", false).Error)
	_, err = s.CreateHandlingUnit(ctx, huRequest, account.ID)
	want(t, err, ErrInvalidInput)
	ok(t, tx.Model(&assignment).Update("is_active", true).Error)
	huRequest.OwnerID = other.ID
	_, err = s.CreateHandlingUnit(ctx, huRequest, account.ID)
	want(t, err, ErrInvalidInput)
	hus, err := s.ListHandlingUnit(ctx, repository.Filter{OwnerID: owner.ID, WarehouseID: warehouse.ID, ParentID: hu.ID, Page: 1, PageSize: 20})
	ok(t, err)
	if hus.TotalItems != 1 || hus.Items[0].ID != child.ID {
		t.Fatal("HU parent filter")
	}
	// Legacy imports may contain cycles. Reject them without hanging.
	ok(t, tx.Model(&model.HandlingUnit{}).Where("handling_unit_id = ?", hu.ID).Update("parent_handling_unit_id", child.ID).Error)
	_, err = s.CreateHandlingUnit(ctx, childRequest, account.ID)
	want(t, err, ErrInvalidInput)
	ok(t, tx.Model(&model.HandlingUnit{}).Where("handling_unit_id = ?", hu.ID).Update("parent_handling_unit_id", nil).Error)
	// A second owner is allowed in the warehouse, but cannot use this owner's parent.
	ok(t, tx.Create(&master.WarehouseOwner{OwnerID: other.ID, WarehouseID: warehouse.ID}).Error)
	childRequest.OwnerID = other.ID
	_, err = s.CreateHandlingUnit(ctx, childRequest, account.ID)
	want(t, err, ErrInvalidInput)
	// Strong composite constraints protect even direct SQL imports.
	rawBad := []interface{}{
		&model.InventoryLot{ID: "BAD-LOT-" + suffix, OwnerID: other.ID, ItemID: item.ID, LotNumber: "bad"},
		&model.SerialNumber{ID: "BAD-SER-" + suffix, OwnerID: other.ID, ItemID: item.ID, SerialNo: "bad"},
		&model.HandlingUnit{ID: "BAD-HU-" + suffix, OwnerID: owner.ID, WarehouseID: warehouse.ID, HandlingUnitTypeID: pallet, CurrentLocationID: &otherLocation.ID, Barcode: "bad-" + suffix},
		&model.HandlingUnit{ID: "BAD-PARENT-" + suffix, OwnerID: other.ID, WarehouseID: warehouse.ID, HandlingUnitTypeID: pallet, ParentHandlingUnitID: &hu.ID, Barcode: "bad-parent-" + suffix},
		&model.InventoryLot{ID: "BAD-DATE-" + suffix, OwnerID: owner.ID, ItemID: item.ID, LotNumber: "bad-date", ManufactureDate: ptr(time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC)), ExpiryDate: ptr(time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC))},
	}
	for _, v := range rawBad {
		err = tx.Transaction(func(nested *gorm.DB) error { return nested.Create(v).Error })
		want(t, repository.Error(err), repository.ErrConstraint)
	}
	// Type master CRUD and repeatable seed preserve edited/deactivated rows.
	catalog := masterservice.NewCatalogService(repos.Catalog)
	custom, err := catalog.CreateHandlingUnitType(ctx, masterdto.CreateHandlingUnitTypeRequest{Code: "ITYPE_" + suffix, Name: "Test", MaxWeight: ptr("100.5")})
	ok(t, err)
	_, err = catalog.UpdateHandlingUnitType(ctx, custom.ID, masterdto.UpdateHandlingUnitTypeRequest{Name: "Edited", IsActive: ptr(false)})
	ok(t, err)
	ok(t, repos.Catalog.HandlingUnitType.SeedDefaults(ctx))
	custom, err = catalog.GetHandlingUnitType(ctx, custom.ID)
	ok(t, err)
	if custom.IsActive || custom.Name != "Edited" {
		t.Fatal("type edits not preserved")
	}
	ok(t, tx.Model(&master.HandlingUnitType{}).Where("handling_unit_type_id = ?", pallet).Updates(map[string]interface{}{"name": "Edited pallet", "is_active": false}).Error)
	ok(t, repos.Catalog.HandlingUnitType.SeedDefaults(ctx))
	palletRead, err := catalog.GetHandlingUnitType(ctx, pallet)
	ok(t, err)
	if palletRead.IsActive || palletRead.Name != "Edited pallet" {
		t.Fatal("seed overwrote existing default")
	}
}
