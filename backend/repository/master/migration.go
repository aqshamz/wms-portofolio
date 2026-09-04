package master

import (
	model "wms-api/models/master"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.Organization{},
		&model.Warehouse{},
		&model.WarehouseOwner{},
		&model.LocationType{},
		&model.WarehouseZone{},
		&model.WarehouseLocation{},
		&model.AccountOwnerAccess{},
		&model.AccountWarehouseAccess{},
	); err != nil {
		return err
	}

	statements := []string{
		`DO $$ BEGIN
			ALTER TABLE organization ADD CONSTRAINT fk_organization_created_by
			FOREIGN KEY (created_by) REFERENCES app_account(account_id);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE organization ADD CONSTRAINT fk_organization_updated_by
			FOREIGN KEY (updated_by) REFERENCES app_account(account_id);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE warehouse ADD CONSTRAINT fk_warehouse_operator
			FOREIGN KEY (operator_id) REFERENCES organization(organization_id);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE warehouse ADD CONSTRAINT fk_warehouse_created_by
			FOREIGN KEY (created_by) REFERENCES app_account(account_id);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE warehouse ADD CONSTRAINT fk_warehouse_updated_by
			FOREIGN KEY (updated_by) REFERENCES app_account(account_id);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE warehouse_owner ADD CONSTRAINT fk_warehouse_owner_warehouse
			FOREIGN KEY (warehouse_id) REFERENCES warehouse(warehouse_id);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE warehouse_owner ADD CONSTRAINT fk_warehouse_owner_owner
			FOREIGN KEY (owner_id) REFERENCES organization(organization_id);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE warehouse_owner ADD CONSTRAINT fk_warehouse_owner_created_by
			FOREIGN KEY (created_by) REFERENCES app_account(account_id);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE warehouse_zone ADD CONSTRAINT fk_warehouse_zone_warehouse
			FOREIGN KEY (warehouse_id) REFERENCES warehouse(warehouse_id);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE warehouse_zone ADD CONSTRAINT fk_warehouse_zone_created_by
			FOREIGN KEY (created_by) REFERENCES app_account(account_id);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE warehouse_location ADD CONSTRAINT fk_location_zone_warehouse
			FOREIGN KEY (zone_id, warehouse_id) REFERENCES warehouse_zone(zone_id, warehouse_id);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE warehouse_location ADD CONSTRAINT fk_location_type
			FOREIGN KEY (location_type_id) REFERENCES location_type(location_type_id);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE warehouse_location ADD CONSTRAINT fk_location_created_by
			FOREIGN KEY (created_by) REFERENCES app_account(account_id);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE warehouse_location ADD CONSTRAINT ck_location_capacity CHECK (
				(max_weight IS NULL OR max_weight >= 0) AND
				(max_volume IS NULL OR max_volume >= 0) AND pick_sequence >= 0
			);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE account_owner_access ADD CONSTRAINT fk_account_owner_access_account
			FOREIGN KEY (account_id) REFERENCES app_account(account_id) ON DELETE CASCADE;
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE account_owner_access ADD CONSTRAINT fk_account_owner_access_owner
			FOREIGN KEY (owner_id) REFERENCES organization(organization_id) ON DELETE CASCADE;
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE account_owner_access ADD CONSTRAINT fk_account_owner_access_granted_by
			FOREIGN KEY (granted_by) REFERENCES app_account(account_id);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE account_warehouse_access ADD CONSTRAINT fk_account_warehouse_access_account
			FOREIGN KEY (account_id) REFERENCES app_account(account_id) ON DELETE CASCADE;
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE account_warehouse_access ADD CONSTRAINT fk_account_warehouse_access_warehouse
			FOREIGN KEY (warehouse_id) REFERENCES warehouse(warehouse_id) ON DELETE CASCADE;
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE account_warehouse_access ADD CONSTRAINT fk_account_warehouse_access_granted_by
			FOREIGN KEY (granted_by) REFERENCES app_account(account_id);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}
