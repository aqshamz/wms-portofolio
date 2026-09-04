package master

import (
	"context"
	"errors"
	"strings"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
	repository "wms-api/repository/master"
)

func (s *CatalogService) CreateItemUOM(ctx context.Context, itemID string, request dto.CreateItemUOMRequest) (response dto.ItemUOMResponse, err error) {
	conversion, err := catalogDecimal(request.ConversionToBase, true)
	if err != nil {
		return response, err
	}
	if err := catalogDimensions(request.Length, request.Width, request.Height, request.Weight); err != nil {
		return response, err
	}
	err = s.repositories.Transaction(ctx, func(repos *repository.CatalogRepositories) error {
		local := NewCatalogService(repos)
		item, err := local.itemForWrite(ctx, itemID)
		if err != nil {
			return err
		}
		if !item.IsActive {
			return invalidCatalog("item is inactive")
		}
		if err := local.requireUOM(ctx, request.UOMID); err != nil {
			return err
		}
		if strings.EqualFold(item.BaseUOMID, request.UOMID) && !decimalIsOne(conversion) {
			return invalidCatalog("base UOM conversion must be 1")
		}
		value := model.ItemUOM{ItemID: item.ID, UOMID: request.UOMID, ConversionToBase: conversion,
			Length: normalizeDecimal(request.Length), Width: normalizeDecimal(request.Width), Height: normalizeDecimal(request.Height), Weight: normalizeDecimal(request.Weight),
			IsReceivingUOM: catalogBool(request.IsReceivingUOM, true), IsPickingUOM: catalogBool(request.IsPickingUOM, true), IsActive: true}
		if err := repos.ItemUOM.Create(ctx, &value); err != nil {
			return err
		}
		response = mapItemUOM(value)
		return nil
	})
	return response, catalogError(err)
}
func (s *CatalogService) ListItemUOM(ctx context.Context, itemID string) ([]dto.ItemUOMResponse, error) {
	if validateID(itemID) != nil {
		return nil, ErrInvalidInput
	}
	if _, err := s.repositories.Item.Get(ctx, itemID); err != nil {
		return nil, catalogError(err)
	}
	rows, _, err := s.repositories.ItemUOM.List(ctx, repository.CatalogFilter{ItemID: itemID})
	result := make([]dto.ItemUOMResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapItemUOM(row))
	}
	return result, catalogError(err)
}
func (s *CatalogService) UpdateItemUOM(ctx context.Context, itemID, id string, request dto.UpdateItemUOMRequest) (response dto.ItemUOMResponse, err error) {
	if validateID(id) != nil || request.IsActive == nil {
		return response, ErrInvalidInput
	}
	conversion, err := catalogDecimal(request.ConversionToBase, true)
	if err != nil {
		return response, err
	}
	if err := catalogDimensions(request.Length, request.Width, request.Height, request.Weight); err != nil {
		return response, err
	}
	err = s.repositories.Transaction(ctx, func(repos *repository.CatalogRepositories) error {
		local := NewCatalogService(repos)
		item, err := local.itemForWrite(ctx, itemID)
		if err != nil {
			return err
		}
		current, err := repos.ItemUOM.Get(ctx, id)
		if err != nil {
			return err
		}
		if !strings.EqualFold(current.ItemID, itemID) {
			return ErrNotFound
		}
		if strings.EqualFold(current.UOMID, item.BaseUOMID) && (!decimalIsOne(conversion) || !*request.IsActive) {
			return invalidCatalog("base UOM must remain active with conversion 1")
		}
		if *request.IsActive {
			if !item.IsActive {
				return invalidCatalog("item is inactive")
			}
			if err := local.requireUOM(ctx, current.UOMID); err != nil {
				return err
			}
		} else if err := local.canDeactivateItemUOM(ctx, item, current); err != nil {
			return err
		}
		value, err := repos.ItemUOM.Update(ctx, id, map[string]interface{}{
			"conversion_to_base": conversion, "length": normalizeDecimal(request.Length), "width": normalizeDecimal(request.Width),
			"height": normalizeDecimal(request.Height), "weight": normalizeDecimal(request.Weight),
			"is_receiving_uom": catalogBool(request.IsReceivingUOM, current.IsReceivingUOM), "is_picking_uom": catalogBool(request.IsPickingUOM, current.IsPickingUOM),
			"is_active": *request.IsActive,
		}, "", nil)
		response = mapItemUOM(value)
		return err
	})
	return response, catalogError(err)
}
func (s *CatalogService) canDeactivateItemUOM(ctx context.Context, item model.Item, unit model.ItemUOM) error {
	if strings.EqualFold(item.BaseUOMID, unit.UOMID) {
		return invalidCatalog("base UOM cannot be deactivated")
	}
	used, err := s.repositories.ItemBarcode.HasActiveUnit(ctx, item.ID, unit.UOMID)
	if err != nil {
		return err
	}
	if used {
		return invalidCatalog("deactivate barcodes using this UOM first")
	}
	return nil
}
func (s *CatalogService) DeactivateItemUOM(ctx context.Context, itemID, id string) (response dto.ItemUOMResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}
	err = s.repositories.Transaction(ctx, func(repos *repository.CatalogRepositories) error {
		local := NewCatalogService(repos)
		item, err := local.itemForWrite(ctx, itemID)
		if err != nil {
			return err
		}
		current, err := repos.ItemUOM.Get(ctx, id)
		if err != nil {
			return err
		}
		if !strings.EqualFold(current.ItemID, itemID) {
			return ErrNotFound
		}
		if err := local.canDeactivateItemUOM(ctx, item, current); err != nil {
			return err
		}
		value, err := repos.ItemUOM.Update(ctx, id, map[string]interface{}{"is_active": false}, "", nil)
		response = mapItemUOM(value)
		return err
	})
	return response, catalogError(err)
}
func (s *CatalogService) requireBarcodeUOM(ctx context.Context, itemID string, uomID *string) error {
	if uomID == nil {
		return nil
	}
	if err := s.requireUOM(ctx, *uomID); err != nil {
		return err
	}
	unit, err := s.repositories.ItemUOM.ForUnit(ctx, itemID, *uomID)
	if errors.Is(catalogError(err), ErrNotFound) {
		return invalidCatalog("barcode UOM must be assigned to this item")
	}
	if err != nil {
		return catalogError(err)
	}
	if !unit.IsActive {
		return invalidCatalog("item UOM is inactive")
	}
	return nil
}
func (s *CatalogService) CreateItemBarcode(ctx context.Context, itemID string, request dto.CreateItemBarcodeRequest) (response dto.ItemBarcodeResponse, err error) {
	barcode := strings.TrimSpace(request.Barcode)
	if barcode == "" {
		return response, ErrInvalidInput
	}
	err = s.repositories.Transaction(ctx, func(repos *repository.CatalogRepositories) error {
		local := NewCatalogService(repos)
		item, err := local.itemForWrite(ctx, itemID)
		if err != nil {
			return err
		}
		if !item.IsActive {
			return invalidCatalog("item is inactive")
		}
		if err := local.requireBarcodeUOM(ctx, itemID, request.UOMID); err != nil {
			return err
		}
		if request.IsPrimary {
			if err := repos.ItemBarcode.ClearPrimary(ctx, itemID); err != nil {
				return err
			}
		}
		value := model.ItemBarcode{ItemID: item.ID, UOMID: request.UOMID, Barcode: barcode, IsPrimary: request.IsPrimary, IsActive: true}
		if err := repos.ItemBarcode.Create(ctx, &value); err != nil {
			return err
		}
		response = mapItemBarcode(value)
		return nil
	})
	return response, catalogError(err)
}
func (s *CatalogService) ListItemBarcode(ctx context.Context, itemID string) ([]dto.ItemBarcodeResponse, error) {
	if validateID(itemID) != nil {
		return nil, ErrInvalidInput
	}
	if _, err := s.repositories.Item.Get(ctx, itemID); err != nil {
		return nil, catalogError(err)
	}
	rows, _, err := s.repositories.ItemBarcode.List(ctx, repository.CatalogFilter{ItemID: itemID})
	result := make([]dto.ItemBarcodeResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapItemBarcode(row))
	}
	return result, catalogError(err)
}
func (s *CatalogService) UpdateItemBarcode(ctx context.Context, itemID, id string, request dto.UpdateItemBarcodeRequest) (response dto.ItemBarcodeResponse, err error) {
	barcode := strings.TrimSpace(request.Barcode)
	if validateID(id) != nil || barcode == "" || request.IsActive == nil {
		return response, ErrInvalidInput
	}
	err = s.repositories.Transaction(ctx, func(repos *repository.CatalogRepositories) error {
		local := NewCatalogService(repos)
		item, err := local.itemForWrite(ctx, itemID)
		if err != nil {
			return err
		}
		current, err := repos.ItemBarcode.Get(ctx, id)
		if err != nil {
			return err
		}
		if !strings.EqualFold(current.ItemID, itemID) {
			return ErrNotFound
		}
		if *request.IsActive {
			if !item.IsActive {
				return invalidCatalog("item is inactive")
			}
			if err := local.requireBarcodeUOM(ctx, itemID, request.UOMID); err != nil {
				return err
			}
		} else if request.UOMID != nil {
			if validateID(*request.UOMID) != nil {
				return ErrInvalidInput
			}
			if _, err := repos.ItemUOM.ForUnit(ctx, itemID, *request.UOMID); err != nil {
				return invalidCatalog("barcode UOM must be assigned to this item")
			}
		}
		value, err := repos.ItemBarcode.Update(ctx, id, map[string]interface{}{"barcode": barcode, "uom_id": request.UOMID, "is_active": *request.IsActive, "is_primary": current.IsPrimary && *request.IsActive}, "", nil)
		response = mapItemBarcode(value)
		return err
	})
	return response, catalogError(err)
}
func (s *CatalogService) DeactivateItemBarcode(ctx context.Context, itemID, id string) (response dto.ItemBarcodeResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}
	err = s.repositories.Transaction(ctx, func(repos *repository.CatalogRepositories) error {
		local := NewCatalogService(repos)
		if _, err := local.itemForWrite(ctx, itemID); err != nil {
			return err
		}
		current, err := repos.ItemBarcode.Get(ctx, id)
		if err != nil {
			return err
		}
		if !strings.EqualFold(current.ItemID, itemID) {
			return ErrNotFound
		}
		value, err := repos.ItemBarcode.Update(ctx, id, map[string]interface{}{"is_active": false, "is_primary": false}, "", nil)
		response = mapItemBarcode(value)
		return err
	})
	return response, catalogError(err)
}
func (s *CatalogService) SetPrimaryItemBarcode(ctx context.Context, itemID, id string) (response dto.ItemBarcodeResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}
	err = s.repositories.Transaction(ctx, func(repos *repository.CatalogRepositories) error {
		local := NewCatalogService(repos)
		item, err := local.itemForWrite(ctx, itemID)
		if err != nil {
			return err
		}
		current, err := repos.ItemBarcode.Get(ctx, id)
		if err != nil {
			return err
		}
		if !strings.EqualFold(current.ItemID, itemID) {
			return ErrNotFound
		}
		if !item.IsActive || !current.IsActive {
			return invalidCatalog("item and barcode must be active")
		}
		if err := local.requireBarcodeUOM(ctx, itemID, current.UOMID); err != nil {
			return err
		}
		if err := repos.ItemBarcode.ClearPrimary(ctx, itemID); err != nil {
			return err
		}
		value, err := repos.ItemBarcode.Update(ctx, id, map[string]interface{}{"is_primary": true}, "", nil)
		response = mapItemBarcode(value)
		return err
	})
	return response, catalogError(err)
}
