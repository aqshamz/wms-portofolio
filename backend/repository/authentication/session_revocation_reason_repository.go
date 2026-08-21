package authentication

import (
	"context"
	model "wms-api/models/authentication"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SessionRevocationReasonRepository struct{ db *gorm.DB }

func NewSessionRevocationReasonRepository(db *gorm.DB) *SessionRevocationReasonRepository {
	return &SessionRevocationReasonRepository{db: db}
}

func (r *SessionRevocationReasonRepository) Seed(ctx context.Context, reasons []model.SessionRevocationReason) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}},
		DoNothing: true,
	}).Create(&reasons).Error
}
