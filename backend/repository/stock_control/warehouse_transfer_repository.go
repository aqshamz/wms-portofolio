package stockcontrol

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	mastermodel "wms-api/models/master"
	model "wms-api/models/stock_control"
	inventoryrepo "wms-api/repository/inventory"
)

type WarehouseTransferFilter struct {
	OwnerID, WarehouseID, Side, StatusCode, Search string
	Page, PageSize                                  int
}
type WarehouseTransferRow struct {
	model.WarehouseTransfer
	StatusCode, OwnerCode, OwnerName                                     string
	SourceWarehouseCode, SourceWarehouseName                             string
	TargetWarehouseCode, TargetWarehouseName, CreatedByDisplayName       string
}
type WarehouseTransferLineRow struct {
	model.WarehouseTransferLine
	ItemCode, ItemName, SourceLocationCode, SourceInventoryStatusCode, UOMCode string
	LotNumber, SerialNumber, HandlingUnitBarcode                              *string
	ReceiptLocationCode, PutawayTargetLocationCode                            *string
	ReceivedBalanceVersionNo                                                  *int64
}
type WarehouseTransferRepository struct{ db *gorm.DB }

func NewWarehouseTransferRepository(db *gorm.DB) *WarehouseTransferRepository {
	return &WarehouseTransferRepository{db: db}
}
func warehouseTransferQuery(db *gorm.DB) *gorm.DB {
	return db.Table("warehouse_transfer transfer").Select(`transfer.*,status.code status_code,
		owner.code owner_code,owner.name owner_name,source.code source_warehouse_code,source.name source_warehouse_name,
		target.code target_warehouse_code,target.name target_warehouse_name,creator.display_name created_by_display_name`).
		Joins("JOIN document_status status ON status.status_id=transfer.status_id").
		Joins("JOIN organization owner ON owner.organization_id=transfer.owner_id").
		Joins("JOIN warehouse source ON source.warehouse_id=transfer.source_warehouse_id").
		Joins("JOIN warehouse target ON target.warehouse_id=transfer.target_warehouse_id").
		Joins("JOIN app_account creator ON creator.account_id=transfer.created_by")
}
func warehouseTransferLineQuery(db *gorm.DB) *gorm.DB {
	return db.Table("warehouse_transfer_line line").Select(`line.*,item.code item_code,item.name item_name,
		source_location.code source_location_code,source_status.code source_inventory_status_code,uom.code uom_code,
		lot.lot_number,serial.serial_no serial_number,hu.barcode handling_unit_barcode,
		receipt_location.code receipt_location_code,putaway_location.code putaway_target_location_code,
		received_balance.version_no received_balance_version_no`).
		Joins("JOIN item ON item.item_id=line.item_id").
		Joins("JOIN warehouse_location source_location ON source_location.location_id=line.source_location_id").
		Joins("JOIN inventory_status source_status ON source_status.inventory_status_id=line.source_inventory_status_id").
		Joins("JOIN uom ON uom.uom_id=line.uom_id").
		Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=line.lot_id").
		Joins("LEFT JOIN serial_number serial ON serial.serial_id=line.serial_id").
		Joins("LEFT JOIN handling_unit hu ON hu.handling_unit_id=line.handling_unit_id").
		Joins("LEFT JOIN warehouse_location receipt_location ON receipt_location.location_id=line.receipt_location_id").
		Joins("LEFT JOIN warehouse_location putaway_location ON putaway_location.location_id=line.putaway_target_location_id").
		Joins("LEFT JOIN inventory_balance received_balance ON received_balance.balance_id=line.received_balance_id")
}
func (r *WarehouseTransferRepository) Create(ctx context.Context, header *model.WarehouseTransfer, line *model.WarehouseTransferLine) error {
	if err := r.db.WithContext(ctx).Create(header).Error; err != nil {
		return inventoryrepo.Error(err)
	}
	return inventoryrepo.Error(r.db.WithContext(ctx).Create(line).Error)
}
func (r *WarehouseTransferRepository) Get(ctx context.Context, id string) (WarehouseTransferRow, error) {
	var row WarehouseTransferRow
	err := warehouseTransferQuery(r.db.WithContext(ctx)).Where("transfer.warehouse_transfer_id=?", id).Take(&row).Error
	return row, inventoryrepo.Error(err)
}
func (r *WarehouseTransferRepository) Lines(ctx context.Context, id string) ([]WarehouseTransferLineRow, error) {
	var rows []WarehouseTransferLineRow
	err := warehouseTransferLineQuery(r.db.WithContext(ctx)).Where("line.warehouse_transfer_id=?", id).Order("line.line_no").Find(&rows).Error
	return rows, inventoryrepo.Error(err)
}
func (r *WarehouseTransferRepository) List(ctx context.Context, f WarehouseTransferFilter) ([]WarehouseTransferRow, int64, error) {
	q := warehouseTransferQuery(r.db.WithContext(ctx)).Where("transfer.owner_id=?", f.OwnerID)
	if f.Side == "SOURCE" {
		q = q.Where("transfer.source_warehouse_id=?", f.WarehouseID)
	} else {
		q = q.Where("transfer.target_warehouse_id=?", f.WarehouseID)
	}
	if f.StatusCode != "" {
		q = q.Where("status.code=?", f.StatusCode)
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where(`transfer.warehouse_transfer_id ILIKE ? OR source.code ILIKE ? OR target.code ILIKE ? OR EXISTS (
			SELECT 1 FROM warehouse_transfer_line search_line JOIN item search_item ON search_item.item_id=search_line.item_id
			WHERE search_line.warehouse_transfer_id=transfer.warehouse_transfer_id AND (search_item.code ILIKE ? OR search_item.name ILIKE ?))`, like, like, like, like, like)
	}
	var total int64
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, inventoryrepo.Error(err)
	}
	var rows []WarehouseTransferRow
	err := q.Order("transfer.created_at DESC,transfer.warehouse_transfer_id DESC").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&rows).Error
	return rows, total, inventoryrepo.Error(err)
}
func (r *WarehouseTransferRepository) Lock(ctx context.Context, id string) (model.WarehouseTransfer, error) {
	var row model.WarehouseTransfer
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("warehouse_transfer_id=?", id).Take(&row).Error
	return row, inventoryrepo.Error(err)
}
func (r *WarehouseTransferRepository) LockLine(ctx context.Context, id string) (model.WarehouseTransferLine, error) {
	var row model.WarehouseTransferLine
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("warehouse_transfer_id=?", id).Take(&row).Error
	return row, inventoryrepo.Error(err)
}
func (r *WarehouseTransferRepository) UpdateHeader(ctx context.Context, id string, version int64, values map[string]any) error {
	values["updated_at"] = gorm.Expr("clock_timestamp()")
	values["version_no"] = gorm.Expr("version_no+1")
	result := r.db.WithContext(ctx).Model(&model.WarehouseTransfer{}).Where("warehouse_transfer_id=? AND version_no=?", id, version).Updates(values)
	if result.Error != nil { return inventoryrepo.Error(result.Error) }
	if result.RowsAffected != 1 { return inventoryrepo.ErrConflict }
	return nil
}
func (r *WarehouseTransferRepository) UpdateLine(ctx context.Context, id string, version int64, values map[string]any) error {
	values["updated_at"] = gorm.Expr("clock_timestamp()")
	values["version_no"] = gorm.Expr("version_no+1")
	result := r.db.WithContext(ctx).Model(&model.WarehouseTransferLine{}).Where("warehouse_transfer_line_id=? AND version_no=?", id, version).Updates(values)
	if result.Error != nil { return inventoryrepo.Error(result.Error) }
	if result.RowsAffected != 1 { return inventoryrepo.ErrConflict }
	return nil
}
func (r *WarehouseTransferRepository) DocumentType(ctx context.Context) (mastermodel.DocumentType, error) {
	var row mastermodel.DocumentType
	err := r.db.WithContext(ctx).Where("code=? AND is_active", "TRANSFER").Take(&row).Error
	return row, inventoryrepo.Error(err)
}
func (r *WarehouseTransferRepository) Status(ctx context.Context, typeID, code string) (mastermodel.DocumentStatus, error) {
	var row mastermodel.DocumentStatus
	err := r.db.WithContext(ctx).Where("document_type_id=? AND code=? AND is_active", typeID, code).Take(&row).Error
	return row, inventoryrepo.Error(err)
}
func (r *WarehouseTransferRepository) InventoryStatus(ctx context.Context, code string) (mastermodel.InventoryStatus, error) {
	var row mastermodel.InventoryStatus
	err := r.db.WithContext(ctx).Where("code=? AND is_active", code).Take(&row).Error
	return row, inventoryrepo.Error(err)
}
