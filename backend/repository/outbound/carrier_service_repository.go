package outbound

import (
	"context"

	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type CarrierServiceRow struct {
	model.CarrierService
	CarrierCode string
}
type CarrierServiceRepository struct{ db *gorm.DB }

func NewCarrierServiceRepository(db *gorm.DB) *CarrierServiceRepository {
	return &CarrierServiceRepository{db: db}
}
func (r *CarrierServiceRepository) Create(ctx context.Context, value *model.CarrierService) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}
func carrierServiceQuery(db *gorm.DB) *gorm.DB {
	return db.Table("carrier_service s").Select("s.*,c.code carrier_code").
		Joins("JOIN carrier c ON c.carrier_id=s.carrier_id")
}
func (r *CarrierServiceRepository) Get(ctx context.Context, id string) (CarrierServiceRow, error) {
	var value CarrierServiceRow
	err := carrierServiceQuery(r.db.WithContext(ctx)).Where("s.carrier_service_id=?", id).Take(&value).Error
	return value, Error(err)
}
func (r *CarrierServiceRepository) List(ctx context.Context, carrierID string, active *bool) ([]CarrierServiceRow, error) {
	query := carrierServiceQuery(r.db.WithContext(ctx)).Where("s.carrier_id=?", carrierID)
	if active != nil {
		query = query.Where("s.is_active=?", *active)
	}
	var values []CarrierServiceRow
	err := query.Order("s.code").Find(&values).Error
	return values, Error(err)
}
func (r *CarrierServiceRepository) Update(ctx context.Context, id, carrierID string, values map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&model.CarrierService{}).
		Where("carrier_service_id=? AND carrier_id=?", id, carrierID).Updates(values)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}
