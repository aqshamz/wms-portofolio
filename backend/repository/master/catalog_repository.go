package master

import (
	"context"
	"time"

	model "wms-api/models/master"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CatalogFilter contains values only. Table/column names are fixed in this package.
type CatalogFilter struct {
	OwnerID, ItemID, CategoryID, PartnerTypeCode, Search string
	Active                                               *bool
	Page, PageSize                                       int
}

// catalogTable shares plumbing; every table still has its own typed repository.
type catalogTable[T any] struct {
	db         *gorm.DB
	key        string
	searchable bool
	audited    bool
}

func (r *catalogTable[T]) Create(ctx context.Context, value *T) error {
	return r.db.WithContext(ctx).Create(value).Error
}

func (r *catalogTable[T]) Seed(ctx context.Context, values []T) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoNothing: true}).Create(&values).Error
}

func (r *catalogTable[T]) Get(ctx context.Context, id string) (T, error) {
	var value T
	err := r.db.WithContext(ctx).Where(r.key+" = ?", id).Take(&value).Error
	return value, err
}

func (r *catalogTable[T]) List(ctx context.Context, filter CatalogFilter) ([]T, int64, error) {
	query := r.db.WithContext(ctx).Model(new(T))
	if filter.OwnerID != "" {
		query = query.Where("owner_id = ?", filter.OwnerID)
	}
	if filter.ItemID != "" {
		query = query.Where("item_id = ?", filter.ItemID)
	}
	if filter.CategoryID != "" {
		query = query.Where("category_id = ?", filter.CategoryID)
	}
	if filter.Active != nil {
		query = query.Where("is_active = ?", *filter.Active)
	}
	if r.searchable && filter.Search != "" {
		query = query.Where("(code ILIKE ? OR name ILIKE ?)", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}
	if filter.PartnerTypeCode != "" {
		query = query.Where("partner_id IN (SELECT bpt.partner_id FROM business_partner_type bpt JOIN partner_type pt ON pt.partner_type_id = bpt.partner_type_id WHERE pt.code = ? AND pt.is_active)", filter.PartnerTypeCode)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]T, 0)
	// PageSize zero is reserved for unpaginated internal child collections.
	if filter.PageSize > 0 {
		query = query.Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize)
	}
	err := query.Order(r.key).Find(&rows).Error
	return rows, total, err
}

func (r *catalogTable[T]) Update(ctx context.Context, id string, changes map[string]interface{}, actor string, expected *time.Time) (T, error) {
	var value T
	query := r.db.WithContext(ctx).Model(&value).Where(r.key+" = ?", id)
	if r.audited {
		if expected == nil {
			return value, ErrConcurrentUpdate
		}
		query = query.Where("updated_at = ?", *expected)
		changes["updated_at"] = gorm.Expr("clock_timestamp()")
		changes["updated_by"] = actor
	}
	result := query.Clauses(clause.Returning{}).Updates(changes)
	if result.Error != nil {
		return value, result.Error
	}
	if result.RowsAffected == 0 {
		if r.audited {
			return value, ErrConcurrentUpdate
		}
		return value, gorm.ErrRecordNotFound
	}
	return value, nil
}

// Transaction keeps GORM out of services while allowing per-table repositories
// to participate in the same atomic business operation.
func (r *CatalogRepositories) Transaction(ctx context.Context, work func(*CatalogRepositories) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return work(NewCatalogRepositories(tx))
	})
}

func (r *CatalogRepositories) Owner(ctx context.Context, id string, lock bool) (model.Organization, error) {
	repo := NewOrganizationRepository(r.db)
	if lock {
		return repo.Lock(ctx, id)
	}
	return repo.FindByID(ctx, id)
}
