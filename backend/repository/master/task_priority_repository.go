package master

import (
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type TaskPriorityRepository struct {
	operationalTable[model.TaskPriority]
}

func NewTaskPriorityRepository(db *gorm.DB) *TaskPriorityRepository {
	return &TaskPriorityRepository{operationalTable[model.TaskPriority]{catalogTable: catalogTable[model.TaskPriority]{db: db, key: "task_priority_id", searchable: true}, parentColumn: "", order: "priority_value DESC, code", scoped: false, moduleScoped: false}}
}
