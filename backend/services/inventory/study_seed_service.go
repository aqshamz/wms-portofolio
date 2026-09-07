package inventory

import (
	"context"
	"errors"
	"fmt"

	dto "wms-api/dto/inventory"
	model "wms-api/models/master"
	repository "wms-api/repository/inventory"
	masterrepository "wms-api/repository/master"
)

var ErrStudyDataMissing = errors.New("study master data is missing or ambiguous")

func studyDataError(label string, err error) error {
	if err != nil {
		return fmt.Errorf("%w: %s: %v", ErrStudyDataMissing, label, err)
	}
	return fmt.Errorf("%w: %s", ErrStudyDataMissing, label)
}

func (s *Service) studyOrganization(ctx context.Context, code string) (model.Organization, error) {
	rows, _, err := s.repositories.Owner.List(ctx, &code, nil, 100, 0)
	if err != nil {
		return model.Organization{}, studyDataError("organization "+code, err)
	}
	for _, row := range rows {
		if row.Code == code {
			return s.repositories.Owner.FindByID(ctx, row.ID)
		}
	}
	return model.Organization{}, studyDataError("organization "+code, nil)
}

func (s *Service) studyWarehouse(ctx context.Context, operatorID, code string) (model.Warehouse, error) {
	rows, _, err := s.repositories.Warehouse.List(ctx, &operatorID, &code, nil, 100, 0)
	if err != nil {
		return model.Warehouse{}, studyDataError("warehouse "+code, err)
	}
	for _, row := range rows {
		if row.Code == code {
			result, findErr := s.repositories.Warehouse.FindByID(ctx, row.ID)
			return result.Warehouse, findErr
		}
	}
	return model.Warehouse{}, studyDataError("warehouse "+code, nil)
}

func (s *Service) studyLocation(ctx context.Context, warehouseID, code string) (model.WarehouseLocation, error) {
	rows, _, err := s.repositories.Location.List(ctx, warehouseID, nil, nil, &code, nil, 100, 0)
	if err != nil {
		return model.WarehouseLocation{}, studyDataError("location "+code, err)
	}
	for _, row := range rows {
		if row.Code == code {
			return row.WarehouseLocation, nil
		}
	}
	return model.WarehouseLocation{}, studyDataError("location "+code, nil)
}

func (s *Service) studyItem(ctx context.Context, ownerID, code string) (model.Item, error) {
	rows, _, err := s.repositories.Catalog.Item.List(ctx, masterrepository.CatalogFilter{OwnerID: ownerID, Search: code, Page: 1, PageSize: 100})
	if err != nil {
		return model.Item{}, studyDataError("item "+code, err)
	}
	for _, row := range rows {
		if row.Code == code {
			return row, nil
		}
	}
	return model.Item{}, studyDataError("item "+code, nil)
}

func (s *Service) studyInventoryStatus(ctx context.Context, code string) (model.InventoryStatus, error) {
	rows, _, err := s.repositories.Catalog.InventoryStatus.List(ctx, masterrepository.CatalogFilter{Search: code, Page: 1, PageSize: 100})
	if err != nil {
		return model.InventoryStatus{}, studyDataError("inventory status "+code, err)
	}
	for _, row := range rows {
		if row.Code == code {
			return row, nil
		}
	}
	return model.InventoryStatus{}, studyDataError("inventory status "+code, nil)
}

func (s *Service) studyQualityStatus(ctx context.Context, code string) (model.QualityStatus, error) {
	rows, _, err := s.repositories.Catalog.QualityStatus.List(ctx, masterrepository.CatalogFilter{Search: code, Page: 1, PageSize: 100})
	if err != nil {
		return model.QualityStatus{}, studyDataError("quality status "+code, err)
	}
	for _, row := range rows {
		if row.Code == code {
			return row, nil
		}
	}
	return model.QualityStatus{}, studyDataError("quality status "+code, nil)
}

func (s *Service) studyHandlingUnitType(ctx context.Context, code string) (model.HandlingUnitType, error) {
	rows, _, err := s.repositories.Catalog.HandlingUnitType.List(ctx, masterrepository.CatalogFilter{Search: code, Page: 1, PageSize: 100})
	if err != nil {
		return model.HandlingUnitType{}, studyDataError("handling-unit type "+code, err)
	}
	for _, row := range rows {
		if row.Code == code {
			return row, nil
		}
	}
	return model.HandlingUnitType{}, studyDataError("handling-unit type "+code, nil)
}

func (s *Service) ensureStudyLot(ctx context.Context, request dto.CreateLotRequest, actor string) (dto.LotResponse, bool, error) {
	rows, total, err := s.repositories.Lot.List(ctx, repository.Filter{OwnerID: request.OwnerID, ItemID: request.ItemID, Number: request.LotNumber, Page: 1, PageSize: 2})
	if err != nil {
		return dto.LotResponse{}, false, err
	}
	if total > 1 {
		return dto.LotResponse{}, false, studyDataError("lot "+request.LotNumber, nil)
	}
	if len(rows) == 1 {
		return mapLot(rows[0]), false, nil
	}
	created, err := s.CreateLot(ctx, request, actor)
	return created, true, err
}

