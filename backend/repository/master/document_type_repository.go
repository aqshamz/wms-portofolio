package master

import (
	"gorm.io/gorm"
	model "wms-api/models/master"
)

type DocumentTypeRepository struct {
	operationalTable[model.DocumentType]
}

func NewDocumentTypeRepository(db *gorm.DB) *DocumentTypeRepository {
	return &DocumentTypeRepository{operationalTable[model.DocumentType]{catalogTable: catalogTable[model.DocumentType]{db: db, key: "document_type_id", searchable: true}, parentColumn: "", order: "document_type_id", scoped: false, moduleScoped: true}}
}
