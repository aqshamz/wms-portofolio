package stockcontrol

import (
	"gorm.io/gorm"
	inventoryrepo "wms-api/repository/inventory"
)

func MigrateGrandStockOpnameFreeze(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := inventoryrepo.MigrateStockFreezes(tx); err != nil {
			return err
		}
		for _, statement := range []string{
			`DO $$ BEGIN ALTER TABLE inventory_stock_freeze ADD CONSTRAINT fk_stock_freeze_count FOREIGN KEY(cycle_count_id) REFERENCES cycle_count(cycle_count_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_stock_freeze ADD CONSTRAINT fk_stock_freeze_owner FOREIGN KEY(owner_id) REFERENCES organization(organization_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE inventory_stock_freeze ADD CONSTRAINT fk_stock_freeze_warehouse FOREIGN KEY(warehouse_id) REFERENCES warehouse(warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`INSERT INTO inventory_stock_freeze(owner_id,warehouse_id,cycle_count_id,frozen_at)
			 SELECT owner_id,warehouse_id,cycle_count_id,clock_timestamp() FROM cycle_count
			 WHERE count_type_code='GRAND' AND completed_at IS NULL AND cancelled_at IS NULL
			 ON CONFLICT (owner_id,warehouse_id) DO NOTHING`,
		} {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
