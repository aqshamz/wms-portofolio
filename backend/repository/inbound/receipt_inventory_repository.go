package inbound

import (
	"context"

	model "wms-api/models/inbound"

	"gorm.io/gorm"
)

type ReceiptInventoryRow struct {
	model.ReceiptInventory
	LotNumber, HandlingUnitBarcode, SerialID, SerialNo                           *string
	ReceivedLocationCode, InitialInventoryStatusCode, SourceUOMCode, BaseUOMCode string
}

type ReceiptInventoryContext struct {
	ReceiptInventoryRow
	ReceiptID, ReceiptStatusCode, OwnerID, WarehouseID, BusinessDate, VendorID string
}

type ReceiptInventoryRepository struct{ db *gorm.DB }

func NewReceiptInventoryRepository(db *gorm.DB) *ReceiptInventoryRepository {
	return &ReceiptInventoryRepository{db: db}
}
func (r *ReceiptInventoryRepository) Create(ctx context.Context, value *model.ReceiptInventory) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}
func (r *ReceiptInventoryRepository) GetContext(ctx context.Context, id string) (ReceiptInventoryContext, error) {
	var value ReceiptInventoryContext
	err := r.db.WithContext(ctx).Table("receipt_inventory batch").Select(`batch.*,receipt.receipt_id,receipt_status.code receipt_status_code,receipt.owner_id,receipt.warehouse_id,receipt.business_date::text business_date,inbound.vendor_id,lot.lot_number,hu.barcode handling_unit_barcode,serial.serial_id,serial.serial_no,location.code received_location_code,status.code initial_inventory_status_code,source_uom.code source_uom_code,base_uom.code base_uom_code`).
		Joins("JOIN receipt_line line ON line.receipt_line_id=batch.receipt_line_id").Joins("JOIN receipt ON receipt.receipt_id=line.receipt_id").Joins("JOIN document_status receipt_status ON receipt_status.status_id=receipt.status_id").Joins("JOIN inbound_order inbound ON inbound.inbound_id=receipt.inbound_id").
		Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=batch.lot_id").Joins("LEFT JOIN handling_unit hu ON hu.handling_unit_id=batch.handling_unit_id").Joins("LEFT JOIN receipt_line_serial link ON link.receipt_inventory_id=batch.receipt_inventory_id").Joins("LEFT JOIN serial_number serial ON serial.serial_id=link.serial_id").
		Joins("JOIN warehouse_location location ON location.location_id=batch.received_location_id").Joins("JOIN inventory_status status ON status.inventory_status_id=batch.initial_inventory_status_id").Joins("JOIN uom source_uom ON source_uom.uom_id=batch.source_uom_id").Joins("JOIN uom base_uom ON base_uom.uom_id=batch.base_uom_id").
		Where("batch.receipt_inventory_id=?", id).Take(&value).Error
	return value, Error(err)
}
func (r *ReceiptInventoryRepository) ListByLine(ctx context.Context, lineID string) ([]ReceiptInventoryRow, error) {
	rows := make([]ReceiptInventoryRow, 0)
	err := r.db.WithContext(ctx).Table("receipt_inventory batch").Select(`batch.*,lot.lot_number,hu.barcode handling_unit_barcode,serial.serial_id,serial.serial_no,location.code received_location_code,status.code initial_inventory_status_code,source_uom.code source_uom_code,base_uom.code base_uom_code`).Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=batch.lot_id").Joins("LEFT JOIN handling_unit hu ON hu.handling_unit_id=batch.handling_unit_id").Joins("LEFT JOIN receipt_line_serial link ON link.receipt_inventory_id=batch.receipt_inventory_id").Joins("LEFT JOIN serial_number serial ON serial.serial_id=link.serial_id").Joins("JOIN warehouse_location location ON location.location_id=batch.received_location_id").Joins("JOIN inventory_status status ON status.inventory_status_id=batch.initial_inventory_status_id").Joins("JOIN uom source_uom ON source_uom.uom_id=batch.source_uom_id").Joins("JOIN uom base_uom ON base_uom.uom_id=batch.base_uom_id").Where("batch.receipt_line_id=?", lineID).Order("batch.receipt_inventory_id").Find(&rows).Error
	return rows, Error(err)
}
func (r *ReceiptInventoryRepository) SetInitialBalance(ctx context.Context, id, balanceID string) error {
	result := r.db.WithContext(ctx).Model(&model.ReceiptInventory{}).Where("receipt_inventory_id=? AND initial_balance_id IS NULL", id).Update("initial_balance_id", balanceID)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConflict
	}
	return nil
}
