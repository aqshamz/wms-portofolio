package stockcontrol

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/stock_control"
	inventoryrepo "wms-api/repository/inventory"
)

type ReplenishmentFilter struct {
	OwnerID, WarehouseID, StatusCode, Search string
	Page, PageSize                           int
}

type ReplenishmentTaskRow struct {
	model.ReplenishmentTask
	TaskStatusCode, TaskPriorityCode, ItemCode, ItemName, UOMCode string
	SourceLocationCode, TargetLocationCode, InventoryStatusCode   string
	LotNumber, SerialNumber, HandlingUnitBarcode                  *string
	AssignedUsername, AssignedDisplayName                         *string
	SourceBalanceVersionNo                                        int64
}

type ReplenishmentAssigneeRow struct{ AccountID, Username, DisplayName string }
type ReplenishmentTargetRow struct{ LocationID, Code, ZoneCode, LocationTypeCode string }

type ReplenishmentTaskRepository struct{ db *gorm.DB }

func NewReplenishmentTaskRepository(db *gorm.DB) *ReplenishmentTaskRepository {
	return &ReplenishmentTaskRepository{db: db}
}

func replenishmentQuery(db *gorm.DB) *gorm.DB {
	return db.Table("replenishment_task task").Select(`task.*, status.code task_status_code,
		priority.code task_priority_code, item.code item_code, item.name item_name,
		uom.code uom_code, source.code source_location_code, target.code target_location_code,
		inventory_status.code inventory_status_code, lot.lot_number,
		serial.serial_no serial_number, hu.barcode handling_unit_barcode,
		assignee.username assigned_username, assignee.display_name assigned_display_name,
		balance.version_no source_balance_version_no`).
		Joins("JOIN task_status status ON status.task_status_id=task.task_status_id").
		Joins("JOIN task_priority priority ON priority.task_priority_id=task.task_priority_id").
		Joins("JOIN inventory_balance balance ON balance.balance_id=task.source_balance_id").
		Joins("JOIN item ON item.item_id=task.item_id").
		Joins("JOIN uom ON uom.uom_id=task.uom_id").
		Joins("JOIN warehouse_location source ON source.location_id=task.source_location_id").
		Joins("JOIN warehouse_location target ON target.location_id=task.target_location_id").
		Joins("JOIN inventory_status ON inventory_status.inventory_status_id=task.inventory_status_id").
		Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=task.lot_id").
		Joins("LEFT JOIN serial_number serial ON serial.serial_id=task.serial_id").
		Joins("LEFT JOIN handling_unit hu ON hu.handling_unit_id=task.handling_unit_id").
		Joins("LEFT JOIN app_account assignee ON assignee.account_id=task.assigned_to")
}

func (r *ReplenishmentTaskRepository) Create(ctx context.Context, task *model.ReplenishmentTask) error {
	return inventoryrepo.Error(r.db.WithContext(ctx).Create(task).Error)
}

func (r *ReplenishmentTaskRepository) Lock(ctx context.Context, id string) (model.ReplenishmentTask, error) {
	var task model.ReplenishmentTask
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("replenishment_task_id=?", id).Take(&task).Error
	return task, inventoryrepo.Error(err)
}

func (r *ReplenishmentTaskRepository) Get(ctx context.Context, id string) (ReplenishmentTaskRow, error) {
	var row ReplenishmentTaskRow
	err := replenishmentQuery(r.db.WithContext(ctx)).Where("task.replenishment_task_id=?", id).Take(&row).Error
	return row, inventoryrepo.Error(err)
}

