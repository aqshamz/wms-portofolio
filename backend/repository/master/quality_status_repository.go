package master

import (
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type QualityStatusRepository struct {
	catalogTable[model.QualityStatus]
}

func NewQualityStatusRepository(db *gorm.DB) *QualityStatusRepository {
	return &QualityStatusRepository{catalogTable[model.QualityStatus]{db: db, key: "quality_status_id", searchable: true, audited: false}}
}
