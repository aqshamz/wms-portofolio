package authentication

import (
	model "wms-api/models/authentication"

	"gorm.io/gorm"
)

func MigrateAdministration(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.AppRole{}, &model.RolePermission{}, &model.AccountRole{}); err != nil {
		return err
	}
	statements := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_app_account_username_ci ON app_account(lower(username))`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_app_account_email_ci ON app_account(lower(email)) WHERE email IS NOT NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_app_role_code_ci ON app_role(lower(code))`,
		`DO $$ BEGIN ALTER TABLE app_role ADD CONSTRAINT fk_app_role_created_by FOREIGN KEY(created_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE app_role ADD CONSTRAINT fk_app_role_updated_by FOREIGN KEY(updated_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE app_role ADD CONSTRAINT ck_app_role_version CHECK(version_no > 0); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE role_permission ADD CONSTRAINT fk_role_permission_role FOREIGN KEY(role_id) REFERENCES app_role(role_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE role_permission ADD CONSTRAINT fk_role_permission_permission FOREIGN KEY(permission_id) REFERENCES app_permission(permission_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE role_permission ADD CONSTRAINT fk_role_permission_granted_by FOREIGN KEY(granted_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE account_role ADD CONSTRAINT fk_account_role_account FOREIGN KEY(account_id) REFERENCES app_account(account_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE account_role ADD CONSTRAINT fk_account_role_role FOREIGN KEY(role_id) REFERENCES app_role(role_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE account_role ADD CONSTRAINT fk_account_role_assigned_by FOREIGN KEY(assigned_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}
