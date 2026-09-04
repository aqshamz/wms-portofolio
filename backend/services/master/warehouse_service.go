package master

import (
	"context"
	"errors"
	"strings"

	dto "wms-api/dto/master"
	model "wms-api/models/master"
	repository "wms-api/repository/master"

	"gorm.io/gorm"
)

type WarehouseService struct {
	warehouses    *repository.WarehouseRepository
	organizations *repository.OrganizationRepository
}

func NewWarehouseService(
	warehouses *repository.WarehouseRepository,
	organizations *repository.OrganizationRepository,
) *WarehouseService {
	return &WarehouseService{warehouses: warehouses, organizations: organizations}
}

func (s *WarehouseService) Create(
	ctx context.Context,
	request dto.CreateWarehouseRequest,
	actorID string,
) (dto.WarehouseResponse, error) {
	code, err := normalizeCode(request.Code)
	if err != nil || strings.TrimSpace(request.Name) == "" || validateID(request.OperatorID) != nil || validateTimezone(request.TimezoneName) != nil {
		return dto.WarehouseResponse{}, ErrInvalidInput
	}
	operator, err := s.organizations.FindByID(ctx, request.OperatorID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.WarehouseResponse{}, ErrInvalidInput
	}
	if err != nil {
		return dto.WarehouseResponse{}, err
	}
	if !operator.IsActive {
		return dto.WarehouseResponse{}, ErrInvalidInput
	}

	warehouse := model.Warehouse{
		OperatorID:   request.OperatorID,
		Code:         code,
		Name:         strings.TrimSpace(request.Name),
		TimezoneName: request.TimezoneName,
		AddressLine1: request.AddressLine1,
		AddressLine2: request.AddressLine2,
		City:         request.City,
		Province:     request.Province,
		PostalCode:   request.PostalCode,
		CountryCode:  normalizeCountry(request.CountryCode),
		IsActive:     true,
		CreatedBy:    &actorID,
		UpdatedBy:    &actorID,
	}
	if err := s.warehouses.Create(ctx, &warehouse); err != nil {
		if repository.IsUniqueViolation(err) {
			return dto.WarehouseResponse{}, ErrConflict
		}
		return dto.WarehouseResponse{}, err
	}
	return warehouseResponse(warehouse, operator.Code, operator.Name), nil
}

func (s *WarehouseService) Get(ctx context.Context, id string) (dto.WarehouseResponse, error) {
	if validateID(id) != nil {
		return dto.WarehouseResponse{}, ErrInvalidInput
	}
	warehouse, err := s.warehouses.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.WarehouseResponse{}, ErrNotFound
	}
	if err != nil {
		return dto.WarehouseResponse{}, err
	}
	return warehouseResponse(warehouse.Warehouse, warehouse.OperatorCode, warehouse.OperatorName), nil
}

func (s *WarehouseService) List(
	ctx context.Context,
	operatorID, search *string,
	active *bool,
	page, pageSize int,
) (dto.PageResponse[dto.WarehouseResponse], error) {
	if operatorID != nil && validateID(*operatorID) != nil {
		return dto.PageResponse[dto.WarehouseResponse]{}, ErrInvalidInput
	}
	if search != nil {
		value := strings.TrimSpace(*search)
		search = &value
	}
	rows, total, err := s.warehouses.List(
		ctx, operatorID, search, active, pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return dto.PageResponse[dto.WarehouseResponse]{}, err
	}
	items := make([]dto.WarehouseResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.WarehouseResponse{
			ID:           row.ID,
			OperatorID:   row.OperatorID,
			OperatorCode: row.OperatorCode,
			OperatorName: row.OperatorName,
			Code:         row.Code,
			Name:         row.Name,
			TimezoneName: row.TimezoneName,
			City:         row.City,
			CountryCode:  row.CountryCode,
			IsActive:     row.IsActive,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		})
	}
	return pageResponse(items, page, pageSize, total), nil
}

func (s *WarehouseService) Update(
	ctx context.Context,
	id string,
	request dto.UpdateWarehouseRequest,
	actorID string,
) (dto.WarehouseResponse, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return dto.WarehouseResponse{}, err
	}
	if strings.TrimSpace(request.Name) == "" || validateTimezone(request.TimezoneName) != nil {
		return dto.WarehouseResponse{}, ErrInvalidInput
	}
	_, err := s.warehouses.Update(ctx, id, request.ExpectedUpdatedAt, map[string]interface{}{
		"name":           strings.TrimSpace(request.Name),
		"timezone_name":  request.TimezoneName,
		"address_line_1": request.AddressLine1,
		"address_line_2": request.AddressLine2,
		"city":           request.City,
		"province":       request.Province,
		"postal_code":    request.PostalCode,
		"country_code":   normalizeCountry(request.CountryCode),
		"is_active":      request.IsActive,
		"updated_at":     gorm.Expr("clock_timestamp()"),
		"updated_by":     actorID,
	})
	if errors.Is(err, repository.ErrConcurrentUpdate) {
		return dto.WarehouseResponse{}, ErrConcurrentUpdate
	}
	if err != nil {
		return dto.WarehouseResponse{}, err
	}
	return s.Get(ctx, id)
}

func (s *WarehouseService) Deactivate(
	ctx context.Context,
	id string,
	request dto.DeactivateRequest,
	actorID string,
) (dto.WarehouseResponse, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return dto.WarehouseResponse{}, err
	}
	if _, err := s.warehouses.Deactivate(ctx, id, actorID, request.ExpectedUpdatedAt); err != nil {
		if errors.Is(err, repository.ErrConcurrentUpdate) {
			return dto.WarehouseResponse{}, ErrConcurrentUpdate
		}
		return dto.WarehouseResponse{}, err
	}
	return s.Get(ctx, id)
}

func warehouseResponse(value model.Warehouse, operatorCode, operatorName string) dto.WarehouseResponse {
	return dto.WarehouseResponse{
		ID:           value.ID,
		OperatorID:   value.OperatorID,
		OperatorCode: operatorCode,
		OperatorName: operatorName,
		Code:         value.Code,
		Name:         value.Name,
		TimezoneName: value.TimezoneName,
		AddressLine1: value.AddressLine1,
		AddressLine2: value.AddressLine2,
		City:         value.City,
		Province:     value.Province,
		PostalCode:   value.PostalCode,
		CountryCode:  value.CountryCode,
		IsActive:     value.IsActive,
		CreatedAt:    value.CreatedAt,
		UpdatedAt:    value.UpdatedAt,
	}
}
