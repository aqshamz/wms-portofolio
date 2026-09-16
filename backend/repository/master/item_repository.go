package master

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/master"
)

type ItemRepository struct{ catalogTable[model.Item] }

func (r *ItemRepository) Lock(ctx context.Context, id string) (model.Item, error) {
	var value model.Item
	query := applyOwnerAccess(ctx, r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}), "owner_id")
	err := query.
		Where("item_id = ?", id).Take(&value).Error
	return value, err
}

func NewItemRepository(db *gorm.DB) *ItemRepository {
	return &ItemRepository{catalogTable[model.Item]{db: db, key: "item_id", searchable: true, audited: true, ownerScoped: true}}
}
