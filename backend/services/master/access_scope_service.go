package master

import (
	"context"
	"errors"

	dto "wms-api/dto/master"
	mastermodel "wms-api/models/master"
	authrepository "wms-api/repository/authentication"
	masterrepository "wms-api/repository/master"

	"gorm.io/gorm"
)

type AccessScopeService struct {
	accounts        *authrepository.AppAccountRepository
	organizations   *masterrepository.OrganizationRepository
	warehouses      *masterrepository.WarehouseRepository
	ownerAccess     *masterrepository.AccountOwnerAccessRepository
	warehouseAccess *masterrepository.AccountWarehouseAccessRepository
}

func NewAccessScopeService(
	accounts *authrepository.AppAccountRepository,
	organizations *masterrepository.OrganizationRepository,
	warehouses *masterrepository.WarehouseRepository,
	ownerAccess *masterrepository.AccountOwnerAccessRepository,
	warehouseAccess *masterrepository.AccountWarehouseAccessRepository,
) *AccessScopeService {
	return &AccessScopeService{accounts: accounts, organizations: organizations,
		warehouses: warehouses, ownerAccess: ownerAccess, warehouseAccess: warehouseAccess}
}

func (s *AccessScopeService) GrantOwner(
	ctx context.Context, ownerID, accountID, actorID string,
) error {
	if validateID(ownerID) != nil || validateID(accountID) != nil {
		return ErrInvalidInput
	}
	if err := s.validateAccount(ctx, accountID); err != nil {
		return err
	}
	owner, err := s.organizations.FindByID(ctx, ownerID)
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && !owner.IsActive) {
		return ErrInvalidInput
	}
	if err != nil {
		return err
	}
	return s.ownerAccess.Grant(ctx, &mastermodel.AccountOwnerAccess{
		AccountID: accountID, OwnerID: ownerID, GrantedBy: &actorID,
	})
}

func (s *AccessScopeService) ListOwners(
	ctx context.Context, accountID string,
) ([]dto.AccountOwnerAccessResponse, error) {
	if err := s.validateAccount(ctx, accountID); err != nil {
		return nil, err
	}
	rows, err := s.ownerAccess.List(ctx, accountID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.AccountOwnerAccessResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, dto.AccountOwnerAccessResponse{
			AccountID: row.AccountID, OwnerID: row.OwnerID, OwnerCode: row.OwnerCode,
			OwnerName: row.OwnerName, GrantedBy: row.GrantedBy, GrantedAt: row.GrantedAt,
		})
	}
	return result, nil
}

func (s *AccessScopeService) RevokeOwner(ctx context.Context, accountID, ownerID string) error {
	if validateID(accountID) != nil || validateID(ownerID) != nil {
		return ErrInvalidInput
	}
	err := s.ownerAccess.Revoke(ctx, accountID, ownerID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *AccessScopeService) GrantWarehouse(
	ctx context.Context, warehouseID, accountID, actorID string,
) error {
	if validateID(warehouseID) != nil || validateID(accountID) != nil {
		return ErrInvalidInput
	}
	if err := s.validateAccount(ctx, accountID); err != nil {
		return err
	}
	warehouse, err := s.warehouses.FindByID(ctx, warehouseID)
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && !warehouse.IsActive) {
		return ErrInvalidInput
	}
	if err != nil {
		return err
	}
	return s.warehouseAccess.Grant(ctx, &mastermodel.AccountWarehouseAccess{
		AccountID: accountID, WarehouseID: warehouseID, GrantedBy: &actorID,
	})
}

func (s *AccessScopeService) ListWarehouses(
	ctx context.Context, accountID string,
) ([]dto.AccountWarehouseAccessResponse, error) {
	if err := s.validateAccount(ctx, accountID); err != nil {
		return nil, err
	}
	rows, err := s.warehouseAccess.List(ctx, accountID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.AccountWarehouseAccessResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, dto.AccountWarehouseAccessResponse{
			AccountID: row.AccountID, WarehouseID: row.WarehouseID,
			WarehouseCode: row.WarehouseCode, WarehouseName: row.WarehouseName,
			GrantedBy: row.GrantedBy, GrantedAt: row.GrantedAt,
		})
	}
	return result, nil
}

func (s *AccessScopeService) RevokeWarehouse(
	ctx context.Context, accountID, warehouseID string,
) error {
	if validateID(accountID) != nil || validateID(warehouseID) != nil {
		return ErrInvalidInput
	}
	err := s.warehouseAccess.Revoke(ctx, accountID, warehouseID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *AccessScopeService) validateAccount(ctx context.Context, accountID string) error {
	if validateID(accountID) != nil {
		return ErrInvalidInput
	}
	_, err := s.accounts.FindByID(ctx, accountID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrInvalidInput
	}
	return err
}
