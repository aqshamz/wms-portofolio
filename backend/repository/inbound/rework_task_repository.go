package inbound

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/inbound"
)

type ReworkTaskRow struct {
	model.ReworkTask
	TaskStatusCode, TaskPriorityCode, QuarantineCaseID, OwnerID, WarehouseID, ItemID, ItemCode string
}

type ReworkTaskRepository struct{ db *gorm.DB }

func NewReworkTaskRepository(db *gorm.DB) *ReworkTaskRepository { return &ReworkTaskRepository{db: db} }

func (r *ReworkTaskRepository) Create(ctx context.Context, value *model.ReworkTask) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}

func (r *ReworkTaskRepository) Lock(ctx context.Context, id string) (model.ReworkTask, error) {
	var value model.ReworkTask
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("rework_task_id=?", id).Take(&value).Error
	return value, Error(err)
}

func reworkTaskQuery(db *gorm.DB) *gorm.DB {
	return db.Table("rework_task task").Select(`task.*,status.code task_status_code,priority.code task_priority_code,disposition.quarantine_case_id,quarantine.owner_id,quarantine.warehouse_id,batch.item_id,item.code item_code`).
		Joins("JOIN task_status status ON status.task_status_id=task.task_status_id").
		Joins("JOIN task_priority priority ON priority.task_priority_id=task.task_priority_id").
		Joins("JOIN quarantine_disposition disposition ON disposition.quarantine_disposition_id=task.quarantine_disposition_id").
		Joins("JOIN quarantine_case quarantine ON quarantine.quarantine_case_id=disposition.quarantine_case_id").
		Joins("JOIN receipt_inventory batch ON batch.receipt_inventory_id=quarantine.receipt_inventory_id").
		Joins("JOIN item ON item.item_id=batch.item_id")
}

func (r *ReworkTaskRepository) Get(ctx context.Context, id string) (ReworkTaskRow, error) {
	var value ReworkTaskRow
	err := reworkTaskQuery(r.db.WithContext(ctx)).Where("task.rework_task_id=?", id).Take(&value).Error
	return value, Error(err)
}

func (r *ReworkTaskRepository) GetByDisposition(ctx context.Context, dispositionID string) (ReworkTaskRow, error) {
	var value ReworkTaskRow
	err := reworkTaskQuery(r.db.WithContext(ctx)).Where("task.quarantine_disposition_id=?", dispositionID).Take(&value).Error
	return value, Error(err)
}

func (r *ReworkTaskRepository) List(ctx context.Context, filter ListFilter) ([]ReworkTaskRow, int64, error) {
	query := reworkTaskQuery(r.db.WithContext(ctx)).Where("quarantine.owner_id=?", filter.OwnerID)
	if filter.WarehouseID != "" {
		query = query.Where("quarantine.warehouse_id=?", filter.WarehouseID)
	}
	if filter.StatusCode != "" {
		query = query.Where("status.code=?", filter.StatusCode)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("task.rework_task_id ILIKE ? OR disposition.quarantine_case_id ILIKE ? OR item.code ILIKE ?", like, like, like)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, Error(err)
	}
	rows := make([]ReworkTaskRow, 0)
	err := query.Order("task.created_at DESC,task.rework_task_id DESC").Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize).Find(&rows).Error
	return rows, total, Error(err)
}

func (r *ReworkTaskRepository) Start(ctx context.Context, id, statusID, actor string, expectedVersion int64) error {
	result := r.db.WithContext(ctx).Model(&model.ReworkTask{}).Where("rework_task_id=? AND version_no=?", id, expectedVersion).Updates(map[string]interface{}{
		"task_status_id": statusID, "assigned_to": actor, "started_at": time.Now(), "updated_at": gorm.Expr("clock_timestamp()"), "updated_by": actor, "version_no": gorm.Expr("version_no+1"),
	})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}

func (r *ReworkTaskRepository) Complete(ctx context.Context, id, statusID, qty, inspectionID, actor string, expectedVersion int64, notes *string) error {
	changes := map[string]interface{}{"task_status_id": statusID, "completed_qty": qty, "completed_at": time.Now(), "reinspection_id": inspectionID, "updated_at": gorm.Expr("clock_timestamp()"), "updated_by": actor, "version_no": gorm.Expr("version_no+1")}
	if notes != nil {
		changes["result_notes"] = notes
	}
	result := r.db.WithContext(ctx).Model(&model.ReworkTask{}).Where("rework_task_id=? AND version_no=?", id, expectedVersion).Updates(changes)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}
