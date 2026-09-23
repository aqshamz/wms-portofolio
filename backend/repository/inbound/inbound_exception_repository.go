package inbound

import (
	"context"

	"gorm.io/gorm"
	model "wms-api/models/inbound"
)

type InboundExceptionRepository struct{ db *gorm.DB }

type InboundExceptionRow struct {
	model.InboundException
	OwnerName, WarehouseName, CreatedByDisplayName *string
}

func inboundExceptionQuery(db *gorm.DB) *gorm.DB {
	return db.Table("inbound_exception exception").
		Select("exception.*,owner.name owner_name,warehouse.name warehouse_name,account.display_name created_by_display_name").
		Joins("LEFT JOIN organization owner ON owner.organization_id=exception.owner_id").
		Joins("LEFT JOIN warehouse ON warehouse.warehouse_id=exception.warehouse_id").
		Joins("LEFT JOIN app_account account ON account.account_id=exception.created_by")
}

func NewInboundExceptionRepository(db *gorm.DB) *InboundExceptionRepository {
	return &InboundExceptionRepository{db: db}
}

func (r *InboundExceptionRepository) Create(ctx context.Context, value *model.InboundException) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}

func (r *InboundExceptionRepository) Get(ctx context.Context, id string) (InboundExceptionRow, error) {
	var value InboundExceptionRow
	err := inboundExceptionQuery(r.db.WithContext(ctx)).Where("exception.inbound_exception_id=?", id).Take(&value).Error
	return value, Error(err)
}

func (r *InboundExceptionRepository) List(ctx context.Context, filter ListFilter) ([]InboundExceptionRow, int64, error) {
	query := inboundExceptionQuery(r.db.WithContext(ctx)).Where("exception.owner_id=?", filter.OwnerID)
	if filter.WarehouseID != "" {
		query = query.Where("exception.warehouse_id=?", filter.WarehouseID)
	}
	if filter.StatusCode != "" {
		query = query.Where("exception.exception_type_code=?", filter.StatusCode)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("exception.source_document_id ILIKE ? OR exception.source_line_id ILIKE ? OR exception.notes ILIKE ?", like, like, like)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, Error(err)
	}
	rows := make([]InboundExceptionRow, 0)
	err := query.Order("exception.created_at DESC,exception.inbound_exception_id DESC").Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize).Find(&rows).Error
	return rows, total, Error(err)
}
