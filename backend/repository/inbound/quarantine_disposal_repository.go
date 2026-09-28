package inbound

import (
	"context"

	"gorm.io/gorm"
	outboundmodel "wms-api/models/outbound"
)

type QuarantineDisposalRepository struct{ db *gorm.DB }

func NewQuarantineDisposalRepository(db *gorm.DB) *QuarantineDisposalRepository {
	return &QuarantineDisposalRepository{db: db}
}

func (r *QuarantineDisposalRepository) Create(ctx context.Context, value *outboundmodel.DisposalTransaction) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}
