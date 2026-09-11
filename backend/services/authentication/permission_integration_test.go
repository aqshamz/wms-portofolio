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
	authmodel "wms-api/models/authentication"
	repository "wms-api/repository/authentication"
	masterrepository "wms-api/repository/master"
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
}

func authOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