func (r *ReplenishmentTaskRepository) List(ctx context.Context, filter ReplenishmentFilter) ([]ReplenishmentTaskRow, int64, error) {
	query := replenishmentQuery(r.db.WithContext(ctx)).Where("task.owner_id=? AND task.warehouse_id=?", filter.OwnerID, filter.WarehouseID)
	if filter.StatusCode != "" {
		query = query.Where("status.code=?", filter.StatusCode)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("task.replenishment_task_id ILIKE ? OR item.code ILIKE ? OR item.name ILIKE ? OR source.code ILIKE ? OR target.code ILIKE ?", like, like, like, like, like)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, inventoryrepo.Error(err)
	}
	rows := make([]ReplenishmentTaskRow, 0)
	err := query.Order("task.created_at DESC,task.replenishment_task_id DESC").Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize).Find(&rows).Error
	return rows, total, inventoryrepo.Error(err)
}

func (r *ReplenishmentTaskRepository) Reference(ctx context.Context, table, code string, out any) error {
	return inventoryrepo.Error(r.db.WithContext(ctx).Table(table).Where("code=? AND is_active", code).Take(out).Error)
}

func (r *ReplenishmentTaskRepository) SourceEligible(ctx context.Context, balanceID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("inventory_balance balance").
		Joins("JOIN warehouse_location location ON location.location_id=balance.location_id AND location.is_active AND NOT location.is_locked AND NOT location.is_pick_face").
		Joins("JOIN warehouse_zone zone ON zone.zone_id=location.zone_id AND zone.is_active").
		Joins("JOIN location_type kind ON kind.location_type_id=location.location_type_id AND kind.is_active AND kind.allows_storage").
		Joins("JOIN inventory_status status ON status.inventory_status_id=balance.inventory_status_id AND status.is_active AND status.is_allocatable AND status.is_pickable").
		Where("balance.balance_id=?", balanceID).Count(&count).Error
	return count == 1, inventoryrepo.Error(err)
}

func (r *ReplenishmentTaskRepository) TargetEligible(ctx context.Context, warehouseID, locationID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("warehouse_location location").
		Joins("JOIN warehouse_zone zone ON zone.zone_id=location.zone_id AND zone.is_active").
		Joins("JOIN location_type kind ON kind.location_type_id=location.location_type_id AND kind.is_active AND kind.allows_storage AND kind.allows_picking").
		Where("location.location_id=? AND location.warehouse_id=? AND location.is_active AND NOT location.is_locked AND location.is_pick_face", locationID, warehouseID).
		Count(&count).Error
	return count == 1, inventoryrepo.Error(err)
}

