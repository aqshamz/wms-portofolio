package stockcontrol

import (
	"context"
	"gorm.io/gorm"
	"time"
	inventorymodel "wms-api/models/inventory"
	model "wms-api/models/master"
	stockmodel "wms-api/models/stock_control"
	inventoryrepo "wms-api/repository/inventory"
)

// legacyInventoryAdjustment freezes migration 18's original table shape.
// Later model changes must not make an already-released migration mutate data.
type legacyInventoryAdjustment struct {
	ID                      string     `gorm:"column:inventory_adjustment_request_id;size:140;primaryKey"`
	DocumentTypeID          string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID                string     `gorm:"column:status_id;type:uuid;not null"`
	OwnerID                 string     `gorm:"column:owner_id;type:uuid;not null"`
	WarehouseID             string     `gorm:"column:warehouse_id;type:uuid;not null"`
	BusinessDate            time.Time  `gorm:"column:business_date;type:date;not null"`
	BalanceID               string     `gorm:"column:balance_id;size:160;not null"`
	PlannedBalanceVersionNo int64      `gorm:"column:planned_balance_version_no;not null"`
	ItemID                  string     `gorm:"column:item_id;type:uuid;not null"`
	LotID                   *string    `gorm:"column:lot_id;size:120"`
	SerialID                *string    `gorm:"column:serial_id;size:160"`
	HandlingUnitID          *string    `gorm:"column:handling_unit_id;size:120"`
	LocationID              string     `gorm:"column:location_id;type:uuid;not null"`
	InventoryStatusID       string     `gorm:"column:inventory_status_id;type:uuid;not null"`
	UOMID                   string     `gorm:"column:uom_id;type:uuid;not null"`
	Direction               string     `gorm:"column:direction;size:10;not null"`
	Quantity                string     `gorm:"column:quantity;type:numeric(20,6);not null"`
	ReasonCodeID            string     `gorm:"column:reason_code_id;type:uuid;not null"`
	Notes                   *string    `gorm:"column:notes;type:text"`
	ApprovedAt              *time.Time `gorm:"column:approved_at"`
	ApprovedBy              *string    `gorm:"column:approved_by;type:uuid"`
	PostedAt                *time.Time `gorm:"column:posted_at"`
	PostedBy                *string    `gorm:"column:posted_by;type:uuid"`
	CancelledAt             *time.Time `gorm:"column:cancelled_at"`
	CancelledBy             *string    `gorm:"column:cancelled_by;type:uuid"`
	CancellationReason      *string    `gorm:"column:cancellation_reason;type:text"`
	InventoryMovementID     *string    `gorm:"column:inventory_movement_id;size:140"`
	ResultingBalanceID      *string    `gorm:"column:resulting_balance_id;size:160"`
	CreatedAt               time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy               string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt               time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy               *string    `gorm:"column:updated_by;type:uuid"`
	VersionNo               int64      `gorm:"column:version_no;not null;default:1"`
}

func (legacyInventoryAdjustment) TableName() string { return "inventory_adjustment_request" }

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.ReasonCode{}); err != nil {
			return err
		}
		if err := tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS uq_stock_reason_module_code ON reason_code(module_code,code)").Error; err != nil {
			return err
		}
		return NewReasonCodeRepository(tx).Seed(context.Background(), []model.ReasonCode{{ModuleCode: "INVENTORY", Code: "DAMAGE", Name: "Damage", RequiresNote: true}, {ModuleCode: "INVENTORY", Code: "EXPIRY", Name: "Expiry"}, {ModuleCode: "INVENTORY", Code: "COUNT_VARIANCE", Name: "Count variance", RequiresNote: true}, {ModuleCode: "INVENTORY", Code: "REPLENISHMENT", Name: "Replenishment"}, {ModuleCode: "INVENTORY", Code: "RELOCATION", Name: "Relocation"}, {ModuleCode: "INVENTORY", Code: "CONSOLIDATION", Name: "Consolidation"}, {ModuleCode: "INVENTORY", Code: "STATUS_HOLD", Name: "Place on hold", RequiresNote: true}, {ModuleCode: "INVENTORY", Code: "STATUS_RELEASE", Name: "Release hold", RequiresNote: true}, {ModuleCode: "INVENTORY", Code: "MANUAL_ADJUSTMENT", Name: "Manual adjustment", RequiresNote: true}, {ModuleCode: "INVENTORY", Code: "TRANSFER_VARIANCE", Name: "Transfer variance", RequiresNote: true}})
	})
}

