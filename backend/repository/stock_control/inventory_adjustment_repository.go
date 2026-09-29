package stockcontrol

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	mastermodel "wms-api/models/master"
	model "wms-api/models/stock_control"
	inventoryrepo "wms-api/repository/inventory"
)

type AdjustmentFilter struct {
	OwnerID, WarehouseID, StatusCode, Search string
	Page, PageSize                           int
}

type InventoryAdjustmentRow struct {
	model.InventoryAdjustment
	StatusCode, OwnerCode, OwnerName, WarehouseCode, WarehouseName       string
	ReasonCode, ReasonName, CreatedByUsername, CreatedByDisplayName      string
	CancelledByDisplayName                                               *string
	TotalLines, PendingLines, PostedLines, RejectedLines, CancelledLines int64
}

type InventoryAdjustmentLineRow struct {
	model.InventoryAdjustmentLine
	ItemCode, ItemName, LocationCode, InventoryStatusCode, UOMCode string
	LotNumber, SerialNumber, HandlingUnitBarcode                   *string
	ApprovedByDisplayName, RejectedByDisplayName                   *string
	CurrentBalanceVersionNo                                        *int64
}

type InventoryAdjustmentRepository struct{ db *gorm.DB }

func NewInventoryAdjustmentRepository(db *gorm.DB) *InventoryAdjustmentRepository {
	return &InventoryAdjustmentRepository{db: db}
}

func adjustmentQuery(db *gorm.DB) *gorm.DB {
	return db.Table("inventory_adjustment_request adjustment").Select(`adjustment.*,
		status.code status_code,owner.code owner_code,owner.name owner_name,
		warehouse.code warehouse_code,warehouse.name warehouse_name,
		reason.code reason_code,reason.name reason_name,
		creator.username created_by_username,creator.display_name created_by_display_name,
		canceller.display_name cancelled_by_display_name,
		(SELECT count(*) FROM inventory_adjustment_request_line l WHERE l.inventory_adjustment_request_id=adjustment.inventory_adjustment_request_id) total_lines,
		(SELECT count(*) FROM inventory_adjustment_request_line l WHERE l.inventory_adjustment_request_id=adjustment.inventory_adjustment_request_id AND l.decision_code='PENDING') pending_lines,
		(SELECT count(*) FROM inventory_adjustment_request_line l WHERE l.inventory_adjustment_request_id=adjustment.inventory_adjustment_request_id AND l.decision_code='POSTED') posted_lines,
		(SELECT count(*) FROM inventory_adjustment_request_line l WHERE l.inventory_adjustment_request_id=adjustment.inventory_adjustment_request_id AND l.decision_code='REJECTED') rejected_lines,
		(SELECT count(*) FROM inventory_adjustment_request_line l WHERE l.inventory_adjustment_request_id=adjustment.inventory_adjustment_request_id AND l.decision_code='CANCELLED') cancelled_lines`).
		Joins("JOIN document_status status ON status.status_id=adjustment.status_id").
		Joins("JOIN organization owner ON owner.organization_id=adjustment.owner_id").
		Joins("JOIN warehouse ON warehouse.warehouse_id=adjustment.warehouse_id").
		Joins("JOIN reason_code reason ON reason.reason_code_id=adjustment.reason_code_id").
		Joins("JOIN app_account creator ON creator.account_id=adjustment.created_by").
		Joins("LEFT JOIN app_account canceller ON canceller.account_id=adjustment.cancelled_by")
}

