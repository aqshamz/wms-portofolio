package master

import (
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type PickingSortMethodRepository struct {
	operationalTable[model.PickingSortMethod]
}

func NewPickingSortMethodRepository(db *gorm.DB) *PickingSortMethodRepository {
	return &PickingSortMethodRepository{operationalTable[model.PickingSortMethod]{catalogTable: catalogTable[model.PickingSortMethod]{db: db, key: "picking_sort_method_id", searchable: true}, parentColumn: "", order: "picking_sort_method_id", scoped: false, moduleScoped: false}}
}
