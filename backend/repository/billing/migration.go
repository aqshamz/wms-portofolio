package billing

import (
	"gorm.io/gorm"
	model "wms-api/models/billing"
)

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.BillingContract{}, &model.RateCard{}, &model.RateCardLine{}, &model.BillableEvent{}, &model.BillingRun{}, &model.BillingCharge{}, &model.Invoice{}, &model.InvoiceLine{}, &model.CreditNote{}, &model.Payment{}); err != nil {
			return err
		}
		for _, sql := range []string{
			"DROP INDEX IF EXISTS uq_billing_contract_scope_period",
			"DROP INDEX IF EXISTS uq_rate_card_contract_period",
			"CREATE INDEX IF NOT EXISTS ix_billing_contract_scope_period ON billing_contract(owner_id,warehouse_id,effective_from)",
			"CREATE INDEX IF NOT EXISTS ix_rate_card_contract_period ON rate_card(billing_contract_id,effective_from)",
			"DROP INDEX IF EXISTS uq_rate_card_line_service",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_rate_card_line_service ON rate_card_line(rate_card_id,service_code,COALESCE(movement_type_id,'00000000-0000-0000-0000-000000000000'::uuid))",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_billable_event_key ON billable_event(event_key)",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_billing_charge_event ON billing_charge(billable_event_id)",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_invoice_run ON invoice(billing_run_id)",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_invoice_line_charge ON invoice_line(billing_charge_id)",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_payment_reference ON payment(reference)",
			"DO $$ BEGIN ALTER TABLE billing_contract ADD CONSTRAINT ck_billing_contract_values CHECK(currency_code=upper(currency_code) AND char_length(currency_code)=3 AND billing_cycle IN ('WEEKLY','MONTHLY') AND payment_term_days BETWEEN 0 AND 365 AND version_no>0 AND (effective_until IS NULL OR effective_until>=effective_from)); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE rate_card ADD CONSTRAINT ck_rate_card_period CHECK(version_no>0 AND (effective_until IS NULL OR effective_until>=effective_from)); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"ALTER TABLE rate_card_line DROP CONSTRAINT IF EXISTS ck_rate_card_line_values",
			"DO $$ BEGIN ALTER TABLE rate_card_line ADD CONSTRAINT ck_rate_card_line_values CHECK(source_kind IN ('MOVEMENT','STORAGE','MANUAL') AND billing_basis IN ('QUANTITY','EVENT') AND unit_rate>=0 AND minimum_charge>=0 AND tax_percent BETWEEN 0 AND 100 AND ((source_kind='MOVEMENT' AND movement_type_id IS NOT NULL) OR (source_kind IN ('STORAGE','MANUAL') AND movement_type_id IS NULL))); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE billable_event ADD CONSTRAINT ck_billable_event_qty CHECK(quantity>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE billing_run ADD CONSTRAINT ck_billing_run_values CHECK(period_until>=period_from AND subtotal_amount>=0 AND tax_amount>=0 AND total_amount=subtotal_amount+tax_amount AND version_no>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE billing_charge ADD CONSTRAINT ck_billing_charge_values CHECK(quantity>0 AND unit_rate>=0 AND subtotal_amount>=0 AND tax_percent BETWEEN 0 AND 100 AND tax_amount>=0 AND total_amount=subtotal_amount+tax_amount); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE invoice ADD CONSTRAINT ck_invoice_values CHECK(due_date>=issue_date AND subtotal_amount>=0 AND tax_amount>=0 AND credit_amount>=0 AND total_amount>=0 AND paid_amount>=0 AND paid_amount<=total_amount AND total_amount=subtotal_amount+tax_amount-credit_amount AND version_no>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE invoice_line ADD CONSTRAINT ck_invoice_line_values CHECK(line_no>0 AND quantity>0 AND unit_rate>=0 AND subtotal_amount>=0 AND tax_amount>=0 AND total_amount=subtotal_amount+tax_amount); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE credit_note ADD CONSTRAINT ck_credit_note_amount CHECK(amount>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE payment ADD CONSTRAINT ck_payment_amount CHECK(amount>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE billing_contract ADD CONSTRAINT fk_billing_contract_status FOREIGN KEY(document_type_id,status_id) REFERENCES document_status(document_type_id,status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE billing_contract ADD CONSTRAINT fk_billing_contract_scope FOREIGN KEY(owner_id,warehouse_id) REFERENCES warehouse_owner(owner_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE rate_card ADD CONSTRAINT fk_rate_card_contract FOREIGN KEY(billing_contract_id) REFERENCES billing_contract(billing_contract_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE rate_card ADD CONSTRAINT fk_rate_card_status FOREIGN KEY(document_type_id,status_id) REFERENCES document_status(document_type_id,status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE rate_card_line ADD CONSTRAINT fk_rate_card_line_card FOREIGN KEY(rate_card_id) REFERENCES rate_card(rate_card_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE rate_card_line ADD CONSTRAINT fk_rate_card_line_movement FOREIGN KEY(movement_type_id) REFERENCES movement_type(movement_type_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE billable_event ADD CONSTRAINT fk_billable_event_status FOREIGN KEY(document_type_id,status_id) REFERENCES document_status(document_type_id,status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE billable_event ADD CONSTRAINT fk_billable_event_rate FOREIGN KEY(rate_card_line_id) REFERENCES rate_card_line(rate_card_line_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE billable_event ADD CONSTRAINT fk_billable_event_movement FOREIGN KEY(inventory_movement_id) REFERENCES inventory_movement(movement_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE billable_event ADD CONSTRAINT fk_billable_event_scope FOREIGN KEY(owner_id,warehouse_id) REFERENCES warehouse_owner(owner_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE billable_event ADD CONSTRAINT fk_billable_event_uom FOREIGN KEY(uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE billing_run ADD CONSTRAINT fk_billing_run_contract FOREIGN KEY(billing_contract_id) REFERENCES billing_contract(billing_contract_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE billing_run ADD CONSTRAINT fk_billing_run_status FOREIGN KEY(document_type_id,status_id) REFERENCES document_status(document_type_id,status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE billing_run ADD CONSTRAINT fk_billing_run_scope FOREIGN KEY(owner_id,warehouse_id) REFERENCES warehouse_owner(owner_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE billing_charge ADD CONSTRAINT fk_billing_charge_run FOREIGN KEY(billing_run_id) REFERENCES billing_run(billing_run_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE billing_charge ADD CONSTRAINT fk_billing_charge_event FOREIGN KEY(billable_event_id) REFERENCES billable_event(billable_event_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE invoice ADD CONSTRAINT fk_invoice_run FOREIGN KEY(billing_run_id) REFERENCES billing_run(billing_run_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE invoice ADD CONSTRAINT fk_invoice_status FOREIGN KEY(document_type_id,status_id) REFERENCES document_status(document_type_id,status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE invoice ADD CONSTRAINT fk_invoice_scope FOREIGN KEY(owner_id,warehouse_id) REFERENCES warehouse_owner(owner_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE invoice_line ADD CONSTRAINT fk_invoice_line_header FOREIGN KEY(invoice_id) REFERENCES invoice(invoice_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE invoice_line ADD CONSTRAINT fk_invoice_line_charge FOREIGN KEY(billing_charge_id) REFERENCES billing_charge(billing_charge_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE credit_note ADD CONSTRAINT fk_credit_note_invoice FOREIGN KEY(invoice_id) REFERENCES invoice(invoice_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE credit_note ADD CONSTRAINT fk_credit_note_status FOREIGN KEY(document_type_id,status_id) REFERENCES document_status(document_type_id,status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE payment ADD CONSTRAINT fk_payment_invoice FOREIGN KEY(invoice_id) REFERENCES invoice(invoice_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE payment ADD CONSTRAINT fk_payment_status FOREIGN KEY(document_type_id,status_id) REFERENCES document_status(document_type_id,status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
		} {
			if err := tx.Exec(sql).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