func (s *Service) ensureStudySerial(ctx context.Context, request dto.CreateSerialRequest, actor string) (dto.SerialResponse, bool, error) {
	rows, total, err := s.repositories.Serial.List(ctx, repository.Filter{OwnerID: request.OwnerID, ItemID: request.ItemID, Number: request.SerialNo, Page: 1, PageSize: 2})
	if err != nil {
		return dto.SerialResponse{}, false, err
	}
	if total > 1 {
		return dto.SerialResponse{}, false, studyDataError("serial "+request.SerialNo, nil)
	}
	if len(rows) == 1 {
		return mapSerial(rows[0]), false, nil
	}
	created, err := s.CreateSerial(ctx, request, actor)
	return created, true, err
}

func (s *Service) ensureStudyHandlingUnit(ctx context.Context, request dto.CreateHandlingUnitRequest, actor string) (dto.HandlingUnitResponse, bool, error) {
	rows, total, err := s.repositories.HandlingUnit.List(ctx, repository.Filter{OwnerID: request.OwnerID, WarehouseID: request.WarehouseID, Number: request.Barcode, Page: 1, PageSize: 2})
	if err != nil {
		return dto.HandlingUnitResponse{}, false, err
	}
	if total > 1 {
		return dto.HandlingUnitResponse{}, false, studyDataError("handling unit "+request.Barcode, nil)
	}
	if len(rows) == 1 {
		return mapHandlingUnit(rows[0]), false, nil
	}
	created, err := s.CreateHandlingUnit(ctx, request, actor)
	return created, true, err
}

func studyTextPointer(value string) *string { return &value }

