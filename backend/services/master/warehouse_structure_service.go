package master

import (
	"context"
	"errors"
	"regexp"
	"strings"

	dto "wms-api/dto/master"
	model "wms-api/models/master"
	repository "wms-api/repository/master"

	"gorm.io/gorm"
)

var decimalPattern = regexp.MustCompile(`^(0|[1-9][0-9]{0,13})(\.[0-9]{1,6})?$`)

type WarehouseStructureService struct {
	warehouses    *repository.WarehouseRepository
	organizations *repository.OrganizationRepository
	owners        *repository.WarehouseOwnerRepository
	locationTypes *repository.LocationTypeRepository
	zones         *repository.WarehouseZoneRepository
	locations     *repository.WarehouseLocationRepository
}

func NewWarehouseStructureService(
	warehouses *repository.WarehouseRepository,
	organizations *repository.OrganizationRepository,
	owners *repository.WarehouseOwnerRepository,
	locationTypes *repository.LocationTypeRepository,
	zones *repository.WarehouseZoneRepository,
	locations *repository.WarehouseLocationRepository,
) *WarehouseStructureService {
	return &WarehouseStructureService{
		warehouses: warehouses, organizations: organizations, owners: owners,
		locationTypes: locationTypes, zones: zones, locations: locations,
	}
}

func (s *WarehouseStructureService) SeedLocationTypes(ctx context.Context) error {
	return s.locationTypes.Seed(ctx, []model.LocationType{
		{Code: "RECEIVING", Name: "Receiving", AllowsReceiving: true, IsActive: true},
		{Code: "STORAGE", Name: "Storage", AllowsStorage: true, IsActive: true},
		{Code: "PICK_FACE", Name: "Pick face", AllowsStorage: true, AllowsPicking: true, IsActive: true},
		{Code: "STAGING", Name: "Staging", AllowsStorage: true, IsActive: true},
		{Code: "CHECKING", Name: "Checking", AllowsStorage: true, IsActive: true},
		{Code: "PACKING", Name: "Packing", IsActive: true},
		{Code: "SHIPPING", Name: "Shipping", AllowsShipping: true, IsActive: true},
		{Code: "QUARANTINE", Name: "Quarantine", AllowsStorage: true, IsActive: true},
	})
}

func (s *WarehouseStructureService) AssignOwner(
	ctx context.Context, warehouseID string, request dto.AssignWarehouseOwnerRequest, actorID string,
) error {
	if validateID(warehouseID) != nil || validateID(request.OwnerID) != nil {
		return ErrInvalidInput
	}
	warehouse, err := s.warehouses.FindByID(ctx, warehouseID)
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && !warehouse.IsActive) {
		return ErrInvalidInput
	}
	if err != nil {
		return err
	}
	owner, err := s.organizations.FindByID(ctx, request.OwnerID)
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && !owner.IsActive) {
		return ErrInvalidInput
	}
	if err != nil {
		return err
	}
	return s.owners.Assign(ctx, &model.WarehouseOwner{
		WarehouseID: warehouseID, OwnerID: request.OwnerID, IsActive: true, CreatedBy: &actorID,
	})
}

func (s *WarehouseStructureService) ListOwners(
	ctx context.Context, warehouseID string,
) ([]dto.WarehouseOwnerResponse, error) {
	if validateID(warehouseID) != nil {
		return nil, ErrInvalidInput
	}
	rows, err := s.owners.List(ctx, warehouseID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.WarehouseOwnerResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, dto.WarehouseOwnerResponse{
			WarehouseID: row.WarehouseID, OwnerID: row.OwnerID, OwnerCode: row.OwnerCode,
			OwnerName: row.OwnerName, IsActive: row.IsActive, CreatedAt: row.CreatedAt,
		})
	}
	return result, nil
}

