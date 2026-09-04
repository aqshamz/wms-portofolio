package master

import (
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type AppModuleRepository struct {
	operationalTable[model.AppModule]
}

func NewAppModuleRepository(db *gorm.DB) *AppModuleRepository {
	return &AppModuleRepository{operationalTable[model.AppModule]{catalogTable: catalogTable[model.AppModule]{db: db, key: "module_id", searchable: true}, parentColumn: "", order: "display_order, code", scoped: false, moduleScoped: false}}
}
