package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var postgresIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func OpenDatabase(cfg DatabaseConfig) (*gorm.DB, error) {
	if !postgresIdentifier.MatchString(cfg.Name) {
		return nil, fmt.Errorf("invalid DB_NAME %q", cfg.Name)
	}
	if !postgresIdentifier.MatchString(cfg.Schema) {
		return nil, fmt.Errorf("invalid DB_SCHEMA %q", cfg.Schema)
	}

	if cfg.AutoCreate {
		if err := createDatabaseIfMissing(cfg); err != nil {
			return nil, err
		}
	}

	db, err := gorm.Open(postgres.Open(cfg.applicationDSN()), &gorm.Config{
		PrepareStmt: true,
		Logger: gormlogger.New(log.New(os.Stdout, "", log.LstdFlags), gormlogger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  gormlogger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("connect to database %q: %w", cfg.Name, err)
	}

	if err := db.Exec(fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, quoteIdentifier(cfg.Schema))).Error; err != nil {
		return nil, fmt.Errorf("ensure schema %q: %w", cfg.Schema, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get database pool: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

func createDatabaseIfMissing(cfg DatabaseConfig) error {
	adminDB, err := gorm.Open(postgres.Open(cfg.adminDSN()), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("connect to PostgreSQL administrative database: %w", err)
	}

	sqlDB, err := adminDB.DB()
	if err != nil {
		return fmt.Errorf("get administrative database pool: %w", err)
	}
	defer sqlDB.Close()

	var exists bool
	if err := adminDB.Raw(`SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = ?)`, cfg.Name).Scan(&exists).Error; err != nil {
		return fmt.Errorf("check database %q: %w", cfg.Name, err)
	}
	if exists {
		return nil
	}

	// PostgreSQL does not allow CREATE DATABASE inside a transaction or with a
	// bind parameter, so the strictly validated identifier is quoted here.
	statement := fmt.Sprintf(`CREATE DATABASE %s`, quoteIdentifier(cfg.Name))
	if err := adminDB.Exec(statement).Error; err != nil {
		return fmt.Errorf("create database %q: %w", cfg.Name, err)
	}

	return nil
}

func (cfg DatabaseConfig) adminDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s TimeZone=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.SSLMode, cfg.Timezone,
	)
}

func (cfg DatabaseConfig) applicationDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s search_path=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode, cfg.Timezone, cfg.Schema,
	)
}

func quoteIdentifier(value string) string {
	return `"` + value + `"`
}
