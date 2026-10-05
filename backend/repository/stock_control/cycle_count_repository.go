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

type CycleCountFilter struct {
	OwnerID, WarehouseID, StatusCode, Search string
	Page, PageSize                           int
}
type CycleCountRow struct {
	model.CycleCount
	StatusCode, OwnerCode, OwnerName, WarehouseCode, WarehouseName, CreatedByDisplayName string
	TotalLines, OpenLines, CountedLines, RecountLines, FinalLines                        int64
}
type CycleCountLineRow struct {
	model.CycleCountLine
	ItemCode, ItemName, LocationCode, InventoryStatusCode, UOMCode             string
	LotNumber, HandlingUnitBarcode, CountedByDisplayName, DecidedByDisplayName *string
	SerialControlled                                                           bool
	CurrentBalanceVersionNo                                                    *int64
}
type CycleCountRepository struct{ db *gorm.DB }

func NewCycleCountRepository(db *gorm.DB) *CycleCountRepository { return &CycleCountRepository{db: db} }
func cycleCountQuery(db *gorm.DB) *gorm.DB {
	return db.Table("cycle_count count").Select(`count.*,status.code status_code,owner.code owner_code,owner.name owner_name,
		warehouse.code warehouse_code,warehouse.name warehouse_name,creator.display_name created_by_display_name,
		(SELECT count(*) FROM cycle_count_line l WHERE l.cycle_count_id=count.cycle_count_id) total_lines,
		(SELECT count(*) FROM cycle_count_line l WHERE l.cycle_count_id=count.cycle_count_id AND l.decision_code='OPEN') open_lines,
		(SELECT count(*) FROM cycle_count_line l WHERE l.cycle_count_id=count.cycle_count_id AND l.decision_code='COUNTED') counted_lines,
		(SELECT count(*) FROM cycle_count_line l WHERE l.cycle_count_id=count.cycle_count_id AND l.requires_recount) recount_lines,
		(SELECT count(*) FROM cycle_count_line l WHERE l.cycle_count_id=count.cycle_count_id AND l.decision_code IN ('NO_VARIANCE','POSTED','REJECTED','CANCELLED')) final_lines`).
		Joins("JOIN document_status status ON status.status_id=count.status_id").
		Joins("JOIN organization owner ON owner.organization_id=count.owner_id").
		Joins("JOIN warehouse ON warehouse.warehouse_id=count.warehouse_id").
		Joins("JOIN app_account creator ON creator.account_id=count.created_by")
}
func cycleCountLineQuery(db *gorm.DB) *gorm.DB {
	return db.Table("cycle_count_line line").Select(`line.*,item.code item_code,item.name item_name,item.serial_controlled,
		location.code location_code,status.code inventory_status_code,uom.code uom_code,lot.lot_number,
		hu.barcode handling_unit_barcode,counter.display_name counted_by_display_name,
		decider.display_name decided_by_display_name,balance.version_no current_balance_version_no`).
		Joins("JOIN item ON item.item_id=line.item_id").Joins("JOIN warehouse_location location ON location.location_id=line.location_id").
		Joins("JOIN inventory_status status ON status.inventory_status_id=line.inventory_status_id").Joins("JOIN uom ON uom.uom_id=line.uom_id").
		Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=line.lot_id").Joins("LEFT JOIN handling_unit hu ON hu.handling_unit_id=line.handling_unit_id").
		Joins("LEFT JOIN app_account counter ON counter.account_id=line.counted_by").Joins("LEFT JOIN app_account decider ON decider.account_id=line.decided_by").
		Joins("LEFT JOIN inventory_balance balance ON balance.balance_id=line.balance_id")
}
func (r *CycleCountRepository) Create(ctx context.Context, header *model.CycleCount, lines []model.CycleCountLine) error {
	if err := r.db.WithContext(ctx).Create(header).Error; err != nil {
		return inventoryrepo.Error(err)
	}
	return inventoryrepo.Error(r.db.WithContext(ctx).Create(&lines).Error)
}
func (r *CycleCountRepository) Get(ctx context.Context, id string) (CycleCountRow, error) {
	var row CycleCountRow
	err := cycleCountQuery(r.db.WithContext(ctx)).Where("count.cycle_count_id=?", id).Take(&row).Error
	return row, inventoryrepo.Error(err)
}
func (r *CycleCountRepository) Lines(ctx context.Context, id string) ([]CycleCountLineRow, error) {
	var rows []CycleCountLineRow
	err := cycleCountLineQuery(r.db.WithContext(ctx)).Where("line.cycle_count_id=?", id).Order("line.line_no").Find(&rows).Error
	return rows, inventoryrepo.Error(err)
}
func (r *CycleCountRepository) List(ctx context.Context, f CycleCountFilter) ([]CycleCountRow, int64, error) {
	q := cycleCountQuery(r.db.WithContext(ctx)).Where("count.owner_id=? AND count.warehouse_id=?", f.OwnerID, f.WarehouseID)
	if f.StatusCode != "" {
		q = q.Where("status.code=?", f.StatusCode)
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where(`count.cycle_count_id ILIKE ? OR EXISTS(SELECT 1 FROM cycle_count_line sl JOIN item si ON si.item_id=sl.item_id JOIN warehouse_location wl ON wl.location_id=sl.location_id WHERE sl.cycle_count_id=count.cycle_count_id AND (si.code ILIKE ? OR si.name ILIKE ? OR wl.code ILIKE ?))`, like, like, like, like)
	}
	var total int64
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, inventoryrepo.Error(err)
	}
	var rows []CycleCountRow
	err := q.Order("count.created_at DESC,count.cycle_count_id DESC").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&rows).Error
	return rows, total, inventoryrepo.Error(err)
}
func (r *CycleCountRepository) Lock(ctx context.Context, id string) (model.CycleCount, error) {
	var row model.CycleCount
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("cycle_count_id=?", id).Take(&row).Error
	return row, inventoryrepo.Error(err)
}
func (r *CycleCountRepository) LockLines(ctx context.Context, id string, lineIDs []string) ([]model.CycleCountLine, error) {
	var rows []model.CycleCountLine
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("cycle_count_id=? AND cycle_count_line_id IN ?", id, lineIDs).Order("line_no").Find(&rows).Error
	return rows, inventoryrepo.Error(err)
}
func (r *CycleCountRepository) DocumentType(ctx context.Context) (mastermodel.DocumentType, error) {
	var row mastermodel.DocumentType
	err := r.db.WithContext(ctx).Where("code=? AND is_active", "STOCK_COUNT").Take(&row).Error
	return row, inventoryrepo.Error(err)
}
func (r *CycleCountRepository) Status(ctx context.Context, typeID, code string) (mastermodel.DocumentStatus, error) {
	var row mastermodel.DocumentStatus
	err := r.db.WithContext(ctx).Where("document_type_id=? AND code=? AND is_active", typeID, code).Take(&row).Error
	return row, inventoryrepo.Error(err)
}
func (r *CycleCountRepository) UpdateHeader(ctx context.Context, id string, version int64, values map[string]any) error {
	values["updated_at"] = gorm.Expr("clock_timestamp()")
	values["version_no"] = gorm.Expr("version_no+1")
	result := r.db.WithContext(ctx).Model(&model.CycleCount{}).Where("cycle_count_id=? AND version_no=?", id, version).Updates(values)
	if result.Error != nil {
		return inventoryrepo.Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return inventoryrepo.ErrConflict
	}
	return nil
}
func (r *CycleCountRepository) Record(ctx context.Context, line model.CycleCountLine, entry model.CycleCountEntry, actor string) error {
	if err := r.db.WithContext(ctx).Create(&entry).Error; err != nil {
		return inventoryrepo.Error(err)
	}
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&model.CycleCountLine{}).Where("cycle_count_line_id=? AND version_no=?", line.ID, line.VersionNo).Updates(map[string]any{"counted_qty": entry.CountedQty, "variance_qty": entry.VarianceQty, "count_attempts": entry.AttemptNo, "requires_recount": line.RequiresRecount, "decision_code": "COUNTED", "counted_at": now, "counted_by": actor, "count_notes": entry.Notes, "updated_at": now, "updated_by": actor, "version_no": gorm.Expr("version_no+1")})
	if result.Error != nil {
		return inventoryrepo.Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return inventoryrepo.ErrConflict
	}
	return nil
}
func (r *CycleCountRepository) Decide(ctx context.Context, lineID, decision, actor, reason string, movementID, resultBalanceID *string) error {
	now := time.Now()
	values := map[string]any{"decision_code": decision, "decided_at": now, "decided_by": actor, "decision_reason": reason, "inventory_movement_id": movementID, "resulting_balance_id": resultBalanceID, "updated_at": now, "updated_by": actor, "version_no": gorm.Expr("version_no+1")}
	result := r.db.WithContext(ctx).Model(&model.CycleCountLine{}).Where("cycle_count_line_id=? AND decision_code='COUNTED' AND NOT requires_recount", lineID).Updates(values)
	if result.Error != nil {
		return inventoryrepo.Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return inventoryrepo.ErrConflict
	}
	return nil
}
func (r *CycleCountRepository) CancelRemaining(ctx context.Context, id, actor, reason string) error {
	return inventoryrepo.Error(r.db.WithContext(ctx).Model(&model.CycleCountLine{}).Where("cycle_count_id=? AND decision_code IN ('OPEN','COUNTED')", id).Updates(map[string]any{"decision_code": "CANCELLED", "decided_at": time.Now(), "decided_by": actor, "decision_reason": reason, "requires_recount": false, "updated_by": actor, "updated_at": time.Now(), "version_no": gorm.Expr("version_no+1")}).Error)
}
func (r *CycleCountRepository) Progress(ctx context.Context, id string) (open, counted, recount, final int64, err error) {
	var rows []struct {
		DecisionCode    string
		RequiresRecount bool
		Count           int64
	}
	err = r.db.WithContext(ctx).Model(&model.CycleCountLine{}).Select("decision_code,requires_recount,count(*) count").Where("cycle_count_id=?", id).Group("decision_code,requires_recount").Scan(&rows).Error
	for _, row := range rows {
		switch row.DecisionCode {
		case "OPEN":
			open += row.Count
		case "COUNTED":
			counted += row.Count
			if row.RequiresRecount {
				recount += row.Count
			}
		default:
			final += row.Count
		}
	}
	return open, counted, recount, final, inventoryrepo.Error(err)
}
