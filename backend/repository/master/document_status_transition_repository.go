package master

import (
	"context"
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type DocumentStatusTransitionRepository struct {
	operationalTable[model.DocumentStatusTransition]
}

func (r *DocumentStatusTransitionRepository) HasOutgoing(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.DocumentStatusTransition{}).Where("from_status_id = ? AND is_active", id).Count(&count).Error
	return count > 0, err
}
func NewDocumentStatusTransitionRepository(db *gorm.DB) *DocumentStatusTransitionRepository {
	return &DocumentStatusTransitionRepository{operationalTable[model.DocumentStatusTransition]{catalogTable: catalogTable[model.DocumentStatusTransition]{db: db, key: "transition_id", searchable: false}, parentColumn: "document_type_id", order: "transition_id", scoped: false, moduleScoped: false}}
}
