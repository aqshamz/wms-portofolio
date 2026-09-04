package master

import (
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type AppPermissionRepository struct {
	operationalTable[model.AppPermission]
}

func NewAppPermissionRepository(db *gorm.DB) *AppPermissionRepository {
	return &AppPermissionRepository{operationalTable[model.AppPermission]{catalogTable: catalogTable[model.AppPermission]{db: db, key: "permission_id", searchable: true}, parentColumn: "", order: "permission_id", scoped: false, moduleScoped: true}}
}
