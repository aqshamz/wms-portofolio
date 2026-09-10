package migration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"wms-api/config"
	model "wms-api/models/system"
)

func TestApplyRunsEachMigrationOncePostgreSQL(t *testing.T) {
	if os.Getenv("WMS_INTEGRATION_TEST") != "1" {
		t.Skip("set WMS_INTEGRATION_TEST=1")
	}
	_ = godotenv.Load("../../.env")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg.Database.AutoCreate = false
	db, err := config.OpenDatabase(cfg.Database)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	tx := db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	schema := fmt.Sprintf("migration_test_%d", time.Now().UnixNano())
	if err := tx.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Exec("SET LOCAL search_path TO " + schema).Error; err != nil {
		t.Fatal(err)
	}

	runs := 0
	steps := []Step{{Version: 1, Name: "sample", Up: func(db *gorm.DB) error {
		runs++
		return db.Exec("CREATE TABLE migration_sample(id integer PRIMARY KEY)").Error
	}}}
	if err := Apply(context.Background(), tx, steps); err != nil {
		t.Fatal(err)
	}
	if err := Apply(context.Background(), tx, steps); err != nil {
		t.Fatal(err)
	}
	if runs != 1 {
		t.Fatalf("migration ran %d times, want once", runs)
	}
	var count int64
	if err := tx.Model(&model.SchemaMigration{}).Where("version = 1").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("recorded migrations=%d want=1", count)
	}
}
