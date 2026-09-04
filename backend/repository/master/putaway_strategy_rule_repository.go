package master

import (
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type PutawayStrategyRuleRepository struct {
	operationalTable[model.PutawayStrategyRule]
}

func NewPutawayStrategyRuleRepository(db *gorm.DB) *PutawayStrategyRuleRepository {
	return &PutawayStrategyRuleRepository{operationalTable[model.PutawayStrategyRule]{catalogTable: catalogTable[model.PutawayStrategyRule]{db: db, key: "rule_id", searchable: false}, parentColumn: "putaway_strategy_id", order: "sequence_no, rule_id", scoped: false, moduleScoped: false}}
}
