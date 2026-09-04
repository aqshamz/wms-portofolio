package master

import (
	"context"
	"time"

	model "wms-api/models/master"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WarehouseOwnerDetail struct {
	WarehouseID string
	OwnerID     string
	OwnerCode   string
	OwnerName   string
	IsActive    bool
	CreatedAt   time.Time
}

type WarehouseOwnerRepository struct{ db *gorm.DB }

func (r *WarehouseOwnerRepository) IsActive(ctx context.Context, warehouseID, ownerID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.WarehouseOwner{}).Where("warehouse_id = ? AND owner_id = ? AND is_active", warehouseID, ownerID).Count(&count).Error
	return count > 0, err
}

func NewWarehouseOwnerRepository(db *gorm.DB) *WarehouseOwnerRepository {
	return &WarehouseOwnerRepository{db: db}
}

func (r *WarehouseOwnerRepository) Assign(ctx context.Context, assignment *model.WarehouseOwner) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "warehouse_id"}, {Name: "owner_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"is_active":  true,
			"created_by": assignment.CreatedBy,
		}),
	}).Create(assignment).Error
}

func (r *WarehouseOwnerRepository) List(ctx context.Context, warehouseID string) ([]WarehouseOwnerDetail, error) {
	var rows []WarehouseOwnerDetail
	err := r.db.WithContext(ctx).Table("warehouse_owner AS wo").
		Select(`wo.warehouse_id, wo.owner_id, owner.code AS owner_code,
			owner.name AS owner_name, wo.is_active, wo.created_at`).
		Joins("JOIN organization owner ON owner.organization_id = wo.owner_id").
		Where("wo.warehouse_id = ?", warehouseID).
		Order("owner.name").
		Scan(&rows).Error
	return rows, err
}

func (r *WarehouseOwnerRepository) Deactivate(ctx context.Context, warehouseID, ownerID string) error {
	result := r.db.WithContext(ctx).Model(&model.WarehouseOwner{}).
		Where("warehouse_id = ? AND owner_id = ?", warehouseID, ownerID).
		Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
