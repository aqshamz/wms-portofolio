package master

import (
	"context"
	model "wms-api/models/master"
	repository "wms-api/repository/master"
)

// SeedCatalog inserts reference codes only, without replacing user edits.
func (s *CatalogService) SeedCatalog(ctx context.Context) error {
	return s.repositories.Transaction(ctx, func(repos *repository.CatalogRepositories) error {
		if err := repos.PartnerType.Seed(ctx, []model.PartnerType{
			{Code: "SUPPLIER", Name: "Supplier", IsActive: true},
			{Code: "FACTORY", Name: "Factory", IsActive: true},
			{Code: "CUSTOMER", Name: "Customer", IsActive: true},
			{Code: "STORE", Name: "Store", IsActive: true},
			{Code: "CARRIER", Name: "Carrier", IsActive: true},
			{Code: "OTHER", Name: "Other", IsActive: true},
		}); err != nil {
			return err
		}
		if err := repos.UOM.Seed(ctx, []model.UOM{
			{Code: "EA", Name: "Each", IsActive: true, DecimalScale: 0},
			{Code: "BOX", Name: "Box", IsActive: true, DecimalScale: 0},
			{Code: "CTN", Name: "Carton", IsActive: true, DecimalScale: 0},
			{Code: "PLT", Name: "Pallet", IsActive: true, DecimalScale: 0},
			{Code: "KG", Name: "Kilogram", IsActive: true, DecimalScale: 3},
			{Code: "G", Name: "Gram", IsActive: true, DecimalScale: 3},
			{Code: "L", Name: "Litre", IsActive: true, DecimalScale: 3},
			{Code: "M", Name: "Metre", IsActive: true, DecimalScale: 3},
		}); err != nil {
			return err
		}
		if err := repos.InventoryStatus.Seed(ctx, []model.InventoryStatus{
			{Code: "QC_PENDING", Name: "QC pending", IsActive: true, IsAllocatable: false, IsPickable: false},
			{Code: "PUTAWAY_PENDING", Name: "Putaway pending", IsActive: true, IsAllocatable: false, IsPickable: false},
			{Code: "AVAILABLE", Name: "Available", IsActive: true, IsAllocatable: true, IsPickable: true},
			{Code: "HOLD", Name: "Hold", IsActive: true, IsAllocatable: false, IsPickable: false},
			{Code: "QUARANTINE", Name: "Quarantine", IsActive: true, IsAllocatable: false, IsPickable: false},
			{Code: "DAMAGED", Name: "Damaged", IsActive: true, IsAllocatable: false, IsPickable: false},
			{Code: "EXPIRED", Name: "Expired", IsActive: true, IsAllocatable: false, IsPickable: false},
		}); err != nil {
			return err
		}
		if err := repos.QualityStatus.Seed(ctx, []model.QualityStatus{
			{Code: "PENDING", Name: "Pending", IsActive: true},
			{Code: "PASSED", Name: "Passed", IsActive: true},
			{Code: "FAILED", Name: "Failed", IsActive: true},
			{Code: "WAIVED", Name: "Waived", IsActive: true},
		}); err != nil {
			return err
		}
		if err := repos.InspectionResult.Seed(ctx, []model.InspectionResult{
			{Code: "ACCEPTED", Name: "Accepted", IsActive: true, IsAccepted: true},
			{Code: "PARTIAL", Name: "Partially accepted", IsActive: true, IsAccepted: true},
			{Code: "REJECTED", Name: "Rejected", IsActive: true, IsAccepted: false},
		}); err != nil {
			return err
		}
		return repos.HandlingUnitType.SeedDefaults(ctx)
	})
}
