package config

import (
	"strings"
	"testing"
)

func TestProductionConfigurationFailsClosed(t *testing.T) {
	setRequiredConfiguration(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("AUTH_RBAC_ENFORCED", "false")
	t.Setenv("DB_AUTO_CREATE", "false")
	t.Setenv("DB_SSLMODE", "require")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "AUTH_RBAC_ENFORCED") {
		t.Fatalf("expected production RBAC validation, got %v", err)
	}
}

func TestProductionConfigurationAcceptsRestrictedSettings(t *testing.T) {
	setRequiredConfiguration(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("AUTH_RBAC_ENFORCED", "true")
	t.Setenv("DB_AUTO_CREATE", "false")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://wms.example, https://admin.example/")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load production config: %v", err)
	}
	if len(cfg.Security.AllowedOrigins) != 2 || cfg.Security.AllowedOrigins[1] != "https://admin.example" {
		t.Fatalf("unexpected normalized origins: %#v", cfg.Security.AllowedOrigins)
	}
}

func TestProductionConfigurationRequiresExplicitOrigins(t *testing.T) {
	setRequiredConfiguration(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("AUTH_RBAC_ENFORCED", "true")
	t.Setenv("DB_AUTO_CREATE", "false")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "CORS_ALLOWED_ORIGINS") {
		t.Fatalf("expected explicit production origins validation, got %v", err)
	}
}

func TestConfigurationRejectsOriginWithPath(t *testing.T) {
	setRequiredConfiguration(t)
	t.Setenv("APP_ENV", "development")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://wms.example/not-an-origin")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "invalid CORS origin") {
		t.Fatalf("expected origin validation, got %v", err)
	}
}

func setRequiredConfiguration(t *testing.T) {
	t.Helper()
	t.Setenv("DB_PASSWORD", "test-password")
	t.Setenv("AUTH_BOOTSTRAP_ADMIN_ENABLED", "false")
	t.Setenv("SECURITY_RATE_LIMIT_PER_MINUTE", "600")
	t.Setenv("SECURITY_MAX_REQUEST_BODY_BYTES", "1048576")
}
