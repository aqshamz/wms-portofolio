package master

import (
	"context"
	"time"

	model "wms-api/models/master"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AccountOwnerAccessDetail struct {
	AccountID string
	OwnerID   string
	OwnerCode string
	OwnerName string
	GrantedBy *string
	GrantedAt time.Time
}

type AccountOwnerAccessRepository struct{ db *gorm.DB }

func NewAccountOwnerAccessRepository(db *gorm.DB) *AccountOwnerAccessRepository {
	return &AccountOwnerAccessRepository{db: db}
}

func (r *AccountOwnerAccessRepository) Grant(ctx context.Context, value *model.AccountOwnerAccess) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "account_id"}, {Name: "owner_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"granted_by": value.GrantedBy,
			"granted_at": gorm.Expr("clock_timestamp()"),
		}),
	}).Create(value).Error
}

func (r *AccountOwnerAccessRepository) List(ctx context.Context, accountID string) ([]AccountOwnerAccessDetail, error) {
	var rows []AccountOwnerAccessDetail
	err := r.db.WithContext(ctx).Table("account_owner_access AS access").
		Select(`access.account_id, access.owner_id, owner.code AS owner_code,
			owner.name AS owner_name, access.granted_by, access.granted_at`).
		Joins("JOIN organization owner ON owner.organization_id = access.owner_id").
		Where("access.account_id = ?", accountID).Order("owner.name").Scan(&rows).Error
	return rows, err
}

func (r *AccountOwnerAccessRepository) Revoke(ctx context.Context, accountID, ownerID string) error {
	result := r.db.WithContext(ctx).Where("account_id = ? AND owner_id = ?", accountID, ownerID).
		Delete(&model.AccountOwnerAccess{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
