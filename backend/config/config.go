package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Security SecurityConfig
}

type AppConfig struct {
	Environment              string
	Host                     string
	Port                     int
	TrustedProxies           []string
	ReadHeaderTimeoutSeconds int
	ReadTimeoutSeconds       int
	WriteTimeoutSeconds      int
	IdleTimeoutSeconds       int
	ShutdownTimeoutSeconds   int
}

type SecurityConfig struct {
	AllowedOrigins        []string
	AuthorizationEnforced bool
	RateLimitPerMinute    int
	MaxRequestBodyBytes   int64
}

type DatabaseConfig struct {
	Host       string
	Port       int
	User       string
	Password   string
	Name       string
	Schema     string
	SSLMode    string
	Timezone   string
	AutoCreate bool
}

type AuthConfig struct {
	BootstrapAdminEnabled     bool
	BootstrapAdminUsername    string
	BootstrapAdminEmail       string
	BootstrapAdminDisplayName string
	BootstrapAdminPassword    string
	BootstrapAdminTimezone    string
}

func Load() (Config, error) {
	// A missing .env is valid in deployed environments where variables are injected.
	_ = godotenv.Load()

	appPort, err := envInt("APP_PORT", 8080)
	if err != nil {
		return Config{}, err
	}

	dbPort, err := envInt("DB_PORT", 7654)
	if err != nil {
		return Config{}, err
	}

	autoCreate, err := envBool("DB_AUTO_CREATE", true)
	if err != nil {
		return Config{}, err
	}
	readHeaderTimeout, err := envInt("APP_READ_HEADER_TIMEOUT_SECONDS", 5)
	if err != nil {
		return Config{}, err
	}
	readTimeout, err := envInt("APP_READ_TIMEOUT_SECONDS", 30)
	if err != nil {
		return Config{}, err
	}
	writeTimeout, err := envInt("APP_WRITE_TIMEOUT_SECONDS", 30)
	if err != nil {
		return Config{}, err
	}
	idleTimeout, err := envInt("APP_IDLE_TIMEOUT_SECONDS", 60)
	if err != nil {
		return Config{}, err
	}
	shutdownTimeout, err := envInt("APP_SHUTDOWN_TIMEOUT_SECONDS", 10)
	if err != nil {
		return Config{}, err
	}
	rateLimit, err := envInt("SECURITY_RATE_LIMIT_PER_MINUTE", 600)
	if err != nil {
		return Config{}, err
	}
	maxBodyBytes, err := envInt64("SECURITY_MAX_REQUEST_BODY_BYTES", 1048576)
	if err != nil {
		return Config{}, err
	}
	authorizationEnforced, err := envBool("AUTH_RBAC_ENFORCED", false)
	if err != nil {
		return Config{}, err
	}
	environment := strings.ToLower(strings.TrimSpace(envString("APP_ENV", "development")))
	originFallback := []string{"http://localhost:3000", "http://localhost:5173"}
	if environment == "production" {
		originFallback = nil
	}

	cfg := Config{
		App: AppConfig{
			Environment:              environment,
			Host:                     envString("APP_HOST", "0.0.0.0"),
			Port:                     appPort,
			TrustedProxies:           envCSV("APP_TRUSTED_PROXIES"),
			ReadHeaderTimeoutSeconds: readHeaderTimeout,
			ReadTimeoutSeconds:       readTimeout,
			WriteTimeoutSeconds:      writeTimeout,
			IdleTimeoutSeconds:       idleTimeout,
			ShutdownTimeoutSeconds:   shutdownTimeout,
		},
		Security: SecurityConfig{AllowedOrigins: envCSVDefault("CORS_ALLOWED_ORIGINS", originFallback), AuthorizationEnforced: authorizationEnforced, RateLimitPerMinute: rateLimit, MaxRequestBodyBytes: maxBodyBytes},
		Database: DatabaseConfig{
			Host:       envString("DB_HOST", "localhost"),
			Port:       dbPort,
			User:       envString("DB_USER", "postgres"),
			Password:   os.Getenv("DB_PASSWORD"),
			Name:       envString("DB_NAME", "wms"),
			Schema:     envString("DB_SCHEMA", "wms"),
			SSLMode:    envString("DB_SSLMODE", "disable"),
			Timezone:   envString("DB_TIMEZONE", "Asia/Jakarta"),
			AutoCreate: autoCreate,
		},
	}

	bootstrapAdmin, err := envBool("AUTH_BOOTSTRAP_ADMIN_ENABLED", false)
	if err != nil {
		return Config{}, err
	}
	cfg.Auth = AuthConfig{
		BootstrapAdminEnabled:     bootstrapAdmin,
		BootstrapAdminUsername:    envString("AUTH_BOOTSTRAP_ADMIN_USERNAME", "admin"),
		BootstrapAdminEmail:       envString("AUTH_BOOTSTRAP_ADMIN_EMAIL", "admin@wms.local"),
		BootstrapAdminDisplayName: envString("AUTH_BOOTSTRAP_ADMIN_DISPLAY_NAME", "WMS Administrator"),
		BootstrapAdminPassword:    os.Getenv("AUTH_BOOTSTRAP_ADMIN_PASSWORD"),
		BootstrapAdminTimezone:    envString("AUTH_BOOTSTRAP_ADMIN_TIMEZONE", "Asia/Jakarta"),
	}

	if cfg.Database.Password == "" {
		return Config{}, errors.New("DB_PASSWORD is required; copy .env.example to .env and set it")
	}
	if cfg.Auth.BootstrapAdminEnabled && len(cfg.Auth.BootstrapAdminPassword) < 12 {
		return Config{}, errors.New("AUTH_BOOTSTRAP_ADMIN_PASSWORD must contain at least 12 characters when bootstrap is enabled")
	}
	if cfg.App.Environment != "development" && cfg.App.Environment != "test" && cfg.App.Environment != "production" {
		return Config{}, errors.New("APP_ENV must be development, test, or production")
	}
	for _, origin := range cfg.Security.AllowedOrigins {
		if origin == "*" {
			continue
		}
		parsed, err := url.Parse(origin)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") ||
			parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" ||
			(parsed.Path != "" && parsed.Path != "/") {
			return Config{}, fmt.Errorf("invalid CORS origin %q", origin)
		}
	}
	if cfg.App.Environment == "production" {
		if !cfg.Security.AuthorizationEnforced {
			return Config{}, errors.New("AUTH_RBAC_ENFORCED must be true in production")
		}
		if cfg.Database.AutoCreate {
			return Config{}, errors.New("DB_AUTO_CREATE must be false in production")
		}
		if strings.EqualFold(cfg.Database.SSLMode, "disable") {
			return Config{}, errors.New("DB_SSLMODE cannot be disable in production")
		}
		if len(cfg.Security.AllowedOrigins) == 0 {
			return Config{}, errors.New("CORS_ALLOWED_ORIGINS is required in production")
		}
		for _, origin := range cfg.Security.AllowedOrigins {
			if origin == "*" {
				return Config{}, errors.New("wildcard CORS origin is forbidden in production")
			}
		}
	}

	return cfg, nil
}

func (c AppConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return parsed, nil
}

func envInt64(key string, fallback int64) (int64, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return parsed, nil
}

func envCSV(key string) []string { return envCSVDefault(key, nil) }
func envCSVDefault(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return append([]string(nil), fallback...)
	}
	seen := map[string]bool{}
	result := make([]string, 0)
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(strings.TrimSuffix(part, "/"))
		if part != "" && !seen[part] {
			seen[part] = true
			result = append(result, part)
		}
	}
	return result
}

func envBool(key string, fallback bool) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false", key)
	}
	return parsed, nil
}
