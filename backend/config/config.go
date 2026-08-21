package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Auth     AuthConfig
}

type AppConfig struct {
	Environment string
	Host        string
	Port        int
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

	cfg := Config{
		App: AppConfig{
			Environment: envString("APP_ENV", "development"),
			Host:        envString("APP_HOST", "0.0.0.0"),
			Port:        appPort,
		},
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