func (s *WarehouseStructureService) DeactivateOwner(ctx context.Context, warehouseID, ownerID string) error {
	if validateID(warehouseID) != nil || validateID(ownerID) != nil {
		return ErrInvalidInput
	}
	err := s.owners.Deactivate(ctx, warehouseID, ownerID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *WarehouseStructureService) CreateLocationType(
	ctx context.Context, request dto.CreateLocationTypeRequest,
) (model.LocationType, error) {
	code, err := normalizeCode(request.Code)
	if err != nil || strings.TrimSpace(request.Name) == "" {
		return model.LocationType{}, ErrInvalidInput
	}
	value := model.LocationType{
		Code: code, Name: strings.TrimSpace(request.Name), Description: request.Description,
		AllowsReceiving: request.AllowsReceiving, AllowsStorage: request.AllowsStorage,
		AllowsPicking: request.AllowsPicking, AllowsShipping: request.AllowsShipping, IsActive: true,
	}
	if err := s.locationTypes.Create(ctx, &value); err != nil {
		if repository.IsUniqueViolation(err) {
			return model.LocationType{}, ErrConflict
		}
		return model.LocationType{}, err
	}
	return value, nil
}

func (s *WarehouseStructureService) ListLocationTypes(
	ctx context.Context, active *bool,
) ([]model.LocationType, error) {
	return s.locationTypes.List(ctx, active)
}

func (s *WarehouseStructureService) UpdateLocationType(
	ctx context.Context, id string, request dto.UpdateLocationTypeRequest,
) (model.LocationType, error) {
	if validateID(id) != nil || strings.TrimSpace(request.Name) == "" {
		return model.LocationType{}, ErrInvalidInput
	}
	value, err := s.locationTypes.Update(ctx, id, map[string]interface{}{
		"name": strings.TrimSpace(request.Name), "description": request.Description,
		"allows_receiving": request.AllowsReceiving, "allows_storage": request.AllowsStorage,
		"allows_picking": request.AllowsPicking, "allows_shipping": request.AllowsShipping,
		"is_active": request.IsActive,
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.LocationType{}, ErrNotFound
	}
	return value, err
}

func (s *WarehouseStructureService) CreateZone(
	ctx context.Context, warehouseID string, request dto.CreateWarehouseZoneRequest, actorID string,
) (dto.WarehouseZoneResponse, error) {
	code, err := normalizeCode(request.Code)
	if err != nil || validateID(warehouseID) != nil || strings.TrimSpace(request.Name) == "" {
		return dto.WarehouseZoneResponse{}, ErrInvalidInput
	}
	warehouse, err := s.warehouses.FindByID(ctx, warehouseID)
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && !warehouse.IsActive) {
		return dto.WarehouseZoneResponse{}, ErrInvalidInput
	}
	if err != nil {
		return dto.WarehouseZoneResponse{}, err
	}
	value := model.WarehouseZone{
		WarehouseID: warehouseID, Code: code, Name: strings.TrimSpace(request.Name),
		Description: request.Description, IsActive: true, CreatedBy: &actorID,
	}
	if err := s.zones.Create(ctx, &value); err != nil {
		if repository.IsUniqueViolation(err) {
			return dto.WarehouseZoneResponse{}, ErrConflict
		}
		return dto.WarehouseZoneResponse{}, err
	}
	return zoneResponse(value, 0), nil
}

func (s *WarehouseStructureService) ListZones(
	ctx context.Context, warehouseID string, search *string, active *bool,
) ([]dto.WarehouseZoneResponse, error) {
	if validateID(warehouseID) != nil {
		return nil, ErrInvalidInput
	}
	rows, err := s.zones.List(ctx, warehouseID, search, active)
	if err != nil {
		return nil, err
	}
	result := make([]dto.WarehouseZoneResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, zoneResponse(row.WarehouseZone, row.LocationCount))
	}
	return result, nil
}

