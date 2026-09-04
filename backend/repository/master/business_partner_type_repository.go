package master

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/master"
)

type BusinessPartnerTypeRepository struct{ db *gorm.DB }

func NewBusinessPartnerTypeRepository(db *gorm.DB) *BusinessPartnerTypeRepository {
	return &BusinessPartnerTypeRepository{db: db}
}

func (r *BusinessPartnerTypeRepository) List(ctx context.Context, partnerID string) ([]model.PartnerType, error) {
	rows := make([]model.PartnerType, 0)
	err := r.db.WithContext(ctx).Table("partner_type AS pt").Select("pt.*").
		Joins("JOIN business_partner_type bpt ON bpt.partner_type_id = pt.partner_type_id").
		Where("bpt.partner_id = ?", partnerID).Order("pt.code").Scan(&rows).Error
	return rows, err
}

func (r *BusinessPartnerTypeRepository) Assign(ctx context.Context, partnerID, typeID string) error {
	value := model.BusinessPartnerType{PartnerID: partnerID, PartnerTypeID: typeID}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&value).Error
}

func (r *BusinessPartnerTypeRepository) Remove(ctx context.Context, partnerID, typeID string) error {
	result := r.db.WithContext(ctx).Where("partner_id = ? AND partner_type_id = ?", partnerID, typeID).
		Delete(&model.BusinessPartnerType{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
