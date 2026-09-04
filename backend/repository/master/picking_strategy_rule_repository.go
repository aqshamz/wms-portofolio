package master

import (
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type PickingStrategyRuleRepository struct {
	operationalTable[model.PickingStrategyRule]
}

func NewPickingStrategyRuleRepository(db *gorm.DB) *PickingStrategyRuleRepository {
	return &PickingStrategyRuleRepository{operationalTable[model.PickingStrategyRule]{catalogTable: catalogTable[model.PickingStrategyRule]{db: db, key: "rule_id", searchable: false}, parentColumn: "picking_strategy_id", order: "sequence_no, rule_id", scoped: false, moduleScoped: false}}
}
