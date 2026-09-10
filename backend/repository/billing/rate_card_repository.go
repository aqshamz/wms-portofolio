package billing

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
	model "wms-api/models/billing"
)

type RateCardRow struct {
	model.RateCard
	StatusCode, OwnerID, OwnerCode, WarehouseID, WarehouseCode, CurrencyCode string
}
type RateCardRepository struct{ db *gorm.DB }

func NewRateCardRepository(db *gorm.DB) *RateCardRepository { return &RateCardRepository{db: db} }
func (r *RateCardRepository) Create(ctx context.Context, v *model.RateCard) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *RateCardRepository) Lock(ctx context.Context, id string) (model.RateCard, error) {
	var v model.RateCard
	e := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("rate_card_id=?", id).Take(&v).Error
	return v, Error(e)
}
func rateCardQuery(db *gorm.DB) *gorm.DB {
	return db.Table("rate_card r").Select("r.*,s.code status_code,c.owner_id,o.code owner_code,c.warehouse_id,w.code warehouse_code,c.currency_code").Joins("JOIN document_status s ON s.status_id=r.status_id").Joins("JOIN billing_contract c ON c.billing_contract_id=r.billing_contract_id").Joins("JOIN organization o ON o.organization_id=c.owner_id").Joins("JOIN warehouse w ON w.warehouse_id=c.warehouse_id")
}
func (r *RateCardRepository) Get(ctx context.Context, id string) (RateCardRow, error) {
	var v RateCardRow
	e := rateCardQuery(r.db.WithContext(ctx)).Where("r.rate_card_id=?", id).Take(&v).Error
	return v, Error(e)
}
func (r *RateCardRepository) List(ctx context.Context, f ListFilter) ([]RateCardRow, int64, error) {
	q := rateCardQuery(r.db.WithContext(ctx)).Where("c.owner_id=?", f.OwnerID)
	if f.WarehouseID != "" {
		q = q.Where("c.warehouse_id=?", f.WarehouseID)
	}
	if f.StatusCode != "" {
		q = q.Where("s.code=?", f.StatusCode)
	}
	if f.Search != "" {
		x := "%" + f.Search + "%"
		q = q.Where("r.rate_card_id ILIKE ? OR r.name ILIKE ?", x, x)
	}
	var n int64
	if e := q.Session(&gorm.Session{}).Count(&n).Error; e != nil {
		return nil, 0, Error(e)
	}
	v := []RateCardRow{}
	e := q.Order("r.effective_from DESC,r.rate_card_id DESC").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&v).Error
	return v, n, Error(e)
}
func (r *RateCardRepository) UpdateDraft(ctx context.Context, id, actor string, version int64, values map[string]interface{}) error {
	values["updated_by"] = actor
	values["updated_at"] = gorm.Expr("clock_timestamp()")
	values["version_no"] = gorm.Expr("version_no+1")
	z := r.db.WithContext(ctx).Model(&model.RateCard{}).Where("rate_card_id=? AND version_no=? AND status_id IN (SELECT status_id FROM document_status WHERE document_type_id=rate_card.document_type_id AND code='DRAFT')", id, version).Updates(values)
	if z.Error != nil {
		return Error(z.Error)
	}
	if z.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}
func (r *RateCardRepository) SetStatus(ctx context.Context, id, status, actor string, version int64) error {
	z := r.db.WithContext(ctx).Model(&model.RateCard{}).Where("rate_card_id=? AND version_no=?", id, version).Updates(map[string]interface{}{"status_id": status, "updated_by": actor, "updated_at": gorm.Expr("clock_timestamp()"), "version_no": gorm.Expr("version_no+1")})
	if z.Error != nil {
		return Error(z.Error)
	}
	if z.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}
func (r *RateCardRepository) ActiveForPeriod(ctx context.Context, contract string, from, until time.Time) ([]RateCardRow, error) {
	v := []RateCardRow{}
	e := rateCardQuery(r.db.WithContext(ctx)).Where("r.billing_contract_id=? AND s.code='ACTIVE' AND r.effective_from<=? AND (r.effective_until IS NULL OR r.effective_until>=?)", contract, until, from).Find(&v).Error
	return v, Error(e)
}
func (r *RateCardRepository) ActiveByScope(ctx context.Context, owner, warehouse string, from, until time.Time) ([]RateCardRow, error) {
	v := []RateCardRow{}
	e := rateCardQuery(r.db.WithContext(ctx)).Joins("JOIN document_status cs ON cs.status_id=c.status_id").Where("c.owner_id=? AND c.warehouse_id=? AND cs.code IN ('ACTIVE','EXPIRED') AND c.effective_from<=? AND (c.effective_until IS NULL OR c.effective_until>=?) AND s.code IN ('ACTIVE','EXPIRED') AND r.effective_from<=? AND (r.effective_until IS NULL OR r.effective_until>=?)", owner, warehouse, until, from, until, from).Find(&v).Error
	return v, Error(e)
}
