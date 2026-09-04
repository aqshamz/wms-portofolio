package master

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	model "wms-api/models/master"
	repository "wms-api/repository/master"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type CatalogService struct {
	repositories *repository.CatalogRepositories
}

func NewCatalogService(repositories *repository.CatalogRepositories) *CatalogService {
	return &CatalogService{repositories: repositories}
}

func catalogError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if errors.Is(err, repository.ErrConcurrentUpdate) {
		return ErrConcurrentUpdate
	}
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.Code {
		case "23505":
			return ErrConflict
		case "23503", "23514", "22003", "22P02":
			return ErrInvalidInput
		}
	}
	return err
}

func invalidCatalog(message string) error { return fmt.Errorf("%w: %s", ErrInvalidInput, message) }

func catalogPage(filter repository.CatalogFilter) error {
	if filter.Page < 1 || filter.PageSize < 1 || filter.PageSize > 100 || filter.Page > 1000000 {
		return invalidCatalog("page must be 1..1000000 and page_size must be 1..100")
	}
	return nil
}

func catalogIdentity(code, name string) (string, string, error) {
	code, err := normalizeCode(code)
	name = strings.TrimSpace(name)
	if err != nil || name == "" {
		return "", "", ErrInvalidInput
	}
	return code, name, nil
}

func catalogBool(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func catalogDecimal(value string, positive bool) (string, error) {
	value = strings.TrimSpace(value)
	if !validCapacity(&value) {
		return "", invalidCatalog("decimal must fit numeric(20,6) and be non-negative")
	}
	number, ok := new(big.Rat).SetString(value)
	if !ok || (positive && number.Sign() <= 0) {
		return "", invalidCatalog("conversion must be greater than zero")
	}
	return value, nil
}

func decimalIsOne(value string) bool {
	number, ok := new(big.Rat).SetString(value)
	return ok && number.Cmp(big.NewRat(1, 1)) == 0
}

func catalogDimensions(values ...*string) error {
	for _, value := range values {
		if value != nil {
			if _, err := catalogDecimal(*value, false); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *CatalogService) requireOwner(ctx context.Context, id string, lock bool) error {
	if validateID(id) != nil {
		return invalidCatalog("invalid owner_id")
	}
	owner, err := s.repositories.Owner(ctx, id, lock)
	if err != nil {
		return catalogError(err)
	}
	if !owner.IsActive {
		return invalidCatalog("owner is inactive")
	}
	return nil
}

func (s *CatalogService) requireUOM(ctx context.Context, id string) error {
	if validateID(id) != nil {
		return invalidCatalog("invalid uom_id")
	}
	unit, err := s.repositories.UOM.Get(ctx, id)
	if err != nil {
		return catalogError(err)
	}
	if !unit.IsActive {
		return invalidCatalog("UOM is inactive")
	}
	return nil
}

// Category writes lock the owner row before walking parents, serializing
// concurrent re-parenting for this owner so two writes cannot form a cycle.
func (s *CatalogService) requireCategory(ctx context.Context, ownerID string, parentID *string, selfID string) error {
	if parentID == nil {
		return nil
	}
	if validateID(*parentID) != nil {
		return invalidCatalog("invalid category ID")
	}
	visited := map[string]bool{}
	cursor := *parentID
	for {
		key := strings.ToLower(cursor)
		if strings.EqualFold(cursor, selfID) || visited[key] {
			return invalidCatalog("category hierarchy would contain a cycle")
		}
		visited[key] = true
		category, err := s.repositories.ItemCategory.Get(ctx, cursor)
		if err != nil {
			return catalogError(err)
		}
		if !strings.EqualFold(category.OwnerID, ownerID) {
			return invalidCatalog("category belongs to another owner")
		}
		if !category.IsActive {
			return invalidCatalog("category is inactive")
		}
		if category.ParentCategoryID == nil {
			break
		}
		cursor = *category.ParentCategoryID
	}
	return nil
}

func (s *CatalogService) itemForWrite(ctx context.Context, id string) (model.Item, error) {
	if validateID(id) != nil {
		return model.Item{}, invalidCatalog("invalid item_id")
	}
	item, err := s.repositories.Item.Lock(ctx, id)
	return item, catalogError(err)
}
