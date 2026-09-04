package master

import (
	"context"

	model "wms-api/models/master"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WarehouseLocationDetail struct {
	model.WarehouseLocation
	WarehouseCode    string `json:"warehouse_code"`
	WarehouseName    string `json:"warehouse_name"`
	ZoneCode         string `json:"zone_code"`
	ZoneName         string `json:"zone_name"`
	LocationTypeCode string `json:"location_type_code"`
	LocationTypeName string `json:"location_type_name"`
}

type WarehouseLocationListItem struct {
	model.WarehouseLocation
	ZoneCode         string
	LocationTypeCode string
	TotalRows        int64
}

type WarehouseLocationRepository struct{ db *gorm.DB }

func NewWarehouseLocationRepository(db *gorm.DB) *WarehouseLocationRepository {
	return &WarehouseLocationRepository{db: db}
}

func (r *WarehouseLocationRepository) Create(ctx context.Context, value *model.WarehouseLocation) error {
	return r.db.WithContext(ctx).Create(value).Error
}

func (r *WarehouseLocationRepository) FindByID(ctx context.Context, id string) (WarehouseLocationDetail, error) {
	var value WarehouseLocationDetail
	err := r.db.WithContext(ctx).Table("warehouse_location AS l").
		Select(`l.*, w.code AS warehouse_code, w.name AS warehouse_name,
			z.code AS zone_code, z.name AS zone_name,
			lt.code AS location_type_code, lt.name AS location_type_name`).
		Joins("JOIN warehouse w ON w.warehouse_id = l.warehouse_id").
		Joins("JOIN warehouse_zone z ON z.zone_id = l.zone_id").
		Joins("JOIN location_type lt ON lt.location_type_id = l.location_type_id").
		Where("l.location_id = ?", id).Take(&value).Error
	return value, err
}

func (r *WarehouseLocationRepository) List(
	ctx context.Context,
	warehouseID string,
	zoneID, locationTypeID, search *string,
	active *bool,
	limit, offset int,
) ([]WarehouseLocationListItem, int64, error) {
	var rows []WarehouseLocationListItem
	query := r.db.WithContext(ctx).Table("warehouse_location AS l").
		Select(`l.*, z.code AS zone_code, lt.code AS location_type_code,
			count(*) OVER () AS total_rows`).
		Joins("JOIN warehouse_zone z ON z.zone_id = l.zone_id").
		Joins("JOIN location_type lt ON lt.location_type_id = l.location_type_id").
		Where("l.warehouse_id = ?", warehouseID)
	if zoneID != nil {
		query = query.Where("l.zone_id = ?", *zoneID)
	}
	if locationTypeID != nil {
		query = query.Where("l.location_type_id = ?", *locationTypeID)
	}
	if search != nil {
		query = query.Where("l.code ILIKE ? OR l.barcode ILIKE ?", "%"+*search+"%", "%"+*search+"%")
	}
	if active != nil {
		query = query.Where("l.is_active = ?", *active)
	}
	err := query.Order("l.code, l.location_id").Limit(limit).Offset(offset).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return rows, 0, err
	}
	return rows, rows[0].TotalRows, nil
}

func (r *WarehouseLocationRepository) Update(
	ctx context.Context, id string, changes map[string]interface{},
) (model.WarehouseLocation, error) {
	var value model.WarehouseLocation
	result := r.db.WithContext(ctx).Model(&value).Clauses(clause.Returning{}).
		Where("location_id = ?", id).Updates(changes)
	if result.Error != nil {
		return value, result.Error
	}
	if result.RowsAffected == 0 {
		return value, gorm.ErrRecordNotFound
	}
	return value, nil
}

func (r *WarehouseLocationRepository) Deactivate(
	ctx context.Context, id string,
) (model.WarehouseLocation, error) {
	return r.Update(ctx, id, map[string]interface{}{"is_active": false})
}
