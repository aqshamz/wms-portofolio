package master

import "gorm.io/gorm"

type CatalogRepositories struct {
	db                  *gorm.DB
	PartnerType         *PartnerTypeRepository
	BusinessPartner     *BusinessPartnerRepository
	BusinessPartnerType *BusinessPartnerTypeRepository
	UOM                 *UOMRepository
	ItemCategory        *ItemCategoryRepository
	Item                *ItemRepository
	ItemUOM             *ItemUOMRepository
	ItemBarcode         *ItemBarcodeRepository
	InventoryStatus     *InventoryStatusRepository
	QualityStatus       *QualityStatusRepository
	InspectionResult    *InspectionResultRepository
	HandlingUnitType    *HandlingUnitTypeRepository
}

func NewCatalogRepositories(db *gorm.DB) *CatalogRepositories {
	return &CatalogRepositories{db: db,
		PartnerType:         NewPartnerTypeRepository(db),
		BusinessPartner:     NewBusinessPartnerRepository(db),
		BusinessPartnerType: NewBusinessPartnerTypeRepository(db),
		UOM:                 NewUOMRepository(db),
		ItemCategory:        NewItemCategoryRepository(db),
		Item:                NewItemRepository(db),
		ItemUOM:             NewItemUOMRepository(db),
		ItemBarcode:         NewItemBarcodeRepository(db),
		InventoryStatus:     NewInventoryStatusRepository(db),
		QualityStatus:       NewQualityStatusRepository(db),
		InspectionResult:    NewInspectionResultRepository(db),
		HandlingUnitType:    NewHandlingUnitTypeRepository(db),
	}
}
