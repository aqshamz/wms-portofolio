package master

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type TaskStatusRepository struct {
	operationalTable[model.TaskStatus]
}

// All writes participate in the same parent/task lock before selecting an initial status.
func (r *TaskStatusRepository) ClearInitial(ctx context.Context) error {
	query := r.db.WithContext(ctx).Model(&model.TaskStatus{})
	query = query.Where("is_initial = true")
	return query.Update("is_initial", false).Error
}
func (r *TaskStatusRepository) InitialExists(ctx context.Context) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.TaskStatus{}).Where("is_initial = true")

	err := query.Count(&count).Error
	return count > 0, err
}
func (r *TaskStatusRepository) LockConfiguration(ctx context.Context) error {
	// Application-wide task-status namespace; released automatically at transaction end.
	return r.db.WithContext(ctx).Exec("SELECT pg_advisory_xact_lock(732491, 1)").Error
}
func NewTaskStatusRepository(db *gorm.DB) *TaskStatusRepository {
	return &TaskStatusRepository{operationalTable[model.TaskStatus]{catalogTable: catalogTable[model.TaskStatus]{db: db, key: "task_status_id", searchable: true}, parentColumn: "", order: "task_status_id", scoped: false, moduleScoped: false}}
}
