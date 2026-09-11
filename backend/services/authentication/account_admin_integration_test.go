package authentication

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"wms-api/config"
	dto "wms-api/dto/authentication"
	authmodel "wms-api/models/authentication"
	mastermodel "wms-api/models/master"
	repository "wms-api/repository/authentication"
	masterrepository "wms-api/repository/master"
	masterservice "wms-api/services/master"
)

func TestAccountAdministrationLifecyclePostgreSQL(t *testing.T) {
	if os.Getenv("WMS_INTEGRATION_TEST") != "1" {
		t.Skip("set WMS_INTEGRATION_TEST=1")
	}
	_ = godotenv.Load("../../.env")
	cfg, err := config.Load()
	authOK(t, err)
	cfg.Database.AutoCreate = false
	db, err := config.OpenDatabase(cfg.Database)
	authOK(t, err)
	pool, err := db.DB()
	authOK(t, err)
	defer pool.Close()
	tx := db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Begin()
	authOK(t, tx.Error)
	defer tx.Rollback()
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	schema := "account_admin_test_" + suffix
	authOK(t, tx.Exec("CREATE SCHEMA "+schema).Error)
	authOK(t, tx.Exec("SET LOCAL search_path TO "+schema).Error)
	authOK(t, repository.Migrate(tx))
	authOK(t, masterrepository.Migrate(tx))
	authOK(t, masterrepository.MigrateCatalog(tx))
	authOK(t, masterrepository.MigrateOperational(tx))
	authOK(t, repository.MigratePermissions(tx))
	authOK(t, repository.MigrateAdministration(tx))

	ctx := context.Background()
	authOK(t, NewBootstrapService(repository.NewAccountStatusRepository(tx),
		repository.NewAuthenticationPolicyRepository(tx), repository.NewSessionRevocationReasonRepository(tx),
		repository.NewAppAccountRepository(tx), repository.NewAccountPermissionRepository(tx)).Seed(ctx, config.AuthConfig{}))
	authOK(t, masterservice.NewCatalogService(masterrepository.NewCatalogRepositories(tx)).SeedCatalog(ctx))
	operational, err := masterservice.NewOperationalService(masterrepository.NewOperationalRepositories(tx), "Asia/Jakarta")
	authOK(t, err)
	authOK(t, operational.SeedOperational(ctx))

	var active authmodel.AccountStatus
	authOK(t, tx.Where("code = 'ACTIVE'").Take(&active).Error)
	adminSubject := "account-admin-" + suffix
	admin := authmodel.AppAccount{Username: "account-admin-" + suffix, DisplayName: "Account admin",
		AccountStatusID: active.ID, ExternalSubject: &adminSubject, VersionNo: 1}
	authOK(t, tx.Create(&admin).Error)
	permissions := repository.NewAccountPermissionRepository(tx)
	authOK(t, permissions.GrantAll(ctx, admin.ID))
	owner := mastermodel.Organization{Code: "AAO_" + suffix, Name: "Account owner", TimezoneName: "Asia/Jakarta"}
	authOK(t, tx.Create(&owner).Error)
	warehouse := mastermodel.Warehouse{OperatorID: owner.ID, Code: "AAW_" + suffix, Name: "Account warehouse", TimezoneName: "Asia/Jakarta"}
	authOK(t, tx.Create(&warehouse).Error)

	repositories := repository.NewAdministrationRepositories(tx)
	accounts := NewAccountAdminService(repositories)
	roles := NewRoleService(repositories)
	workerPassword := "Worker-Password-2026!"
	worker, err := accounts.Create(ctx, dto.CreateAccountRequest{Username: "worker." + suffix,
		Email: stringPointer("worker." + suffix + "@example.com"), DisplayName: "Warehouse Worker",
		Password: workerPassword, AccountStatusID: &active.ID,
		PreferredTimezone: stringPointer("Asia/Jakarta")}, admin.ID)
	authOK(t, err)
	if worker.VersionNo != 1 || worker.Status.Code != "ACTIVE" {
		t.Fatalf("unexpected created worker: %#v", worker.AccountSummaryResponse)
	}
	_, err = accounts.Create(ctx, dto.CreateAccountRequest{Username: worker.Username,
		DisplayName: "Duplicate", Password: workerPassword}, admin.ID)
	if !errors.Is(err, ErrAccountConflict) {
		t.Fatalf("expected duplicate account conflict, got %v", err)
	}

	available, err := permissions.Available(ctx)
	authOK(t, err)
	permissionIDs := make([]string, 0, 2)
	for _, permission := range available {
		if permission.Code == "INBOUND.READ" || permission.Code == "INBOUND.WRITE" {
			permissionIDs = append(permissionIDs, permission.PermissionID)
		}
	}
	role, err := roles.Create(ctx, dto.CreateRoleRequest{Code: "RECEIVER", Name: "Receiver", PermissionIDs: permissionIDs}, admin.ID)
	authOK(t, err)
	authOK(t, roles.Assign(ctx, worker.AccountID, role.RoleID, admin.ID))
	authOK(t, accounts.GrantOwner(ctx, worker.AccountID, owner.ID, admin.ID))
	authOK(t, accounts.GrantWarehouse(ctx, worker.AccountID, warehouse.ID, admin.ID))
	worker, err = accounts.Get(ctx, worker.AccountID)
	authOK(t, err)
	if len(worker.Roles) != 1 || len(worker.OwnerAccess) != 1 || len(worker.WarehouseAccess) != 1 ||
		!hasPermissionCode(worker.EffectivePermissions, "INBOUND.WRITE") {
		t.Fatalf("aggregate relationships missing: %#v", worker)
	}
	role, err = roles.Deactivate(ctx, role.RoleID, role.VersionNo, admin.ID)
	authOK(t, err)
	worker, err = accounts.Get(ctx, worker.AccountID)
	authOK(t, err)
	if hasPermissionCode(worker.EffectivePermissions, "INBOUND.WRITE") || len(worker.Roles) != 1 {
		t.Fatal("inactive role remained effective or its historical assignment was deleted")
	}
	role, err = roles.ChangeStatus(ctx, role.RoleID, dto.ChangeRoleStatusRequest{IsActive: true, ExpectedVersion: role.VersionNo}, admin.ID)
	authOK(t, err)

	authentication := NewService(repositories)
	_, err = authentication.Login(ctx, dto.LoginRequest{Identifier: worker.Username, Password: "wrong-password"}, ClientInfo{})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	var failedLoginAccount authmodel.AppAccount
	authOK(t, tx.Where("account_id = ?", worker.AccountID).Take(&failedLoginAccount).Error)
	if failedLoginAccount.FailedLoginCount != 1 {
		t.Fatalf("failed login update was not committed: %d", failedLoginAccount.FailedLoginCount)
	}
	login, err := authentication.Login(ctx, dto.LoginRequest{Identifier: worker.Username, Password: workerPassword}, ClientInfo{})
	authOK(t, err)
	authenticated, err := authentication.Authenticate(ctx, login.Token)
	authOK(t, err)
	if !hasPermissionCode(authenticated.Permissions, "INBOUND.WRITE") {
		t.Fatal("active role permission missing from authenticated session")
	}
	role, err = roles.Deactivate(ctx, role.RoleID, role.VersionNo, admin.ID)
	authOK(t, err)
	authenticated, err = authentication.Authenticate(ctx, login.Token)
	authOK(t, err)
	if hasPermissionCode(authenticated.Permissions, "INBOUND.WRITE") {
		t.Fatal("role revocation did not take effect for an existing session")
	}
	role, err = roles.ChangeStatus(ctx, role.RoleID, dto.ChangeRoleStatusRequest{
		IsActive: true, ExpectedVersion: role.VersionNo}, admin.ID)
	authOK(t, err)
	worker, err = accounts.Get(ctx, worker.AccountID)
	authOK(t, err)
	var disabled authmodel.AccountStatus
	authOK(t, tx.Where("code = 'DISABLED'").Take(&disabled).Error)
	worker, err = accounts.ChangeStatus(ctx, worker.AccountID, dto.ChangeAccountStatusRequest{
		AccountStatusID: disabled.ID, ExpectedVersion: worker.VersionNo}, admin.ID)
	authOK(t, err)
	if worker.Status.AllowsLogin {
		t.Fatal("disabled account still allows login")
	}
	if _, err := authentication.Authenticate(ctx, login.Token); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("disabled account session remained valid: %v", err)
	}
	if len(worker.Roles) != 1 || len(worker.OwnerAccess) != 1 || len(worker.WarehouseAccess) != 1 {
		t.Fatal("deactivation deleted historical access relationships")
	}

	worker, err = accounts.ChangeStatus(ctx, worker.AccountID, dto.ChangeAccountStatusRequest{
		AccountStatusID: active.ID, ExpectedVersion: worker.VersionNo}, admin.ID)
	authOK(t, err)
	if _, err := authentication.Authenticate(ctx, login.Token); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("revoked session became valid after account reactivation: %v", err)
	}
	login, err = authentication.Login(ctx, dto.LoginRequest{Identifier: worker.Username, Password: workerPassword}, ClientInfo{})
	authOK(t, err)
	worker, err = accounts.Get(ctx, worker.AccountID)
	authOK(t, err)
	if worker.ActiveSessionCount != 1 {
		t.Fatalf("expected one active session, got %d", worker.ActiveSessionCount)
	}
	newPassword := "New-Worker-Password-2026!"
	authOK(t, accounts.ResetPassword(ctx, worker.AccountID, dto.ResetAccountPasswordRequest{
		Password: newPassword, ExpectedVersion: worker.VersionNo}, admin.ID))
	if _, err := authentication.Authenticate(ctx, login.Token); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("password reset did not revoke session: %v", err)
	}
	login, err = authentication.Login(ctx, dto.LoginRequest{Identifier: worker.Username, Password: newPassword}, ClientInfo{})
	authOK(t, err)
	for attempt := 0; attempt < 5; attempt++ {
		_, err = authentication.Login(ctx, dto.LoginRequest{Identifier: worker.Username, Password: "wrong-password"}, ClientInfo{})
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("failed-login attempt %d returned %v", attempt+1, err)
		}
	}
	if _, err := authentication.Authenticate(ctx, login.Token); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("account lockout did not revoke the existing session: %v", err)
	}
	worker, err = accounts.Get(ctx, worker.AccountID)
	authOK(t, err)
	if worker.LockedUntil == nil || worker.ActiveSessionCount != 0 {
		t.Fatalf("unexpected lockout state: %#v", worker.AccountSummaryResponse)
	}
	worker, err = accounts.Unlock(ctx, worker.AccountID, worker.VersionNo, admin.ID)
	authOK(t, err)
	login, err = authentication.Login(ctx, dto.LoginRequest{Identifier: worker.Username, Password: newPassword}, ClientInfo{})
	authOK(t, err)
	worker, err = accounts.Get(ctx, worker.AccountID)
	authOK(t, err)
	policies, err := accounts.Policies(ctx)
	authOK(t, err)
	if len(policies) == 0 {
		t.Fatal("authentication policy seed missing")
	}
	policyID := policies[0].AuthenticationPolicyID
	worker, err = accounts.Update(ctx, worker.AccountID, dto.UpdateAccountRequest{
		Username: worker.Username, Email: worker.Email, DisplayName: worker.DisplayName,
		AuthenticationPolicyID: &policyID, PreferredTimezone: worker.PreferredTimezone,
		ExpectedVersion: worker.VersionNo,
	}, admin.ID)
	authOK(t, err)
	if _, err := authentication.Authenticate(ctx, login.Token); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("authentication policy change did not revoke the existing session: %v", err)
	}
	login, err = authentication.Login(ctx, dto.LoginRequest{Identifier: worker.Username, Password: newPassword}, ClientInfo{})
	authOK(t, err)
	authOK(t, accounts.RevokeSessions(ctx, worker.AccountID))
	if _, err := authentication.Authenticate(ctx, login.Token); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("manual session revocation left session valid: %v", err)
	}
	worker, err = accounts.Get(ctx, worker.AccountID)
	authOK(t, err)
	if worker.ActiveSessionCount != 0 {
		t.Fatalf("expected no active sessions after revocation, got %d", worker.ActiveSessionCount)
	}
}
