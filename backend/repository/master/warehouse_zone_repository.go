package master

import (
	"context"

	model "wms-api/models/master"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WarehouseZoneDetail struct {
	model.WarehouseZone
	LocationCount int64 `json:"location_count"`
}

type WarehouseZoneRepository struct{ db *gorm.DB }

func NewWarehouseZoneRepository(db *gorm.DB) *WarehouseZoneRepository {
	return &WarehouseZoneRepository{db: db}
}

func (r *WarehouseZoneRepository) Create(ctx context.Context, value *model.WarehouseZone) error {
	return r.db.WithContext(ctx).Create(value).Error
}

func (r *WarehouseZoneRepository) FindByID(ctx context.Context, id string) (model.WarehouseZone, error) {
	var value model.WarehouseZone
	err := r.db.WithContext(ctx).Where("zone_id = ?", id).Take(&value).Error
	return value, err
}

func (r *WarehouseZoneRepository) List(
	ctx context.Context, warehouseID string, search *string, active *bool,
) ([]WarehouseZoneDetail, error) {
	var rows []WarehouseZoneDetail
	query := r.db.WithContext(ctx).Table("warehouse_zone AS z").
		Select("z.*, count(l.location_id) AS location_count").
		Joins("LEFT JOIN warehouse_location l ON l.zone_id = z.zone_id").
		Where("z.warehouse_id = ?", warehouseID)
	if search != nil {
		query = query.Where("z.code ILIKE ? OR z.name ILIKE ?", "%"+*search+"%", "%"+*search+"%")
	}
	if active != nil {
		query = query.Where("z.is_active = ?", *active)
	}
	err := query.Group("z.zone_id").Order("z.code").Scan(&rows).Error
	return rows, err
}

func (r *WarehouseZoneRepository) Update(
	ctx context.Context, id string, changes map[string]interface{},
) (model.WarehouseZone, error) {
	var value model.WarehouseZone
	result := r.db.WithContext(ctx).Model(&value).Clauses(clause.Returning{}).
		Where("zone_id = ?", id).Updates(changes)
	if result.Error != nil {
		return value, result.Error
	}
	if result.RowsAffected == 0 {
		return value, gorm.ErrRecordNotFound
	}
	return value, nil
}

func (r *WarehouseZoneRepository) Deactivate(ctx context.Context, id string) (model.WarehouseZone, error) {
	return r.Update(ctx, id, map[string]interface{}{"is_active": false})
}
