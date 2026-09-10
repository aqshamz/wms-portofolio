package migration

import (
	"context"
	"fmt"
	"sort"

	"gorm.io/gorm"
	model "wms-api/models/system"
)

type Step struct {
	Version int64
	Name    string
	Up      func(*gorm.DB) error
}

// Apply serializes startup migrations and records each successful version.
// Existing installations safely bootstrap because all module migrations are
// idempotent before their version is recorded.
func Apply(ctx context.Context, db *gorm.DB, steps []Step) error {
	steps = append([]Step(nil), steps...)
	sort.Slice(steps, func(i, j int) bool { return steps[i].Version < steps[j].Version })
	seen := map[int64]bool{}
	for _, step := range steps {
		if step.Version <= 0 || step.Name == "" || step.Up == nil {
			return fmt.Errorf("invalid schema migration step")
		}
		if seen[step.Version] {
			return fmt.Errorf("duplicate schema migration version %d", step.Version)
		}
		seen[step.Version] = true
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", int64(873246210)).Error; err != nil {
			return err
		}
		if err := tx.AutoMigrate(&model.SchemaMigration{}); err != nil {
			return err
		}
		for _, step := range steps {
			var count int64
			if err := tx.Model(&model.SchemaMigration{}).Where("version = ?", step.Version).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				continue
			}
			if err := step.Up(tx); err != nil {
				return fmt.Errorf("migration %d %s: %w", step.Version, step.Name, err)
			}
			if err := tx.Create(&model.SchemaMigration{Version: step.Version, Name: step.Name}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
