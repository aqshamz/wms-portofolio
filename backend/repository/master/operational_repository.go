package master

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OperationalFilter struct {
	ParentID, OwnerID, WarehouseID, ModuleCode, Search string
	Active                                             *bool
	Page, PageSize                                     int
}
type operationalTable[T any] struct {
	catalogTable[T]
	parentColumn, order  string
	scoped, moduleScoped bool
}

func (r *operationalTable[T]) List(ctx context.Context, filter OperationalFilter) ([]T, int64, error) {
	query := r.db.WithContext(ctx).Model(new(T))
	if r.parentColumn != "" {
		query = query.Where(r.parentColumn+" = ?", filter.ParentID)
	}
	if r.scoped {
		if filter.OwnerID != "" {
			query = query.Where("owner_id = ?", filter.OwnerID)
		}
		if filter.WarehouseID != "" {
			query = query.Where("warehouse_id = ?", filter.WarehouseID)
		}
	}
	if r.moduleScoped && filter.ModuleCode != "" {
		query = query.Where("module_code = ?", filter.ModuleCode)
	}
	if r.searchable && filter.Search != "" {
		query = query.Where("(code ILIKE ? OR name ILIKE ?)", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}
	if filter.Active != nil {
		query = query.Where("is_active = ?", *filter.Active)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]T, 0)
	if filter.PageSize > 0 {
		query = query.Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize)
	}
	err := query.Order(r.order).Find(&rows).Error
	return rows, count, err
}
func (r *operationalTable[T]) Lock(ctx context.Context, id string) (T, error) {
	var value T
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where(r.key+" = ?", id).Take(&value).Error
	return value, err
}
func (r *operationalTable[T]) ByCode(ctx context.Context, code string) (T, error) {
	var value T
	err := r.db.WithContext(ctx).Where("code = ?", code).Take(&value).Error
	return value, err
}
func (r *operationalTable[T]) SeedOne(ctx context.Context, value *T) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(value).Error
}

type OperationalRepositories struct {
	db                       *gorm.DB
	AppModule                *AppModuleRepository
	AppPermission            *AppPermissionRepository
	DocumentType             *DocumentTypeRepository
	DocumentStatus           *DocumentStatusRepository
	DocumentStatusTransition *DocumentStatusTransitionRepository
	DocumentNumberRule       *DocumentNumberRuleRepository
	DocumentDailyCounter     *DocumentDailyCounterRepository
	TaskType                 *TaskTypeRepository
	TaskStatus               *TaskStatusRepository
	TaskStatusTransition     *TaskStatusTransitionRepository
	TaskPriority             *TaskPriorityRepository
	ReasonCode               *ReasonCodeRepository
	ValidationSeverity       *ValidationSeverityRepository
	PickingSortMethod        *PickingSortMethodRepository
	PickingStrategy          *PickingStrategyRepository
	PickingStrategyRule      *PickingStrategyRuleRepository
	PutawayStrategy          *PutawayStrategyRepository
	PutawayStrategyRule      *PutawayStrategyRuleRepository
	Organization             *OrganizationRepository
	Warehouse                *WarehouseRepository
	WarehouseOwner           *WarehouseOwnerRepository
	Zone                     *WarehouseZoneRepository
	LocationType             *LocationTypeRepository
	Catalog                  *CatalogRepositories
}

func NewOperationalRepositories(db *gorm.DB) *OperationalRepositories {
	return &OperationalRepositories{db: db,
		AppModule:                NewAppModuleRepository(db),
		AppPermission:            NewAppPermissionRepository(db),
		DocumentType:             NewDocumentTypeRepository(db),
		DocumentStatus:           NewDocumentStatusRepository(db),
		DocumentStatusTransition: NewDocumentStatusTransitionRepository(db),
		DocumentNumberRule:       NewDocumentNumberRuleRepository(db),
		DocumentDailyCounter:     NewDocumentDailyCounterRepository(db),
		TaskType:                 NewTaskTypeRepository(db),
		TaskStatus:               NewTaskStatusRepository(db),
		TaskStatusTransition:     NewTaskStatusTransitionRepository(db),
		TaskPriority:             NewTaskPriorityRepository(db),
		ReasonCode:               NewReasonCodeRepository(db),
		ValidationSeverity:       NewValidationSeverityRepository(db),
		PickingSortMethod:        NewPickingSortMethodRepository(db),
		PickingStrategy:          NewPickingStrategyRepository(db),
		PickingStrategyRule:      NewPickingStrategyRuleRepository(db),
		PutawayStrategy:          NewPutawayStrategyRepository(db),
		PutawayStrategyRule:      NewPutawayStrategyRuleRepository(db),
		Organization:             NewOrganizationRepository(db), Warehouse: NewWarehouseRepository(db), WarehouseOwner: NewWarehouseOwnerRepository(db),
		Zone: NewWarehouseZoneRepository(db), LocationType: NewLocationTypeRepository(db), Catalog: NewCatalogRepositories(db)}
}
func (r *OperationalRepositories) Transaction(ctx context.Context, work func(*OperationalRepositories) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return work(NewOperationalRepositories(tx)) })
}
