package master

import (
	"context"
	"strings"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
	repository "wms-api/repository/master"
)

func (s *CatalogService) CreateItemCategory(ctx context.Context, request dto.CreateItemCategoryRequest) (response dto.ItemCategoryResponse, err error) {
	code, name, err := catalogIdentity(request.Code, request.Name)
	if err != nil {
		return response, err
	}
	err = s.repositories.Transaction(ctx, func(repos *repository.CatalogRepositories) error {
		local := NewCatalogService(repos)
		if err := local.requireOwner(ctx, request.OwnerID, true); err != nil {
			return err
		}
		if err := local.requireCategory(ctx, request.OwnerID, request.ParentCategoryID, ""); err != nil {
			return err
		}
		value := model.ItemCategory{OwnerID: request.OwnerID, ParentCategoryID: request.ParentCategoryID, Code: code, Name: name, IsActive: true}
		if err := repos.ItemCategory.Create(ctx, &value); err != nil {
			return err
		}
		response = mapItemCategory(value)
		return nil
	})
	return response, catalogError(err)
}
func (s *CatalogService) GetItemCategory(ctx context.Context, id string) (dto.ItemCategoryResponse, error) {
	if validateID(id) != nil {
		return dto.ItemCategoryResponse{}, ErrInvalidInput
	}
	value, err := s.repositories.ItemCategory.Get(ctx, id)
	return mapItemCategory(value), catalogError(err)
}
func (s *CatalogService) ListItemCategory(ctx context.Context, filter repository.CatalogFilter) (dto.PageResponse[dto.ItemCategoryResponse], error) {
	if err := catalogPage(filter); err != nil {
		return dto.PageResponse[dto.ItemCategoryResponse]{}, err
	}
	if filter.OwnerID != "" && validateID(filter.OwnerID) != nil {
		return dto.PageResponse[dto.ItemCategoryResponse]{}, ErrInvalidInput
	}
	rows, total, err := s.repositories.ItemCategory.List(ctx, repository.CatalogFilter{OwnerID: filter.OwnerID, Search: filter.Search, Active: filter.Active, Page: filter.Page, PageSize: filter.PageSize})
	if err != nil {
		return dto.PageResponse[dto.ItemCategoryResponse]{}, catalogError(err)
	}
	items := make([]dto.ItemCategoryResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapItemCategory(row))
	}
	return pageResponse(items, filter.Page, filter.PageSize, total), nil
}
func (s *CatalogService) UpdateItemCategory(ctx context.Context, id string, request dto.UpdateItemCategoryRequest) (response dto.ItemCategoryResponse, err error) {
	if validateID(id) != nil || strings.TrimSpace(request.Name) == "" || request.IsActive == nil {
		return response, ErrInvalidInput
	}
	err = s.repositories.Transaction(ctx, func(repos *repository.CatalogRepositories) error {
		current, err := repos.ItemCategory.Get(ctx, id)
		if err != nil {
			return err
		}
		local := NewCatalogService(repos)
		if err := local.requireOwner(ctx, current.OwnerID, true); err != nil {
			return err
		}
		if err := local.requireCategory(ctx, current.OwnerID, request.ParentCategoryID, id); err != nil {
			return err
		}
		value, err := repos.ItemCategory.Update(ctx, id, map[string]interface{}{"name": strings.TrimSpace(request.Name), "parent_category_id": request.ParentCategoryID, "is_active": *request.IsActive}, "", nil)
		response = mapItemCategory(value)
		return err
	})
	return response, catalogError(err)
}
func (s *CatalogService) DeactivateItemCategory(ctx context.Context, id string) (response dto.ItemCategoryResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}
	err = s.repositories.Transaction(ctx, func(repos *repository.CatalogRepositories) error {
		current, err := repos.ItemCategory.Get(ctx, id)
		if err != nil {
			return err
		}
		// Serialize with category re-parenting without requiring the owner to be active.
		if _, err := repos.Owner(ctx, current.OwnerID, true); err != nil {
			return err
		}
		value, err := repos.ItemCategory.Update(ctx, id, map[string]interface{}{"is_active": false}, "", nil)
		response = mapItemCategory(value)
		return err
	})
	return response, catalogError(err)
}
