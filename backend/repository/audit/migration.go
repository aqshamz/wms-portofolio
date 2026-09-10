package audit

import (
	"gorm.io/gorm"
	model "wms-api/models/audit"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.APIAuditLog{}); err != nil {
		return err
	}
	for _, statement := range []string{
		"DROP INDEX IF EXISTS uq_api_audit_request",
		"CREATE INDEX IF NOT EXISTS ix_api_audit_request ON api_audit_log(request_id)",
		"CREATE INDEX IF NOT EXISTS ix_api_audit_account_time ON api_audit_log(account_id,occurred_at DESC)",
		"CREATE INDEX IF NOT EXISTS ix_api_audit_route_time ON api_audit_log(route,occurred_at DESC)",
		`DO $$ BEGIN ALTER TABLE api_audit_log ADD CONSTRAINT fk_api_audit_account FOREIGN KEY(account_id) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE api_audit_log ADD CONSTRAINT ck_api_audit_values CHECK(status_code BETWEEN 100 AND 599 AND duration_ms>=0); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func MigrateRequestIndex(db *gorm.DB) error {
	if err := db.Exec("DROP INDEX IF EXISTS uq_api_audit_request").Error; err != nil {
		return err
	}
	return db.Exec("CREATE INDEX IF NOT EXISTS ix_api_audit_request ON api_audit_log(request_id)").Error
}