// SeedStudyInventory is an explicit development helper. It requires the
// additive STUDY_ master dataset and posts opening quantities through the same
// inventory core used by real workflows. Stable operation keys make reruns safe.
func (s *Service) SeedStudyInventory(ctx context.Context, actor string) (dto.StudySeedResponse, error) {
	if !uuid(actor) {
		return dto.StudySeedResponse{}, invalid("study seed actor must be a UUID")
	}
	operator, err := s.studyOrganization(ctx, "STUDY_OP")
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	owner, err := s.studyOrganization(ctx, "STUDY_OWNER")
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	warehouse, err := s.studyWarehouse(ctx, operator.ID, "STUDY_WH")
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	transferWarehouse, err := s.studyWarehouse(ctx, operator.ID, "STUDY_WH_2")
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	bulk, err := s.studyLocation(ctx, warehouse.ID, "STUDY_BULK_01")
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	pick, err := s.studyLocation(ctx, warehouse.ID, "STUDY_PICK_01")
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	qc, err := s.studyLocation(ctx, warehouse.ID, "STUDY_QC_01")
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	transferLocation, err := s.studyLocation(ctx, transferWarehouse.ID, "STUDY_BULK_01")
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	coffee250, err := s.studyItem(ctx, owner.ID, "STUDY_COFFEE_250G")
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	coffee1KG, err := s.studyItem(ctx, owner.ID, "STUDY_COFFEE_1KG")
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	scanner, err := s.studyItem(ctx, owner.ID, "STUDY_SCANNER")
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	available, err := s.studyInventoryStatus(ctx, "AVAILABLE")
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	qcPending, err := s.studyInventoryStatus(ctx, "QC_PENDING")
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	passed, err := s.studyQualityStatus(ctx, "PASSED")
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	pending, err := s.studyQualityStatus(ctx, "PENDING")
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	palletType, err := s.studyHandlingUnitType(ctx, "PALLET")
	if err != nil {
		return dto.StudySeedResponse{}, err
	}

	result := dto.StudySeedResponse{
		OwnerID: owner.ID, PrimaryWarehouseID: warehouse.ID, TransferWarehouseID: transferWarehouse.ID,
		Locations: map[string]string{"bulk": bulk.ID, "pick": pick.ID, "qc": qc.ID, "transfer_bulk": transferLocation.ID},
		Items:     map[string]string{"coffee_250g": coffee250.ID, "coffee_1kg": coffee1KG.ID, "scanner": scanner.ID},
		Lots:      map[string]string{}, HandlingUnits: map[string]string{}, Serials: map[string]string{},
		Balances: make([]dto.BalanceResponse, 0), Movements: make([]dto.MovementResponse, 0, 6),
	}
	countIdentity := func(created bool) {
		if created {
			result.CreatedIdentities++
		} else {
			result.ReusedIdentities++
		}
	}

	lot250A, created, err := s.ensureStudyLot(ctx, dto.CreateLotRequest{OwnerID: owner.ID, ItemID: coffee250.ID, LotNumber: "STUDY-250G-2026-08-A", ManufactureDate: studyTextPointer("2026-08-01"), ExpiryDate: studyTextPointer("2027-08-01"), QualityStatusID: &passed.ID}, actor)
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	countIdentity(created)
	result.Lots["coffee_250g_lot_a"] = lot250A.ID
	lot250B, created, err := s.ensureStudyLot(ctx, dto.CreateLotRequest{OwnerID: owner.ID, ItemID: coffee250.ID, LotNumber: "STUDY-250G-2026-08-B", ManufactureDate: studyTextPointer("2026-08-15"), ExpiryDate: studyTextPointer("2027-08-15"), QualityStatusID: &passed.ID}, actor)
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	countIdentity(created)
	result.Lots["coffee_250g_lot_b"] = lot250B.ID
	lot1KG, created, err := s.ensureStudyLot(ctx, dto.CreateLotRequest{OwnerID: owner.ID, ItemID: coffee1KG.ID, LotNumber: "STUDY-1KG-2026-09-A", ManufactureDate: studyTextPointer("2026-09-01"), ExpiryDate: studyTextPointer("2027-09-01"), QualityStatusID: &pending.ID}, actor)
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	countIdentity(created)
	result.Lots["coffee_1kg_lot_a"] = lot1KG.ID

	pallet, created, err := s.ensureStudyHandlingUnit(ctx, dto.CreateHandlingUnitRequest{WarehouseID: warehouse.ID, OwnerID: owner.ID, HandlingUnitTypeID: palletType.ID, CurrentLocationID: &bulk.ID, Barcode: "STUDY-INV-PLT-0001"}, actor)
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	countIdentity(created)
	result.HandlingUnits["pallet_0001"] = pallet.ID

	serialIDs := make([]string, 0, 3)
	for i := 1; i <= 3; i++ {
		serialNo := fmt.Sprintf("STUDY-SCN-%04d", i)
		serial, wasCreated, serialErr := s.ensureStudySerial(ctx, dto.CreateSerialRequest{OwnerID: owner.ID, ItemID: scanner.ID, SerialNo: serialNo}, actor)
		if serialErr != nil {
			return dto.StudySeedResponse{}, serialErr
		}
		countIdentity(wasCreated)
		result.Serials[serialNo] = serial.ID
		serialIDs = append(serialIDs, serial.ID)
	}

	post := func(request dto.PostingRequest) error {
		posted, postErr := s.PostMovement(ctx, request, actor)
		if postErr != nil {
			return postErr
		}
		result.Movements = append(result.Movements, posted.Movement)
		if posted.IdempotentReplay {
			result.ReplayedMovements++
		} else {
			result.PostedMovements++
		}
		return nil
	}
	opening := []dto.PostingRequest{
		{OperationKey: "study.inventory.opening.coffee250.a", MovementTypeCode: "RECEIVE", OwnerID: owner.ID, WarehouseID: warehouse.ID, BusinessDate: "2026-09-07", ItemID: coffee250.ID, LotID: &lot250A.ID, HandlingUnitID: &pallet.ID, To: &dto.BalanceDimension{LocationID: bulk.ID, InventoryStatusID: available.ID}, Quantity: "240", SourceDocumentID: "STUDY-OPENING", SourceLineID: studyTextPointer("COFFEE-250-A")},
		{OperationKey: "study.inventory.opening.coffee250.b", MovementTypeCode: "RECEIVE", OwnerID: owner.ID, WarehouseID: warehouse.ID, BusinessDate: "2026-09-07", ItemID: coffee250.ID, LotID: &lot250B.ID, To: &dto.BalanceDimension{LocationID: pick.ID, InventoryStatusID: available.ID}, Quantity: "96", SourceDocumentID: "STUDY-OPENING", SourceLineID: studyTextPointer("COFFEE-250-B")},
		{OperationKey: "study.inventory.opening.coffee1kg.a", MovementTypeCode: "RECEIVE", OwnerID: owner.ID, WarehouseID: warehouse.ID, BusinessDate: "2026-09-07", ItemID: coffee1KG.ID, LotID: &lot1KG.ID, To: &dto.BalanceDimension{LocationID: qc.ID, InventoryStatusID: qcPending.ID}, Quantity: "60", SourceDocumentID: "STUDY-OPENING", SourceLineID: studyTextPointer("COFFEE-1KG-A")},
	}
	for _, request := range opening {
		if err := post(request); err != nil {
			return dto.StudySeedResponse{}, err
		}
	}
	for i, serialID := range serialIDs {
		line := fmt.Sprintf("SCANNER-%04d", i+1)
		if err := post(dto.PostingRequest{OperationKey: fmt.Sprintf("study.inventory.opening.scanner.%04d", i+1), MovementTypeCode: "RECEIVE", OwnerID: owner.ID, WarehouseID: warehouse.ID, BusinessDate: "2026-09-07", ItemID: scanner.ID, To: &dto.BalanceDimension{LocationID: pick.ID, InventoryStatusID: available.ID}, Quantity: "1", SerialIDs: []string{serialID}, SourceDocumentID: "STUDY-OPENING", SourceLineID: &line}); err != nil {
			return dto.StudySeedResponse{}, err
		}
	}

	balances, _, err := s.repositories.Balance.List(ctx, repository.BalanceFilter{OwnerID: owner.ID, IncludeZero: true, Page: 1, PageSize: 100})
	if err != nil {
		return dto.StudySeedResponse{}, err
	}
	for _, balance := range balances {
		result.Balances = append(result.Balances, mapBalance(balance))
	}
	return result, nil
}