func (s *WarehouseStructureService) UpdateZone(
	ctx context.Context, id string, request dto.UpdateWarehouseZoneRequest,
) (dto.WarehouseZoneResponse, error) {
	if validateID(id) != nil || strings.TrimSpace(request.Name) == "" {
		return dto.WarehouseZoneResponse{}, ErrInvalidInput
	}
	value, err := s.zones.Update(ctx, id, map[string]interface{}{
		"name": strings.TrimSpace(request.Name), "description": request.Description,
		"is_active": request.IsActive,
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.WarehouseZoneResponse{}, ErrNotFound
	}
	return zoneResponse(value, 0), err
}

func (s *WarehouseStructureService) DeactivateZone(ctx context.Context, id string) error {
	if validateID(id) != nil {
		return ErrInvalidInput
	}
	_, err := s.zones.Deactivate(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *WarehouseStructureService) CreateLocation(
	ctx context.Context, warehouseID string, request dto.CreateWarehouseLocationRequest, actorID string,
) (dto.WarehouseLocationResponse, error) {
	code, err := normalizeCode(request.Code)
	if err != nil || validateID(warehouseID) != nil || validateID(request.ZoneID) != nil ||
		validateID(request.LocationTypeID) != nil || !validCapacity(request.MaxWeight, request.MaxVolume) {
		return dto.WarehouseLocationResponse{}, ErrInvalidInput
	}
	warehouse, err := s.warehouses.FindByID(ctx, warehouseID)
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && !warehouse.IsActive) {
		return dto.WarehouseLocationResponse{}, ErrInvalidInput
	}
	if err != nil {
		return dto.WarehouseLocationResponse{}, err
	}
	zone, err := s.zones.FindByID(ctx, request.ZoneID)
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && (!zone.IsActive || zone.WarehouseID != warehouseID)) {
		return dto.WarehouseLocationResponse{}, ErrInvalidInput
	}
	if err != nil {
		return dto.WarehouseLocationResponse{}, err
	}
	locationType, err := s.locationTypes.FindByID(ctx, request.LocationTypeID)
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && !locationType.IsActive) {
		return dto.WarehouseLocationResponse{}, ErrInvalidInput
	}
	if err != nil {
		return dto.WarehouseLocationResponse{}, err
	}
	value := model.WarehouseLocation{
		WarehouseID: warehouseID, ZoneID: request.ZoneID, LocationTypeID: request.LocationTypeID,
		Code: code, Barcode: request.Barcode, Aisle: request.Aisle, Bay: request.Bay,
		LevelNo: request.LevelNo, PositionNo: request.PositionNo, PickSequence: request.PickSequence,
		MaxWeight: normalizeDecimal(request.MaxWeight), MaxVolume: normalizeDecimal(request.MaxVolume),
		IsPickFace: request.IsPickFace,
		IsActive:   true, CreatedBy: &actorID,
	}
	if err := s.locations.Create(ctx, &value); err != nil {
		if repository.IsUniqueViolation(err) {
			return dto.WarehouseLocationResponse{}, ErrConflict
		}
		return dto.WarehouseLocationResponse{}, err
	}
	return s.GetLocation(ctx, value.ID)
}

func (s *WarehouseStructureService) GetLocation(
	ctx context.Context, id string,
) (dto.WarehouseLocationResponse, error) {
	if validateID(id) != nil {
		return dto.WarehouseLocationResponse{}, ErrInvalidInput
	}
	value, err := s.locations.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.WarehouseLocationResponse{}, ErrNotFound
	}
	if err != nil {
		return dto.WarehouseLocationResponse{}, err
	}
	return locationResponse(value.WarehouseLocation, value.WarehouseCode, value.WarehouseName,
		value.ZoneCode, value.ZoneName, value.LocationTypeCode, value.LocationTypeName), nil
}

func (s *WarehouseStructureService) ListLocations(
	ctx context.Context, warehouseID string, zoneID, locationTypeID, search *string,
	active *bool, page, pageSize int,
) (dto.PageResponse[dto.WarehouseLocationResponse], error) {
	if validateID(warehouseID) != nil || (zoneID != nil && validateID(*zoneID) != nil) ||
		(locationTypeID != nil && validateID(*locationTypeID) != nil) {
		return dto.PageResponse[dto.WarehouseLocationResponse]{}, ErrInvalidInput
	}
	rows, total, err := s.locations.List(ctx, warehouseID, zoneID, locationTypeID, search,
		active, pageSize, (page-1)*pageSize)
	if err != nil {
		return dto.PageResponse[dto.WarehouseLocationResponse]{}, err
	}
	items := make([]dto.WarehouseLocationResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, locationResponse(row.WarehouseLocation, "", "",
			row.ZoneCode, "", row.LocationTypeCode, ""))
	}
	return pageResponse(items, page, pageSize, total), nil
}

