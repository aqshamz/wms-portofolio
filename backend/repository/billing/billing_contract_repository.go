package billing

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
	model "wms-api/models/billing"
)

type ContractRow struct {
	model.BillingContract
	StatusCode, OwnerCode, WarehouseCode string
}
type BillingContractRepository struct{ db *gorm.DB }

func NewBillingContractRepository(db *gorm.DB) *BillingContractRepository {
	return &BillingContractRepository{db: db}
}
func (r *BillingContractRepository) Create(ctx context.Context, v *model.BillingContract) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}
func (r *BillingContractRepository) Lock(ctx context.Context, id string) (model.BillingContract, error) {
	var v model.BillingContract
	e := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("billing_contract_id=?", id).Take(&v).Error
	return v, Error(e)
}
func contractQuery(db *gorm.DB) *gorm.DB {
	return db.Table("billing_contract c").Select("c.*,s.code status_code,o.code owner_code,w.code warehouse_code").Joins("JOIN document_status s ON s.status_id=c.status_id").Joins("JOIN organization o ON o.organization_id=c.owner_id").Joins("JOIN warehouse w ON w.warehouse_id=c.warehouse_id")
}
func (r *BillingContractRepository) Get(ctx context.Context, id string) (ContractRow, error) {
	var v ContractRow
	e := contractQuery(r.db.WithContext(ctx)).Where("c.billing_contract_id=?", id).Take(&v).Error
	return v, Error(e)
}
func (r *BillingContractRepository) List(ctx context.Context, f ListFilter) ([]ContractRow, int64, error) {
	q := contractQuery(r.db.WithContext(ctx)).Where("c.owner_id=?", f.OwnerID)
	if f.WarehouseID != "" {
		q = q.Where("c.warehouse_id=?", f.WarehouseID)
	}
	if f.StatusCode != "" {
		q = q.Where("s.code=?", f.StatusCode)
	}
	if f.Search != "" {
		x := "%" + f.Search + "%"
		q = q.Where("c.billing_contract_id ILIKE ? OR o.name ILIKE ?", x, x)
	}
	var n int64
	if e := q.Session(&gorm.Session{}).Count(&n).Error; e != nil {
		return nil, 0, Error(e)
	}
	v := []ContractRow{}
	e := q.Order("c.effective_from DESC,c.billing_contract_id DESC").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&v).Error
	return v, n, Error(e)
}
func (r *BillingContractRepository) UpdateDraft(ctx context.Context, id, actor string, version int64, values map[string]interface{}) error {
	values["updated_by"] = actor
	values["updated_at"] = gorm.Expr("clock_timestamp()")
	values["version_no"] = gorm.Expr("version_no+1")
	z := r.db.WithContext(ctx).Model(&model.BillingContract{}).Where("billing_contract_id=? AND version_no=? AND status_id IN (SELECT status_id FROM document_status WHERE document_type_id=billing_contract.document_type_id AND code='DRAFT')", id, version).Updates(values)
	if z.Error != nil {
		return Error(z.Error)
	}
	if z.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}
func (r *BillingContractRepository) SetStatus(ctx context.Context, id, status, actor string, version int64) error {
	z := r.db.WithContext(ctx).Model(&model.BillingContract{}).Where("billing_contract_id=? AND version_no=?", id, version).Updates(map[string]interface{}{"status_id": status, "updated_by": actor, "updated_at": gorm.Expr("clock_timestamp()"), "version_no": gorm.Expr("version_no+1")})
	if z.Error != nil {
		return Error(z.Error)
	}
	if z.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}

func (r *BillingContractRepository) ActiveForPeriod(ctx context.Context, owner, warehouse string, from, until time.Time) ([]ContractRow, error) {
	rows := []ContractRow{}
	err := contractQuery(r.db.WithContext(ctx)).Where("c.owner_id=? AND c.warehouse_id=? AND s.code='ACTIVE' AND c.effective_from<=? AND (c.effective_until IS NULL OR c.effective_until>=?)", owner, warehouse, until, from).Find(&rows).Error
	return rows, Error(err)
}
