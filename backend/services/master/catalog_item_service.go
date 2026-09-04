package master

import (
	"context"
	"strings"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
	repository "wms-api/repository/master"
)

func (s *CatalogService) CreateItem(ctx context.Context, request dto.CreateItemRequest, actor string) (response dto.ItemResponse, err error) {
	code, name, err := catalogIdentity(request.Code, request.Name)
	if err != nil || validateID(actor) != nil {
		return response, ErrInvalidInput
	}

	if err := catalogDimensions(request.Weight, request.Volume); err != nil {
		return response, err
	}
	if err := catalogShelfLife(request.ShelfLifeDays, request.MinimumReceiveDays); err != nil {
		return response, err
	}
	err = s.repositories.Transaction(ctx, func(repos *repository.CatalogRepositories) error {
		local := NewCatalogService(repos)
		if err := local.requireOwner(ctx, request.OwnerID, true); err != nil {
			return err
		}

		if err := local.requireCategory(ctx, request.OwnerID, request.CategoryID, ""); err != nil {
			return err
		}
		if err := local.requireUOM(ctx, request.BaseUOMID); err != nil {
			return err
		}
		value := model.Item{OwnerID: request.OwnerID,
			CategoryID:         request.CategoryID,
			Code:               code,
			Name:               name,
			Description:        request.Description,
			BaseUOMID:          request.BaseUOMID,
			Weight:             normalizeDecimal(request.Weight),
			Volume:             normalizeDecimal(request.Volume),
			LotControlled:      request.LotControlled,
			SerialControlled:   request.SerialControlled,
			ShelfLifeDays:      request.ShelfLifeDays,
			MinimumReceiveDays: request.MinimumReceiveDays,
			IsActive:           true, CreatedBy: &actor, UpdatedBy: &actor,
		}
		if err := repos.Item.Create(ctx, &value); err != nil {
			return err
		}
		base := model.ItemUOM{ItemID: value.ID, UOMID: value.BaseUOMID, ConversionToBase: "1", IsReceivingUOM: true, IsPickingUOM: true, IsActive: true}
		if err := repos.ItemUOM.Create(ctx, &base); err != nil {
			return err
		}
		response = mapItem(value)
		return nil
	})
	return response, catalogError(err)
}
func (s *CatalogService) GetItem(ctx context.Context, id string) (response dto.ItemDetailResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}
	value, err := s.repositories.Item.Get(ctx, id)
	if err != nil {
		return response, catalogError(err)
	}
	response.ItemResponse = mapItem(value)
	response.UOMs, err = s.ListItemUOM(ctx, id)
	if err != nil {
		return response, err
	}
	response.Barcodes, err = s.ListItemBarcode(ctx, id)
	return response, err
}
func (s *CatalogService) ListItem(ctx context.Context, filter repository.CatalogFilter) (dto.PageResponse[dto.ItemResponse], error) {
	if err := catalogPage(filter); err != nil {
		return dto.PageResponse[dto.ItemResponse]{}, err
	}
	if filter.OwnerID != "" && validateID(filter.OwnerID) != nil {
		return dto.PageResponse[dto.ItemResponse]{}, ErrInvalidInput
	}
	if filter.CategoryID != "" && validateID(filter.CategoryID) != nil {
		return dto.PageResponse[dto.ItemResponse]{}, ErrInvalidInput
	}
	filter.PartnerTypeCode = ""
	filter.ItemID = ""
	rows, total, err := s.repositories.Item.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.ItemResponse]{}, catalogError(err)
	}
	items := make([]dto.ItemResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapItem(row))
	}
	return pageResponse(items, filter.Page, filter.PageSize, total), nil
}
func (s *CatalogService) UpdateItem(ctx context.Context, id string, request dto.UpdateItemRequest, actor string) (response dto.ItemResponse, err error) {
	name := strings.TrimSpace(request.Name)
	if validateID(id) != nil || validateID(actor) != nil || name == "" || request.IsActive == nil || request.ExpectedUpdatedAt.IsZero() {
		return response, ErrInvalidInput
	}

	if err := catalogDimensions(request.Weight, request.Volume); err != nil {
		return response, err
	}
	if err := catalogShelfLife(request.ShelfLifeDays, request.MinimumReceiveDays); err != nil {
		return response, err
	}
	err = s.repositories.Transaction(ctx, func(repos *repository.CatalogRepositories) error {
		current, err := repos.Item.Get(ctx, id)
		if err != nil {
			return err
		}
		local := NewCatalogService(repos)
		if err := local.requireOwner(ctx, current.OwnerID, true); err != nil {
			return err
		}
		if err := local.requireCategory(ctx, current.OwnerID, request.CategoryID, ""); err != nil {
			return err
		}
		if *request.IsActive {
			if err := local.requireUOM(ctx, current.BaseUOMID); err != nil {
				return err
			}
		}
		value, err := repos.Item.Update(ctx, id, map[string]interface{}{
			"category_id":          request.CategoryID,
			"name":                 name,
			"description":          request.Description,
			"weight":               normalizeDecimal(request.Weight),
			"volume":               normalizeDecimal(request.Volume),
			"lot_controlled":       request.LotControlled,
			"serial_controlled":    request.SerialControlled,
			"shelf_life_days":      request.ShelfLifeDays,
			"minimum_receive_days": request.MinimumReceiveDays,
			"is_active":            *request.IsActive,
		}, actor, &request.ExpectedUpdatedAt)
		response = mapItem(value)
		return err
	})
	return response, catalogError(err)
}
func (s *CatalogService) DeactivateItem(ctx context.Context, id string, request dto.DeactivateRequest, actor string) (dto.ItemResponse, error) {
	if validateID(id) != nil || validateID(actor) != nil || request.ExpectedUpdatedAt.IsZero() {
		return dto.ItemResponse{}, ErrInvalidInput
	}
	if _, err := s.repositories.Item.Get(ctx, id); err != nil {
		return dto.ItemResponse{}, catalogError(err)
	}
	value, err := s.repositories.Item.Update(ctx, id, map[string]interface{}{"is_active": false}, actor, &request.ExpectedUpdatedAt)
	return mapItem(value), catalogError(err)
}

func catalogShelfLife(shelf, minimum *int) error {
	if shelf != nil && *shelf < 0 || minimum != nil && *minimum < 0 {
		return invalidCatalog("shelf-life days must be non-negative")
	}
	if shelf != nil && minimum != nil && *minimum > *shelf {
		return invalidCatalog("minimum_receive_days exceeds shelf_life_days")
	}
	return nil
}
