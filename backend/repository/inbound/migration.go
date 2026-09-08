package inbound

import (
	"context"

	model "wms-api/models/inbound"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.PurchaseOrder{}, &model.PurchaseOrderLine{}, &model.InboundOrder{}, &model.InboundOrderLine{}, &model.Receipt{}, &model.ReceiptLine{}, &model.ReceiptInventory{}, &model.ReceiptLineSerial{}, &model.QualityInspection{}, &model.PutawayTask{}, &model.QuarantineDispositionType{}, &model.QuarantineCase{}, &model.QuarantineDisposition{}, &model.InboundException{}, &model.ReworkTask{}); err != nil {
			return err
		}
		statements := []string{
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
			"DO $$ BEGIN ALTER TABLE inbound_order ADD CONSTRAINT ck_inbound_version CHECK (version_no>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inbound_order_line ADD CONSTRAINT ck_inbound_line_qty CHECK (line_no>0 AND expected_qty>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt ADD CONSTRAINT ck_receipt_version CHECK (version_no>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt_line ADD CONSTRAINT ck_receipt_line_qty CHECK (line_no>0 AND received_qty>0 AND rejected_qty>=0 AND rejected_qty<=received_qty); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
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
			"DO $$ BEGIN ALTER TABLE inbound_order_line ADD CONSTRAINT fk_inbound_line_header FOREIGN KEY (inbound_id) REFERENCES inbound_order(inbound_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inbound_order_line ADD CONSTRAINT fk_inbound_line_po_line FOREIGN KEY (purchase_order_line_id) REFERENCES purchase_order_line(purchase_order_line_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt ADD CONSTRAINT fk_receipt_inbound FOREIGN KEY (inbound_id) REFERENCES inbound_order(inbound_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt_line ADD CONSTRAINT fk_receipt_line_header FOREIGN KEY (receipt_id) REFERENCES receipt(receipt_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE receipt_line ADD CONSTRAINT fk_receipt_line_inbound FOREIGN KEY (inbound_line_id) REFERENCES inbound_order_line(inbound_line_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
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
