package master

import (
	"context"
	"strings"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
	repository "wms-api/repository/master"
)

func mapHandlingUnitType(v model.HandlingUnitType) dto.HandlingUnitTypeResponse {
	return dto.HandlingUnitTypeResponse{ID: v.ID, Code: v.Code, Name: v.Name, MaxWeight: v.MaxWeight, MaxVolume: v.MaxVolume, IsActive: v.IsActive}
}
func (s *CatalogService) CreateHandlingUnitType(ctx context.Context, r dto.CreateHandlingUnitTypeRequest) (dto.HandlingUnitTypeResponse, error) {
	code, name, err := catalogIdentity(r.Code, r.Name)
	if err != nil {
		return dto.HandlingUnitTypeResponse{}, err
	}
	if err = catalogDimensions(r.MaxWeight, r.MaxVolume); err != nil {
		return dto.HandlingUnitTypeResponse{}, err
	}
	v := model.HandlingUnitType{Code: code, Name: name, MaxWeight: r.MaxWeight, MaxVolume: r.MaxVolume, IsActive: true}
	err = s.repositories.HandlingUnitType.Create(ctx, &v)
	return mapHandlingUnitType(v), catalogError(err)
}
func (s *CatalogService) GetHandlingUnitType(ctx context.Context, id string) (dto.HandlingUnitTypeResponse, error) {
	if validateID(id) != nil {
		return dto.HandlingUnitTypeResponse{}, ErrInvalidInput
	}
	v, err := s.repositories.HandlingUnitType.Get(ctx, id)
	return mapHandlingUnitType(v), catalogError(err)
}
func (s *CatalogService) ListHandlingUnitType(ctx context.Context, f repository.CatalogFilter) (dto.PageResponse[dto.HandlingUnitTypeResponse], error) {
	if err := catalogPage(f); err != nil {
		return dto.PageResponse[dto.HandlingUnitTypeResponse]{}, err
	}
	rows, total, err := s.repositories.HandlingUnitType.List(ctx, repository.CatalogFilter{Page: f.Page, PageSize: f.PageSize, Search: f.Search, Active: f.Active})
	items := make([]dto.HandlingUnitTypeResponse, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapHandlingUnitType(v))
	}
	return pageResponse(items, f.Page, f.PageSize, total), catalogError(err)
}
func (s *CatalogService) UpdateHandlingUnitType(ctx context.Context, id string, r dto.UpdateHandlingUnitTypeRequest) (dto.HandlingUnitTypeResponse, error) {
	if validateID(id) != nil || strings.TrimSpace(r.Name) == "" || r.IsActive == nil {
		return dto.HandlingUnitTypeResponse{}, ErrInvalidInput
	}
	if err := catalogDimensions(r.MaxWeight, r.MaxVolume); err != nil {
		return dto.HandlingUnitTypeResponse{}, err
	}
	v, err := s.repositories.HandlingUnitType.Update(ctx, id, map[string]interface{}{"name": strings.TrimSpace(r.Name), "max_weight": r.MaxWeight, "max_volume": r.MaxVolume, "is_active": *r.IsActive}, "", nil)
	return mapHandlingUnitType(v), catalogError(err)
}
func (s *CatalogService) DeactivateHandlingUnitType(ctx context.Context, id string) (dto.HandlingUnitTypeResponse, error) {
	if validateID(id) != nil {
		return dto.HandlingUnitTypeResponse{}, ErrInvalidInput
	}
	v, err := s.repositories.HandlingUnitType.Update(ctx, id, map[string]interface{}{"is_active": false}, "", nil)
	return mapHandlingUnitType(v), catalogError(err)
}
