package authentication

import (
	"context"
	"errors"
	"regexp"
	"strings"

	dto "wms-api/dto/authentication"
	model "wms-api/models/authentication"
	repository "wms-api/repository/authentication"

	"gorm.io/gorm"
)

var (
	ErrInvalidRole          = errors.New("invalid role data")
	ErrRoleNotFound         = errors.New("role not found")
	ErrRoleConflict         = errors.New("role code already exists")
	ErrConcurrentRoleUpdate = errors.New("role changed; reload it before updating")
	ErrAccountRoleNotFound  = errors.New("account role not found")
)

var roleCodePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]{1,59}$`)

type RoleService struct {
	repositories *repository.AdministrationRepositories
}

func NewRoleService(repositories *repository.AdministrationRepositories) *RoleService {
	return &RoleService{repositories: repositories}
}

func (s *RoleService) Create(ctx context.Context, request dto.CreateRoleRequest, actorID string) (dto.RoleResponse, error) {
	code := strings.ToUpper(strings.TrimSpace(request.Code))
	name := strings.TrimSpace(request.Name)
	if !roleCodePattern.MatchString(code) || name == "" || !validAccountID(actorID) {
		return dto.RoleResponse{}, ErrInvalidRole
	}
	permissionIDs, err := s.validatePermissions(ctx, request.PermissionIDs)
	if err != nil {
		return dto.RoleResponse{}, err
	}
	exists, err := s.repositories.Roles.CodeExists(ctx, code, "")
	if err != nil {
		return dto.RoleResponse{}, err
	}
	if exists {
		return dto.RoleResponse{}, ErrRoleConflict
	}
	actor := actorID
	value := model.AppRole{Code: code, Name: name, Description: request.Description,
		IsActive: true, CreatedBy: &actor, UpdatedBy: &actor, VersionNo: 1}
	err = s.repositories.Transaction(ctx, func(local *repository.AdministrationRepositories) error {
		if createErr := local.Roles.Create(ctx, &value); createErr != nil {
			return createErr
		}
		return local.RolePermissions.Replace(ctx, value.ID, permissionIDs, actorID)
	})
	if isUniqueViolation(err) {
		return dto.RoleResponse{}, ErrRoleConflict
	}
	if err != nil {
		return dto.RoleResponse{}, err
	}
	return s.Get(ctx, value.ID)
}

func (s *RoleService) List(ctx context.Context, search string, active *bool, page, pageSize int) (dto.RolePageResponse, error) {
	if page < 1 || pageSize < 1 || pageSize > 200 {
		return dto.RolePageResponse{}, ErrInvalidRole
	}
	rows, total, err := s.repositories.Roles.List(ctx, repository.RoleFilter{
		Search: strings.TrimSpace(search), Active: active, Page: page, PageSize: pageSize,
	})
	if err != nil {
		return dto.RolePageResponse{}, err
	}
	items := make([]dto.RoleResponse, 0, len(rows))
	for _, row := range rows {
		response, mapErr := s.mapRole(ctx, row)
		if mapErr != nil {
			return dto.RolePageResponse{}, mapErr
		}
		items = append(items, response)
	}
	return dto.RolePageResponse{Items: items, Page: page, PageSize: pageSize,
		TotalItems: total, TotalPages: pageCount(total, pageSize)}, nil
}

func (s *RoleService) Get(ctx context.Context, roleID string) (dto.RoleResponse, error) {
	if !validAccountID(roleID) {
		return dto.RoleResponse{}, ErrInvalidRole
	}
	value, err := s.repositories.Roles.Get(ctx, roleID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.RoleResponse{}, ErrRoleNotFound
	}
	if err != nil {
		return dto.RoleResponse{}, err
	}
	return s.mapRole(ctx, value)
}

func (s *RoleService) Update(ctx context.Context, roleID string, request dto.UpdateRoleRequest, actorID string) (dto.RoleResponse, error) {
	if !validAccountID(roleID) || !validAccountID(actorID) || strings.TrimSpace(request.Name) == "" || request.ExpectedVersion < 1 {
		return dto.RoleResponse{}, ErrInvalidRole
	}
	if _, err := s.Get(ctx, roleID); err != nil {
		return dto.RoleResponse{}, err
	}
	_, err := s.repositories.Roles.Update(ctx, roleID, request.ExpectedVersion, map[string]interface{}{
		"name": strings.TrimSpace(request.Name), "description": request.Description, "updated_by": actorID,
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.RoleResponse{}, ErrConcurrentRoleUpdate
	}
	if err != nil {
		return dto.RoleResponse{}, err
	}
	return s.Get(ctx, roleID)
}

func (s *RoleService) ReplacePermissions(ctx context.Context, roleID string, request dto.ReplaceRolePermissionsRequest, actorID string) (dto.RoleResponse, error) {
	if !validAccountID(roleID) || !validAccountID(actorID) || request.PermissionIDs == nil || request.ExpectedVersion < 1 {
		return dto.RoleResponse{}, ErrInvalidRole
	}
	role, err := s.repositories.Roles.Get(ctx, roleID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.RoleResponse{}, ErrRoleNotFound
	}
	if err != nil {
		return dto.RoleResponse{}, err
	}
	permissionIDs, err := s.validatePermissions(ctx, *request.PermissionIDs)
	if err != nil {
		return dto.RoleResponse{}, err
	}
	err = s.repositories.Transaction(ctx, func(local *repository.AdministrationRepositories) error {
		if replaceErr := local.RolePermissions.Replace(ctx, roleID, permissionIDs, actorID); replaceErr != nil {
			return replaceErr
		}
		if _, updateErr := local.Roles.Update(ctx, roleID, request.ExpectedVersion, map[string]interface{}{"updated_by": actorID}); updateErr != nil {
			return updateErr
		}
		if role.IsActive {
			return local.Permissions.EnsureSecurityAdministrator(ctx)
		}
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.RoleResponse{}, ErrConcurrentRoleUpdate
	}
	if errors.Is(err, repository.ErrLastSecurityAdministrator) {
		return dto.RoleResponse{}, ErrLastSecurityAdmin
	}
	if err != nil {
		return dto.RoleResponse{}, err
	}
	return s.Get(ctx, roleID)
}

func (s *RoleService) Deactivate(ctx context.Context, roleID string, expectedVersion int, actorID string) (dto.RoleResponse, error) {
	return s.ChangeStatus(ctx, roleID, dto.ChangeRoleStatusRequest{IsActive: false, ExpectedVersion: expectedVersion}, actorID)
}

func (s *RoleService) ChangeStatus(ctx context.Context, roleID string, request dto.ChangeRoleStatusRequest, actorID string) (dto.RoleResponse, error) {
	if !validAccountID(roleID) || !validAccountID(actorID) || request.ExpectedVersion < 1 {
		return dto.RoleResponse{}, ErrInvalidRole
	}
	role, err := s.repositories.Roles.Get(ctx, roleID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.RoleResponse{}, ErrRoleNotFound
	}
	if err != nil {
		return dto.RoleResponse{}, err
	}
	err = s.repositories.Transaction(ctx, func(local *repository.AdministrationRepositories) error {
		if _, updateErr := local.Roles.Update(ctx, roleID, request.ExpectedVersion, map[string]interface{}{"is_active": request.IsActive, "updated_by": actorID}); updateErr != nil {
			return updateErr
		}
		if role.IsActive && !request.IsActive {
			return local.Permissions.EnsureSecurityAdministrator(ctx)
		}
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.RoleResponse{}, ErrConcurrentRoleUpdate
	}
	if errors.Is(err, repository.ErrLastSecurityAdministrator) {
		return dto.RoleResponse{}, ErrLastSecurityAdmin
	}
	if err != nil {
		return dto.RoleResponse{}, err
	}
	return s.Get(ctx, roleID)
}

func (s *RoleService) Assign(ctx context.Context, accountID, roleID, actorID string) error {
	if !validAccountID(accountID) || !validAccountID(roleID) || !validAccountID(actorID) {
		return ErrInvalidRole
	}
	if _, err := s.repositories.Accounts.FindByID(ctx, accountID); errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrAccountNotFound
	} else if err != nil {
		return err
	}
	active, err := s.repositories.Roles.ActiveExists(ctx, roleID)
	if err != nil {
		return err
	}
	if !active {
		return ErrInvalidRole
	}
	return s.repositories.AccountRoles.Assign(ctx, &model.AccountRole{AccountID: accountID, RoleID: roleID, AssignedBy: &actorID})
}

func (s *RoleService) Revoke(ctx context.Context, accountID, roleID string) error {
	if !validAccountID(accountID) || !validAccountID(roleID) {
		return ErrInvalidRole
	}
	err := s.repositories.Transaction(ctx, func(local *repository.AdministrationRepositories) error {
		if revokeErr := local.AccountRoles.Revoke(ctx, accountID, roleID); revokeErr != nil {
			return revokeErr
		}
		return local.Permissions.EnsureSecurityAdministrator(ctx)
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrAccountRoleNotFound
	}
	if errors.Is(err, repository.ErrLastSecurityAdministrator) {
		return ErrLastSecurityAdmin
	}
	return err
}

func (s *RoleService) validatePermissions(ctx context.Context, values []string) ([]string, error) {
	result := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, id := range values {
		if !validAccountID(id) || seen[id] {
			return nil, ErrInvalidRole
		}
		active, err := s.repositories.Permissions.ActivePermissionExists(ctx, id)
		if err != nil {
			return nil, err
		}
		if !active {
			return nil, ErrInvalidRole
		}
		seen[id] = true
		result = append(result, id)
	}
	return result, nil
}

func (s *RoleService) mapRole(ctx context.Context, value model.AppRole) (dto.RoleResponse, error) {
	rows, err := s.repositories.RolePermissions.List(ctx, value.ID)
	if err != nil {
		return dto.RoleResponse{}, err
	}
	permissions := make([]dto.PermissionResponse, 0, len(rows))
	for _, row := range rows {
		permissions = append(permissions, dto.PermissionResponse{PermissionID: row.PermissionID,
			Code: row.Code, Name: row.Name, ModuleCode: row.ModuleCode})
	}
	return dto.RoleResponse{RoleID: value.ID, Code: value.Code, Name: value.Name,
		Description: value.Description, IsActive: value.IsActive, Permissions: permissions,
		CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, VersionNo: value.VersionNo}, nil
}
