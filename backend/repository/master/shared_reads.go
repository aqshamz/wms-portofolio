package master

import (
	"context"
	"gorm.io/gorm/clause"
	model "wms-api/models/master"
)

func (r *catalogTable[T]) GetShared(ctx context.Context, id string) (T, error) {
	var v T
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "SHARE"}).Where(r.key+" = ?", id).Take(&v).Error
	return v, err
}
func (r *OrganizationRepository) GetShared(ctx context.Context, id string) (model.Organization, error) {
	var v model.Organization
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "SHARE"}).Where("organization_id = ?", id).Take(&v).Error
	return v, err
}
func (r *WarehouseRepository) GetShared(ctx context.Context, id string) (model.Warehouse, error) {
	var v model.Warehouse
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "SHARE"}).Where("warehouse_id = ?", id).Take(&v).Error
	return v, err
}
func (r *WarehouseLocationRepository) GetShared(ctx context.Context, id string) (model.WarehouseLocation, error) {
	var v model.WarehouseLocation
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "SHARE"}).Where("location_id = ?", id).Take(&v).Error
	return v, err
}
func (r *WarehouseZoneRepository) GetShared(ctx context.Context, id string) (model.WarehouseZone, error) {
	var v model.WarehouseZone
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "SHARE"}).Where("zone_id = ?", id).Take(&v).Error
	return v, err
}
func (r *LocationTypeRepository) GetShared(ctx context.Context, id string) (model.LocationType, error) {
	var v model.LocationType
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "SHARE"}).Where("location_type_id = ?", id).Take(&v).Error
	return v, err
}
func (r *WarehouseOwnerRepository) GetShared(ctx context.Context, warehouseID, ownerID string) (model.WarehouseOwner, error) {
	var v model.WarehouseOwner
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "SHARE"}).Where("warehouse_id = ? AND owner_id = ?", warehouseID, ownerID).Take(&v).Error
	return v, err
}
