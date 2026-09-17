package inbound

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/inbound"
	mastermodel "wms-api/models/master"
)

type PutawayTaskRow struct {
	model.PutawayTask
	TaskStatusCode, TaskPriorityCode, ItemCode, LotNumber string
	SourceLocationCode, TargetLocationCode                string
	BaseUOMCode                                           string
	AssignedDisplayName                                   *string
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
	return db.Table("putaway_task task").Select(`task.*,status.code task_status_code,priority.code task_priority_code,item.code item_code,COALESCE(lot.lot_number,'') lot_number,source.code source_location_code,target.code target_location_code,uom.code base_uom_code,assignee.display_name assigned_display_name`).
		Joins("JOIN task_status status ON status.task_status_id=task.task_status_id").Joins("JOIN task_priority priority ON priority.task_priority_id=task.task_priority_id").Joins("JOIN item ON item.item_id=task.item_id").Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=task.lot_id").Joins("JOIN warehouse_location source ON source.location_id=task.source_location_id").Joins("JOIN warehouse_location target ON target.location_id=task.target_location_id").Joins("JOIN uom ON uom.uom_id=task.uom_id").Joins("LEFT JOIN app_account assignee ON assignee.account_id=task.assigned_to")
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

func (r *PutawayTaskRepository) Cancel(ctx context.Context, id, statusID, inspectionID, actor string, expectedVersion int64) error {
	result := r.db.WithContext(ctx).Model(&model.PutawayTask{}).Where("putaway_task_id=? AND version_no=?", id, expectedVersion).Updates(map[string]interface{}{"task_status_id": statusID, "replacement_inspection_id": inspectionID, "updated_at": gorm.Expr("clock_timestamp()"), "updated_by": actor, "version_no": gorm.Expr("version_no+1")})
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

func (r *PutawayTaskRepository) AccountCanPutaway(ctx context.Context, accountID, ownerID, warehouseID string) (bool, error) {
	var count int64
	err := putawayAccountQuery(r.db.WithContext(ctx), ownerID, warehouseID).Where("account.account_id=?", accountID).Count(&count).Error
	return count == 1, Error(err)
}

func putawayAccountQuery(db *gorm.DB, ownerID, warehouseID string) *gorm.DB {
	return db.Table("app_account account").Joins("JOIN account_status status ON status.account_status_id=account.account_status_id").Where("status.allows_login AND (account.locked_until IS NULL OR account.locked_until<=clock_timestamp())").
		Where(`EXISTS (
			SELECT 1 FROM app_permission permission
			WHERE permission.is_active AND permission.code IN ('INBOUND.PUTAWAY','*')
			AND (permission.code='*' OR (
				EXISTS (SELECT 1 FROM account_owner_access oa WHERE oa.account_id=account.account_id AND oa.owner_id=?)
				AND EXISTS (SELECT 1 FROM account_warehouse_access wa WHERE wa.account_id=account.account_id AND wa.warehouse_id=?)
			)) AND (
				EXISTS (SELECT 1 FROM account_permission direct WHERE direct.account_id=account.account_id AND direct.permission_id=permission.permission_id)
				OR EXISTS (SELECT 1 FROM account_role assignment
					JOIN app_role role ON role.role_id=assignment.role_id AND role.is_active
					JOIN role_permission role_grant ON role_grant.role_id=role.role_id
					WHERE assignment.account_id=account.account_id AND role_grant.permission_id=permission.permission_id)
			)
		)`, ownerID, warehouseID)
}

type PutawayAssigneeRow struct{ AccountID, Username, DisplayName string }

func (r *PutawayTaskRepository) ListAssignees(ctx context.Context, ownerID, warehouseID, search string, page, size int) ([]PutawayAssigneeRow, int64, error) {
	query := putawayAccountQuery(r.db.WithContext(ctx), ownerID, warehouseID)
	if search != "" {
		query = query.Where("account.username ILIKE ? OR account.display_name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, Error(err)
	}
	rows := make([]PutawayAssigneeRow, 0)
	err := query.Select("account.account_id,account.username,account.display_name").Order("account.display_name,account.account_id").Limit(size).Offset((page - 1) * size).Find(&rows).Error
	return rows, total, Error(err)
}

type PutawayTargetRow struct{ LocationID, Code, ZoneCode, LocationTypeCode string }

func (r *PutawayTaskRepository) ListTargets(ctx context.Context, warehouseID string, item mastermodel.Item, rules []mastermodel.PutawayStrategyRule, search string, page, size int) ([]PutawayTargetRow, int64, error) {
	query := r.db.WithContext(ctx).Table("warehouse_location location").Joins("JOIN location_type kind ON kind.location_type_id=location.location_type_id").Joins("JOIN warehouse_zone zone ON zone.zone_id=location.zone_id").Where("location.warehouse_id=? AND location.is_active AND NOT location.is_locked AND kind.is_active AND kind.allows_storage", warehouseID)
	conditions := make([]string, 0)
	args := make([]interface{}, 0)
	for _, rule := range rules {
		if rule.CategoryID != nil && (item.CategoryID == nil || *rule.CategoryID != *item.CategoryID) {
			continue
		}
		parts := []string{"TRUE"}
		if rule.LocationTypeID != nil {
			parts = append(parts, "location.location_type_id=?")
			args = append(args, *rule.LocationTypeID)
		}
		if rule.ZoneID != nil {
			parts = append(parts, "location.zone_id=?")
			args = append(args, *rule.ZoneID)
		}
		conditions = append(conditions, "("+strings.Join(parts, " AND ")+")")
	}
	if len(conditions) == 0 {
		query = query.Where("FALSE")
	} else {
		query = query.Where("("+strings.Join(conditions, " OR ")+")", args...)
	}
	if search != "" {
		query = query.Where("location.code ILIKE ? OR zone.code ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, Error(err)
	}
	rows := make([]PutawayTargetRow, 0)
	err := query.Select("location.location_id,location.code,zone.code zone_code,kind.code location_type_code").Order("location.code,location.location_id").Limit(size).Offset((page - 1) * size).Find(&rows).Error
	return rows, total, Error(err)
}
