package inbound

import (
	"context"

	model "wms-api/models/inbound"

	"gorm.io/gorm"
)

// migrateQuantitySnapshots upgrades existing document lines before AutoMigrate
// sees the new non-null snapshot fields. Batch ratios are preferred for receipt
// history because they preserve the conversion actually used when stock posted.
func migrateQuantitySnapshots(tx *gorm.DB) error {
	statements := []string{
		`DO $$ BEGIN
			IF to_regclass(current_schema() || '.purchase_order_line') IS NOT NULL THEN
				ALTER TABLE purchase_order_line ADD COLUMN IF NOT EXISTS uom_conversion_to_base numeric(20,6);
				ALTER TABLE purchase_order_line ADD COLUMN IF NOT EXISTS ordered_base_qty numeric(20,6);
				ALTER TABLE purchase_order_line ADD COLUMN IF NOT EXISTS base_uom_id uuid;
				UPDATE purchase_order_line line SET
					uom_conversion_to_base=COALESCE(line.uom_conversion_to_base,unit.conversion_to_base),
					ordered_base_qty=COALESCE(line.ordered_base_qty,round(line.ordered_qty*COALESCE(line.uom_conversion_to_base,unit.conversion_to_base),6)),
					base_uom_id=COALESCE(line.base_uom_id,item.base_uom_id)
				FROM item,item_uom unit
				WHERE item.item_id=line.item_id AND unit.item_id=line.item_id AND unit.uom_id=line.uom_id
					AND (line.uom_conversion_to_base IS NULL OR line.ordered_base_qty IS NULL OR line.base_uom_id IS NULL);
				IF EXISTS (SELECT 1 FROM purchase_order_line WHERE uom_conversion_to_base IS NULL OR uom_conversion_to_base<=0 OR ordered_base_qty IS NULL OR ordered_base_qty<=0 OR base_uom_id IS NULL) THEN
					RAISE EXCEPTION 'cannot backfill purchase-order quantity snapshots';
				END IF;
				ALTER TABLE purchase_order_line ALTER COLUMN uom_conversion_to_base SET NOT NULL;
				ALTER TABLE purchase_order_line ALTER COLUMN ordered_base_qty SET NOT NULL;
				ALTER TABLE purchase_order_line ALTER COLUMN base_uom_id SET NOT NULL;
			END IF;
		END $$`,
		`DO $$ BEGIN
			IF to_regclass(current_schema() || '.inbound_order_line') IS NOT NULL THEN
				ALTER TABLE inbound_order_line ADD COLUMN IF NOT EXISTS uom_conversion_to_base numeric(20,6);
				ALTER TABLE inbound_order_line ADD COLUMN IF NOT EXISTS expected_base_qty numeric(20,6);
				ALTER TABLE inbound_order_line ADD COLUMN IF NOT EXISTS base_uom_id uuid;
				WITH snapshot AS (
					SELECT line.inbound_line_id,
						COALESCE(po.uom_conversion_to_base,unit.conversion_to_base) conversion,
						COALESCE(po.base_uom_id,item.base_uom_id) base_uom_id
					FROM inbound_order_line line
					JOIN item ON item.item_id=line.item_id
					JOIN item_uom unit ON unit.item_id=line.item_id AND unit.uom_id=line.uom_id
					LEFT JOIN purchase_order_line po ON po.purchase_order_line_id=line.purchase_order_line_id
				)
				UPDATE inbound_order_line line SET
					uom_conversion_to_base=COALESCE(line.uom_conversion_to_base,snapshot.conversion),
					expected_base_qty=COALESCE(line.expected_base_qty,round(line.expected_qty*COALESCE(line.uom_conversion_to_base,snapshot.conversion),6)),
					base_uom_id=COALESCE(line.base_uom_id,snapshot.base_uom_id)
				FROM snapshot WHERE snapshot.inbound_line_id=line.inbound_line_id
					AND (line.uom_conversion_to_base IS NULL OR line.expected_base_qty IS NULL OR line.base_uom_id IS NULL);
				IF EXISTS (SELECT 1 FROM inbound_order_line WHERE uom_conversion_to_base IS NULL OR uom_conversion_to_base<=0 OR expected_base_qty IS NULL OR expected_base_qty<=0 OR base_uom_id IS NULL) THEN
					RAISE EXCEPTION 'cannot backfill inbound-order quantity snapshots';
				END IF;
				ALTER TABLE inbound_order_line ALTER COLUMN uom_conversion_to_base SET NOT NULL;
				ALTER TABLE inbound_order_line ALTER COLUMN expected_base_qty SET NOT NULL;
				ALTER TABLE inbound_order_line ALTER COLUMN base_uom_id SET NOT NULL;
			END IF;
		END $$`,
		`DO $$ BEGIN
			IF to_regclass(current_schema() || '.receipt_line') IS NOT NULL THEN
				ALTER TABLE receipt_line ADD COLUMN IF NOT EXISTS uom_conversion_to_base numeric(20,6);
				ALTER TABLE receipt_line ADD COLUMN IF NOT EXISTS received_base_qty numeric(20,6);
				ALTER TABLE receipt_line ADD COLUMN IF NOT EXISTS rejected_base_qty numeric(20,6);
				ALTER TABLE receipt_line ADD COLUMN IF NOT EXISTS base_uom_id uuid;
				WITH snapshot AS (
					SELECT line.receipt_line_id,
						round(COALESCE((SELECT batch.base_qty/NULLIF(batch.source_qty,0) FROM receipt_inventory batch WHERE batch.receipt_line_id=line.receipt_line_id ORDER BY batch.receipt_inventory_id LIMIT 1),unit.conversion_to_base),6) conversion,
						COALESCE((SELECT batch.base_uom_id FROM receipt_inventory batch WHERE batch.receipt_line_id=line.receipt_line_id ORDER BY batch.receipt_inventory_id LIMIT 1),item.base_uom_id) base_uom_id
					FROM receipt_line line
					JOIN item ON item.item_id=line.item_id
					JOIN item_uom unit ON unit.item_id=line.item_id AND unit.uom_id=line.uom_id
				)
				UPDATE receipt_line line SET
					uom_conversion_to_base=COALESCE(line.uom_conversion_to_base,snapshot.conversion),
					received_base_qty=COALESCE(line.received_base_qty,round(line.received_qty*COALESCE(line.uom_conversion_to_base,snapshot.conversion),6)),
					rejected_base_qty=COALESCE(line.rejected_base_qty,round(line.rejected_qty*COALESCE(line.uom_conversion_to_base,snapshot.conversion),6)),
					base_uom_id=COALESCE(line.base_uom_id,snapshot.base_uom_id)
				FROM snapshot WHERE snapshot.receipt_line_id=line.receipt_line_id
					AND (line.uom_conversion_to_base IS NULL OR line.received_base_qty IS NULL OR line.rejected_base_qty IS NULL OR line.base_uom_id IS NULL);
				IF EXISTS (SELECT 1 FROM receipt_line WHERE uom_conversion_to_base IS NULL OR uom_conversion_to_base<=0 OR received_base_qty IS NULL OR received_base_qty<=0 OR rejected_base_qty IS NULL OR rejected_base_qty<0 OR base_uom_id IS NULL) THEN
					RAISE EXCEPTION 'cannot backfill receipt quantity snapshots';
				END IF;
				ALTER TABLE receipt_line ALTER COLUMN uom_conversion_to_base SET NOT NULL;
				ALTER TABLE receipt_line ALTER COLUMN received_base_qty SET NOT NULL;
				ALTER TABLE receipt_line ALTER COLUMN rejected_base_qty SET NOT NULL;
				ALTER TABLE receipt_line ALTER COLUMN base_uom_id SET NOT NULL;
			END IF;
		END $$`,
	}
	for _, statement := range statements {
		if err := tx.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

// MigrateQuantitySnapshots is a separately versioned upgrade for installations
// that already recorded the original inbound migration.
func MigrateQuantitySnapshots(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := migrateQuantitySnapshots(tx); err != nil {
			return err
		}
		for _, statement := range []string{
			"DO $$ BEGIN ALTER TABLE purchase_order_line ADD CONSTRAINT ck_purchase_order_line_snapshot CHECK (uom_conversion_to_base>0 AND ordered_base_qty>0 AND ordered_base_qty=round(ordered_qty*uom_conversion_to_base,6)); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inbound_order_line ADD CONSTRAINT ck_inbound_line_snapshot CHECK (uom_conversion_to_base>0 AND expected_base_qty>0 AND expected_base_qty=round(expected_qty*uom_conversion_to_base,6)); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt_line ADD CONSTRAINT ck_receipt_line_snapshot CHECK (uom_conversion_to_base>0 AND received_base_qty>0 AND rejected_base_qty>=0 AND rejected_base_qty<=received_base_qty AND received_base_qty=round(received_qty*uom_conversion_to_base,6) AND rejected_base_qty=round(rejected_qty*uom_conversion_to_base,6)); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE purchase_order_line ADD CONSTRAINT fk_po_line_base_uom FOREIGN KEY (base_uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inbound_order_line ADD CONSTRAINT fk_inbound_line_base_uom FOREIGN KEY (base_uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt_line ADD CONSTRAINT fk_receipt_line_base_uom FOREIGN KEY (base_uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
		} {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := migrateQuantitySnapshots(tx); err != nil {
			return err
		}
		if err := tx.AutoMigrate(&model.PurchaseOrder{}, &model.PurchaseOrderLine{}, &model.InboundOrder{}, &model.InboundOrderLine{}, &model.Receipt{}, &model.ReceiptLine{}, &model.ReceiptInventory{}, &model.ReceiptLineSerial{}, &model.QualityInspection{}, &model.PutawayTask{}, &model.QuarantineDispositionType{}, &model.QuarantineCase{}, &model.QuarantineDisposition{}, &model.InboundException{}, &model.ReworkTask{}); err != nil {
			return err
		}
		statements := []string{
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_purchase_order_successor ON purchase_order(supersedes_purchase_order_id) WHERE supersedes_purchase_order_id IS NOT NULL",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_inbound_order_successor ON inbound_order(supersedes_inbound_id) WHERE supersedes_inbound_id IS NOT NULL",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_receipt_successor ON receipt(supersedes_receipt_id) WHERE supersedes_receipt_id IS NOT NULL",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_purchase_order_id_owner ON purchase_order(purchase_order_id,owner_id)",
			"CREATE INDEX IF NOT EXISTS ix_purchase_order_owner_vendor_date ON purchase_order(owner_id,vendor_id,business_date DESC)",
			"CREATE INDEX IF NOT EXISTS ix_inbound_owner_warehouse_date ON inbound_order(owner_id,warehouse_id,business_date DESC)",
			"CREATE INDEX IF NOT EXISTS ix_receipt_owner_warehouse_date ON receipt(owner_id,warehouse_id,business_date DESC)",
			"CREATE INDEX IF NOT EXISTS ix_receipt_inventory_line ON receipt_inventory(receipt_line_id,item_id)",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_quality_inspection_receipt_inventory ON quality_inspection(receipt_inventory_id) WHERE parent_inspection_id IS NULL",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_putaway_task_inspection ON putaway_task(inspection_id)",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_putaway_task_movement ON putaway_task(inventory_movement_id) WHERE inventory_movement_id IS NOT NULL",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_putaway_task_reversal_movement ON putaway_task(reversal_movement_id) WHERE reversal_movement_id IS NOT NULL",
			"CREATE INDEX IF NOT EXISTS ix_putaway_owner_status ON putaway_task(owner_id,warehouse_id,task_status_id)",
			"CREATE INDEX IF NOT EXISTS ix_quarantine_owner_status ON quarantine_case(owner_id,warehouse_id,status_id)",
			"CREATE INDEX IF NOT EXISTS ix_quarantine_disposition_case ON quarantine_disposition(quarantine_case_id,decided_at)",
			"CREATE INDEX IF NOT EXISTS ix_inbound_exception_source ON inbound_exception(owner_id,warehouse_id,source_document_id,created_at DESC)",
			"CREATE INDEX IF NOT EXISTS ix_rework_owner_status ON rework_task(task_status_id,created_at DESC)",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_quarantine_case_inspection ON quarantine_case(inspection_id)",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_quarantine_disposition_movement ON quarantine_disposition(inventory_movement_id) WHERE inventory_movement_id IS NOT NULL",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_quarantine_disposition_type_code ON quarantine_disposition_type(code)",
			"DO $$ BEGIN ALTER TABLE purchase_order ADD CONSTRAINT ck_purchase_order_version CHECK (version_no>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE purchase_order_line ADD CONSTRAINT ck_purchase_order_line_qty CHECK (line_no>0 AND ordered_qty>0 AND over_receipt_tolerance_pct>=0 AND over_receipt_tolerance_pct<=100 AND under_receipt_tolerance_pct>=0 AND under_receipt_tolerance_pct<=100); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE purchase_order_line ADD CONSTRAINT ck_purchase_order_line_tolerance CHECK (over_receipt_tolerance_pct>=0 AND over_receipt_tolerance_pct<=100 AND under_receipt_tolerance_pct>=0 AND under_receipt_tolerance_pct<=100); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE purchase_order_line ADD CONSTRAINT ck_purchase_order_line_snapshot CHECK (uom_conversion_to_base>0 AND ordered_base_qty>0 AND ordered_base_qty=round(ordered_qty*uom_conversion_to_base,6)); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inbound_order ADD CONSTRAINT ck_inbound_version CHECK (version_no>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inbound_order_line ADD CONSTRAINT ck_inbound_line_qty CHECK (line_no>0 AND expected_qty>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inbound_order_line ADD CONSTRAINT ck_inbound_line_snapshot CHECK (uom_conversion_to_base>0 AND expected_base_qty>0 AND expected_base_qty=round(expected_qty*uom_conversion_to_base,6)); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt ADD CONSTRAINT ck_receipt_version CHECK (version_no>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt_line ADD CONSTRAINT ck_receipt_line_qty CHECK (line_no>0 AND received_qty>0 AND rejected_qty>=0 AND rejected_qty<=received_qty); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt_line ADD CONSTRAINT ck_receipt_line_snapshot CHECK (uom_conversion_to_base>0 AND received_base_qty>0 AND rejected_base_qty>=0 AND rejected_base_qty<=received_base_qty AND received_base_qty=round(received_qty*uom_conversion_to_base,6) AND rejected_base_qty=round(rejected_qty*uom_conversion_to_base,6)); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt_inventory ADD CONSTRAINT ck_receipt_inventory_qty CHECK (source_qty>0 AND base_qty>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE quality_inspection ADD CONSTRAINT ck_quality_inspection_qty CHECK (inspected_qty>0 AND passed_qty>=0 AND failed_qty>=0 AND passed_qty+failed_qty<=inspected_qty); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE quality_inspection ADD CONSTRAINT ck_quality_inspection_version CHECK (version_no>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE putaway_task ADD CONSTRAINT ck_putaway_task_qty CHECK (planned_qty>0 AND completed_qty>=0 AND completed_qty<=planned_qty); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE putaway_task ADD CONSTRAINT ck_putaway_task_version CHECK (version_no>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE quarantine_disposition_type ADD CONSTRAINT ck_quarantine_disposition_action CHECK (((releases_to_available)::int+(requires_reinspection)::int+(removes_inventory)::int)=1); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE quarantine_case ADD CONSTRAINT ck_quarantine_case_qty CHECK (quarantine_qty>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE quarantine_case ADD CONSTRAINT ck_quarantine_case_version CHECK (version_no>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE quarantine_disposition ADD CONSTRAINT ck_quarantine_disposition_qty CHECK (disposition_qty>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN IF EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid='inbound_exception'::regclass AND conname='ck_inbound_exception_type' AND pg_get_constraintdef(oid) NOT LIKE '%WRONG_ITEM%') THEN ALTER TABLE inbound_exception DROP CONSTRAINT ck_inbound_exception_type; END IF; BEGIN ALTER TABLE inbound_exception ADD CONSTRAINT ck_inbound_exception_type CHECK (exception_type_code IN ('OVER_RECEIPT','UNDER_RECEIPT','REJECTED_AT_DOCK','DAMAGED','WRONG_ITEM','CANCELLATION','REVERSAL')); EXCEPTION WHEN duplicate_object THEN NULL; END; END $$",
			"DO $$ BEGIN ALTER TABLE rework_task ADD CONSTRAINT ck_rework_task_qty CHECK (planned_qty>0 AND completed_qty>=0 AND completed_qty<=planned_qty); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE rework_task ADD CONSTRAINT ck_rework_task_version CHECK (version_no>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE purchase_order_line ADD CONSTRAINT fk_po_line_header FOREIGN KEY (purchase_order_id) REFERENCES purchase_order(purchase_order_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE purchase_order ADD CONSTRAINT fk_purchase_order_supersedes FOREIGN KEY (supersedes_purchase_order_id) REFERENCES purchase_order(purchase_order_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inbound_order_line ADD CONSTRAINT fk_inbound_line_header FOREIGN KEY (inbound_id) REFERENCES inbound_order(inbound_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inbound_order ADD CONSTRAINT fk_inbound_order_supersedes FOREIGN KEY (supersedes_inbound_id) REFERENCES inbound_order(inbound_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inbound_order_line ADD CONSTRAINT fk_inbound_line_po_line FOREIGN KEY (purchase_order_line_id) REFERENCES purchase_order_line(purchase_order_line_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt ADD CONSTRAINT fk_receipt_inbound FOREIGN KEY (inbound_id) REFERENCES inbound_order(inbound_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt ADD CONSTRAINT fk_receipt_supersedes FOREIGN KEY (supersedes_receipt_id) REFERENCES receipt(receipt_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt_line ADD CONSTRAINT fk_receipt_line_header FOREIGN KEY (receipt_id) REFERENCES receipt(receipt_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt_line ADD CONSTRAINT fk_receipt_line_inbound FOREIGN KEY (inbound_line_id) REFERENCES inbound_order_line(inbound_line_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE purchase_order_line ADD CONSTRAINT fk_po_line_base_uom FOREIGN KEY (base_uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inbound_order_line ADD CONSTRAINT fk_inbound_line_base_uom FOREIGN KEY (base_uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt_line ADD CONSTRAINT fk_receipt_line_base_uom FOREIGN KEY (base_uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt_inventory ADD CONSTRAINT fk_receipt_inventory_line FOREIGN KEY (receipt_line_id) REFERENCES receipt_line(receipt_line_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt_inventory ADD CONSTRAINT fk_receipt_inventory_balance FOREIGN KEY (initial_balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt_line_serial ADD CONSTRAINT fk_receipt_serial_batch FOREIGN KEY (receipt_inventory_id) REFERENCES receipt_inventory(receipt_inventory_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt_line_serial ADD CONSTRAINT fk_receipt_serial_identity FOREIGN KEY (serial_id) REFERENCES serial_number(serial_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE quality_inspection ADD CONSTRAINT fk_quality_inspection_batch FOREIGN KEY (receipt_inventory_id) REFERENCES receipt_inventory(receipt_inventory_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE quality_inspection ADD CONSTRAINT fk_quality_inspection_parent FOREIGN KEY (parent_inspection_id) REFERENCES quality_inspection(inspection_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE quality_inspection ADD CONSTRAINT fk_quality_inspection_source_balance FOREIGN KEY (source_balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE quality_inspection ADD CONSTRAINT fk_quality_inspection_status FOREIGN KEY (quality_status_id) REFERENCES quality_status(quality_status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE quality_inspection ADD CONSTRAINT fk_quality_inspection_result FOREIGN KEY (inspection_result_id) REFERENCES inspection_result(inspection_result_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE putaway_task ADD CONSTRAINT fk_putaway_inspection FOREIGN KEY (inspection_id) REFERENCES quality_inspection(inspection_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE putaway_task ADD CONSTRAINT fk_putaway_batch FOREIGN KEY (receipt_inventory_id) REFERENCES receipt_inventory(receipt_inventory_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE putaway_task ADD CONSTRAINT fk_putaway_task_type FOREIGN KEY (task_type_id) REFERENCES task_type(task_type_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE putaway_task ADD CONSTRAINT fk_putaway_task_status FOREIGN KEY (task_status_id) REFERENCES task_status(task_status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE putaway_task ADD CONSTRAINT fk_putaway_task_priority FOREIGN KEY (task_priority_id) REFERENCES task_priority(task_priority_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE putaway_task ADD CONSTRAINT fk_putaway_source_balance FOREIGN KEY (source_balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE putaway_task ADD CONSTRAINT fk_putaway_source_location FOREIGN KEY (source_location_id) REFERENCES warehouse_location(location_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE putaway_task ADD CONSTRAINT fk_putaway_target_location FOREIGN KEY (target_location_id) REFERENCES warehouse_location(location_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE putaway_task ADD CONSTRAINT fk_putaway_movement FOREIGN KEY (inventory_movement_id) REFERENCES inventory_movement(movement_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE putaway_task ADD CONSTRAINT fk_putaway_result_balance FOREIGN KEY (resulting_balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE putaway_task ADD CONSTRAINT fk_putaway_reversal_movement FOREIGN KEY (reversal_movement_id) REFERENCES inventory_movement(movement_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE putaway_task ADD CONSTRAINT fk_putaway_replacement_inspection FOREIGN KEY (replacement_inspection_id) REFERENCES quality_inspection(inspection_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE quarantine_disposition_type ADD CONSTRAINT fk_quarantine_type_movement FOREIGN KEY (removal_movement_type_id) REFERENCES movement_type(movement_type_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE quarantine_case ADD CONSTRAINT fk_quarantine_case_batch FOREIGN KEY (receipt_inventory_id) REFERENCES receipt_inventory(receipt_inventory_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE quarantine_case ADD CONSTRAINT fk_quarantine_case_inspection FOREIGN KEY (inspection_id) REFERENCES quality_inspection(inspection_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE quarantine_case ADD CONSTRAINT fk_quarantine_case_balance FOREIGN KEY (quarantine_balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE quarantine_disposition ADD CONSTRAINT fk_quarantine_disposition_case FOREIGN KEY (quarantine_case_id) REFERENCES quarantine_case(quarantine_case_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE quarantine_disposition ADD CONSTRAINT fk_quarantine_disposition_type FOREIGN KEY (quarantine_disposition_type_id) REFERENCES quarantine_disposition_type(quarantine_disposition_type_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE quarantine_disposition ADD CONSTRAINT fk_quarantine_disposition_target FOREIGN KEY (target_location_id) REFERENCES warehouse_location(location_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inbound_exception ADD CONSTRAINT fk_inbound_exception_owner FOREIGN KEY (owner_id) REFERENCES organization(organization_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inbound_exception ADD CONSTRAINT fk_inbound_exception_warehouse FOREIGN KEY (warehouse_id) REFERENCES warehouse(warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inbound_exception ADD CONSTRAINT fk_inbound_exception_actor FOREIGN KEY (created_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE rework_task ADD CONSTRAINT fk_rework_disposition FOREIGN KEY (quarantine_disposition_id) REFERENCES quarantine_disposition(quarantine_disposition_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE rework_task ADD CONSTRAINT fk_rework_task_type FOREIGN KEY (task_type_id) REFERENCES task_type(task_type_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE rework_task ADD CONSTRAINT fk_rework_task_status FOREIGN KEY (task_status_id) REFERENCES task_status(task_status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE rework_task ADD CONSTRAINT fk_rework_task_priority FOREIGN KEY (task_priority_id) REFERENCES task_priority(task_priority_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE rework_task ADD CONSTRAINT fk_rework_source_balance FOREIGN KEY (source_balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE rework_task ADD CONSTRAINT fk_rework_inspection FOREIGN KEY (reinspection_id) REFERENCES quality_inspection(inspection_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
		}
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func SeedReferenceData(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		description := func(value string) *string { return &value }
		repository := NewQuarantineDispositionTypeRepository(tx)
		if err := repository.Seed(context.Background(), []model.QuarantineDispositionType{
			{Code: "ACCEPT", Name: "Accept", Description: description("Client accepts quarantined stock for storage."), ReleasesToAvailable: true, IsActive: true},
			{Code: "REWORK", Name: "Rework", Description: description("Stock must be reworked and inspected again."), RequiresReinspection: true, IsActive: true},
			{Code: "RETURN", Name: "Return", Description: description("Return stock to the vendor or factory."), RemovesInventory: true, IsActive: true},
			{Code: "DISPOSE", Name: "Dispose", Description: description("Destroy or otherwise dispose of stock."), RemovesInventory: true, IsActive: true},
		}); err != nil {
			return err
		}
		if err := repository.LinkRemovalMovement(context.Background(), "RETURN", "RETURN_TO_VENDOR"); err != nil {
			return err
		}
		return repository.LinkRemovalMovement(context.Background(), "DISPOSE", "DISPOSE")
	})
}
