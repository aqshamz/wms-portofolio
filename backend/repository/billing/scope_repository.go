package billing

import (
	"context"
	"fmt"
	"gorm.io/gorm"
)

type ScopeRepository struct{ db *gorm.DB }

func NewScopeRepository(db *gorm.DB) *ScopeRepository { return &ScopeRepository{db: db} }
func (r *ScopeRepository) Allowed(ctx context.Context, accountID, ownerID, warehouseID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("account_owner_access ao").Joins("JOIN account_warehouse_access aw ON aw.account_id=ao.account_id AND aw.warehouse_id=?", warehouseID).Joins("JOIN warehouse_owner wo ON wo.owner_id=ao.owner_id AND wo.warehouse_id=aw.warehouse_id AND wo.is_active").Where("ao.account_id=? AND ao.owner_id=?", accountID, ownerID).Count(&count).Error
	return count > 0, Error(err)
}
func (r *ScopeRepository) ResourceScope(ctx context.Context, kind, id string) (string, string, error) {
	queries := map[string]string{
		"contracts":    "SELECT owner_id::text,warehouse_id::text FROM billing_contract WHERE billing_contract_id=?",
		"rate-cards":   "SELECT c.owner_id::text,c.warehouse_id::text FROM rate_card r JOIN billing_contract c ON c.billing_contract_id=r.billing_contract_id WHERE r.rate_card_id=?",
		"events":       "SELECT owner_id::text,warehouse_id::text FROM billable_event WHERE billable_event_id=?",
		"runs":         "SELECT owner_id::text,warehouse_id::text FROM billing_run WHERE billing_run_id=?",
		"invoices":     "SELECT owner_id::text,warehouse_id::text FROM invoice WHERE invoice_id=?",
		"credit-notes": "SELECT i.owner_id::text,i.warehouse_id::text FROM credit_note c JOIN invoice i ON i.invoice_id=c.invoice_id WHERE c.credit_note_id=?",
		"payments":     "SELECT i.owner_id::text,i.warehouse_id::text FROM payment p JOIN invoice i ON i.invoice_id=p.invoice_id WHERE p.payment_id=?",
	}
	q, ok := queries[kind]
	if !ok {
		return "", "", fmt.Errorf("unknown billing resource kind")
	}
	var v struct{ OwnerID, WarehouseID string }
	err := r.db.WithContext(ctx).Raw(q, id).Take(&v).Error
	return v.OwnerID, v.WarehouseID, Error(err)
}
