package inbound

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/inbound"
)

type PutawayTaskRow struct {
	model.PutawayTask
	TaskStatusCode, TaskPriorityCode, ItemCode, LotNumber string
	SourceLocationCode, TargetLocationCode                string
}

type PutawayTaskRepository struct{ db *gorm.DB }

func NewPutawayTaskRepository(db *gorm.DB) *PutawayTaskRepository {
	return &PutawayTaskRepository{db: db}
}
func (r *PutawayTaskRepository) Create(ctx context.Context, value *model.PutawayTask) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}
func (r *PutawayTaskRepository) Lock(ctx context.Context, id string) (model.PutawayTask, error) {
	var value model.PutawayTask
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("putaway_task_id=?", id).Take(&value).Error
	return value, Error(err)
}
func putawayTaskQuery(db *gorm.DB) *gorm.DB {
	return db.Table("putaway_task task").Select(`task.*,status.code task_status_code,priority.code task_priority_code,item.code item_code,COALESCE(lot.lot_number,'') lot_number,source.code source_location_code,target.code target_location_code`).
		Joins("JOIN task_status status ON status.task_status_id=task.task_status_id").Joins("JOIN task_priority priority ON priority.task_priority_id=task.task_priority_id").Joins("JOIN item ON item.item_id=task.item_id").Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=task.lot_id").Joins("JOIN warehouse_location source ON source.location_id=task.source_location_id").Joins("JOIN warehouse_location target ON target.location_id=task.target_location_id")
}
func (r *PutawayTaskRepository) Get(ctx context.Context, id string) (PutawayTaskRow, error) {
	var value PutawayTaskRow
	err := putawayTaskQuery(r.db.WithContext(ctx)).Where("task.putaway_task_id=?", id).Take(&value).Error
	return value, Error(err)
}
func (r *PutawayTaskRepository) GetByInspection(ctx context.Context, inspectionID string) (PutawayTaskRow, error) {
	var value PutawayTaskRow
	err := putawayTaskQuery(r.db.WithContext(ctx)).Where("task.inspection_id=?", inspectionID).Take(&value).Error
	return value, Error(err)
}
func (r *PutawayTaskRepository) List(ctx context.Context, filter ListFilter) ([]PutawayTaskRow, int64, error) {
	query := putawayTaskQuery(r.db.WithContext(ctx)).Where("task.owner_id=?", filter.OwnerID)
	if filter.WarehouseID != "" {
		query = query.Where("task.warehouse_id=?", filter.WarehouseID)
	}
	if filter.StatusCode != "" {
		query = query.Where("status.code=?", filter.StatusCode)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("task.putaway_task_id ILIKE ? OR item.code ILIKE ? OR source.code ILIKE ? OR target.code ILIKE ?", like, like, like, like)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, Error(err)
	}
	rows := make([]PutawayTaskRow, 0)
	err := query.Order("task.created_at DESC,task.putaway_task_id DESC").Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize).Find(&rows).Error
	return rows, total, Error(err)
}
func (r *PutawayTaskRepository) Start(ctx context.Context, id, statusID, actor string, expectedVersion int64) error {
	result := r.db.WithContext(ctx).Model(&model.PutawayTask{}).Where("putaway_task_id=? AND version_no=?", id, expectedVersion).Updates(map[string]interface{}{"task_status_id": statusID, "assigned_to": actor, "started_at": time.Now(), "updated_at": gorm.Expr("clock_timestamp()"), "updated_by": actor, "version_no": gorm.Expr("version_no+1")})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}
func (r *PutawayTaskRepository) Complete(ctx context.Context, id, statusID, movementID, balanceID, qty, actor string, expectedVersion int64) error {
	result := r.db.WithContext(ctx).Model(&model.PutawayTask{}).Where("putaway_task_id=? AND version_no=?", id, expectedVersion).Updates(map[string]interface{}{"task_status_id": statusID, "completed_qty": qty, "completed_at": time.Now(), "inventory_movement_id": movementID, "resulting_balance_id": balanceID, "updated_at": gorm.Expr("clock_timestamp()"), "updated_by": actor, "version_no": gorm.Expr("version_no+1")})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}

func (r *PutawayTaskRepository) Assign(ctx context.Context, id, statusID, accountID, actor string, expectedVersion int64) error {
	result := r.db.WithContext(ctx).Model(&model.PutawayTask{}).Where("putaway_task_id=? AND version_no=?", id, expectedVersion).Updates(map[string]interface{}{"task_status_id": statusID, "assigned_to": accountID, "updated_at": gorm.Expr("clock_timestamp()"), "updated_by": actor, "version_no": gorm.Expr("version_no+1")})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}

func (r *PutawayTaskRepository) Retarget(ctx context.Context, id, targetID, actor string, expectedVersion int64) error {
	result := r.db.WithContext(ctx).Model(&model.PutawayTask{}).Where("putaway_task_id=? AND version_no=?", id, expectedVersion).Updates(map[string]interface{}{"target_location_id": targetID, "updated_at": gorm.Expr("clock_timestamp()"), "updated_by": actor, "version_no": gorm.Expr("version_no+1")})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}

func (r *PutawayTaskRepository) Cancel(ctx context.Context, id, statusID, actor string, expectedVersion int64) error {
	result := r.db.WithContext(ctx).Model(&model.PutawayTask{}).Where("putaway_task_id=? AND version_no=?", id, expectedVersion).Updates(map[string]interface{}{"task_status_id": statusID, "updated_at": gorm.Expr("clock_timestamp()"), "updated_by": actor, "version_no": gorm.Expr("version_no+1")})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}

func (r *PutawayTaskRepository) Reverse(ctx context.Context, id, statusID, movementID, inspectionID, actor, reason string, expectedVersion int64) error {
	result := r.db.WithContext(ctx).Model(&model.PutawayTask{}).Where("putaway_task_id=? AND version_no=?", id, expectedVersion).Updates(map[string]interface{}{"task_status_id": statusID, "reversal_movement_id": movementID, "replacement_inspection_id": inspectionID, "reversed_at": time.Now(), "reversed_by": actor, "reversal_reason": reason, "updated_at": gorm.Expr("clock_timestamp()"), "updated_by": actor, "version_no": gorm.Expr("version_no+1")})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}

func (r *PutawayTaskRepository) AccountCanAccess(ctx context.Context, accountID, ownerID, warehouseID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("app_account account").Joins("JOIN account_status status ON status.account_status_id=account.account_status_id").Where("account.account_id=? AND status.allows_login AND (account.locked_until IS NULL OR account.locked_until<=clock_timestamp())", accountID).
		Where(`EXISTS (SELECT 1 FROM account_owner_access owner_access WHERE owner_access.account_id=account.account_id AND owner_access.owner_id=?)`, ownerID).
		Where(`EXISTS (SELECT 1 FROM account_warehouse_access warehouse_access WHERE warehouse_access.account_id=account.account_id AND warehouse_access.warehouse_id=?)`, warehouseID).
		Count(&count).Error
	return count == 1, Error(err)
}
