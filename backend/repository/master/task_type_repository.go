package master

import (
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type TaskTypeRepository struct {
	operationalTable[model.TaskType]
}

func NewTaskTypeRepository(db *gorm.DB) *TaskTypeRepository {
	return &TaskTypeRepository{operationalTable[model.TaskType]{catalogTable: catalogTable[model.TaskType]{db: db, key: "task_type_id", searchable: true}, parentColumn: "", order: "task_type_id", scoped: false, moduleScoped: false}}
}
