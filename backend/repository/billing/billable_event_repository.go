package billing

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/billing"
)

type EventRow struct {
	model.BillableEvent
	StatusCode, ServiceCode   string
	MovementTypeCode, UOMCode *string
}
type BillableEventRepository struct{ db *gorm.DB }

func NewBillableEventRepository(db *gorm.DB) *BillableEventRepository {
	return &BillableEventRepository{db: db}
}
func (r *BillableEventRepository) CreateIgnore(ctx context.Context, v *model.BillableEvent) (bool, error) {
	z := r.db.WithContext(ctx).Exec(`INSERT INTO billable_event (billable_event_id,document_type_id,status_id,owner_id,warehouse_id,rate_card_line_id,inventory_movement_id,event_key,business_date,source_document_id,source_line_id,quantity,uom_id,notes,created_at,created_by) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,clock_timestamp(),?) ON CONFLICT(event_key) DO NOTHING`, v.ID, v.DocumentTypeID, v.StatusID, v.OwnerID, v.WarehouseID, v.RateCardLineID, v.InventoryMovementID, v.EventKey, v.BusinessDate, v.SourceDocumentID, v.SourceLineID, v.Quantity, v.UOMID, v.Notes, v.CreatedBy)
	return z.RowsAffected == 1, Error(z.Error)
}
func eventQuery(db *gorm.DB) *gorm.DB {
	return db.Table("billable_event e").Select("e.*,s.code status_code,l.service_code,m.code movement_type_code,u.code uom_code").Joins("JOIN document_status s ON s.status_id=e.status_id").Joins("JOIN rate_card_line l ON l.rate_card_line_id=e.rate_card_line_id").Joins("LEFT JOIN movement_type m ON m.movement_type_id=l.movement_type_id").Joins("LEFT JOIN uom u ON u.uom_id=e.uom_id")
}
func (r *BillableEventRepository) Get(ctx context.Context, id string) (EventRow, error) {
	var v EventRow
	e := eventQuery(r.db.WithContext(ctx)).Where("e.billable_event_id=?", id).Take(&v).Error
	return v, Error(e)
}
func (r *BillableEventRepository) List(ctx context.Context, f ListFilter) ([]EventRow, int64, error) {
	q := eventQuery(r.db.WithContext(ctx)).Where("e.owner_id=?", f.OwnerID)
	if f.WarehouseID != "" {
		q = q.Where("e.warehouse_id=?", f.WarehouseID)
	}
	if f.StatusCode != "" {
		q = q.Where("s.code=?", f.StatusCode)
	}
	if f.DateFrom != "" {
		q = q.Where("e.business_date>=?", f.DateFrom)
	}
	if f.DateUntil != "" {
		q = q.Where("e.business_date<=?", f.DateUntil)
	}
	if f.Search != "" {
		x := "%" + f.Search + "%"
		q = q.Where("e.billable_event_id ILIKE ? OR e.source_document_id ILIKE ? OR l.service_code ILIKE ?", x, x, x)
	}
	var n int64
	if e := q.Session(&gorm.Session{}).Count(&n).Error; e != nil {
		return nil, 0, Error(e)
	}
	v := []EventRow{}
	e := q.Order("e.business_date DESC,e.billable_event_id DESC").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&v).Error
	return v, n, Error(e)
}
func (r *BillableEventRepository) PendingForRun(ctx context.Context, contract string, from, until interface{}) ([]EventRow, error) {
	v := []EventRow{}
	e := eventQuery(r.db.WithContext(ctx)).Joins("JOIN rate_card rc ON rc.rate_card_id=l.rate_card_id").Where("rc.billing_contract_id=? AND s.code='PENDING' AND e.business_date BETWEEN ? AND ?", contract, from, until).Order("e.business_date,e.billable_event_id").Find(&v).Error
	return v, Error(e)
}
func (r *BillableEventRepository) SetStatus(ctx context.Context, id, status string, notes *string) error {
	return Error(r.db.WithContext(ctx).Model(&model.BillableEvent{}).Where("billable_event_id=?", id).Updates(map[string]interface{}{"status_id": status, "notes": notes}).Error)
}
func (r *BillableEventRepository) ResetRunEvents(ctx context.Context, runID, status string) error {
	return Error(r.db.WithContext(ctx).Exec("UPDATE billable_event SET status_id=? WHERE billable_event_id IN (SELECT billable_event_id FROM billing_charge WHERE billing_run_id=?)", status, runID).Error)
}
