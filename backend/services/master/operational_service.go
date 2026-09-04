package master

import (
	"context"
	"errors"
	"math/big"
	"regexp"
	"strings"
	"time"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
	repository "wms-api/repository/master"
)

var ErrCounterExhausted = repository.ErrCounterExhausted

type OperationalService struct {
	repositories *repository.OperationalRepositories
	location     *time.Location
}

func NewOperationalService(repositories *repository.OperationalRepositories, timezone string) (*OperationalService, error) {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}
	return &OperationalService{repositories: repositories, location: location}, nil
}
func (s *OperationalService) transaction(ctx context.Context, work func(*OperationalService) error) error {
	return catalogError(s.repositories.Transaction(ctx, func(repos *repository.OperationalRepositories) error {
		return work(&OperationalService{repositories: repos, location: s.location})
	}))
}
func operationalFilter(request dto.OperationalListRequest, parent string) (repository.OperationalFilter, error) {
	if err := catalogPage(repository.CatalogFilter{Page: request.Page, PageSize: request.PageSize}); err != nil {
		return repository.OperationalFilter{}, err
	}
	for _, id := range []string{request.OwnerID, request.WarehouseID, parent} {
		if id != "" && validateID(id) != nil {
			return repository.OperationalFilter{}, ErrInvalidInput
		}
	}
	return repository.OperationalFilter{ParentID: parent, OwnerID: request.OwnerID, WarehouseID: request.WarehouseID,
		ModuleCode: strings.ToUpper(strings.TrimSpace(request.ModuleCode)), Search: strings.TrimSpace(request.Search), Active: request.Active, Page: request.Page, PageSize: request.PageSize}, nil
}
func statusFlags(initial, final, cancelled, active bool) error {
	if initial && (!active || final || cancelled) {
		return invalidCatalog("an initial status must be active and non-terminal")
	}
	if cancelled && !final {
		return invalidCatalog("a cancelled status must also be final")
	}
	return nil
}
func (s *OperationalService) module(ctx context.Context, code string) error {
	value, err := s.repositories.AppModule.ByCode(ctx, code)
	if err != nil {
		return catalogError(err)
	}
	if !value.IsActive {
		return invalidCatalog("module is inactive")
	}
	return nil
}
func (s *OperationalService) permission(ctx context.Context, id *string) error {
	if id == nil {
		return nil
	}
	if validateID(*id) != nil {
		return ErrInvalidInput
	}
	value, err := s.repositories.AppPermission.Get(ctx, *id)
	if err != nil {
		return catalogError(err)
	}
	if !value.IsActive {
		return invalidCatalog("required permission is inactive")
	}
	return nil
}
func (s *OperationalService) scope(ctx context.Context, ownerID, warehouseID *string) error {
	if ownerID != nil {
		if validateID(*ownerID) != nil {
			return ErrInvalidInput
		}
		value, err := s.repositories.Organization.FindByID(ctx, *ownerID)
		if err != nil {
			return catalogError(err)
		}
		if !value.IsActive {
			return invalidCatalog("owner is inactive")
		}
	}
	if warehouseID != nil {
		if validateID(*warehouseID) != nil {
			return ErrInvalidInput
		}
		value, err := s.repositories.Warehouse.FindByID(ctx, *warehouseID)
		if err != nil {
			return catalogError(err)
		}
		if !value.IsActive {
			return invalidCatalog("warehouse is inactive")
		}
	}
	if ownerID != nil && warehouseID != nil {
		assigned, err := s.repositories.WarehouseOwner.IsActive(ctx, *warehouseID, *ownerID)
		if err != nil {
			return err
		}
		if !assigned {
			return invalidCatalog("owner is not assigned to warehouse")
		}
	}
	return nil
}
func (s *OperationalService) zone(ctx context.Context, zoneID, warehouseID *string) error {
	if zoneID == nil {
		return nil
	}
	if validateID(*zoneID) != nil {
		return ErrInvalidInput
	}
	if warehouseID == nil {
		return invalidCatalog("zone-specific rules require a warehouse-scoped strategy")
	}
	zone, err := s.repositories.Zone.FindByID(ctx, *zoneID)
	if err != nil {
		return catalogError(err)
	}
	if !zone.IsActive || !strings.EqualFold(zone.WarehouseID, *warehouseID) {
		return invalidCatalog("zone must be active and belong to strategy warehouse")
	}
	return nil
}

