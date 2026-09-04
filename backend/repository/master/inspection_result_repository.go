package master

import (
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type InspectionResultRepository struct {
	catalogTable[model.InspectionResult]
}

func NewInspectionResultRepository(db *gorm.DB) *InspectionResultRepository {
	return &InspectionResultRepository{catalogTable[model.InspectionResult]{db: db, key: "inspection_result_id", searchable: true, audited: false}}
}
