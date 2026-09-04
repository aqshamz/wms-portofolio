package master

import (
	"context"
	"strings"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
	repository "wms-api/repository/master"
)

func (s *CatalogService) CreatePartnerType(ctx context.Context, request dto.CreatePartnerTypeRequest) (dto.PartnerTypeResponse, error) {
	code, name, err := catalogIdentity(request.Code, request.Name)
	if err != nil {
		return dto.PartnerTypeResponse{}, err
	}

	value := model.PartnerType{Code: code, Name: name, IsActive: true,
		Description: request.Description,
	}
	err = s.repositories.PartnerType.Create(ctx, &value)
	return mapPartnerType(value), catalogError(err)
}
func (s *CatalogService) GetPartnerType(ctx context.Context, id string) (dto.PartnerTypeResponse, error) {
	if validateID(id) != nil {
		return dto.PartnerTypeResponse{}, ErrInvalidInput
	}
	value, err := s.repositories.PartnerType.Get(ctx, id)
	return mapPartnerType(value), catalogError(err)
}
func (s *CatalogService) ListPartnerType(ctx context.Context, filter repository.CatalogFilter) (dto.PageResponse[dto.PartnerTypeResponse], error) {
	if err := catalogPage(filter); err != nil {
		return dto.PageResponse[dto.PartnerTypeResponse]{}, err
	}
	rows, total, err := s.repositories.PartnerType.List(ctx, repository.CatalogFilter{Page: filter.Page, PageSize: filter.PageSize, Search: filter.Search, Active: filter.Active})
	if err != nil {
		return dto.PageResponse[dto.PartnerTypeResponse]{}, catalogError(err)
	}
	items := make([]dto.PartnerTypeResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapPartnerType(row))
	}
	return pageResponse(items, filter.Page, filter.PageSize, total), nil
}
func (s *CatalogService) UpdatePartnerType(ctx context.Context, id string, request dto.UpdatePartnerTypeRequest) (dto.PartnerTypeResponse, error) {
	if validateID(id) != nil || strings.TrimSpace(request.Name) == "" || request.IsActive == nil {
		return dto.PartnerTypeResponse{}, ErrInvalidInput
	}

	value, err := s.repositories.PartnerType.Update(ctx, id, map[string]interface{}{
		"name":        strings.TrimSpace(request.Name),
		"description": request.Description,
		"is_active":   *request.IsActive,
	}, "", nil)
	return mapPartnerType(value), catalogError(err)
}
func (s *CatalogService) DeactivatePartnerType(ctx context.Context, id string) (dto.PartnerTypeResponse, error) {
	if validateID(id) != nil {
		return dto.PartnerTypeResponse{}, ErrInvalidInput
	}
	value, err := s.repositories.PartnerType.Update(ctx, id, map[string]interface{}{"is_active": false}, "", nil)
	return mapPartnerType(value), catalogError(err)
}

func (s *CatalogService) CreateUOM(ctx context.Context, request dto.CreateUOMRequest) (dto.UOMResponse, error) {
	code, name, err := catalogIdentity(request.Code, request.Name)
	if err != nil {
		return dto.UOMResponse{}, err
	}
	if request.DecimalScale < 0 || request.DecimalScale > 6 {
		return dto.UOMResponse{}, ErrInvalidInput
	}
	value := model.UOM{Code: code, Name: name, IsActive: true,
		DecimalScale: request.DecimalScale,
	}
	err = s.repositories.UOM.Create(ctx, &value)
	return mapUOM(value), catalogError(err)
}
func (s *CatalogService) GetUOM(ctx context.Context, id string) (dto.UOMResponse, error) {
	if validateID(id) != nil {
		return dto.UOMResponse{}, ErrInvalidInput
	}
	value, err := s.repositories.UOM.Get(ctx, id)
	return mapUOM(value), catalogError(err)
}
func (s *CatalogService) ListUOM(ctx context.Context, filter repository.CatalogFilter) (dto.PageResponse[dto.UOMResponse], error) {
	if err := catalogPage(filter); err != nil {
		return dto.PageResponse[dto.UOMResponse]{}, err
	}
	rows, total, err := s.repositories.UOM.List(ctx, repository.CatalogFilter{Page: filter.Page, PageSize: filter.PageSize, Search: filter.Search, Active: filter.Active})
	if err != nil {
		return dto.PageResponse[dto.UOMResponse]{}, catalogError(err)
	}
	items := make([]dto.UOMResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapUOM(row))
	}
	return pageResponse(items, filter.Page, filter.PageSize, total), nil
}
func (s *CatalogService) UpdateUOM(ctx context.Context, id string, request dto.UpdateUOMRequest) (dto.UOMResponse, error) {
	if validateID(id) != nil || strings.TrimSpace(request.Name) == "" || request.IsActive == nil {
		return dto.UOMResponse{}, ErrInvalidInput
	}
	if request.DecimalScale < 0 || request.DecimalScale > 6 {
		return dto.UOMResponse{}, ErrInvalidInput
	}
	value, err := s.repositories.UOM.Update(ctx, id, map[string]interface{}{
		"name":          strings.TrimSpace(request.Name),
		"decimal_scale": request.DecimalScale,
		"is_active":     *request.IsActive,
	}, "", nil)
	return mapUOM(value), catalogError(err)
}
func (s *CatalogService) DeactivateUOM(ctx context.Context, id string) (dto.UOMResponse, error) {
	if validateID(id) != nil {
		return dto.UOMResponse{}, ErrInvalidInput
	}
	value, err := s.repositories.UOM.Update(ctx, id, map[string]interface{}{"is_active": false}, "", nil)
	return mapUOM(value), catalogError(err)
}

