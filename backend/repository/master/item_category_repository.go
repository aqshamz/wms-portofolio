package master

import (
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type ItemCategoryRepository struct {
	catalogTable[model.ItemCategory]
}

func NewItemCategoryRepository(db *gorm.DB) *ItemCategoryRepository {
	return &ItemCategoryRepository{catalogTable[model.ItemCategory]{db: db, key: "category_id", searchable: true, audited: false}}
}
