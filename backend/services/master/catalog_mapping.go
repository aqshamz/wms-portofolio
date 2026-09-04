package master

import (
	dto "wms-api/dto/master"
	model "wms-api/models/master"
)

func mapPartnerType(v model.PartnerType) dto.PartnerTypeResponse {
	return dto.PartnerTypeResponse{
		ID:          v.ID,
		Code:        v.Code,
		Name:        v.Name,
		Description: v.Description,
		IsActive:    v.IsActive,
	}
}

func mapBusinessPartner(v model.BusinessPartner) dto.BusinessPartnerResponse {
	return dto.BusinessPartnerResponse{
		ID:           v.ID,
		OwnerID:      v.OwnerID,
		Code:         v.Code,
		Name:         v.Name,
		LegalName:    v.LegalName,
		TaxNumber:    v.TaxNumber,
		Email:        v.Email,
		Phone:        v.Phone,
		AddressLine1: v.AddressLine1,
		AddressLine2: v.AddressLine2,
		City:         v.City,
		Province:     v.Province,
		PostalCode:   v.PostalCode,
		CountryCode:  v.CountryCode,
		IsActive:     v.IsActive,
		CreatedAt:    v.CreatedAt,
		CreatedBy:    v.CreatedBy,
		UpdatedAt:    v.UpdatedAt,
		UpdatedBy:    v.UpdatedBy,
	}
}

func mapUOM(v model.UOM) dto.UOMResponse {
	return dto.UOMResponse{
		ID:           v.ID,
		Code:         v.Code,
		Name:         v.Name,
		DecimalScale: v.DecimalScale,
		IsActive:     v.IsActive,
	}
}

func mapItemCategory(v model.ItemCategory) dto.ItemCategoryResponse {
	return dto.ItemCategoryResponse{
		ID:               v.ID,
		OwnerID:          v.OwnerID,
		ParentCategoryID: v.ParentCategoryID,
		Code:             v.Code,
		Name:             v.Name,
		IsActive:         v.IsActive,
	}
}

func mapItem(v model.Item) dto.ItemResponse {
	return dto.ItemResponse{
		ID:                 v.ID,
		OwnerID:            v.OwnerID,
		CategoryID:         v.CategoryID,
		Code:               v.Code,
		Name:               v.Name,
		Description:        v.Description,
		BaseUOMID:          v.BaseUOMID,
		Weight:             v.Weight,
		Volume:             v.Volume,
		LotControlled:      v.LotControlled,
		SerialControlled:   v.SerialControlled,
		ShelfLifeDays:      v.ShelfLifeDays,
		MinimumReceiveDays: v.MinimumReceiveDays,
		IsActive:           v.IsActive,
		CreatedAt:          v.CreatedAt,
		CreatedBy:          v.CreatedBy,
		UpdatedAt:          v.UpdatedAt,
		UpdatedBy:          v.UpdatedBy,
	}
}

func mapItemUOM(v model.ItemUOM) dto.ItemUOMResponse {
	return dto.ItemUOMResponse{
		ID:               v.ID,
		ItemID:           v.ItemID,
		UOMID:            v.UOMID,
		ConversionToBase: v.ConversionToBase,
		Length:           v.Length,
		Width:            v.Width,
		Height:           v.Height,
		Weight:           v.Weight,
		IsReceivingUOM:   v.IsReceivingUOM,
		IsPickingUOM:     v.IsPickingUOM,
		IsActive:         v.IsActive,
	}
}

func mapItemBarcode(v model.ItemBarcode) dto.ItemBarcodeResponse {
	return dto.ItemBarcodeResponse{
		ID:        v.ID,
		ItemID:    v.ItemID,
		UOMID:     v.UOMID,
		Barcode:   v.Barcode,
		IsPrimary: v.IsPrimary,
		IsActive:  v.IsActive,
	}
}

func mapInventoryStatus(v model.InventoryStatus) dto.InventoryStatusResponse {
	return dto.InventoryStatusResponse{
		ID:            v.ID,
		Code:          v.Code,
		Name:          v.Name,
		Description:   v.Description,
		IsAllocatable: v.IsAllocatable,
		IsPickable:    v.IsPickable,
		IsActive:      v.IsActive,
	}
}

func mapQualityStatus(v model.QualityStatus) dto.QualityStatusResponse {
	return dto.QualityStatusResponse{
		ID:          v.ID,
		Code:        v.Code,
		Name:        v.Name,
		Description: v.Description,
		IsActive:    v.IsActive,
	}
}

func mapInspectionResult(v model.InspectionResult) dto.InspectionResultResponse {
	return dto.InspectionResultResponse{
		ID:          v.ID,
		Code:        v.Code,
		Name:        v.Name,
		Description: v.Description,
		IsAccepted:  v.IsAccepted,
		IsActive:    v.IsActive,
	}
}
