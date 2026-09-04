package master

import (
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type PutawayStrategyRepository struct {
	operationalTable[model.PutawayStrategy]
}

func NewPutawayStrategyRepository(db *gorm.DB) *PutawayStrategyRepository {
	return &PutawayStrategyRepository{operationalTable[model.PutawayStrategy]{catalogTable: catalogTable[model.PutawayStrategy]{db: db, key: "putaway_strategy_id", searchable: true}, parentColumn: "", order: "putaway_strategy_id", scoped: true, moduleScoped: false}}
}
