package master

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type TaskStatusTransitionRepository struct {
	operationalTable[model.TaskStatusTransition]
}

func (r *TaskStatusTransitionRepository) HasOutgoing(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.TaskStatusTransition{}).Where("from_status_id = ? AND is_active", id).Count(&count).Error
	return count > 0, err
}
func NewTaskStatusTransitionRepository(db *gorm.DB) *TaskStatusTransitionRepository {
	return &TaskStatusTransitionRepository{operationalTable[model.TaskStatusTransition]{catalogTable: catalogTable[model.TaskStatusTransition]{db: db, key: "task_status_transition_id", searchable: false}, parentColumn: "", order: "task_status_transition_id", scoped: false, moduleScoped: false}}
}
