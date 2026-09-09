package outbound

import (
	"context"

	"gorm.io/gorm"
	model "wms-api/models/outbound"
)

type CarrierRepository struct{ db *gorm.DB }

func NewCarrierRepository(db *gorm.DB) *CarrierRepository { return &CarrierRepository{db: db} }

func (r *CarrierRepository) Create(ctx context.Context, value *model.Carrier) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}
func (r *CarrierRepository) Get(ctx context.Context, id string) (model.Carrier, error) {
	var value model.Carrier
	err := r.db.WithContext(ctx).Where("carrier_id=?", id).Take(&value).Error
	return value, Error(err)
}
func (r *CarrierRepository) List(ctx context.Context, active *bool, search string) ([]model.Carrier, error) {
	query := r.db.WithContext(ctx).Model(&model.Carrier{})
	if active != nil {
		query = query.Where("is_active=?", *active)
	}
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("code ILIKE ? OR name ILIKE ?", like, like)
	}
	var values []model.Carrier
	err := query.Order("code").Find(&values).Error
	return values, Error(err)
}
func (r *CarrierRepository) Update(ctx context.Context, id string, values map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&model.Carrier{}).Where("carrier_id=?", id).Updates(values)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}
