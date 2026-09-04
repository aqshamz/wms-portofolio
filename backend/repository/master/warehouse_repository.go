package master

import (
	"context"
	"time"

	model "wms-api/models/master"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WarehouseDetail struct {
	model.Warehouse
	OperatorCode string `json:"operator_code"`
	OperatorName string `json:"operator_name"`
}

type WarehouseListItem struct {
	ID           string
	OperatorID   string
	Code         string
	Name         string
	OperatorCode string
	OperatorName string
	TimezoneName string
	City         *string
	CountryCode  *string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	TotalRows    int64
}

type WarehouseRepository struct{ db *gorm.DB }

func NewWarehouseRepository(db *gorm.DB) *WarehouseRepository {
	return &WarehouseRepository{db: db}
}

func (r *WarehouseRepository) Create(ctx context.Context, warehouse *model.Warehouse) error {
	return r.db.WithContext(ctx).Create(warehouse).Error
}

func (r *WarehouseRepository) FindByID(ctx context.Context, id string) (WarehouseDetail, error) {
	var warehouse WarehouseDetail
	err := r.db.WithContext(ctx).
		Table("warehouse AS w").
		Select("w.*, operator.code AS operator_code, operator.name AS operator_name").
		Joins("JOIN organization operator ON operator.organization_id = w.operator_id").
		Where("w.warehouse_id = ?", id).
		Take(&warehouse).Error
	return warehouse, err
}

func (r *WarehouseRepository) List(
	ctx context.Context,
	operatorID, search *string,
	active *bool,
	limit, offset int,
) ([]WarehouseListItem, int64, error) {
	var rows []WarehouseListItem
	query := r.db.WithContext(ctx).
		Table("warehouse AS w").
		Select(`w.warehouse_id AS id, w.operator_id, w.code, w.name,
			operator.code AS operator_code, operator.name AS operator_name,
			w.timezone_name, w.city, w.country_code, w.is_active, w.created_at, w.updated_at,
			count(*) OVER () AS total_rows`).
		Joins("JOIN organization operator ON operator.organization_id = w.operator_id")
	if operatorID != nil {
		query = query.Where("w.operator_id = ?", *operatorID)
	}
	if search != nil {
		query = query.Where("w.code ILIKE ? OR w.name ILIKE ?", "%"+*search+"%", "%"+*search+"%")
	}
	if active != nil {
		query = query.Where("w.is_active = ?", *active)
	}
	err := query.Order("w.name, w.warehouse_id").Limit(limit).Offset(offset).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return rows, 0, err
	}
	return rows, rows[0].TotalRows, nil
}

func (r *WarehouseRepository) Update(
	ctx context.Context,
	id string,
	expectedUpdatedAt time.Time,
	changes map[string]interface{},
) (model.Warehouse, error) {
	var warehouse model.Warehouse
	result := r.db.WithContext(ctx).Model(&warehouse).
		Clauses(clause.Returning{}).
		Where("warehouse_id = ? AND updated_at = ?", id, expectedUpdatedAt).
		Updates(changes)
	if result.Error != nil {
		return warehouse, result.Error
	}
	if result.RowsAffected == 0 {
		return warehouse, ErrConcurrentUpdate
	}
	return warehouse, nil
}

func (r *WarehouseRepository) Deactivate(
	ctx context.Context,
	id, actorID string,
	expectedUpdatedAt time.Time,
) (model.Warehouse, error) {
	return r.Update(ctx, id, expectedUpdatedAt, map[string]interface{}{
		"is_active":  false,
		"updated_at": gorm.Expr("clock_timestamp()"),
		"updated_by": actorID,
	})
}
