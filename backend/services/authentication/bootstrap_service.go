package authentication

import (
	"context"
	"errors"
	"fmt"

	"wms-api/config"
	model "wms-api/models/authentication"
	repository "wms-api/repository/authentication"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type BootstrapService struct {
	statuses    *repository.AccountStatusRepository
	policies    *repository.AuthenticationPolicyRepository
	reasons     *repository.SessionRevocationReasonRepository
	accounts    *repository.AppAccountRepository
	permissions *repository.AccountPermissionRepository
}

func NewBootstrapService(
	statuses *repository.AccountStatusRepository,
	policies *repository.AuthenticationPolicyRepository,
	reasons *repository.SessionRevocationReasonRepository,
	accounts *repository.AppAccountRepository,
	permissions *repository.AccountPermissionRepository,
) *BootstrapService {
	return &BootstrapService{
		statuses:    statuses,
		policies:    policies,
		reasons:     reasons,
		accounts:    accounts,
		permissions: permissions,
	}
}

func (s *BootstrapService) Seed(ctx context.Context, cfg config.AuthConfig) error {
	if err := s.policies.Seed(ctx, []model.AuthenticationPolicy{{
		Code:              "DEFAULT",
		Name:              "Default authentication policy",
		MaxFailedAttempts: 5,
		LockoutSeconds:    900,
		SessionTTLSeconds: 28800,
		IsDefault:         true,
		IsActive:          true,
	}}); err != nil {
		return fmt.Errorf("seed authentication policy: %w", err)
	}

	if err := s.statuses.Seed(ctx, []model.AccountStatus{
		{Code: "ACTIVE", Name: "Active", AllowsLogin: true, IsActive: true},
		{Code: "PENDING", Name: "Pending", AllowsLogin: false, IsActive: true},
		{Code: "LOCKED", Name: "Locked", AllowsLogin: false, IsActive: true},
		{Code: "DISABLED", Name: "Disabled", AllowsLogin: false, IsActive: true},
	}); err != nil {
		return fmt.Errorf("seed account statuses: %w", err)
	}

	if err := s.reasons.Seed(ctx, []model.SessionRevocationReason{
		{Code: "USER_LOGOUT", Name: "User logout", IsActive: true},
		{Code: "USER_LOGOUT_ALL", Name: "Logout all", IsActive: true},
		{Code: "PASSWORD_CHANGED", Name: "Password changed", IsActive: true},
		{Code: "ACCOUNT_DISABLED", Name: "Account disabled", IsActive: true},
		{Code: "ADMIN_REVOKED", Name: "Administrator revoked", IsActive: true},
	}); err != nil {
		return fmt.Errorf("seed session revocation reasons: %w", err)
	}

	if !cfg.BootstrapAdminEnabled {
		return nil
	}
	if existing, err := s.accounts.FindByUsername(ctx, cfg.BootstrapAdminUsername); err == nil {
		return s.permissions.GrantAll(ctx, existing.ID)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("find bootstrap admin: %w", err)
	}

	status, err := s.statuses.FindByCode(ctx, "ACTIVE")
	if err != nil {
		return fmt.Errorf("find active account status: %w", err)
	}
	policy, err := s.policies.FindDefault(ctx)
	if err != nil {
		return fmt.Errorf("find default authentication policy: %w", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.BootstrapAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash bootstrap password: %w", err)
	}

	passwordHash := string(hash)
	email := cfg.BootstrapAdminEmail
	policyID := policy.ID
	timezone := cfg.BootstrapAdminTimezone
	account := model.AppAccount{
		Username:               cfg.BootstrapAdminUsername,
		Email:                  &email,
		DisplayName:            cfg.BootstrapAdminDisplayName,
		PasswordHash:           &passwordHash,
		AccountStatusID:        status.ID,
		AuthenticationPolicyID: &policyID,
		PreferredTimezone:      &timezone,
		VersionNo:              1,
	}
	if err := s.accounts.Create(ctx, &account); err != nil {
		return err
	}
	return s.permissions.GrantAll(ctx, account.ID)
}
