package master

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type HandlingUnitTypeRepository struct {
	catalogTable[model.HandlingUnitType]
}

func NewHandlingUnitTypeRepository(db *gorm.DB) *HandlingUnitTypeRepository {
	return &HandlingUnitTypeRepository{catalogTable[model.HandlingUnitType]{db: db, key: "handling_unit_type_id", searchable: true}}
}
func MigrateHandlingUnitTypes(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.HandlingUnitType{}); err != nil {
			return err
		}
		for _, sql := range []string{
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_handling_unit_type_code ON handling_unit_type(code)",
			"DO $$ BEGIN ALTER TABLE handling_unit_type ADD CONSTRAINT ck_hu_type_capacity CHECK ((max_weight IS NULL OR max_weight >= 0) AND (max_volume IS NULL OR max_volume >= 0)); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
		} {
			if err := tx.Exec(sql).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (r *HandlingUnitTypeRepository) SeedDefaults(ctx context.Context) error {
	return r.Seed(ctx, []model.HandlingUnitType{
		{Code: "PALLET", Name: "Pallet"}, {Code: "CARTON", Name: "Carton"}, {Code: "TOTE", Name: "Tote"}, {Code: "BIN", Name: "Bin"},
	})
}
