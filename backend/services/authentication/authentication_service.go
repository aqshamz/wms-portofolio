package authentication

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	dto "wms-api/dto/authentication"
	model "wms-api/models/authentication"
	repository "wms-api/repository/authentication"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid username/email or password")
	ErrUnauthorized       = errors.New("invalid or expired session")
	dummyPasswordHash, _  = bcrypt.GenerateFromPassword([]byte("not-a-real-password"), bcrypt.DefaultCost)
)

type ClientInfo struct {
	IPAddress string
	UserAgent string
}

type Service struct {
	repositories *repository.AdministrationRepositories
}

func NewService(repositories *repository.AdministrationRepositories) *Service {
	return &Service{repositories: repositories}
}

func (s *Service) Login(
	ctx context.Context,
	request dto.LoginRequest,
	client ClientInfo,
) (dto.LoginResponse, error) {
	rawToken, err := secureRandomString(32)
	if err != nil {
		return dto.LoginResponse{}, fmt.Errorf("generate session token: %w", err)
	}
	sessionIDPart, err := secureRandomString(24)
	if err != nil {
		return dto.LoginResponse{}, fmt.Errorf("generate session ID: %w", err)
	}

	var account repository.LoginAccount
	var permissions []string
	var now, expiresAt time.Time
	var loginErr error
	err = s.repositories.Transaction(ctx, func(local *repository.AdministrationRepositories) error {
		account, err = local.Accounts.FindForLoginForUpdate(ctx, strings.TrimSpace(request.Identifier))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				compareDummyPassword(request.Password)
				loginErr = ErrInvalidCredentials
				return nil
			}
			return fmt.Errorf("find login account: %w", err)
		}

		passwordValid := account.PasswordHash != nil &&
			bcrypt.CompareHashAndPassword([]byte(*account.PasswordHash), []byte(request.Password)) == nil
		locked := account.LockedUntil != nil && account.LockedUntil.After(time.Now())
		if !passwordValid {
			if account.StatusIsActive && account.AllowsLogin && !locked {
				if recordErr := local.Accounts.RecordFailedLogin(ctx, account.ID); recordErr != nil {
					return fmt.Errorf("record failed login: %w", recordErr)
				}
				if account.FailedLoginCount+1 >= account.MaxFailedAttempts {
					if revokeErr := local.Sessions.RevokeByAccount(ctx, account.ID, "ACCOUNT_LOCKED"); revokeErr != nil {
						return fmt.Errorf("revoke locked account sessions: %w", revokeErr)
					}
				}
			}
			loginErr = ErrInvalidCredentials
			return nil
		}
		if !account.StatusIsActive || !account.AllowsLogin || locked {
			loginErr = ErrInvalidCredentials
			return nil
		}

		permissions, err = local.Permissions.Codes(ctx, account.ID)
		if err != nil {
			return fmt.Errorf("load account permissions: %w", err)
		}
		now = time.Now()
		expiresAt = now.Add(time.Duration(account.SessionTTLSeconds) * time.Second)
		session := model.AppSession{
			ID:         "ses_" + sessionIDPart,
			AccountID:  account.ID,
			TokenHash:  HashToken(rawToken),
			IssuedAt:   now,
			ExpiresAt:  expiresAt,
			LastSeenAt: now,
			IPAddress:  optionalString(client.IPAddress),
			UserAgent:  optionalString(client.UserAgent),
		}
		if completeErr := local.Accounts.RecordSuccessfulLogin(ctx, account.ID); completeErr != nil {
			return fmt.Errorf("complete login: %w", completeErr)
		}
		if createErr := local.Sessions.Create(ctx, &session); createErr != nil {
			return fmt.Errorf("create session: %w", createErr)
		}
		return nil
	})
	if err != nil {
		return dto.LoginResponse{}, err
	}
	if loginErr != nil {
		return dto.LoginResponse{}, loginErr
	}

	return dto.LoginResponse{
		Token:     rawToken,
		TokenType: "Bearer",
		ExpiresAt: expiresAt,
		User: dto.UserResponse{
			AccountID:   account.ID,
			Username:    account.Username,
			Email:       account.Email,
			DisplayName: account.DisplayName,
			LastLoginAt: &now,
			Permissions: permissions,
		},
	}, nil
}

func (s *Service) Authenticate(ctx context.Context, rawToken string) (dto.UserResponse, error) {
	account, err := s.repositories.Sessions.FindValid(ctx, HashToken(rawToken))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.UserResponse{}, ErrUnauthorized
		}
		return dto.UserResponse{}, fmt.Errorf("validate session: %w", err)
	}

	if err := s.repositories.Sessions.Touch(ctx, HashToken(rawToken)); err != nil {
		return dto.UserResponse{}, fmt.Errorf("touch session: %w", err)
	}
	permissions, err := s.repositories.Permissions.Codes(ctx, account.AccountID)
	if err != nil {
		return dto.UserResponse{}, fmt.Errorf("load account permissions: %w", err)
	}

	return dto.UserResponse{
		AccountID:         account.AccountID,
		Username:          account.Username,
		Email:             account.Email,
		DisplayName:       account.DisplayName,
		PreferredTimezone: account.PreferredTimezone,
		Permissions:       permissions,
	}, nil
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	err := s.repositories.Sessions.Revoke(ctx, HashToken(rawToken), "USER_LOGOUT")
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

func (s *Service) LogoutAll(ctx context.Context, rawToken string) error {
	err := s.repositories.Sessions.RevokeAll(ctx, HashToken(rawToken), "USER_LOGOUT_ALL")
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

func HashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

func secureRandomString(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func compareDummyPassword(password string) {
	_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(password))
}
