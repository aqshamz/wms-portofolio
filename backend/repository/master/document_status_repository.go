package master

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type DocumentStatusRepository struct {
	operationalTable[model.DocumentStatus]
}

// All writes participate in the same parent/task lock before selecting an initial status.
func (r *DocumentStatusRepository) ClearInitial(ctx context.Context, typeID string) error {
	query := r.db.WithContext(ctx).Model(&model.DocumentStatus{})
	query = query.Where("document_type_id = ?", typeID)
	return query.Update("is_initial", false).Error
}
func (r *DocumentStatusRepository) InitialExists(ctx context.Context, typeID string) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.DocumentStatus{}).Where("is_initial = true")
	query = query.Where("document_type_id = ?", typeID)
	err := query.Count(&count).Error
	return count > 0, err
}
func NewDocumentStatusRepository(db *gorm.DB) *DocumentStatusRepository {
	return &DocumentStatusRepository{operationalTable[model.DocumentStatus]{catalogTable: catalogTable[model.DocumentStatus]{db: db, key: "status_id", searchable: true}, parentColumn: "document_type_id", order: "display_order, code", scoped: false, moduleScoped: false}}
}
