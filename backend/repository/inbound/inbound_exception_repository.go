package inbound

import (
	"context"

	"gorm.io/gorm"
	model "wms-api/models/inbound"
)

type InboundExceptionRepository struct{ db *gorm.DB }

func NewInboundExceptionRepository(db *gorm.DB) *InboundExceptionRepository {
	return &InboundExceptionRepository{db: db}
}

func (r *InboundExceptionRepository) Create(ctx context.Context, value *model.InboundException) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}

func (r *InboundExceptionRepository) Get(ctx context.Context, id string) (model.InboundException, error) {
	var value model.InboundException
	err := r.db.WithContext(ctx).Where("inbound_exception_id=?", id).Take(&value).Error
	return value, Error(err)
}

func (r *InboundExceptionRepository) List(ctx context.Context, filter ListFilter) ([]model.InboundException, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.InboundException{}).Where("owner_id=?", filter.OwnerID)
	if filter.WarehouseID != "" {
		query = query.Where("warehouse_id=?", filter.WarehouseID)
	}
	if filter.StatusCode != "" {
		query = query.Where("exception_type_code=?", filter.StatusCode)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("source_document_id ILIKE ? OR source_line_id ILIKE ? OR notes ILIKE ?", like, like, like)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, Error(err)
	}
	rows := make([]model.InboundException, 0)
	err := query.Order("created_at DESC,inbound_exception_id DESC").Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize).Find(&rows).Error
	return rows, total, Error(err)
}
