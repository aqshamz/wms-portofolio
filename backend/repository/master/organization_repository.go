package master

import (
	"context"
	"time"

	model "wms-api/models/master"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrganizationListItem struct {
	ID           string
	Code         string
	Name         string
	LegalName    *string
	TimezoneName string
	City         *string
	CountryCode  *string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	TotalRows    int64
}

type OrganizationRepository struct{ db *gorm.DB }

func NewOrganizationRepository(db *gorm.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Lock(ctx context.Context, id string) (model.Organization, error) {
	var value model.Organization
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("organization_id = ?", id).Take(&value).Error
	return value, err
}

func (r *OrganizationRepository) Create(ctx context.Context, organization *model.Organization) error {
	return r.db.WithContext(ctx).Create(organization).Error
}

func (r *OrganizationRepository) FindByID(ctx context.Context, id string) (model.Organization, error) {
	var organization model.Organization
	err := r.db.WithContext(ctx).Where("organization_id = ?", id).Take(&organization).Error
	return organization, err
}

func (r *OrganizationRepository) List(
	ctx context.Context,
	search *string,
	active *bool,
	limit, offset int,
) ([]OrganizationListItem, int64, error) {
	var rows []OrganizationListItem
	query := r.db.WithContext(ctx).
		Table("organization").
		Select(`organization_id AS id, code, name, legal_name, timezone_name,
			city, country_code, is_active, created_at, updated_at,
			count(*) OVER () AS total_rows`)
	if search != nil {
		query = query.Where("code ILIKE ? OR name ILIKE ?", "%"+*search+"%", "%"+*search+"%")
	}
	if active != nil {
		query = query.Where("is_active = ?", *active)
	}
	err := query.Order("name, organization_id").Limit(limit).Offset(offset).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return rows, 0, err
	}
	return rows, rows[0].TotalRows, nil
}

func (r *OrganizationRepository) Update(
	ctx context.Context,
	id string,
	expectedUpdatedAt time.Time,
	changes map[string]interface{},
) (model.Organization, error) {
	var organization model.Organization
	result := r.db.WithContext(ctx).Model(&organization).
		Clauses(clause.Returning{}).
		Where("organization_id = ? AND updated_at = ?", id, expectedUpdatedAt).
		Updates(changes)
	if result.Error != nil {
		return organization, result.Error
	}
	if result.RowsAffected == 0 {
		return organization, ErrConcurrentUpdate
	}
	return organization, nil
}

func (r *OrganizationRepository) Deactivate(
	ctx context.Context,
	id, actorID string,
	expectedUpdatedAt time.Time,
) (model.Organization, error) {
	return r.Update(ctx, id, expectedUpdatedAt, map[string]interface{}{
		"is_active":  false,
		"updated_at": gorm.Expr("clock_timestamp()"),
		"updated_by": actorID,
	})
}
