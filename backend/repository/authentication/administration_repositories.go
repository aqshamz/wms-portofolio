package authentication

import (
	"context"

	masterrepository "wms-api/repository/master"

	"gorm.io/gorm"
)

type AdministrationRepositories struct {
	db              *gorm.DB
	Accounts        *AppAccountRepository
	Statuses        *AccountStatusRepository
	Policies        *AuthenticationPolicyRepository
	Sessions        *AppSessionRepository
	Permissions     *AccountPermissionRepository
	Roles           *AppRoleRepository
	RolePermissions *RolePermissionRepository
	AccountRoles    *AccountRoleRepository
	OwnerAccess     *masterrepository.AccountOwnerAccessRepository
	WarehouseAccess *masterrepository.AccountWarehouseAccessRepository
	Organizations   *masterrepository.OrganizationRepository
	Warehouses      *masterrepository.WarehouseRepository
}

func NewAdministrationRepositories(db *gorm.DB) *AdministrationRepositories {
	return &AdministrationRepositories{
		db: db, Accounts: NewAppAccountRepository(db), Statuses: NewAccountStatusRepository(db),
		Policies: NewAuthenticationPolicyRepository(db), Sessions: NewAppSessionRepository(db),
		Permissions: NewAccountPermissionRepository(db), Roles: NewAppRoleRepository(db),
		RolePermissions: NewRolePermissionRepository(db), AccountRoles: NewAccountRoleRepository(db),
		OwnerAccess:     masterrepository.NewAccountOwnerAccessRepository(db),
		WarehouseAccess: masterrepository.NewAccountWarehouseAccessRepository(db),
		Organizations:   masterrepository.NewOrganizationRepository(db),
		Warehouses:      masterrepository.NewWarehouseRepository(db),
	}
}

func (r *AdministrationRepositories) Transaction(ctx context.Context, work func(*AdministrationRepositories) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return work(NewAdministrationRepositories(tx))
	})
}
