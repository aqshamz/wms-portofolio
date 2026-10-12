package outbound

import (
	"time"

	"gorm.io/gorm"
)

// Migration 9 retains its original line schema. Snapshot columns are added and
// backfilled by migration 27, including on databases with existing orders.
type legacyOutboundOrderLine struct {
	ID                    string    `gorm:"column:outbound_line_id;size:150;primaryKey"`
	OutboundID            string    `gorm:"column:outbound_id;size:120;not null"`
	LineNo                int       `gorm:"column:line_no;not null"`
	ItemID                string    `gorm:"column:item_id;type:uuid;not null"`
	OrderedQty            string    `gorm:"column:ordered_qty;type:numeric(20,6);not null"`
	AllocatedQty          string    `gorm:"column:allocated_qty;type:numeric(20,6);not null;default:0"`
	PickedQty             string    `gorm:"column:picked_qty;type:numeric(20,6);not null;default:0"`
	CheckedQty            string    `gorm:"column:checked_qty;type:numeric(20,6);not null;default:0"`
	PackedQty             string    `gorm:"column:packed_qty;type:numeric(20,6);not null;default:0"`
	ShippedQty            string    `gorm:"column:shipped_qty;type:numeric(20,6);not null;default:0"`
	DeliveredQty          string    `gorm:"column:delivered_qty;type:numeric(20,6);not null;default:0"`
	RejectedQty           string    `gorm:"column:rejected_qty;type:numeric(20,6);not null;default:0"`
	ShortAcceptedQty      string    `gorm:"column:short_accepted_qty;type:numeric(20,6);not null;default:0"`
	UOMID                 string    `gorm:"column:uom_id;type:uuid;not null"`
	RequestedLotNo        *string   `gorm:"column:requested_lot_no;size:100"`
	CustomerLineReference *string   `gorm:"column:customer_line_reference;size:100"`
	Notes                 *string   `gorm:"column:notes;type:text"`
	CreatedAt             time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy             string    `gorm:"column:created_by;type:uuid;not null"`
}

func (legacyOutboundOrderLine) TableName() string { return "outbound_order_line" }

func MigrateQuantitySnapshots(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		statements := []string{
			`ALTER TABLE outbound_order_line ADD COLUMN IF NOT EXISTS uom_conversion_to_base numeric(20,6), ADD COLUMN IF NOT EXISTS ordered_base_qty numeric(20,6), ADD COLUMN IF NOT EXISTS base_uom_id uuid`,
			// Legacy orders were base-unit-only. Today's packaging must never
			// reinterpret quantities already allocated, picked or shipped.
			`UPDATE outbound_order_line SET uom_conversion_to_base=1, ordered_base_qty=ordered_qty, base_uom_id=uom_id WHERE uom_conversion_to_base IS NULL AND ordered_base_qty IS NULL AND base_uom_id IS NULL`,
			`ALTER TABLE outbound_order_line ALTER COLUMN uom_conversion_to_base SET NOT NULL, ALTER COLUMN ordered_base_qty SET NOT NULL, ALTER COLUMN base_uom_id SET NOT NULL`,
			// SQL-bootstrap installations have a second execution-quantity
			// constraint. Preserve its rules, changing only the demand UOM.
			`DO $$ BEGIN
			IF EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid='outbound_order_line'::regclass AND conname='ck_outbound_ordered_qty') THEN
				ALTER TABLE outbound_order_line DROP CONSTRAINT ck_outbound_ordered_qty;
				ALTER TABLE outbound_order_line ADD CONSTRAINT ck_outbound_ordered_qty CHECK (
					ordered_qty>0 AND allocated_qty>=0 AND picked_qty>=0 AND checked_qty>=0 AND packed_qty>=0
					AND shipped_qty>=0 AND delivered_qty>=0 AND rejected_qty>=0 AND rejected_qty<=picked_qty
					AND short_accepted_qty>=0 AND short_accepted_qty<=rejected_qty AND delivered_qty<=shipped_qty
					AND shipped_qty<=packed_qty AND packed_qty<=checked_qty AND checked_qty<=picked_qty-rejected_qty
					AND checked_qty<=ordered_base_qty-short_accepted_qty AND picked_qty<=allocated_qty
					AND allocated_qty<=ordered_base_qty+rejected_qty);
			END IF;
			END $$`,
			`ALTER TABLE outbound_order_line DROP CONSTRAINT IF EXISTS ck_outbound_line_qty`,
			`ALTER TABLE outbound_order_line ADD CONSTRAINT ck_outbound_line_qty CHECK(line_no>0 AND ordered_qty>0 AND uom_conversion_to_base>0 AND ordered_base_qty>0 AND ordered_base_qty=ordered_qty*uom_conversion_to_base AND (uom_id<>base_uom_id OR uom_conversion_to_base=1) AND allocated_qty>=0 AND picked_qty>=0 AND picked_qty<=allocated_qty AND allocated_qty<=ordered_base_qty)`,
			`DO $$ BEGIN ALTER TABLE outbound_order_line ADD CONSTRAINT fk_outbound_line_base_uom FOREIGN KEY(base_uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`UPDATE outbound_validation_rule SET name='Valid quantity snapshots', description='Order quantities must have a valid frozen conversion to the item base UOM.' WHERE handler_code='REQUIRE_BASE_UOM'`,
		}
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
