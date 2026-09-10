package audit

import (
	"context"

	"gorm.io/gorm"
	model "wms-api/models/audit"
)

type APIAuditLogRepository struct{ db *gorm.DB }

func NewAPIAuditLogRepository(db *gorm.DB) *APIAuditLogRepository {
	return &APIAuditLogRepository{db: db}
}

func (r *APIAuditLogRepository) Create(ctx context.Context, value *model.APIAuditLog) error {
	return r.db.WithContext(ctx).Create(value).Error
}
