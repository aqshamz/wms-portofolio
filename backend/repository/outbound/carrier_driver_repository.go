package outbound

import (
	"context"

	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type CarrierDriverRow struct {
	model.CarrierDriver
	CarrierCode string
}
type CarrierDriverRepository struct{ db *gorm.DB }

func NewCarrierDriverRepository(db *gorm.DB) *CarrierDriverRepository {
	return &CarrierDriverRepository{db: db}
}
func (r *CarrierDriverRepository) Create(ctx context.Context, value *model.CarrierDriver) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}
func carrierDriverQuery(db *gorm.DB) *gorm.DB {
	return db.Table("carrier_driver d").Select("d.*,c.code carrier_code").
		Joins("JOIN carrier c ON c.carrier_id=d.carrier_id")
}
func (r *CarrierDriverRepository) Get(ctx context.Context, id string) (CarrierDriverRow, error) {
	var value CarrierDriverRow
	err := carrierDriverQuery(r.db.WithContext(ctx)).Where("d.driver_id=?", id).Take(&value).Error
	return value, Error(err)
}
func (r *CarrierDriverRepository) List(ctx context.Context, carrierID string, active *bool) ([]CarrierDriverRow, error) {
	query := carrierDriverQuery(r.db.WithContext(ctx)).Where("d.carrier_id=?", carrierID)
	if active != nil {
		query = query.Where("d.is_active=?", *active)
	}
	var values []CarrierDriverRow
	err := query.Order("d.code").Find(&values).Error
	return values, Error(err)
}
func (r *CarrierDriverRepository) Update(ctx context.Context, id, carrierID string, values map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&model.CarrierDriver{}).
		Where("driver_id=? AND carrier_id=?", id, carrierID).Updates(values)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}