func (s *WarehouseStructureService) UpdateLocation(
	ctx context.Context, id string, request dto.UpdateWarehouseLocationRequest,
) (dto.WarehouseLocationResponse, error) {
	if validateID(id) != nil {
		return dto.WarehouseLocationResponse{}, ErrInvalidInput
	}
	current, err := s.locations.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.WarehouseLocationResponse{}, ErrNotFound
	}
	if err != nil || validateID(request.ZoneID) != nil || validateID(request.LocationTypeID) != nil ||
		!validCapacity(request.MaxWeight, request.MaxVolume) {
		if err != nil {
			return dto.WarehouseLocationResponse{}, err
		}
		return dto.WarehouseLocationResponse{}, ErrInvalidInput
	}
	zone, err := s.zones.FindByID(ctx, request.ZoneID)
	if err != nil || !zone.IsActive || zone.WarehouseID != current.WarehouseID {
		return dto.WarehouseLocationResponse{}, ErrInvalidInput
	}
	locationType, err := s.locationTypes.FindByID(ctx, request.LocationTypeID)
	if err != nil || !locationType.IsActive {
		return dto.WarehouseLocationResponse{}, ErrInvalidInput
	}
	_, err = s.locations.Update(ctx, id, map[string]interface{}{
		"zone_id": request.ZoneID, "location_type_id": request.LocationTypeID,
		"barcode": request.Barcode, "aisle": request.Aisle, "bay": request.Bay,
		"level_no": request.LevelNo, "position_no": request.PositionNo,
		"pick_sequence": request.PickSequence, "max_weight": normalizeDecimal(request.MaxWeight),
		"max_volume": normalizeDecimal(request.MaxVolume), "is_pick_face": request.IsPickFace,
		"is_locked": request.IsLocked, "is_active": request.IsActive,
	})
	if err != nil {
		if repository.IsUniqueViolation(err) {
			return dto.WarehouseLocationResponse{}, ErrConflict
		}
		return dto.WarehouseLocationResponse{}, err
	}
	return s.GetLocation(ctx, id)
}

func (s *WarehouseStructureService) DeactivateLocation(ctx context.Context, id string) error {
	if validateID(id) != nil {
		return ErrInvalidInput
	}
	_, err := s.locations.Deactivate(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func validCapacity(values ...*string) bool {
	for _, value := range values {
		if value != nil && !decimalPattern.MatchString(strings.TrimSpace(*value)) {
			return false
		}
	}
	return true
}

func normalizeDecimal(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	return &normalized
}

func zoneResponse(value model.WarehouseZone, count int64) dto.WarehouseZoneResponse {
	return dto.WarehouseZoneResponse{ID: value.ID, WarehouseID: value.WarehouseID, Code: value.Code,
		Name: value.Name, Description: value.Description, IsActive: value.IsActive,
		LocationCount: count, CreatedAt: value.CreatedAt}
}

func locationResponse(value model.WarehouseLocation, warehouseCode, warehouseName,
	zoneCode, zoneName, typeCode, typeName string) dto.WarehouseLocationResponse {
	return dto.WarehouseLocationResponse{
		ID: value.ID, WarehouseID: value.WarehouseID, WarehouseCode: warehouseCode,
		WarehouseName: warehouseName, ZoneID: value.ZoneID, ZoneCode: zoneCode, ZoneName: zoneName,
		LocationTypeID: value.LocationTypeID, LocationTypeCode: typeCode, LocationTypeName: typeName,
		Code: value.Code, Barcode: value.Barcode, Aisle: value.Aisle, Bay: value.Bay,
		LevelNo: value.LevelNo, PositionNo: value.PositionNo, PickSequence: value.PickSequence,
		MaxWeight: value.MaxWeight, MaxVolume: value.MaxVolume, IsPickFace: value.IsPickFace,
		IsLocked: value.IsLocked, IsActive: value.IsActive, CreatedAt: value.CreatedAt,
	}
}