var emptyPercentPattern = regexp.MustCompile(`^(0|[1-9][0-9]{0,2})(\.[0-9]{1,4})?$`)

func validateEmptyPercent(value *string) error {
	if value == nil {
		return nil
	}
	if !emptyPercentPattern.MatchString(*value) {
		return invalidCatalog("minimum_empty_percent must be a decimal string with up to 4 decimal places")
	}
	number, ok := new(big.Rat).SetString(*value)
	if !ok || number.Cmp(big.NewRat(100, 1)) > 0 {
		return invalidCatalog("minimum_empty_percent must be between 0 and 100")
	}
	return nil
}

// prepareWrite is called only inside the relevant parent/configuration lock.
func (s *OperationalService) prepareWrite(ctx context.Context, input interface{}) error {
	switch value := input.(type) {
	case *model.AppModule:
		if value.DisplayOrder < 0 {
			return ErrInvalidInput
		}
	case *model.DocumentType:
		value.ModuleCode = strings.ToUpper(strings.TrimSpace(value.ModuleCode))
		return s.module(ctx, value.ModuleCode)
	case *model.TaskPriority:
		if value.PriorityValue < 0 {
			return ErrInvalidInput
		}
	case *model.DocumentStatus:
		if value.DisplayOrder < 0 {
			return ErrInvalidInput
		}
		if err := statusFlags(value.IsInitial, value.IsFinal, value.IsCancelled, value.IsActive); err != nil {
			return err
		}
		if value.IsFinal && value.ID != "" {
			outgoing, err := s.repositories.DocumentStatusTransition.HasOutgoing(ctx, value.ID)
			if err != nil {
				return err
			}
			if outgoing {
				return invalidCatalog("deactivate outgoing transitions before marking a status final")
			}
		}
		if value.IsInitial {
			return s.repositories.DocumentStatus.ClearInitial(ctx, value.DocumentTypeID)
		}
	case *model.TaskStatus:
		if err := statusFlags(value.IsInitial, value.IsFinal, value.IsCancelled, value.IsActive); err != nil {
			return err
		}
		if value.IsFinal && value.ID != "" {
			outgoing, err := s.repositories.TaskStatusTransition.HasOutgoing(ctx, value.ID)
			if err != nil {
				return err
			}
			if outgoing {
				return invalidCatalog("deactivate outgoing transitions before marking a status final")
			}
		}
		if value.IsInitial {
			return s.repositories.TaskStatus.ClearInitial(ctx)
		}
	case *model.DocumentStatusTransition:
		if validateID(value.FromStatusID) != nil || validateID(value.ToStatusID) != nil || strings.EqualFold(value.FromStatusID, value.ToStatusID) {
			return ErrInvalidInput
		}
		from, err := s.repositories.DocumentStatus.Get(ctx, value.FromStatusID)
		if err != nil {
			return catalogError(err)
		}
		to, err := s.repositories.DocumentStatus.Get(ctx, value.ToStatusID)
		if err != nil {
			return catalogError(err)
		}
		if !strings.EqualFold(from.DocumentTypeID, value.DocumentTypeID) || !strings.EqualFold(to.DocumentTypeID, value.DocumentTypeID) {
			return invalidCatalog("both statuses must belong to this document type")
		}
		if value.IsActive && (!from.IsActive || !to.IsActive || from.IsFinal) {
			return invalidCatalog("transition requires active statuses and a non-final source")
		}
		return s.permission(ctx, value.RequiredPermissionID)
	case *model.TaskStatusTransition:
		if validateID(value.FromStatusID) != nil || validateID(value.ToStatusID) != nil || strings.EqualFold(value.FromStatusID, value.ToStatusID) {
			return ErrInvalidInput
		}
		from, err := s.repositories.TaskStatus.Get(ctx, value.FromStatusID)
		if err != nil {
			return catalogError(err)
		}
		to, err := s.repositories.TaskStatus.Get(ctx, value.ToStatusID)
		if err != nil {
			return catalogError(err)
		}
		if value.IsActive && (!from.IsActive || !to.IsActive || from.IsFinal) {
			return invalidCatalog("transition requires active statuses and a non-final source")
		}
		return s.permission(ctx, value.RequiredPermissionID)
	case *model.PickingStrategy:
		return s.scope(ctx, value.OwnerID, value.WarehouseID)
	case *model.PutawayStrategy:
		return s.scope(ctx, value.OwnerID, value.WarehouseID)
	case *model.PickingStrategyRule:
		if value.SequenceNo < 1 || validateID(value.PickingSortMethodID) != nil {
			return ErrInvalidInput
		}
		parent, err := s.repositories.PickingStrategy.Get(ctx, value.PickingStrategyID)
		if err != nil {
			return catalogError(err)
		}
		if err := s.scope(ctx, parent.OwnerID, parent.WarehouseID); err != nil {
			return err
		}
		if err := s.zone(ctx, value.ZoneID, parent.WarehouseID); err != nil {
			return err
		}
		method, err := s.repositories.PickingSortMethod.Get(ctx, value.PickingSortMethodID)
		if err != nil {
			return catalogError(err)
		}
		if !method.IsActive {
			return invalidCatalog("picking sort method is inactive")
		}
		if value.InventoryStatusID != nil {
			if validateID(*value.InventoryStatusID) != nil {
				return ErrInvalidInput
			}
			status, err := s.repositories.Catalog.InventoryStatus.Get(ctx, *value.InventoryStatusID)
			if err != nil {
				return catalogError(err)
			}
			if !status.IsActive || !status.IsAllocatable || !status.IsPickable {
				return invalidCatalog("picking inventory status must be active, allocatable and pickable")
			}
		}
	case *model.PutawayStrategyRule:
		if value.SequenceNo < 1 {
			return ErrInvalidInput
		}
		value.MinimumEmptyPercent = normalizeDecimal(value.MinimumEmptyPercent)
		if err := validateEmptyPercent(value.MinimumEmptyPercent); err != nil {
			return err
		}
		parent, err := s.repositories.PutawayStrategy.Get(ctx, value.PutawayStrategyID)
		if err != nil {
			return catalogError(err)
		}
		if err := s.scope(ctx, parent.OwnerID, parent.WarehouseID); err != nil {
			return err
		}
		if err := s.zone(ctx, value.ZoneID, parent.WarehouseID); err != nil {
			return err
		}
		if value.CategoryID != nil {
			if parent.OwnerID == nil {
				return invalidCatalog("category-specific rules require an owner-scoped strategy")
			}
			if err := NewCatalogService(s.repositories.Catalog).requireCategory(ctx, *parent.OwnerID, value.CategoryID, ""); err != nil {
				return err
			}
		}
		if value.LocationTypeID != nil {
			if validateID(*value.LocationTypeID) != nil {
				return ErrInvalidInput
			}
			location, err := s.repositories.LocationType.FindByID(ctx, *value.LocationTypeID)
			if err != nil {
				return catalogError(err)
			}
			if !location.IsActive {
				return invalidCatalog("location type is inactive")
			}
		}
	}
	return nil
}
func operationalError(err error) error {
	if errors.Is(err, repository.ErrCounterExhausted) {
		return ErrCounterExhausted
	}
	return catalogError(err)
}
