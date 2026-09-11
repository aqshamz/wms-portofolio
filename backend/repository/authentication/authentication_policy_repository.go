package authentication

import (
	"context"
	model "wms-api/models/authentication"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AuthenticationPolicyRepository struct{ db *gorm.DB }

func NewAuthenticationPolicyRepository(db *gorm.DB) *AuthenticationPolicyRepository {
	return &AuthenticationPolicyRepository{db: db}
}

func (r *AuthenticationPolicyRepository) Seed(ctx context.Context, policies []model.AuthenticationPolicy) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}},
		DoNothing: true,
	}).Create(&policies).Error
}

func (r *AuthenticationPolicyRepository) FindDefault(ctx context.Context) (model.AuthenticationPolicy, error) {
	var policy model.AuthenticationPolicy
	err := r.db.WithContext(ctx).
		Where("is_default AND is_active").
		First(&policy).Error
	return policy, err
}

func (r *AuthenticationPolicyRepository) FindByID(ctx context.Context, id string) (model.AuthenticationPolicy, error) {
	var policy model.AuthenticationPolicy
	err := r.db.WithContext(ctx).Where("authentication_policy_id = ? AND is_active", id).First(&policy).Error
	return policy, err
}

func (r *AuthenticationPolicyRepository) List(ctx context.Context) ([]model.AuthenticationPolicy, error) {
	rows := make([]model.AuthenticationPolicy, 0)
	err := r.db.WithContext(ctx).Where("is_active").Order("code").Find(&rows).Error
	return rows, err
}