func (s *CatalogService) CreateInventoryStatus(ctx context.Context, request dto.CreateInventoryStatusRequest) (dto.InventoryStatusResponse, error) {
	code, name, err := catalogIdentity(request.Code, request.Name)
	if err != nil {
		return dto.InventoryStatusResponse{}, err
	}

	value := model.InventoryStatus{Code: code, Name: name, IsActive: true,
		Description:   request.Description,
		IsAllocatable: request.IsAllocatable,
		IsPickable:    request.IsPickable,
	}
	err = s.repositories.InventoryStatus.Create(ctx, &value)
	return mapInventoryStatus(value), catalogError(err)
}
func (s *CatalogService) GetInventoryStatus(ctx context.Context, id string) (dto.InventoryStatusResponse, error) {
	if validateID(id) != nil {
		return dto.InventoryStatusResponse{}, ErrInvalidInput
	}
	value, err := s.repositories.InventoryStatus.Get(ctx, id)
	return mapInventoryStatus(value), catalogError(err)
}
func (s *CatalogService) ListInventoryStatus(ctx context.Context, filter repository.CatalogFilter) (dto.PageResponse[dto.InventoryStatusResponse], error) {
	if err := catalogPage(filter); err != nil {
		return dto.PageResponse[dto.InventoryStatusResponse]{}, err
	}
	rows, total, err := s.repositories.InventoryStatus.List(ctx, repository.CatalogFilter{Page: filter.Page, PageSize: filter.PageSize, Search: filter.Search, Active: filter.Active})
	if err != nil {
		return dto.PageResponse[dto.InventoryStatusResponse]{}, catalogError(err)
	}
	items := make([]dto.InventoryStatusResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapInventoryStatus(row))
	}
	return pageResponse(items, filter.Page, filter.PageSize, total), nil
}
func (s *CatalogService) UpdateInventoryStatus(ctx context.Context, id string, request dto.UpdateInventoryStatusRequest) (dto.InventoryStatusResponse, error) {
	if validateID(id) != nil || strings.TrimSpace(request.Name) == "" || request.IsActive == nil {
		return dto.InventoryStatusResponse{}, ErrInvalidInput
	}

	value, err := s.repositories.InventoryStatus.Update(ctx, id, map[string]interface{}{
		"name":           strings.TrimSpace(request.Name),
		"description":    request.Description,
		"is_allocatable": request.IsAllocatable,
		"is_pickable":    request.IsPickable,
		"is_active":      *request.IsActive,
	}, "", nil)
	return mapInventoryStatus(value), catalogError(err)
}
func (s *CatalogService) DeactivateInventoryStatus(ctx context.Context, id string) (dto.InventoryStatusResponse, error) {
	if validateID(id) != nil {
		return dto.InventoryStatusResponse{}, ErrInvalidInput
	}
	value, err := s.repositories.InventoryStatus.Update(ctx, id, map[string]interface{}{"is_active": false}, "", nil)
	return mapInventoryStatus(value), catalogError(err)
}