func lineQuery(db *gorm.DB) *gorm.DB {
	return db.Table("inventory_adjustment_request_line line").Select(`line.*,
		item.code item_code,item.name item_name,location.code location_code,
		inventory_status.code inventory_status_code,uom.code uom_code,
		lot.lot_number,serial.serial_no serial_number,hu.barcode handling_unit_barcode,
		approver.display_name approved_by_display_name,rejecter.display_name rejected_by_display_name,
		balance.version_no current_balance_version_no`).
		Joins("JOIN item ON item.item_id=line.item_id").
		Joins("JOIN warehouse_location location ON location.location_id=line.location_id").
		Joins("JOIN inventory_status ON inventory_status.inventory_status_id=line.inventory_status_id").
		Joins("JOIN uom ON uom.uom_id=line.uom_id").
		Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=line.lot_id").
		Joins("LEFT JOIN serial_number serial ON serial.serial_id=line.serial_id").
		Joins("LEFT JOIN handling_unit hu ON hu.handling_unit_id=line.handling_unit_id").
		Joins("LEFT JOIN app_account approver ON approver.account_id=line.approved_by").
		Joins("LEFT JOIN app_account rejecter ON rejecter.account_id=line.rejected_by").
		Joins("LEFT JOIN inventory_balance balance ON balance.balance_id=line.balance_id")
}

