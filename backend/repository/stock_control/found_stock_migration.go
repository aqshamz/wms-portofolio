package stockcontrol

import (
	"gorm.io/gorm"
	model "wms-api/models/stock_control"
)

func MigrateCountFoundStock(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("ALTER TABLE cycle_count_line ALTER COLUMN balance_id DROP NOT NULL").Error; err != nil {
			return err
		}
		if err := tx.AutoMigrate(&model.CycleCountLine{}); err != nil {
			return err
		}
		for _, statement := range []string{
			`DO $$ BEGIN ALTER TABLE cycle_count_line ADD CONSTRAINT ck_cycle_count_found_stock CHECK ((found_stock AND system_qty=0 AND snapshot_version_no>=0 AND (balance_id IS NOT NULL OR snapshot_version_no=0)) OR (NOT found_stock AND balance_id IS NOT NULL AND snapshot_version_no>0)); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_cycle_count_line_identity ON cycle_count_line(cycle_count_id,item_id,location_id,inventory_status_id,COALESCE(lot_id,''),COALESCE(handling_unit_id,''))`,
		} {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
