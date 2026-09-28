package inbound

import (
	"context"

	"gorm.io/gorm"
	outboundmodel "wms-api/models/outbound"
)

type QuarantineVendorReturnRepository struct{ db *gorm.DB }

func NewQuarantineVendorReturnRepository(db *gorm.DB) *QuarantineVendorReturnRepository {
	return &QuarantineVendorReturnRepository{db: db}
}

func (r *QuarantineVendorReturnRepository) Create(ctx context.Context, value *outboundmodel.VendorReturnTransaction) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}
