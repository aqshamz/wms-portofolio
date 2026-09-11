package authentication

import (
	"context"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"

	dto "wms-api/dto/authentication"
	authmodel "wms-api/models/authentication"
	mastermodel "wms-api/models/master"
	repository "wms-api/repository/authentication"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidAccount          = errors.New("invalid account data")
	ErrAccountNotFound         = errors.New("account not found")
	ErrAccountConflict         = errors.New("username or email already exists")
	ErrConcurrentAccountUpdate = errors.New("account changed; reload it before updating")
	ErrAccessNotFound          = errors.New("account access not found")
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{2,99}$`)

type AccountAdminService struct {
	repositories *repository.AdministrationRepositories
}

func NewAccountAdminService(repositories *repository.AdministrationRepositories) *AccountAdminService {
	return &AccountAdminService{repositories: repositories}
}

func (s *AccountAdminService) Create(ctx context.Context, request dto.CreateAccountRequest, actorID string) (dto.AccountDetailResponse, error) {
	username := strings.TrimSpace(request.Username)
	displayName := strings.TrimSpace(request.DisplayName)
	email, err := normalizeAccountEmail(request.Email)
	if err != nil || !usernamePattern.MatchString(username) || displayName == "" || !validAccountID(actorID) || len(request.Password) < 12 || len(request.Password) > 72 {
		return dto.AccountDetailResponse{}, ErrInvalidAccount
	}
	statusID, err := s.resolveStatus(ctx, request.AccountStatusID)
	if err != nil {
		return dto.AccountDetailResponse{}, err
	}
	if err := s.validatePolicy(ctx, request.AuthenticationPolicyID); err != nil {
		return dto.AccountDetailResponse{}, err
	}
	timezone, err := normalizeAccountTimezone(request.PreferredTimezone)
	if err != nil {
		return dto.AccountDetailResponse{}, err
	}
	exists, err := s.repositories.Accounts.IdentityExists(ctx, username, email, "")
	if err != nil {
		return dto.AccountDetailResponse{}, err
	}
	if exists {
		return dto.AccountDetailResponse{}, ErrAccountConflict
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return dto.AccountDetailResponse{}, err
	}
	actor := actorID
	account := authmodel.AppAccount{
		Username: username, Email: email, DisplayName: displayName, PasswordHash: stringPointer(string(hash)),
		AccountStatusID: statusID, AuthenticationPolicyID: request.AuthenticationPolicyID,
		PreferredTimezone: timezone, CreatedBy: &actor, UpdatedBy: &actor, VersionNo: 1,
	}
	if err := s.repositories.Accounts.Create(ctx, &account); err != nil {
		if isUniqueViolation(err) {
			return dto.AccountDetailResponse{}, ErrAccountConflict
		}
		return dto.AccountDetailResponse{}, err
	}
	return s.Get(ctx, account.ID)
}

func (s *AccountAdminService) List(ctx context.Context, search, statusID string, page, pageSize int) (dto.AccountPageResponse, error) {
	if page < 1 || pageSize < 1 || pageSize > 200 || (statusID != "" && !validAccountID(statusID)) {
		return dto.AccountPageResponse{}, ErrInvalidAccount
	}
	rows, total, err := s.repositories.Accounts.List(ctx, repository.AccountFilter{
		Search: strings.TrimSpace(search), StatusID: statusID, Page: page, PageSize: pageSize,
	})
	if err != nil {
		return dto.AccountPageResponse{}, err
	}
	items := make([]dto.AccountSummaryResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapAccountSummary(row))
	}
	return dto.AccountPageResponse{Items: items, Page: page, PageSize: pageSize,
		TotalItems: total, TotalPages: pageCount(total, pageSize)}, nil
}

func (s *AccountAdminService) Get(ctx context.Context, accountID string) (dto.AccountDetailResponse, error) {
	if !validAccountID(accountID) {
		return dto.AccountDetailResponse{}, ErrInvalidAccount
	}
	account, err := s.repositories.Accounts.Details(ctx, accountID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.AccountDetailResponse{}, ErrAccountNotFound
	}
	if err != nil {
		return dto.AccountDetailResponse{}, err
	}
	roles, err := s.repositories.AccountRoles.List(ctx, accountID)
	if err != nil {
		return dto.AccountDetailResponse{}, err
	}
	permissions, err := s.repositories.Permissions.List(ctx, accountID)
	if err != nil {
		return dto.AccountDetailResponse{}, err
	}
	effective, err := s.repositories.Permissions.Codes(ctx, accountID)
	if err != nil {
		return dto.AccountDetailResponse{}, err
	}
	owners, err := s.repositories.OwnerAccess.List(ctx, accountID)
	if err != nil {
		return dto.AccountDetailResponse{}, err
	}
	warehouses, err := s.repositories.WarehouseAccess.List(ctx, accountID)
	if err != nil {
		return dto.AccountDetailResponse{}, err
	}
	activeSessions, err := s.repositories.Sessions.ActiveCountByAccount(ctx, accountID)
	if err != nil {
		return dto.AccountDetailResponse{}, err
	}
	response := dto.AccountDetailResponse{AccountSummaryResponse: mapAccountSummary(account),
		ActiveSessionCount:   activeSessions,
		Roles:                make([]dto.AccountRoleResponse, 0, len(roles)),
		DirectPermissions:    make([]dto.AccountPermissionResponse, 0, len(permissions)),
		EffectivePermissions: effective,
		OwnerAccess:          make([]dto.AccountOwnerAccessResponse, 0, len(owners)),
		WarehouseAccess:      make([]dto.AccountWarehouseAccessResponse, 0, len(warehouses)),
	}
	for _, row := range roles {
		response.Roles = append(response.Roles, dto.AccountRoleResponse{AccountID: row.AccountID,
			RoleID: row.RoleID, Code: row.Code, Name: row.Name, IsActive: row.IsActive,
			AssignedAt: row.AssignedAt, AssignedBy: row.AssignedBy})
	}
	for _, row := range permissions {
		response.DirectPermissions = append(response.DirectPermissions, dto.AccountPermissionResponse{
			AccountID: row.AccountID, PermissionID: row.PermissionID, Code: row.Code,
			Name: row.Name, ModuleCode: row.ModuleCode, GrantedAt: row.GrantedAt, GrantedBy: row.GrantedBy,
		})
	}
	for _, row := range owners {
		response.OwnerAccess = append(response.OwnerAccess, dto.AccountOwnerAccessResponse{
			OwnerID: row.OwnerID, OwnerCode: row.OwnerCode, OwnerName: row.OwnerName,
			GrantedAt: row.GrantedAt, GrantedBy: row.GrantedBy,
		})
	}
	for _, row := range warehouses {
		response.WarehouseAccess = append(response.WarehouseAccess, dto.AccountWarehouseAccessResponse{
			WarehouseID: row.WarehouseID, WarehouseCode: row.WarehouseCode,
			WarehouseName: row.WarehouseName, GrantedAt: row.GrantedAt, GrantedBy: row.GrantedBy,
		})
	}
	return response, nil
}

func (s *AccountAdminService) RevokeSessions(ctx context.Context, accountID string) error {
	if !validAccountID(accountID) {
		return ErrInvalidAccount
	}
	if _, err := s.account(ctx, accountID); err != nil {
		return err
	}
	return s.repositories.Sessions.RevokeByAccount(ctx, accountID, "ADMIN_REVOKED")
}

func (s *AccountAdminService) Update(ctx context.Context, accountID string, request dto.UpdateAccountRequest, actorID string) (dto.AccountDetailResponse, error) {
	if !validAccountID(accountID) || !validAccountID(actorID) || request.ExpectedVersion < 1 {
		return dto.AccountDetailResponse{}, ErrInvalidAccount
	}
	current, err := s.account(ctx, accountID)
	if err != nil {
		return dto.AccountDetailResponse{}, err
	}
	username := strings.TrimSpace(request.Username)
	displayName := strings.TrimSpace(request.DisplayName)
	email, err := normalizeAccountEmail(request.Email)
	if err != nil || !usernamePattern.MatchString(username) || displayName == "" {
		return dto.AccountDetailResponse{}, ErrInvalidAccount
	}
	if err := s.validatePolicy(ctx, request.AuthenticationPolicyID); err != nil {
		return dto.AccountDetailResponse{}, err
	}
	timezone, err := normalizeAccountTimezone(request.PreferredTimezone)
	if err != nil {
		return dto.AccountDetailResponse{}, err
	}
	exists, err := s.repositories.Accounts.IdentityExists(ctx, username, email, accountID)
	if err != nil {
		return dto.AccountDetailResponse{}, err
	}
	if exists {
		return dto.AccountDetailResponse{}, ErrAccountConflict
	}
	err = s.repositories.Transaction(ctx, func(local *repository.AdministrationRepositories) error {
		_, updateErr := local.Accounts.Update(ctx, accountID, request.ExpectedVersion, map[string]interface{}{
			"username": username, "email": email, "display_name": displayName,
			"authentication_policy_id": request.AuthenticationPolicyID,
			"preferred_timezone":       timezone, "updated_by": actorID,
		})
		if updateErr != nil {
			return updateErr
		}
		if !equalOptionalString(current.AuthenticationPolicyID, request.AuthenticationPolicyID) {
			return local.Sessions.RevokeByAccount(ctx, accountID, "AUTH_POLICY_CHANGED")
		}
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.AccountDetailResponse{}, ErrConcurrentAccountUpdate
	}
	if isUniqueViolation(err) {
		return dto.AccountDetailResponse{}, ErrAccountConflict
	}
	if err != nil {
		return dto.AccountDetailResponse{}, err
	}
	return s.Get(ctx, accountID)
}

func (s *AccountAdminService) ChangeStatus(ctx context.Context, accountID string, request dto.ChangeAccountStatusRequest, actorID string) (dto.AccountDetailResponse, error) {
	if !validAccountID(accountID) || !validAccountID(actorID) || !validAccountID(request.AccountStatusID) || request.ExpectedVersion < 1 {
		return dto.AccountDetailResponse{}, ErrInvalidAccount
	}
	if _, err := s.account(ctx, accountID); err != nil {
		return dto.AccountDetailResponse{}, err
	}
	status, err := s.repositories.Statuses.FindByID(ctx, request.AccountStatusID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.AccountDetailResponse{}, ErrInvalidAccount
	}
	if err != nil {
		return dto.AccountDetailResponse{}, err
	}
	err = s.repositories.Transaction(ctx, func(local *repository.AdministrationRepositories) error {
		_, updateErr := local.Accounts.Update(ctx, accountID, request.ExpectedVersion, map[string]interface{}{
			"account_status_id": request.AccountStatusID, "updated_by": actorID,
		})
		if updateErr != nil {
			return updateErr
		}
		if !status.AllowsLogin {
			if revokeErr := local.Sessions.RevokeByAccount(ctx, accountID, "ACCOUNT_DISABLED"); revokeErr != nil {
				return revokeErr
			}
			return local.Permissions.EnsureSecurityAdministrator(ctx)
		}
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.AccountDetailResponse{}, ErrConcurrentAccountUpdate
	}
	if errors.Is(err, repository.ErrLastSecurityAdministrator) {
		return dto.AccountDetailResponse{}, ErrLastSecurityAdmin
	}
	if err != nil {
		return dto.AccountDetailResponse{}, err
	}
	return s.Get(ctx, accountID)
}

func (s *AccountAdminService) Deactivate(ctx context.Context, accountID string, expectedVersion int, actorID string) (dto.AccountDetailResponse, error) {
	status, err := s.repositories.Statuses.FindByCode(ctx, "DISABLED")
	if err != nil {
		return dto.AccountDetailResponse{}, err
	}
	return s.ChangeStatus(ctx, accountID, dto.ChangeAccountStatusRequest{AccountStatusID: status.ID, ExpectedVersion: expectedVersion}, actorID)
}

func (s *AccountAdminService) ResetPassword(ctx context.Context, accountID string, request dto.ResetAccountPasswordRequest, actorID string) error {
	if !validAccountID(accountID) || !validAccountID(actorID) || request.ExpectedVersion < 1 || len(request.Password) < 12 || len(request.Password) > 72 {
		return ErrInvalidAccount
	}
	if _, err := s.account(ctx, accountID); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	err = s.repositories.Transaction(ctx, func(local *repository.AdministrationRepositories) error {
		if updateErr := local.Accounts.SetPassword(ctx, accountID, string(hash), actorID, request.ExpectedVersion); updateErr != nil {
			return updateErr
		}
		return local.Sessions.RevokeByAccount(ctx, accountID, "PASSWORD_CHANGED")
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrConcurrentAccountUpdate
	}
	return err
}

func (s *AccountAdminService) Unlock(ctx context.Context, accountID string, expectedVersion int, actorID string) (dto.AccountDetailResponse, error) {
	if !validAccountID(accountID) || !validAccountID(actorID) || expectedVersion < 1 {
		return dto.AccountDetailResponse{}, ErrInvalidAccount
	}
	if _, err := s.account(ctx, accountID); err != nil {
		return dto.AccountDetailResponse{}, err
	}
	err := s.repositories.Accounts.Unlock(ctx, accountID, actorID, expectedVersion)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.AccountDetailResponse{}, ErrConcurrentAccountUpdate
	}
	if err != nil {
		return dto.AccountDetailResponse{}, err
	}
	return s.Get(ctx, accountID)
}

func (s *AccountAdminService) Statuses(ctx context.Context) ([]dto.AccountStatusResponse, error) {
	rows, err := s.repositories.Statuses.List(ctx)
	if err != nil {
		return nil, err
	}
	response := make([]dto.AccountStatusResponse, 0, len(rows))
	for _, row := range rows {
		response = append(response, dto.AccountStatusResponse{AccountStatusID: row.ID, Code: row.Code, Name: row.Name, AllowsLogin: row.AllowsLogin})
	}
	return response, nil
}

func (s *AccountAdminService) Policies(ctx context.Context) ([]dto.AuthenticationPolicyResponse, error) {
	rows, err := s.repositories.Policies.List(ctx)
	if err != nil {
		return nil, err
	}
	response := make([]dto.AuthenticationPolicyResponse, 0, len(rows))
	for _, row := range rows {
		response = append(response, dto.AuthenticationPolicyResponse{AuthenticationPolicyID: row.ID,
			Code: row.Code, Name: row.Name, MaxFailedAttempts: row.MaxFailedAttempts,
			LockoutSeconds: row.LockoutSeconds, SessionTTLSeconds: row.SessionTTLSeconds, IsDefault: row.IsDefault})
	}
	return response, nil
}

func (s *AccountAdminService) GrantOwner(ctx context.Context, accountID, ownerID, actorID string) error {
	if !validAccountID(accountID) || !validAccountID(ownerID) || !validAccountID(actorID) {
		return ErrInvalidAccount
	}
	if _, err := s.account(ctx, accountID); err != nil {
		return err
	}
	owner, err := s.repositories.Organizations.FindByID(ctx, ownerID)
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && !owner.IsActive) {
		return ErrInvalidAccount
	}
	if err != nil {
		return err
	}
	return s.repositories.OwnerAccess.Grant(ctx, &mastermodel.AccountOwnerAccess{AccountID: accountID, OwnerID: ownerID, GrantedBy: &actorID})
}

func (s *AccountAdminService) RevokeOwner(ctx context.Context, accountID, ownerID string) error {
	if !validAccountID(accountID) || !validAccountID(ownerID) {
		return ErrInvalidAccount
	}
	err := s.repositories.OwnerAccess.Revoke(ctx, accountID, ownerID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrAccessNotFound
	}
	return err
}

func (s *AccountAdminService) GrantWarehouse(ctx context.Context, accountID, warehouseID, actorID string) error {
	if !validAccountID(accountID) || !validAccountID(warehouseID) || !validAccountID(actorID) {
		return ErrInvalidAccount
	}
	if _, err := s.account(ctx, accountID); err != nil {
		return err
	}
	warehouse, err := s.repositories.Warehouses.FindByID(ctx, warehouseID)
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && !warehouse.IsActive) {
		return ErrInvalidAccount
	}
	if err != nil {
		return err
	}
	return s.repositories.WarehouseAccess.Grant(ctx, &mastermodel.AccountWarehouseAccess{AccountID: accountID, WarehouseID: warehouseID, GrantedBy: &actorID})
}

func (s *AccountAdminService) RevokeWarehouse(ctx context.Context, accountID, warehouseID string) error {
	if !validAccountID(accountID) || !validAccountID(warehouseID) {
		return ErrInvalidAccount
	}
	err := s.repositories.WarehouseAccess.Revoke(ctx, accountID, warehouseID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrAccessNotFound
	}
	return err
}

func (s *AccountAdminService) account(ctx context.Context, id string) (authmodel.AppAccount, error) {
	if !validAccountID(id) {
		return authmodel.AppAccount{}, ErrInvalidAccount
	}
	value, err := s.repositories.Accounts.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return authmodel.AppAccount{}, ErrAccountNotFound
	}
	return value, err
}

func (s *AccountAdminService) resolveStatus(ctx context.Context, requested *string) (string, error) {
	if requested == nil {
		status, err := s.repositories.Statuses.FindByCode(ctx, "PENDING")
		return status.ID, err
	}
	status, err := s.repositories.Statuses.FindByID(ctx, *requested)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrInvalidAccount
	}
	return status.ID, err
}

func (s *AccountAdminService) validatePolicy(ctx context.Context, id *string) error {
	if id == nil {
		return nil
	}
	if !validAccountID(*id) {
		return ErrInvalidAccount
	}
	_, err := s.repositories.Policies.FindByID(ctx, *id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrInvalidAccount
	}
	return err
}

func mapAccountSummary(value repository.AccountDetails) dto.AccountSummaryResponse {
	return dto.AccountSummaryResponse{AccountID: value.ID, Username: value.Username, Email: value.Email,
		DisplayName: value.DisplayName, Status: dto.AccountStatusResponse{AccountStatusID: value.AccountStatusID,
			Code: value.StatusCode, Name: value.StatusName, AllowsLogin: value.StatusAllowsLogin},
		AuthenticationPolicyID: value.AuthenticationPolicyID, AuthenticationPolicyCode: value.AuthenticationPolicyCode,
		PreferredTimezone: value.PreferredTimezone, FailedLoginCount: value.FailedLoginCount,
		LockedUntil: value.LockedUntil, LastLoginAt: value.LastLoginAt, CreatedAt: value.CreatedAt,
		UpdatedAt: value.UpdatedAt, VersionNo: value.VersionNo}
}

func normalizeAccountEmail(value *string) (*string, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	normalized := strings.ToLower(strings.TrimSpace(*value))
	parsed, err := mail.ParseAddress(normalized)
	if err != nil || parsed.Address != normalized || len(normalized) > 254 {
		return nil, ErrInvalidAccount
	}
	return &normalized, nil
}

func normalizeAccountTimezone(value *string) (*string, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	normalized := strings.TrimSpace(*value)
	if _, err := time.LoadLocation(normalized); err != nil {
		return nil, ErrInvalidAccount
	}
	return &normalized, nil
}

func validAccountID(value string) bool   { return permissionUUIDPattern.MatchString(value) }
func stringPointer(value string) *string { return &value }
func equalOptionalString(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
func pageCount(total int64, pageSize int) int {
	if total == 0 {
		return 0
	}
	return int((total + int64(pageSize) - 1) / int64(pageSize))
}
func hasPermissionCode(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
func isUniqueViolation(err error) bool {
	var databaseError *pgconn.PgError
	return errors.As(err, &databaseError) && databaseError.Code == "23505"
}