// MigrateReplenishments is deliberately a separate migration step so existing
// installations that already applied the original stock-control migration get
// the task table and movement type as well.
func MigrateReplenishments(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&stockmodel.ReplenishmentTask{}); err != nil {
			return err
		}
		for _, statement := range []string{
			`CREATE INDEX IF NOT EXISTS ix_replenishment_scope ON replenishment_task(owner_id,warehouse_id,created_at DESC)`,
			`CREATE INDEX IF NOT EXISTS ix_replenishment_source ON replenishment_task(source_balance_id)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_replenishment_active_serial ON replenishment_task(serial_id) WHERE serial_id IS NOT NULL AND completed_at IS NULL AND cancelled_at IS NULL`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_replenishment_active_hu ON replenishment_task(handling_unit_id) WHERE handling_unit_id IS NOT NULL AND completed_at IS NULL AND cancelled_at IS NULL`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT ck_replenishment_quantity CHECK(planned_qty>0 AND completed_qty>=0 AND completed_qty<=planned_qty); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT ck_replenishment_locations CHECK(source_location_id<>target_location_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_source_balance FOREIGN KEY(source_balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_task_type FOREIGN KEY(task_type_id) REFERENCES task_type(task_type_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_task_status FOREIGN KEY(task_status_id) REFERENCES task_status(task_status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_task_priority FOREIGN KEY(task_priority_id) REFERENCES task_priority(task_priority_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_owner FOREIGN KEY(owner_id) REFERENCES organization(organization_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_warehouse FOREIGN KEY(warehouse_id) REFERENCES warehouse(warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_item FOREIGN KEY(owner_id,item_id) REFERENCES item(owner_id,item_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_lot FOREIGN KEY(lot_id,owner_id,item_id) REFERENCES inventory_lot(lot_id,owner_id,item_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_serial FOREIGN KEY(serial_id,owner_id,item_id) REFERENCES serial_number(serial_id,owner_id,item_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_hu FOREIGN KEY(handling_unit_id,owner_id,warehouse_id) REFERENCES handling_unit(handling_unit_id,owner_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_status FOREIGN KEY(inventory_status_id) REFERENCES inventory_status(inventory_status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_source_location FOREIGN KEY(source_location_id,warehouse_id) REFERENCES warehouse_location(location_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_target_location FOREIGN KEY(target_location_id,warehouse_id) REFERENCES warehouse_location(location_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_uom FOREIGN KEY(uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_assignee FOREIGN KEY(assigned_to) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_created_by FOREIGN KEY(created_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_updated_by FOREIGN KEY(updated_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_cancelled_by FOREIGN KEY(cancelled_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_movement FOREIGN KEY(inventory_movement_id) REFERENCES inventory_movement(movement_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE replenishment_task ADD CONSTRAINT fk_replenishment_result_balance FOREIGN KEY(resulting_balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		} {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return inventoryrepo.NewMovementTypeRepository(tx).Seed(context.Background(), []inventorymodel.MovementType{{Code: "REPLENISHMENT", Name: "Replenishment"}})
	})
}

// MigrateInventoryAdjustments adds the maker-checker document that guards the
// existing low-level adjustment posting command.
func MigrateInventoryAdjustments(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&legacyInventoryAdjustment{}); err != nil {
			return err
		}
		for _, statement := range []string{
			`CREATE INDEX IF NOT EXISTS ix_inventory_adjustment_request_scope ON inventory_adjustment_request(owner_id,warehouse_id,created_at DESC)`,
			`CREATE INDEX IF NOT EXISTS ix_inventory_adjustment_request_balance ON inventory_adjustment_request(balance_id,created_at DESC)`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT ck_inventory_adjustment_request_quantity CHECK(quantity>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT ck_inventory_adjustment_request_direction CHECK(direction IN ('INCREASE','DECREASE')); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT ck_inventory_adjustment_request_approval CHECK((approved_at IS NULL)=(approved_by IS NULL)); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT ck_inventory_adjustment_request_posting CHECK((posted_at IS NULL)=(posted_by IS NULL) AND (posted_at IS NULL)=(inventory_movement_id IS NULL)); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT ck_inventory_adjustment_request_cancellation CHECK((cancelled_at IS NULL)=(cancelled_by IS NULL)); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT fk_adjustment_request_document_type FOREIGN KEY(document_type_id) REFERENCES document_type(document_type_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT fk_adjustment_request_status FOREIGN KEY(status_id) REFERENCES document_status(status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT fk_adjustment_request_owner FOREIGN KEY(owner_id) REFERENCES organization(organization_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT fk_adjustment_request_warehouse FOREIGN KEY(warehouse_id) REFERENCES warehouse(warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT fk_adjustment_request_balance FOREIGN KEY(balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT fk_adjustment_request_item FOREIGN KEY(owner_id,item_id) REFERENCES item(owner_id,item_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT fk_adjustment_request_lot FOREIGN KEY(lot_id,owner_id,item_id) REFERENCES inventory_lot(lot_id,owner_id,item_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT fk_adjustment_request_serial FOREIGN KEY(serial_id,owner_id,item_id) REFERENCES serial_number(serial_id,owner_id,item_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT fk_adjustment_request_hu FOREIGN KEY(handling_unit_id,owner_id,warehouse_id) REFERENCES handling_unit(handling_unit_id,owner_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT fk_adjustment_request_location FOREIGN KEY(location_id,warehouse_id) REFERENCES warehouse_location(location_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT fk_adjustment_request_inventory_status FOREIGN KEY(inventory_status_id) REFERENCES inventory_status(inventory_status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT fk_adjustment_request_uom FOREIGN KEY(uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT fk_adjustment_request_reason FOREIGN KEY(reason_code_id) REFERENCES reason_code(reason_code_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT fk_adjustment_request_creator FOREIGN KEY(created_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT fk_adjustment_request_approver FOREIGN KEY(approved_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT fk_adjustment_request_poster FOREIGN KEY(posted_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT fk_adjustment_request_canceller FOREIGN KEY(cancelled_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT fk_adjustment_request_movement FOREIGN KEY(inventory_movement_id) REFERENCES inventory_movement(movement_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request ADD CONSTRAINT fk_adjustment_request_result_balance FOREIGN KEY(resulting_balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		} {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// MigrateInventoryAdjustmentLines upgrades the original single-balance request
// without dropping its legacy columns. Existing requests are copied into line 1.
func MigrateInventoryAdjustmentLines(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, column := range []string{"balance_id", "planned_balance_version_no", "item_id", "location_id", "inventory_status_id", "uom_id", "quantity"} {
			if err := tx.Exec("ALTER TABLE inventory_adjustment_request ALTER COLUMN " + column + " DROP NOT NULL").Error; err != nil {
				return err
			}
		}
		if err := tx.AutoMigrate(&stockmodel.InventoryAdjustment{}, &stockmodel.InventoryAdjustmentLine{}); err != nil {
			return err
		}
		for _, statement := range []string{
			`CREATE INDEX IF NOT EXISTS ix_adjustment_line_document ON inventory_adjustment_request_line(inventory_adjustment_request_id,line_no)`,
			`CREATE INDEX IF NOT EXISTS ix_adjustment_line_balance ON inventory_adjustment_request_line(balance_id)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_adjustment_line_number ON inventory_adjustment_request_line(inventory_adjustment_request_id,line_no)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_adjustment_line_balance ON inventory_adjustment_request_line(inventory_adjustment_request_id,balance_id)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_adjustment_line_serial ON inventory_adjustment_request_line(inventory_adjustment_request_id,serial_id) WHERE serial_id IS NOT NULL`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT ck_adjustment_line_quantity CHECK(quantity>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT ck_adjustment_line_version CHECK(planned_balance_version_no>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT ck_adjustment_line_decision CHECK(decision_code IN ('PENDING','POSTED','REJECTED','CANCELLED')); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT ck_adjustment_line_approval CHECK((approved_at IS NULL)=(approved_by IS NULL)); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT ck_adjustment_line_rejection CHECK((rejected_at IS NULL)=(rejected_by IS NULL)); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT ck_adjustment_line_cancellation CHECK((cancelled_at IS NULL)=(cancelled_by IS NULL)); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT fk_adjustment_line_document FOREIGN KEY(inventory_adjustment_request_id) REFERENCES inventory_adjustment_request(inventory_adjustment_request_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT fk_adjustment_line_balance FOREIGN KEY(balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT fk_adjustment_line_item FOREIGN KEY(item_id) REFERENCES item(item_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT fk_adjustment_line_lot FOREIGN KEY(lot_id) REFERENCES inventory_lot(lot_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT fk_adjustment_line_serial FOREIGN KEY(serial_id) REFERENCES serial_number(serial_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT fk_adjustment_line_hu FOREIGN KEY(handling_unit_id) REFERENCES handling_unit(handling_unit_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT fk_adjustment_line_location FOREIGN KEY(location_id) REFERENCES warehouse_location(location_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT fk_adjustment_line_status FOREIGN KEY(inventory_status_id) REFERENCES inventory_status(inventory_status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT fk_adjustment_line_uom FOREIGN KEY(uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT fk_adjustment_line_approver FOREIGN KEY(approved_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT fk_adjustment_line_rejecter FOREIGN KEY(rejected_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT fk_adjustment_line_canceller FOREIGN KEY(cancelled_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT fk_adjustment_line_creator FOREIGN KEY(created_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT fk_adjustment_line_updater FOREIGN KEY(updated_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT fk_adjustment_line_movement FOREIGN KEY(inventory_movement_id) REFERENCES inventory_movement(movement_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_adjustment_request_line ADD CONSTRAINT fk_adjustment_line_result_balance FOREIGN KEY(resulting_balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`INSERT INTO inventory_adjustment_request_line (
				inventory_adjustment_request_line_id,inventory_adjustment_request_id,line_no,balance_id,
				planned_balance_version_no,item_id,lot_id,serial_id,handling_unit_id,location_id,
				inventory_status_id,uom_id,quantity,decision_code,approved_at,approved_by,
				cancelled_at,cancelled_by,cancellation_reason,inventory_movement_id,resulting_balance_id,
				created_at,created_by,updated_at,updated_by,version_no)
			SELECT inventory_adjustment_request_id||'-L0001',inventory_adjustment_request_id,1,balance_id,
				planned_balance_version_no,item_id,lot_id,serial_id,handling_unit_id,location_id,
				inventory_status_id,uom_id,quantity,
				CASE WHEN posted_at IS NOT NULL THEN 'POSTED' WHEN cancelled_at IS NOT NULL THEN 'CANCELLED' ELSE 'PENDING' END,
				approved_at,approved_by,cancelled_at,cancelled_by,cancellation_reason,
				inventory_movement_id,resulting_balance_id,created_at,created_by,updated_at,updated_by,1
			FROM inventory_adjustment_request WHERE balance_id IS NOT NULL
			ON CONFLICT (inventory_adjustment_request_line_id) DO NOTHING`,
			`UPDATE inventory_adjustment_request SET completed_at=COALESCE(posted_at,cancelled_at)
			WHERE completed_at IS NULL AND (posted_at IS NOT NULL OR cancelled_at IS NOT NULL)`,
		} {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
