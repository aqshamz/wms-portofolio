package outbound

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/outbound"
)

type PickTaskRow struct {
	model.PickTask
	TaskStatusCode, PriorityCode, OutboundID, ClientDeliveryOrderNo, ItemID, ItemCode, OwnerID, WarehouseID, WarehouseCode, SourceLocationCode, TargetLocationCode, LotNumber string
	BusinessDate                                                                                                                                                              string
	BalanceID                                                                                                                                                                 string
	BalanceVersion                                                                                                                                                            int64
}
type PickTaskRepository struct{ db *gorm.DB }

func NewPickTaskRepository(db *gorm.DB) *PickTaskRepository { return &PickTaskRepository{db: db} }
func (r *PickTaskRepository) CreateBatch(ctx context.Context, v []model.PickTask) error {
	return Error(r.db.WithContext(ctx).Create(&v).Error)
}
func pickTaskQuery(db *gorm.DB) *gorm.DB {
	return db.Table("pick_task p").Select("p.*,ts.code task_status_code,tp.code priority_code,o.outbound_id,o.client_delivery_order_no,o.owner_id,o.warehouse_id,o.business_date,wh.code warehouse_code,l.item_id,i.code item_code,src.code source_location_code,COALESCE(dst.code,'') target_location_code,COALESCE(lot.lot_number,'') lot_number,r.balance_id,b.version_no balance_version").Joins("JOIN task_status ts ON ts.task_status_id=p.task_status_id").Joins("JOIN task_priority tp ON tp.task_priority_id=p.task_priority_id").Joins("JOIN outbound_order_line l ON l.outbound_line_id=p.outbound_line_id").Joins("JOIN outbound_order o ON o.outbound_id=l.outbound_id").Joins("JOIN warehouse wh ON wh.warehouse_id=o.warehouse_id").Joins("JOIN item i ON i.item_id=l.item_id").Joins("JOIN inventory_reservation r ON r.reservation_id=p.reservation_id").Joins("JOIN inventory_balance b ON b.balance_id=r.balance_id").Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=b.lot_id").Joins("JOIN warehouse_location src ON src.location_id=p.source_location_id").Joins("LEFT JOIN warehouse_location dst ON dst.location_id=p.target_location_id")
}
func (r *PickTaskRepository) Get(ctx context.Context, id string) (PickTaskRow, error) {
	var v PickTaskRow
	err := pickTaskQuery(r.db.WithContext(ctx)).Where("p.pick_task_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *PickTaskRepository) GetByReservation(ctx context.Context, reservationID string) (PickTaskRow, error) {
	var v PickTaskRow
	err := pickTaskQuery(r.db.WithContext(ctx)).Where("p.reservation_id=?", reservationID).Take(&v).Error
	return v, Error(err)
}
func (r *PickTaskRepository) Lock(ctx context.Context, id string) (model.PickTask, error) {
	var v model.PickTask
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("pick_task_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *PickTaskRepository) List(ctx context.Context, f ListFilter, assignee string) ([]PickTaskRow, int64, error) {
	q := pickTaskQuery(r.db.WithContext(ctx)).Where("o.owner_id=?", f.OwnerID)
	if f.WarehouseID != "" {
		q = q.Where("o.warehouse_id=?", f.WarehouseID)
	}
	if f.StatusCode != "" {
		q = q.Where("ts.code=?", f.StatusCode)
	}
	if assignee != "" {
		q = q.Where("p.assigned_to=?", assignee)
	}
	if f.Search != "" {
		x := "%" + f.Search + "%"
		q = q.Where("p.pick_task_id ILIKE ? OR o.client_delivery_order_no ILIKE ? OR i.code ILIKE ?", x, x, x)
	}
	var n int64
	if err := q.Session(&gorm.Session{}).Count(&n).Error; err != nil {
		return nil, 0, Error(err)
	}
	var v []PickTaskRow
	err := q.Order("tp.priority_value DESC,p.created_at,p.pick_task_id").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&v).Error
	return v, n, Error(err)
}
func (r *PickTaskRepository) Start(ctx context.Context, id, statusID, actor string) error {
	res := r.db.WithContext(ctx).Model(&model.PickTask{}).Where("pick_task_id=? AND (assigned_to IS NULL OR assigned_to=?)", id, actor).Updates(map[string]interface{}{"task_status_id": statusID, "assigned_to": actor, "started_at": gorm.Expr("COALESCE(started_at,clock_timestamp())")})
	if res.Error != nil {
		return Error(res.Error)
	}
	if res.RowsAffected != 1 {
		return ErrConflict
	}
	return nil
}
func (r *PickTaskRepository) AddPicked(ctx context.Context, id, statusID, qty string, completed bool) error {
	v := map[string]interface{}{"picked_qty": gorm.Expr("picked_qty+?", qty), "task_status_id": statusID}
	if completed {
		v["completed_at"] = gorm.Expr("clock_timestamp()")
	}
	return Error(r.db.WithContext(ctx).Model(&model.PickTask{}).Where("pick_task_id=?", id).Updates(v).Error)
}
func (r *PickTaskRepository) CloseShort(ctx context.Context, id, statusID, reasonID, qty string, notes *string) error {
	return Error(r.db.WithContext(ctx).Model(&model.PickTask{}).Where("pick_task_id=?", id).Updates(map[string]interface{}{"short_qty": qty, "short_reason_code_id": reasonID, "result_notes": notes, "task_status_id": statusID, "completed_at": gorm.Expr("clock_timestamp()")}).Error)
}
func (r *PickTaskRepository) RemainingByWave(ctx context.Context, id string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("pick_task p").Joins("JOIN task_status s ON s.task_status_id=p.task_status_id").Where("p.wave_id=? AND NOT s.is_final", id).Count(&n).Error
	return n, Error(err)
}
func (r *PickTaskRepository) RemainingByOrderWave(ctx context.Context, outboundID, waveID string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("pick_task p").Joins("JOIN task_status s ON s.task_status_id=p.task_status_id").Joins("JOIN outbound_order_line l ON l.outbound_line_id=p.outbound_line_id").Where("l.outbound_id=? AND p.wave_id=? AND NOT s.is_final", outboundID, waveID).Count(&n).Error
	return n, Error(err)
}
func (r *PickTaskRepository) ExistsForReservation(ctx context.Context, id string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.PickTask{}).Where("reservation_id=?", id).Count(&n).Error
	return n > 0, Error(err)
}
