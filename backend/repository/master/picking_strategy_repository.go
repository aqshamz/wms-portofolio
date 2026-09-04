package master

import (
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type PickingStrategyRepository struct {
	operationalTable[model.PickingStrategy]
}

func NewPickingStrategyRepository(db *gorm.DB) *PickingStrategyRepository {
	return &PickingStrategyRepository{operationalTable[model.PickingStrategy]{catalogTable: catalogTable[model.PickingStrategy]{db: db, key: "picking_strategy_id", searchable: true}, parentColumn: "", order: "picking_strategy_id", scoped: true, moduleScoped: false}}
}
