package outbound

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/outbound"
)

type ReturnPolicyRow struct {
	model.OutboundReturnPolicy
	ReturnLocationCode, ReturnInventoryStatusCode string
	StatusAllocatable                             bool
}
type OutboundReturnPolicyRepository struct{ db *gorm.DB }

func NewOutboundReturnPolicyRepository(db *gorm.DB) *OutboundReturnPolicyRepository {
	return &OutboundReturnPolicyRepository{db: db}
}
func returnPolicyQuery(db *gorm.DB) *gorm.DB {
	return db.Table("outbound_return_policy p").Select("p.*,l.code return_location_code,s.code return_inventory_status_code,s.is_allocatable status_allocatable").Joins("JOIN warehouse_location l ON l.location_id=p.return_location_id").Joins("JOIN inventory_status s ON s.inventory_status_id=p.return_inventory_status_id")
}
func (r *OutboundReturnPolicyRepository) Get(ctx context.Context, owner, warehouse string) (ReturnPolicyRow, error) {
	var v ReturnPolicyRow
	err := returnPolicyQuery(r.db.WithContext(ctx)).Where("p.owner_id=? AND p.warehouse_id=?", owner, warehouse).Take(&v).Error
	return v, Error(err)
}
func (r *OutboundReturnPolicyRepository) Upsert(ctx context.Context, v *model.OutboundReturnPolicy) error {
	return Error(r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "owner_id"}, {Name: "warehouse_id"}}, DoUpdates: clause.AssignmentColumns([]string{"return_location_id", "return_inventory_status_id", "is_active", "updated_at", "updated_by"})}).Create(v).Error)
}
func (r *OutboundReturnPolicyRepository) Validate(ctx context.Context, owner, warehouse, location, status string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("warehouse_owner wo").Joins("JOIN warehouse_location l ON l.warehouse_id=wo.warehouse_id AND l.location_id=? AND l.is_active", location).Joins("JOIN inventory_status s ON s.inventory_status_id=? AND s.is_active AND NOT s.is_allocatable", status).Where("wo.owner_id=? AND wo.warehouse_id=? AND wo.is_active", owner, warehouse).Count(&n).Error
	return n == 1, Error(err)
}