func (r *ReplenishmentTaskRepository) ListTargets(ctx context.Context, warehouseID, search string, page, size int) ([]ReplenishmentTargetRow, int64, error) {
	query := r.db.WithContext(ctx).Table("warehouse_location location").
		Joins("JOIN warehouse_zone zone ON zone.zone_id=location.zone_id AND zone.is_active").
		Joins("JOIN location_type kind ON kind.location_type_id=location.location_type_id AND kind.is_active AND kind.allows_storage AND kind.allows_picking").
		Where("location.warehouse_id=? AND location.is_active AND NOT location.is_locked AND location.is_pick_face", warehouseID)
	if search != "" {
		query = query.Where("location.code ILIKE ? OR zone.code ILIKE ? OR kind.code ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, inventoryrepo.Error(err)
	}
	rows := make([]ReplenishmentTargetRow, 0)
	err := query.Select("location.location_id,location.code,zone.code zone_code,kind.code location_type_code").Order("location.pick_sequence,location.code").Limit(size).Offset((page - 1) * size).Find(&rows).Error
	return rows, total, inventoryrepo.Error(err)
}

func replenishmentAccountQuery(db *gorm.DB, ownerID, warehouseID string) *gorm.DB {
	return db.Table("app_account account").
		Joins("JOIN account_status status ON status.account_status_id=account.account_status_id").
		Where("status.allows_login AND (account.locked_until IS NULL OR account.locked_until<=clock_timestamp())").
		Where(`EXISTS (
			SELECT 1 FROM app_permission permission
			WHERE permission.is_active AND permission.code IN ('INVENTORY.MOVE','*')
			AND (permission.code='*' OR (
				EXISTS (SELECT 1 FROM account_owner_access oa WHERE oa.account_id=account.account_id AND oa.owner_id=?)
				AND EXISTS (SELECT 1 FROM account_warehouse_access wa WHERE wa.account_id=account.account_id AND wa.warehouse_id=?)))
			AND (EXISTS (SELECT 1 FROM account_permission direct WHERE direct.account_id=account.account_id AND direct.permission_id=permission.permission_id)
				OR EXISTS (SELECT 1 FROM account_role assignment JOIN app_role role ON role.role_id=assignment.role_id AND role.is_active
					JOIN role_permission grant_row ON grant_row.role_id=role.role_id
					WHERE assignment.account_id=account.account_id AND grant_row.permission_id=permission.permission_id)))`, ownerID, warehouseID)
}

func (r *ReplenishmentTaskRepository) AccountCanExecute(ctx context.Context, accountID, ownerID, warehouseID string) (bool, error) {
	var count int64
	err := replenishmentAccountQuery(r.db.WithContext(ctx), ownerID, warehouseID).Where("account.account_id=?", accountID).Count(&count).Error
	return count == 1, inventoryrepo.Error(err)
}

func (r *ReplenishmentTaskRepository) ListAssignees(ctx context.Context, ownerID, warehouseID, search string, page, size int) ([]ReplenishmentAssigneeRow, int64, error) {
	query := replenishmentAccountQuery(r.db.WithContext(ctx), ownerID, warehouseID)
	if search != "" {
		query = query.Where("account.username ILIKE ? OR account.display_name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, inventoryrepo.Error(err)
	}
	rows := make([]ReplenishmentAssigneeRow, 0)
	err := query.Select("account.account_id,account.username,account.display_name").Order("account.display_name,account.account_id").Limit(size).Offset((page - 1) * size).Find(&rows).Error
	return rows, total, inventoryrepo.Error(err)
}

func (r *ReplenishmentTaskRepository) Update(ctx context.Context, id string, expectedVersion int64, values map[string]any) error {
	values["updated_at"] = gorm.Expr("clock_timestamp()")
	values["version_no"] = gorm.Expr("version_no+1")
	result := r.db.WithContext(ctx).Model(&model.ReplenishmentTask{}).Where("replenishment_task_id=? AND version_no=?", id, expectedVersion).Updates(values)
	if result.Error != nil {
		return inventoryrepo.Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return inventoryrepo.ErrConflict
	}
	return nil
}

func (r *ReplenishmentTaskRepository) Start(ctx context.Context, id, statusID, actor string, expectedVersion int64) error {
	return r.Update(ctx, id, expectedVersion, map[string]any{"task_status_id": statusID, "assigned_to": actor, "started_at": time.Now(), "updated_by": actor})
}

func (r *ReplenishmentTaskRepository) Assign(ctx context.Context, id, statusID, accountID, actor string, expectedVersion int64) error {
	return r.Update(ctx, id, expectedVersion, map[string]any{"task_status_id": statusID, "assigned_to": accountID, "updated_by": actor})
}

func (r *ReplenishmentTaskRepository) Complete(ctx context.Context, id, statusID, movementID, balanceID, qty, actor string, expectedVersion int64) error {
	return r.Update(ctx, id, expectedVersion, map[string]any{"task_status_id": statusID, "completed_qty": qty, "completed_at": time.Now(), "inventory_movement_id": movementID, "resulting_balance_id": balanceID, "updated_by": actor})
}

func (r *ReplenishmentTaskRepository) Cancel(ctx context.Context, id, statusID, actor, reason string, expectedVersion int64) error {
	return r.Update(ctx, id, expectedVersion, map[string]any{"task_status_id": statusID, "cancelled_at": time.Now(), "cancelled_by": actor, "cancellation_reason": reason, "updated_by": actor})
}
