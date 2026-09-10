package authentication

import (
	"context"
	"errors"
	"regexp"

	dto "wms-api/dto/authentication"
	model "wms-api/models/authentication"
	repository "wms-api/repository/authentication"

	"gorm.io/gorm"
)

var permissionUUIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

var (
	ErrInvalidPermissionGrant = errors.New("invalid account or permission")
	ErrPermissionGrantMissing = errors.New("account permission not found")
	ErrLastSecurityAdmin      = errors.New("cannot revoke the last security administrator")
)

type PermissionService struct {
	accounts    *repository.AppAccountRepository
	permissions *repository.AccountPermissionRepository
}

func NewPermissionService(
	accounts *repository.AppAccountRepository,
	permissions *repository.AccountPermissionRepository,
) *PermissionService {
	return &PermissionService{accounts: accounts, permissions: permissions}
}

func (s *PermissionService) Available(ctx context.Context) ([]dto.PermissionResponse, error) {
	rows, err := s.permissions.Available(ctx)
	if err != nil {
		return nil, err
	}
	response := make([]dto.PermissionResponse, 0, len(rows))
	for _, row := range rows {
		response = append(response, dto.PermissionResponse{
			PermissionID: row.PermissionID, Code: row.Code,
			Name: row.Name, ModuleCode: row.ModuleCode,
		})
	}
	return response, nil
}

func (s *PermissionService) List(ctx context.Context, accountID string) ([]dto.AccountPermissionResponse, error) {
	if err := s.validateAccount(ctx, accountID); err != nil {
		return nil, err
	}
	rows, err := s.permissions.List(ctx, accountID)
	if err != nil {
		return nil, err
	}
	response := make([]dto.AccountPermissionResponse, 0, len(rows))
	for _, row := range rows {
		response = append(response, dto.AccountPermissionResponse{
			AccountID: row.AccountID, PermissionID: row.PermissionID,
			Code: row.Code, Name: row.Name, ModuleCode: row.ModuleCode,
			GrantedAt: row.GrantedAt, GrantedBy: row.GrantedBy,
		})
	}
	return response, nil
}

func (s *PermissionService) Grant(ctx context.Context, accountID, permissionID, actorID string) error {
	if err := s.validateAccount(ctx, accountID); err != nil {
		return err
	}
	if !permissionUUIDPattern.MatchString(permissionID) || !permissionUUIDPattern.MatchString(actorID) {
		return ErrInvalidPermissionGrant
	}
	exists, err := s.permissions.ActivePermissionExists(ctx, permissionID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrInvalidPermissionGrant
	}
	return s.permissions.Grant(ctx, &model.AccountPermission{
		AccountID: accountID, PermissionID: permissionID, GrantedBy: &actorID,
	})
}

func (s *PermissionService) Revoke(ctx context.Context, accountID, permissionID string) error {
	if !permissionUUIDPattern.MatchString(accountID) || !permissionUUIDPattern.MatchString(permissionID) {
		return ErrInvalidPermissionGrant
	}
	err := s.permissions.Revoke(ctx, accountID, permissionID)
	if errors.Is(err, repository.ErrLastSecurityAdministrator) {
		return ErrLastSecurityAdmin
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrPermissionGrantMissing
	}
	return err
}

func (s *PermissionService) validateAccount(ctx context.Context, accountID string) error {
	if !permissionUUIDPattern.MatchString(accountID) {
		return ErrInvalidPermissionGrant
	}
	_, err := s.accounts.FindByID(ctx, accountID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrInvalidPermissionGrant
	}
	return err
}
