package inventory

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/inventory"
	master "wms-api/repository/master"
)

func defaultMovementTypes() []model.MovementType {
	return []model.MovementType{
		{Code: "RECEIVE", Name: "Receive"}, {Code: "PUTAWAY", Name: "Putaway"},
		{Code: "PICK", Name: "Pick"}, {Code: "PACK", Name: "Pack"},
		{Code: "SHIP", Name: "Ship"}, {Code: "TRANSFER", Name: "Transfer"},
		{Code: "INTERNAL_MOVE", Name: "Internal move"}, {Code: "TRANSFER_OUT", Name: "Transfer out"},
		{Code: "TRANSFER_IN", Name: "Transfer in"}, {Code: "ADJUSTMENT", Name: "Adjustment"},
		{Code: "STATUS_CHANGE", Name: "Status change"}, {Code: "COUNT_CORRECTION", Name: "Count correction"},
		{Code: "RETURN_TO_VENDOR", Name: "Return to vendor"}, {Code: "DISPOSE", Name: "Dispose"},
		{Code: "DELIVERY_RETURN", Name: "Delivery return"},
		{Code: "OUTBOUND_CHECK_CORRECTION", Name: "Outbound check correction"},
	}
}

// Migrate is additive. Conflicting legacy data causes a rollback, never cleanup.
func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := master.MigrateHandlingUnitTypes(tx); err != nil {
			return err
		}
		if err := tx.AutoMigrate(&model.InventoryLot{}, &model.SerialNumber{}, &model.HandlingUnit{}, &model.MovementType{}, &model.InventoryBalance{}, &model.InventoryMovement{}, &model.SerialInventory{}); err != nil {
			return err
		}
		for _, sql := range []string{
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_identity_lot_business ON inventory_lot(owner_id,item_id,lot_number)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_identity_serial_business ON serial_number(owner_id,item_id,serial_no)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_identity_hu_barcode ON handling_unit(barcode)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_identity_hu_scope ON handling_unit(handling_unit_id,owner_id,warehouse_id)`,
			`CREATE INDEX IF NOT EXISTS ix_identity_hu_location ON handling_unit(warehouse_id,current_location_id)`,
			`CREATE INDEX IF NOT EXISTS ix_identity_hu_parent ON handling_unit(parent_handling_unit_id)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_core_balance_identity ON inventory_balance(owner_id,warehouse_id,location_id,item_id,lot_id,handling_unit_id,inventory_status_id) NULLS NOT DISTINCT`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_core_balance_scope ON inventory_balance(balance_id,owner_id,item_id)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_core_lot_scope ON inventory_lot(lot_id,owner_id,item_id)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_core_serial_scope ON serial_number(serial_id,owner_id,item_id)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_core_movement_operation ON inventory_movement(operation_key) WHERE operation_key IS NOT NULL`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_core_movement_type_code ON movement_type(code)`,
			`CREATE INDEX IF NOT EXISTS ix_core_balance_inquiry ON inventory_balance(owner_id,warehouse_id,item_id,inventory_status_id)`,
			`CREATE INDEX IF NOT EXISTS ix_core_movement_inquiry ON inventory_movement(owner_id,warehouse_id,occurred_at DESC)`,
			`CREATE INDEX IF NOT EXISTS ix_core_serial_balance ON serial_inventory(balance_id)`,
			`DO $$ BEGIN ALTER TABLE inventory_lot ADD CONSTRAINT fk_identity_inventory_lot_item FOREIGN KEY(owner_id,item_id) REFERENCES item(owner_id,item_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_lot ADD CONSTRAINT fk_identity_inventory_lot_owner FOREIGN KEY(owner_id) REFERENCES organization(organization_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_lot ADD CONSTRAINT fk_identity_inventory_lot_actor FOREIGN KEY(created_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE serial_number ADD CONSTRAINT fk_identity_serial_number_item FOREIGN KEY(owner_id,item_id) REFERENCES item(owner_id,item_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE serial_number ADD CONSTRAINT fk_identity_serial_number_owner FOREIGN KEY(owner_id) REFERENCES organization(organization_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE serial_number ADD CONSTRAINT fk_identity_serial_number_actor FOREIGN KEY(created_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_lot ADD CONSTRAINT fk_identity_lot_quality FOREIGN KEY(quality_status_id) REFERENCES quality_status(quality_status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_lot ADD CONSTRAINT ck_identity_lot_dates CHECK (manufacture_date IS NULL OR expiry_date IS NULL OR expiry_date >= manufacture_date); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE handling_unit ADD CONSTRAINT fk_identity_hu_owner FOREIGN KEY(owner_id) REFERENCES organization(organization_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE handling_unit ADD CONSTRAINT fk_identity_hu_warehouse FOREIGN KEY(warehouse_id) REFERENCES warehouse(warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE handling_unit ADD CONSTRAINT fk_identity_hu_assignment FOREIGN KEY(owner_id,warehouse_id) REFERENCES warehouse_owner(owner_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE handling_unit ADD CONSTRAINT fk_identity_hu_location FOREIGN KEY(current_location_id,warehouse_id) REFERENCES warehouse_location(location_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE handling_unit ADD CONSTRAINT fk_identity_hu_parent FOREIGN KEY(parent_handling_unit_id,owner_id,warehouse_id) REFERENCES handling_unit(handling_unit_id,owner_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE handling_unit ADD CONSTRAINT fk_identity_hu_type FOREIGN KEY(handling_unit_type_id) REFERENCES handling_unit_type(handling_unit_type_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE handling_unit ADD CONSTRAINT fk_identity_hu_actor FOREIGN KEY(created_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE handling_unit ADD CONSTRAINT ck_identity_hu_parent CHECK (parent_handling_unit_id IS NULL OR parent_handling_unit_id <> handling_unit_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_lot ADD CONSTRAINT ck_identity_lot_number_nonempty CHECK (length(btrim(lot_number)) > 0); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE serial_number ADD CONSTRAINT ck_identity_serial_no_nonempty CHECK (length(btrim(serial_no)) > 0); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE handling_unit ADD CONSTRAINT ck_identity_barcode_nonempty CHECK (length(btrim(barcode)) > 0); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_balance ADD CONSTRAINT fk_core_balance_item FOREIGN KEY(owner_id,item_id) REFERENCES item(owner_id,item_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_balance ADD CONSTRAINT fk_core_balance_location FOREIGN KEY(location_id,warehouse_id) REFERENCES warehouse_location(location_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_balance ADD CONSTRAINT fk_core_balance_lot FOREIGN KEY(lot_id,owner_id,item_id) REFERENCES inventory_lot(lot_id,owner_id,item_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_balance ADD CONSTRAINT fk_core_balance_hu FOREIGN KEY(handling_unit_id,owner_id,warehouse_id) REFERENCES handling_unit(handling_unit_id,owner_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_balance ADD CONSTRAINT ck_core_balance_qty CHECK(on_hand_qty>=0 AND reserved_qty>=0 AND reserved_qty<=on_hand_qty); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_balance ADD CONSTRAINT ck_core_balance_version CHECK(version_no>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_movement ADD CONSTRAINT fk_core_movement_item FOREIGN KEY(owner_id,item_id) REFERENCES item(owner_id,item_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_movement ADD CONSTRAINT fk_core_movement_lot FOREIGN KEY(lot_id,owner_id,item_id) REFERENCES inventory_lot(lot_id,owner_id,item_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_movement ADD CONSTRAINT fk_core_movement_serial FOREIGN KEY(serial_id,owner_id,item_id) REFERENCES serial_number(serial_id,owner_id,item_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_movement ADD CONSTRAINT fk_core_movement_hu FOREIGN KEY(handling_unit_id,owner_id,warehouse_id) REFERENCES handling_unit(handling_unit_id,owner_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_movement ADD CONSTRAINT fk_core_movement_from_location FOREIGN KEY(from_location_id,warehouse_id) REFERENCES warehouse_location(location_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_movement ADD CONSTRAINT fk_core_movement_to_location FOREIGN KEY(to_location_id,warehouse_id) REFERENCES warehouse_location(location_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_movement ADD CONSTRAINT ck_core_movement_qty CHECK(quantity>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_movement ADD CONSTRAINT ck_core_movement_effect CHECK(from_location_id IS DISTINCT FROM to_location_id OR from_status_id IS DISTINCT FROM to_status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE serial_inventory ADD CONSTRAINT fk_core_serial_identity FOREIGN KEY(serial_id,owner_id,item_id) REFERENCES serial_number(serial_id,owner_id,item_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE serial_inventory ADD CONSTRAINT fk_core_serial_balance FOREIGN KEY(balance_id,owner_id,item_id) REFERENCES inventory_balance(balance_id,owner_id,item_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE serial_inventory ADD CONSTRAINT ck_core_serial_version CHECK(version_no>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		} {
			if err := tx.Exec(sql).Error; err != nil {
				return err
			}
		}
		return NewMovementTypeRepository(tx).Seed(context.Background(), defaultMovementTypes())
	})
}
