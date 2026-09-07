package stockcontrol

import (
	"context"
	"gorm.io/gorm"
	inventory "wms-api/repository/inventory"
)

type Repositories struct {
	db       *gorm.DB
	Balances *inventory.InventoryBalanceRepository
	Reasons  *ReasonCodeRepository
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{db: db, Balances: inventory.NewInventoryBalanceRepository(db), Reasons: NewReasonCodeRepository(db)}
}
func (r *Repositories) Transaction(ctx context.Context, work func(*Repositories, *inventory.Repositories) error) error {
	return inventory.Error(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return work(NewRepositories(tx), inventory.NewRepositories(tx)) }))
}
