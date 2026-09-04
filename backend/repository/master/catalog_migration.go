package master

import (
	"gorm.io/gorm"
	model "wms-api/models/master"
)

// MigrateCatalog is additive and follows the existing SQL schema's dependency order.
func MigrateCatalog(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.PartnerType{}, &model.BusinessPartner{}, &model.BusinessPartnerType{}, &model.UOM{}, &model.ItemCategory{}, &model.Item{}, &model.ItemUOM{}, &model.ItemBarcode{}, &model.InventoryStatus{}, &model.QualityStatus{}, &model.InspectionResult{}); err != nil {
			return err
		}
		statements := []string{
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_catalog_partner_type_code ON partner_type (code)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_catalog_partner_owner_code ON business_partner (owner_id, code)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_catalog_partner_owner_id ON business_partner (owner_id, partner_id)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_catalog_uom_code ON uom (code)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_catalog_category_owner_code ON item_category (owner_id, code)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_catalog_category_owner_id ON item_category (owner_id, category_id)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_catalog_item_owner_code ON item (owner_id, code)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_catalog_item_owner_id ON item (owner_id, item_id)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_catalog_item_unit ON item_uom (item_id, uom_id)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_catalog_barcode ON item_barcode (barcode)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_item_primary_barcode ON item_barcode (item_id) WHERE is_primary AND is_active`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_catalog_inventory_status_code ON inventory_status (code)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_catalog_quality_status_code ON quality_status (code)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_catalog_inspection_result_code ON inspection_result (code)`,
			`DO $$ BEGIN ALTER TABLE business_partner ADD CONSTRAINT fk_catalog_business_partner_owner_id FOREIGN KEY (owner_id) REFERENCES organization(organization_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE business_partner ADD CONSTRAINT fk_catalog_business_partner_created_by FOREIGN KEY (created_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE business_partner ADD CONSTRAINT fk_catalog_business_partner_updated_by FOREIGN KEY (updated_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE item ADD CONSTRAINT fk_catalog_item_created_by FOREIGN KEY (created_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE item ADD CONSTRAINT fk_catalog_item_updated_by FOREIGN KEY (updated_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE business_partner_type ADD CONSTRAINT fk_catalog_business_partner_type_partner_id FOREIGN KEY (partner_id) REFERENCES business_partner(partner_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE business_partner_type ADD CONSTRAINT fk_catalog_business_partner_type_partner_type_id FOREIGN KEY (partner_type_id) REFERENCES partner_type(partner_type_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE item_category ADD CONSTRAINT fk_catalog_item_category_owner_id FOREIGN KEY (owner_id) REFERENCES organization(organization_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE item_category ADD CONSTRAINT fk_catalog_item_category_owner_id_parent_category_id FOREIGN KEY (owner_id, parent_category_id) REFERENCES item_category(owner_id, category_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE item ADD CONSTRAINT fk_catalog_item_owner_id FOREIGN KEY (owner_id) REFERENCES organization(organization_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE item ADD CONSTRAINT fk_catalog_item_owner_id_category_id FOREIGN KEY (owner_id, category_id) REFERENCES item_category(owner_id, category_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE item ADD CONSTRAINT fk_catalog_item_base_uom_id FOREIGN KEY (base_uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE item_uom ADD CONSTRAINT fk_catalog_item_uom_item_id FOREIGN KEY (item_id) REFERENCES item(item_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE item_uom ADD CONSTRAINT fk_catalog_item_uom_uom_id FOREIGN KEY (uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE item_barcode ADD CONSTRAINT fk_catalog_item_barcode_item_id FOREIGN KEY (item_id) REFERENCES item(item_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE item_barcode ADD CONSTRAINT fk_catalog_item_barcode_uom_id FOREIGN KEY (uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE item_barcode ADD CONSTRAINT fk_catalog_item_barcode_item_id_uom_id FOREIGN KEY (item_id, uom_id) REFERENCES item_uom(item_id, uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE uom ADD CONSTRAINT ck_uom_scale CHECK (decimal_scale BETWEEN 0 AND 6); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE item_category ADD CONSTRAINT ck_category_not_own_parent CHECK (parent_category_id IS NULL OR parent_category_id <> category_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE item ADD CONSTRAINT ck_item_measurements CHECK ((weight IS NULL OR weight >= 0) AND (volume IS NULL OR volume >= 0) AND (shelf_life_days IS NULL OR shelf_life_days >= 0) AND (minimum_receive_days IS NULL OR minimum_receive_days >= 0)); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE item_uom ADD CONSTRAINT ck_item_uom_conversion CHECK (conversion_to_base > 0); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE item_uom ADD CONSTRAINT ck_catalog_item_uom_dimensions CHECK ((length IS NULL OR length >= 0) AND (width IS NULL OR width >= 0) AND (height IS NULL OR height >= 0) AND (weight IS NULL OR weight >= 0)); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		}
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