func (s *CatalogService) CreateQualityStatus(ctx context.Context, request dto.CreateQualityStatusRequest) (dto.QualityStatusResponse, error) {
	code, name, err := catalogIdentity(request.Code, request.Name)
	if err != nil {
		return dto.QualityStatusResponse{}, err
	}

	value := model.QualityStatus{Code: code, Name: name, IsActive: true,
		Description: request.Description,
	}
	err = s.repositories.QualityStatus.Create(ctx, &value)
	return mapQualityStatus(value), catalogError(err)
}
func (s *CatalogService) GetQualityStatus(ctx context.Context, id string) (dto.QualityStatusResponse, error) {
	if validateID(id) != nil {
		return dto.QualityStatusResponse{}, ErrInvalidInput
	}
	value, err := s.repositories.QualityStatus.Get(ctx, id)
	return mapQualityStatus(value), catalogError(err)
}
func (s *CatalogService) ListQualityStatus(ctx context.Context, filter repository.CatalogFilter) (dto.PageResponse[dto.QualityStatusResponse], error) {
	if err := catalogPage(filter); err != nil {
		return dto.PageResponse[dto.QualityStatusResponse]{}, err
	}
	rows, total, err := s.repositories.QualityStatus.List(ctx, repository.CatalogFilter{Page: filter.Page, PageSize: filter.PageSize, Search: filter.Search, Active: filter.Active})
	if err != nil {
		return dto.PageResponse[dto.QualityStatusResponse]{}, catalogError(err)
	}
	items := make([]dto.QualityStatusResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapQualityStatus(row))
	}
	return pageResponse(items, filter.Page, filter.PageSize, total), nil
}
func (s *CatalogService) UpdateQualityStatus(ctx context.Context, id string, request dto.UpdateQualityStatusRequest) (dto.QualityStatusResponse, error) {
	if validateID(id) != nil || strings.TrimSpace(request.Name) == "" || request.IsActive == nil {
		return dto.QualityStatusResponse{}, ErrInvalidInput
	}

	value, err := s.repositories.QualityStatus.Update(ctx, id, map[string]interface{}{
		"name":        strings.TrimSpace(request.Name),
		"description": request.Description,
		"is_active":   *request.IsActive,
	}, "", nil)
	return mapQualityStatus(value), catalogError(err)
}
func (s *CatalogService) DeactivateQualityStatus(ctx context.Context, id string) (dto.QualityStatusResponse, error) {
	if validateID(id) != nil {
		return dto.QualityStatusResponse{}, ErrInvalidInput
	}
	value, err := s.repositories.QualityStatus.Update(ctx, id, map[string]interface{}{"is_active": false}, "", nil)
	return mapQualityStatus(value), catalogError(err)
}

func (s *CatalogService) CreateInspectionResult(ctx context.Context, request dto.CreateInspectionResultRequest) (dto.InspectionResultResponse, error) {
	code, name, err := catalogIdentity(request.Code, request.Name)
	if err != nil {
		return dto.InspectionResultResponse{}, err
	}

	value := model.InspectionResult{Code: code, Name: name, IsActive: true,
		Description: request.Description,
		IsAccepted:  request.IsAccepted,
	}
	err = s.repositories.InspectionResult.Create(ctx, &value)
	return mapInspectionResult(value), catalogError(err)
}
func (s *CatalogService) GetInspectionResult(ctx context.Context, id string) (dto.InspectionResultResponse, error) {
	if validateID(id) != nil {
		return dto.InspectionResultResponse{}, ErrInvalidInput
	}
	value, err := s.repositories.InspectionResult.Get(ctx, id)
	return mapInspectionResult(value), catalogError(err)
}
func (s *CatalogService) ListInspectionResult(ctx context.Context, filter repository.CatalogFilter) (dto.PageResponse[dto.InspectionResultResponse], error) {
	if err := catalogPage(filter); err != nil {
		return dto.PageResponse[dto.InspectionResultResponse]{}, err
	}
	rows, total, err := s.repositories.InspectionResult.List(ctx, repository.CatalogFilter{Page: filter.Page, PageSize: filter.PageSize, Search: filter.Search, Active: filter.Active})
	if err != nil {
		return dto.PageResponse[dto.InspectionResultResponse]{}, catalogError(err)
	}
	items := make([]dto.InspectionResultResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapInspectionResult(row))
	}
	return pageResponse(items, filter.Page, filter.PageSize, total), nil
}
func (s *CatalogService) UpdateInspectionResult(ctx context.Context, id string, request dto.UpdateInspectionResultRequest) (dto.InspectionResultResponse, error) {
	if validateID(id) != nil || strings.TrimSpace(request.Name) == "" || request.IsActive == nil {
		return dto.InspectionResultResponse{}, ErrInvalidInput
	}

	value, err := s.repositories.InspectionResult.Update(ctx, id, map[string]interface{}{
		"name":        strings.TrimSpace(request.Name),
		"description": request.Description,
		"is_accepted": request.IsAccepted,
		"is_active":   *request.IsActive,
	}, "", nil)
	return mapInspectionResult(value), catalogError(err)
}
func (s *CatalogService) DeactivateInspectionResult(ctx context.Context, id string) (dto.InspectionResultResponse, error) {
	if validateID(id) != nil {
		return dto.InspectionResultResponse{}, ErrInvalidInput
	}
	value, err := s.repositories.InspectionResult.Update(ctx, id, map[string]interface{}{"is_active": false}, "", nil)
	return mapInspectionResult(value), catalogError(err)
}
