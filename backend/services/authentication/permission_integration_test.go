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
	billingrepository "wms-api/repository/billing"
	inboundrepository "wms-api/repository/inbound"
	masterrepository "wms-api/repository/master"
	outboundrepository "wms-api/repository/outbound"
	"wms-api/requestscope"
	masterservice "wms-api/services/master"
)

func TestPermissionAdministrationPostgreSQL(t *testing.T) {
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
	schema := fmt.Sprintf("permission_test_%d", time.Now().UnixNano())
	authOK(t, tx.Exec("CREATE SCHEMA "+schema).Error)
	authOK(t, tx.Exec("SET LOCAL search_path TO "+schema).Error)
	authOK(t, repository.Migrate(tx))
	authOK(t, masterrepository.Migrate(tx))
	authOK(t, masterrepository.MigrateCatalog(tx))
	authOK(t, masterrepository.MigrateOperational(tx))
	authOK(t, repository.MigratePermissions(tx))
	authOK(t, repository.MigrateAdministration(tx))

	ctx := context.Background()
	operational, err := masterservice.NewOperationalService(masterrepository.NewOperationalRepositories(tx), "Asia/Jakarta")
	authOK(t, err)
	authOK(t, operational.SeedOperational(ctx))
	status := authmodel.AccountStatus{Code: "ACTIVE", Name: "Active", AllowsLogin: true, IsActive: true}
	authOK(t, tx.Create(&status).Error)
	firstSubject, secondSubject := "first-admin", "second-admin"
	first := authmodel.AppAccount{Username: "first-admin", DisplayName: "First", AccountStatusID: status.ID, ExternalSubject: &firstSubject}
	second := authmodel.AppAccount{Username: "second-admin", DisplayName: "Second", AccountStatusID: status.ID, ExternalSubject: &secondSubject}
	authOK(t, tx.Create(&first).Error)
	authOK(t, tx.Create(&second).Error)

	permissionRepository := repository.NewAccountPermissionRepository(tx)
	authOK(t, permissionRepository.GrantAll(ctx, first.ID))
	service := NewPermissionService(repository.NewAppAccountRepository(tx), permissionRepository)
	available, err := service.Available(ctx)
	authOK(t, err)
	var securityWrite string
	for _, permission := range available {
		if permission.Code == "SECURITY.WRITE" {
			securityWrite = permission.PermissionID
		}
	}
	if securityWrite == "" {
		t.Fatal("SECURITY.WRITE was not seeded")
	}
	if err := service.Revoke(ctx, first.ID, securityWrite); !errors.Is(err, ErrLastSecurityAdmin) {
		t.Fatalf("expected last-admin protection, got %v", err)
	}
	authOK(t, service.Grant(ctx, second.ID, securityWrite, first.ID))
	authOK(t, service.Revoke(ctx, first.ID, securityWrite))
	codes, err := permissionRepository.Codes(ctx, first.ID)
	authOK(t, err)
	for _, code := range codes {
		if code == "SECURITY.WRITE" {
			t.Fatal("revoked permission is still present")
		}
	}
	// The full-access role is additive and idempotent, not a username bypass.
	authOK(t, permissionRepository.EnsureSuperadmin(ctx, first.ID))
	authOK(t, permissionRepository.EnsureSuperadmin(ctx, first.ID))
	codes, err = permissionRepository.Codes(ctx, first.ID)
	authOK(t, err)
	if !hasPermissionCode(codes, "*") {
		t.Fatal("superadmin wildcard is missing")
	}
	authOK(t, permissionRepository.GrantAll(ctx, second.ID))
	secondCodes, err := permissionRepository.Codes(ctx, second.ID)
	authOK(t, err)
	if hasPermissionCode(secondCodes, "*") {
		t.Fatal("granting all named permissions implicitly granted superadmin")
	}
	var superRole authmodel.AppRole
	authOK(t, tx.Where("code=?", "SUPERADMIN").Take(&superRole).Error)
	var roleCount int64
	authOK(t, tx.Model(&authmodel.AccountRole{}).Where("account_id=? AND role_id=?", first.ID, superRole.ID).Count(&roleCount).Error)
	if roleCount != 1 {
		t.Fatalf("duplicate superadmin assignments: %d", roleCount)
	}
	owner := mastermodel.Organization{Code: "SUPER_OWNER", Name: "Superadmin owner", TimezoneName: "Asia/Jakarta"}
	authOK(t, tx.Create(&owner).Error)
	warehouse := mastermodel.Warehouse{OperatorID: owner.ID, Code: "SUPER_WH", Name: "Superadmin warehouse", TimezoneName: "Asia/Jakarta"}
	authOK(t, tx.Create(&warehouse).Error)
	warehouseRepository := masterrepository.NewWarehouseRepository(tx)
	ordinaryContext := requestscope.WithPrincipal(ctx, requestscope.Principal{AccountID: second.ID})
	rows, total, err := warehouseRepository.List(ordinaryContext, nil, nil, nil, 20, 0)
	authOK(t, err)
	if total != 0 || len(rows) != 0 {
		t.Fatal("ordinary administrator unexpectedly bypassed warehouse scope")
	}
	superContext := requestscope.WithPrincipal(ctx, requestscope.Principal{AccountID: first.ID, Unrestricted: true})
	rows, total, err = warehouseRepository.List(superContext, nil, nil, nil, 20, 0)
	authOK(t, err)
	if total != 1 || len(rows) != 1 || rows[0].ID != warehouse.ID {
		t.Fatal("superadmin could not see warehouses without individual scope grants")
	}
	for name, scope := range map[string]interface {
		Allowed(context.Context, string, string, string) (bool, error)
	}{
		"inbound":  inboundrepository.NewInboundScopeRepository(tx),
		"outbound": outboundrepository.NewOutboundScopeRepository(tx),
		"billing":  billingrepository.NewScopeRepository(tx),
	} {
		allowed, err := scope.Allowed(superContext, first.ID, owner.ID, warehouse.ID)
		authOK(t, err)
		if !allowed {
			t.Fatalf("superadmin denied %s scope without access grants", name)
		}
		allowed, err = scope.Allowed(ordinaryContext, second.ID, owner.ID, warehouse.ID)
		authOK(t, err)
		if allowed {
			t.Fatalf("ordinary administrator bypassed %s scope", name)
		}
	}
	eligible, err := inboundrepository.NewPutawayTaskRepository(tx).AccountCanPutaway(ctx, first.ID, owner.ID, warehouse.ID)
	authOK(t, err)
	if !eligible {
		t.Fatal("superadmin without scope grants cannot be assigned putaway")
	}
	// Wildcard-only access still counts for last-security-administrator safety.
	var wildcard mastermodel.AppPermission
	authOK(t, tx.Where("code=?", "*").Take(&wildcard).Error)
	roles := NewRoleService(repository.NewAdministrationRepositories(tx))
	wildcardIDs := []string{wildcard.ID}
	updatedRole, err := roles.ReplacePermissions(ctx, superRole.ID, dto.ReplaceRolePermissionsRequest{PermissionIDs: &wildcardIDs, ExpectedVersion: superRole.VersionNo}, first.ID)
	authOK(t, err)
	authOK(t, service.Revoke(ctx, second.ID, securityWrite))
	if err := roles.Revoke(ctx, first.ID, superRole.ID); !errors.Is(err, ErrLastSecurityAdmin) {
		t.Fatalf("last wildcard admin could be revoked: %v", err)
	}
	if _, err := roles.Deactivate(ctx, superRole.ID, updatedRole.VersionNo, first.ID); !errors.Is(err, ErrLastSecurityAdmin) {
		t.Fatalf("last wildcard admin role could be disabled: %v", err)
	}
	// With another administrator available, deactivation removes wildcard access.
	authOK(t, service.Grant(ctx, second.ID, securityWrite, first.ID))
	_, err = roles.Deactivate(ctx, superRole.ID, updatedRole.VersionNo, first.ID)
	authOK(t, err)
	codes, err = permissionRepository.Codes(ctx, first.ID)
	authOK(t, err)
	if hasPermissionCode(codes, "*") {
		t.Fatal("inactive superadmin role retained unrestricted access")
	}
	eligible, err = inboundrepository.NewPutawayTaskRepository(tx).AccountCanPutaway(ctx, first.ID, owner.ID, warehouse.ID)
	authOK(t, err)
	if eligible {
		t.Fatal("inactive superadmin remained assignable without scope grants")
	}
}

func authOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
