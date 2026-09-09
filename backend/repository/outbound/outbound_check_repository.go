package outbound

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
	model "wms-api/models/outbound"
)

type CheckRow struct {
	model.OutboundCheck
	StatusCode, OutboundID, ClientDeliveryOrderNo, OwnerID, WarehouseID string
	LineCount                                                           int64
}

type OutboundCheckRepository struct{ db *gorm.DB }

func NewOutboundCheckRepository(db *gorm.DB) *OutboundCheckRepository {
	return &OutboundCheckRepository{db: db}
}
func checkQuery(db *gorm.DB) *gorm.DB {
	return db.Table("outbound_check c").Select("c.*,s.code status_code,st.outbound_id,o.client_delivery_order_no,o.owner_id,o.warehouse_id,(SELECT count(*) FROM outbound_check_line l WHERE l.outbound_check_id=c.outbound_check_id) line_count").Joins("JOIN document_status s ON s.status_id=c.status_id").Joins("JOIN outbound_staging st ON st.staging_id=c.staging_id").Joins("JOIN outbound_order o ON o.outbound_id=st.outbound_id")
}
func (r *OutboundCheckRepository) Create(ctx context.Context, v *model.OutboundCheck) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *OutboundCheckRepository) Get(ctx context.Context, id string) (CheckRow, error) {
	var v CheckRow
	err := checkQuery(r.db.WithContext(ctx)).Where("c.outbound_check_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *OutboundCheckRepository) Lock(ctx context.Context, id string) (model.OutboundCheck, error) {
	var v model.OutboundCheck
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("outbound_check_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *OutboundCheckRepository) List(ctx context.Context, f ListFilter) ([]CheckRow, int64, error) {
	q := checkQuery(r.db.WithContext(ctx)).Where("o.owner_id=?", f.OwnerID)
	if f.WarehouseID != "" {
		q = q.Where("o.warehouse_id=?", f.WarehouseID)
	}
	if f.StatusCode != "" {
		q = q.Where("s.code=?", f.StatusCode)
	}
	if f.Search != "" {
		x := "%" + f.Search + "%"
		q = q.Where("c.outbound_check_id ILIKE ? OR o.client_delivery_order_no ILIKE ?", x, x)
	}
	var total int64
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, Error(err)
	}
	var rows []CheckRow
	err := q.Order("c.created_at DESC,c.outbound_check_id").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&rows).Error
	return rows, total, Error(err)
}
func (r *OutboundCheckRepository) Complete(ctx context.Context, id, statusID, actor string) error {
	res := r.db.WithContext(ctx).Model(&model.OutboundCheck{}).Where("outbound_check_id=?", id).Updates(map[string]interface{}{"status_id": statusID, "checked_at": time.Now(), "checked_by": actor})
	if res.Error != nil {
		return Error(res.Error)
	}
	if res.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}
func (r *OutboundCheckRepository) HasUnresolvedParent(ctx context.Context, parentID string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("outbound_check_exception e").Joins("JOIN outbound_check_line l ON l.outbound_check_line_id=e.outbound_check_line_id").Joins("JOIN outbound_check_exception_status s ON s.outbound_check_exception_status_id=e.status_id").Where("l.outbound_check_id=? AND NOT s.is_final", parentID).Count(&n).Error
	return n > 0, Error(err)
}
func (r *OutboundCheckRepository) HasOpenForStaging(ctx context.Context, stagingID string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("outbound_check c").Joins("JOIN document_status s ON s.status_id=c.status_id").Where("c.staging_id=? AND s.code='OPEN'", stagingID).Count(&n).Error
	return n > 0, Error(err)
}
