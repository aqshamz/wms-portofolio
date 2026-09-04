package master

import (
	"context"
	"gorm.io/gorm"
	"time"
	model "wms-api/models/master"
)

type DocumentNumberRuleRepository struct {
	operationalTable[model.DocumentNumberRule]
}

func (r *DocumentNumberRuleRepository) Create(ctx context.Context, value *model.DocumentNumberRule) error {
	partner, warehouse, separator := value.IncludePartnerCode, value.IncludeWarehouseCode, value.Separator
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(value).Error; err != nil {
			return err
		}
		value.IncludePartnerCode, value.IncludeWarehouseCode, value.Separator = partner, warehouse, separator
		return tx.Model(value).Updates(map[string]interface{}{"include_partner_code": partner, "include_warehouse_code": warehouse, "separator": separator}).Error
	})
}
func (r *DocumentNumberRuleRepository) Retire(ctx context.Context, typeID string, end time.Time) error {
	return r.db.WithContext(ctx).Model(&model.DocumentNumberRule{}).Where("document_type_id = ? AND is_active", typeID).
		Updates(map[string]interface{}{"is_active": false, "effective_until": gorm.Expr("GREATEST(effective_from, ?::date)", end)}).Error
}
func NewDocumentNumberRuleRepository(db *gorm.DB) *DocumentNumberRuleRepository {
	return &DocumentNumberRuleRepository{operationalTable[model.DocumentNumberRule]{catalogTable: catalogTable[model.DocumentNumberRule]{db: db, key: "document_number_rule_id", searchable: false}, parentColumn: "document_type_id", order: "effective_from DESC, created_at DESC", scoped: false, moduleScoped: false}}
}
