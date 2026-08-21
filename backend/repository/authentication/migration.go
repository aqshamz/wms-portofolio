package authentication

import (
	model "wms-api/models/authentication"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.AccountStatus{},
		&model.AuthenticationPolicy{},
		&model.SessionRevocationReason{},
		&model.AppAccount{},
		&model.AppSession{},
	); err != nil {
		return err
	}

	statements := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_authentication_policy_default
			ON authentication_policy (is_default) WHERE is_default AND is_active`,
		`DO $$ BEGIN
			ALTER TABLE authentication_policy ADD CONSTRAINT ck_auth_policy_values
			CHECK (max_failed_attempts > 0 AND lockout_seconds > 0 AND session_ttl_seconds > 0);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE app_account ADD CONSTRAINT fk_app_account_status
			FOREIGN KEY (account_status_id) REFERENCES account_status(account_status_id);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE app_account ADD CONSTRAINT fk_app_account_auth_policy
			FOREIGN KEY (authentication_policy_id) REFERENCES authentication_policy(authentication_policy_id);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE app_account ADD CONSTRAINT ck_account_identity
			CHECK (password_hash IS NOT NULL OR external_subject IS NOT NULL);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE app_account ADD CONSTRAINT ck_account_failed_login_count
			CHECK (failed_login_count >= 0);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE app_account ADD CONSTRAINT ck_account_version CHECK (version_no > 0);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE app_session ADD CONSTRAINT fk_app_session_account
			FOREIGN KEY (account_id) REFERENCES app_account(account_id) ON DELETE CASCADE;
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE app_session ADD CONSTRAINT fk_app_session_revocation_reason
			FOREIGN KEY (session_revocation_reason_id)
			REFERENCES session_revocation_reason(session_revocation_reason_id);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE app_session ADD CONSTRAINT ck_session_period CHECK (expires_at > issued_at);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		`DO $$ BEGIN
			ALTER TABLE app_session ADD CONSTRAINT ck_session_revocation
			CHECK (revoked_at IS NULL OR revoked_at >= issued_at);
		EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
	}

	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}
