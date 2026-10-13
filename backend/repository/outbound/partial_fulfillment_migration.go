package outbound

import "gorm.io/gorm"

// Preserve original demand and historical document membership. Fulfilled shipment
// links are history, not cancelled/removed links, and no longer claim the next cycle.
func MigratePartialFulfillment(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, sql := range []string{
			`ALTER TABLE outbound_order ADD COLUMN IF NOT EXISTS closed_short_at timestamptz, ADD COLUMN IF NOT EXISTS closed_short_by uuid, ADD COLUMN IF NOT EXISTS closed_short_reason text`,
			`ALTER TABLE outbound_order_line ADD COLUMN IF NOT EXISTS closed_short_qty numeric(20,6) NOT NULL DEFAULT 0`,
			`DO $$ BEGIN ALTER TABLE outbound_order_line ADD CONSTRAINT ck_outbound_closed_short CHECK (closed_short_qty>=0 AND allocated_qty+closed_short_qty<=ordered_base_qty AND shipped_qty+short_accepted_qty+closed_short_qty<=ordered_base_qty); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE outbound_order ADD CONSTRAINT ck_outbound_close_audit CHECK ((closed_short_at IS NULL AND closed_short_by IS NULL AND closed_short_reason IS NULL) OR (closed_short_at IS NOT NULL AND closed_short_by IS NOT NULL AND closed_short_reason IS NOT NULL AND length(btrim(closed_short_reason))>0)); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE outbound_order ADD CONSTRAINT fk_outbound_closed_short_by FOREIGN KEY(closed_short_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`ALTER TABLE packing ADD COLUMN IF NOT EXISTS outbound_check_id varchar(120)`,
			`DO $$ BEGIN ALTER TABLE packing ADD CONSTRAINT fk_packing_source_check FOREIGN KEY(outbound_check_id) REFERENCES outbound_check(outbound_check_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN IF EXISTS (SELECT 1 FROM packing_line pl JOIN outbound_check_line cl ON cl.outbound_check_line_id=pl.outbound_check_line_id GROUP BY pl.packing_id HAVING count(DISTINCT cl.outbound_check_id)>1) THEN RAISE EXCEPTION 'Legacy packing spans multiple checks; reconcile its source before partial fulfillment migration'; END IF; END $$`,
			// Posted lines identify the legacy source; empty packing documents use
			// the most recent passed check that existed when packing was created.
			`UPDATE packing p SET outbound_check_id=(SELECT cl.outbound_check_id FROM packing_line pl JOIN outbound_check_line cl ON cl.outbound_check_line_id=pl.outbound_check_line_id WHERE pl.packing_id=p.packing_id ORDER BY cl.created_at DESC LIMIT 1) WHERE p.outbound_check_id IS NULL`,
			`UPDATE packing p SET outbound_check_id=(SELECT c.outbound_check_id FROM outbound_check c JOIN outbound_staging st ON st.staging_id=c.staging_id JOIN document_status ds ON ds.status_id=c.status_id WHERE st.outbound_id=p.outbound_id AND ds.code='PASSED' AND c.created_at<=p.created_at ORDER BY c.created_at DESC LIMIT 1) WHERE p.outbound_check_id IS NULL`,
			`ALTER TABLE shipment_order ADD COLUMN IF NOT EXISTS fulfilled_at timestamptz`,
			`UPDATE shipment_order so SET fulfilled_at=COALESCE(s.shipped_at,s.created_at) FROM shipment s JOIN document_status ds ON ds.status_id=s.status_id WHERE so.shipment_id=s.shipment_id AND ds.code='SHIPPED' AND so.fulfilled_at IS NULL`,
			`DROP INDEX IF EXISTS uq_active_shipment_order`,
			`CREATE UNIQUE INDEX uq_active_shipment_order ON shipment_order(outbound_id) WHERE removed_at IS NULL AND fulfilled_at IS NULL`,
		} {
			if err := tx.Exec(sql).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