func (r *InventoryAdjustmentRepository) Create(ctx context.Context, header *model.InventoryAdjustment, lines []model.InventoryAdjustmentLine) error {
	if err := r.db.WithContext(ctx).Create(header).Error; err != nil {
		return inventoryrepo.Error(err)
	}
	return inventoryrepo.Error(r.db.WithContext(ctx).Create(&lines).Error)
}
func (r *InventoryAdjustmentRepository) Get(ctx context.Context, id string) (InventoryAdjustmentRow, error) {
	var value InventoryAdjustmentRow
	err := adjustmentQuery(r.db.WithContext(ctx)).Where("adjustment.inventory_adjustment_request_id=?", id).Take(&value).Error
	return value, inventoryrepo.Error(err)
}
func (r *InventoryAdjustmentRepository) Lines(ctx context.Context, id string) ([]InventoryAdjustmentLineRow, error) {
	var rows []InventoryAdjustmentLineRow
	err := lineQuery(r.db.WithContext(ctx)).Where("line.inventory_adjustment_request_id=?", id).Order("line.line_no").Find(&rows).Error
	return rows, inventoryrepo.Error(err)
}
func (r *InventoryAdjustmentRepository) Lock(ctx context.Context, id string) (model.InventoryAdjustment, error) {
	var value model.InventoryAdjustment
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("inventory_adjustment_request_id=?", id).Take(&value).Error
	return value, inventoryrepo.Error(err)
}
func (r *InventoryAdjustmentRepository) LockLines(ctx context.Context, id string, lineIDs []string) ([]model.InventoryAdjustmentLine, error) {
	var values []model.InventoryAdjustmentLine
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("inventory_adjustment_request_id=? AND inventory_adjustment_request_line_id IN ?", id, lineIDs).
		Order("line_no").Find(&values).Error
	return values, inventoryrepo.Error(err)
}
func (r *InventoryAdjustmentRepository) List(ctx context.Context, filter AdjustmentFilter) ([]InventoryAdjustmentRow, int64, error) {
	query := adjustmentQuery(r.db.WithContext(ctx)).Where("adjustment.owner_id=? AND adjustment.warehouse_id=?", filter.OwnerID, filter.WarehouseID)
	if filter.StatusCode != "" {
		query = query.Where("status.code=?", filter.StatusCode)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where(`adjustment.inventory_adjustment_request_id ILIKE ? OR reason.code ILIKE ? OR EXISTS (
			SELECT 1 FROM inventory_adjustment_request_line search_line
			JOIN item search_item ON search_item.item_id=search_line.item_id
			JOIN warehouse_location search_location ON search_location.location_id=search_line.location_id
			WHERE search_line.inventory_adjustment_request_id=adjustment.inventory_adjustment_request_id
			AND (search_item.code ILIKE ? OR search_item.name ILIKE ? OR search_location.code ILIKE ?))`, like, like, like, like, like)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, inventoryrepo.Error(err)
	}
	var rows []InventoryAdjustmentRow
	err := query.Order("adjustment.created_at DESC,adjustment.inventory_adjustment_request_id DESC").Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize).Find(&rows).Error
	return rows, total, inventoryrepo.Error(err)
}
func (r *InventoryAdjustmentRepository) DocumentType(ctx context.Context) (mastermodel.DocumentType, error) {
	var value mastermodel.DocumentType
	err := r.db.WithContext(ctx).Where("code=? AND is_active", "INVENTORY_ADJUSTMENT").Take(&value).Error
	return value, inventoryrepo.Error(err)
}
func (r *InventoryAdjustmentRepository) Status(ctx context.Context, typeID, code string) (mastermodel.DocumentStatus, error) {
	var value mastermodel.DocumentStatus
	err := r.db.WithContext(ctx).Where("document_type_id=? AND code=? AND is_active", typeID, code).Take(&value).Error
	return value, inventoryrepo.Error(err)
}
func (r *InventoryAdjustmentRepository) Update(ctx context.Context, id string, expectedVersion int64, values map[string]any) error {
	values["updated_at"], values["version_no"] = gorm.Expr("clock_timestamp()"), gorm.Expr("version_no+1")
	result := r.db.WithContext(ctx).Model(&model.InventoryAdjustment{}).Where("inventory_adjustment_request_id=? AND version_no=?", id, expectedVersion).Updates(values)
	if result.Error != nil {
		return inventoryrepo.Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return inventoryrepo.ErrConflict
	}
	return nil
}
func (r *InventoryAdjustmentRepository) PostLine(ctx context.Context, lineID, movementID, balanceID, actor string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&model.InventoryAdjustmentLine{}).
		Where("inventory_adjustment_request_line_id=? AND decision_code='PENDING'", lineID).
		Updates(map[string]any{"decision_code": "POSTED", "approved_at": now, "approved_by": actor, "inventory_movement_id": movementID, "resulting_balance_id": balanceID, "updated_at": now, "updated_by": actor, "version_no": gorm.Expr("version_no+1")})
	if result.Error != nil {
		return inventoryrepo.Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return inventoryrepo.ErrConflict
	}
	return nil
}
func (r *InventoryAdjustmentRepository) DecideLines(ctx context.Context, id string, lineIDs []string, decision, actor, reason string) error {
	now := time.Now()
	values := map[string]any{"decision_code": decision, "updated_at": now, "updated_by": actor, "version_no": gorm.Expr("version_no+1")}
	if decision == "REJECTED" {
		values["rejected_at"], values["rejected_by"], values["rejection_reason"] = now, actor, reason
	} else {
		values["cancelled_at"], values["cancelled_by"], values["cancellation_reason"] = now, actor, reason
	}
	result := r.db.WithContext(ctx).Model(&model.InventoryAdjustmentLine{}).Where("inventory_adjustment_request_id=? AND inventory_adjustment_request_line_id IN ? AND decision_code='PENDING'", id, lineIDs).Updates(values)
	if result.Error != nil {
		return inventoryrepo.Error(result.Error)
	}
	if result.RowsAffected != int64(len(lineIDs)) {
		return inventoryrepo.ErrConflict
	}
	return nil
}
func (r *InventoryAdjustmentRepository) PendingLineIDs(ctx context.Context, id string) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).Model(&model.InventoryAdjustmentLine{}).Where("inventory_adjustment_request_id=? AND decision_code='PENDING'", id).Pluck("inventory_adjustment_request_line_id", &ids).Error
	return ids, inventoryrepo.Error(err)
}
func (r *InventoryAdjustmentRepository) Counts(ctx context.Context, id string) (pending, posted int64, err error) {
	var rows []struct {
		DecisionCode string
		Count        int64
	}
	err = r.db.WithContext(ctx).Model(&model.InventoryAdjustmentLine{}).Select("decision_code,count(*) count").Where("inventory_adjustment_request_id=?", id).Group("decision_code").Scan(&rows).Error
	for _, row := range rows {
		if row.DecisionCode == "PENDING" {
			pending = row.Count
		}
		if row.DecisionCode == "POSTED" {
			posted = row.Count
		}
	}
	return pending, posted, inventoryrepo.Error(err)
}
