package stockcontrol

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/master"
)

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.ReasonCode{}); err != nil {
			return err
		}
		if err := tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS uq_stock_reason_module_code ON reason_code(module_code,code)").Error; err != nil {
			return err
		}
		return NewReasonCodeRepository(tx).Seed(context.Background(), []model.ReasonCode{{ModuleCode: "INVENTORY", Code: "DAMAGE", Name: "Damage", RequiresNote: true}, {ModuleCode: "INVENTORY", Code: "EXPIRY", Name: "Expiry"}, {ModuleCode: "INVENTORY", Code: "COUNT_VARIANCE", Name: "Count variance", RequiresNote: true}, {ModuleCode: "INVENTORY", Code: "REPLENISHMENT", Name: "Replenishment"}, {ModuleCode: "INVENTORY", Code: "RELOCATION", Name: "Relocation"}, {ModuleCode: "INVENTORY", Code: "CONSOLIDATION", Name: "Consolidation"}, {ModuleCode: "INVENTORY", Code: "STATUS_HOLD", Name: "Place on hold", RequiresNote: true}, {ModuleCode: "INVENTORY", Code: "STATUS_RELEASE", Name: "Release hold", RequiresNote: true}, {ModuleCode: "INVENTORY", Code: "MANUAL_ADJUSTMENT", Name: "Manual adjustment", RequiresNote: true}, {ModuleCode: "INVENTORY", Code: "TRANSFER_VARIANCE", Name: "Transfer variance", RequiresNote: true}})
	})
}
