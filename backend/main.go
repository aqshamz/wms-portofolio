package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"wms-api/config"
	authcontroller "wms-api/controller/authentication"
	billingcontroller "wms-api/controller/billing"
	inboundcontroller "wms-api/controller/inbound"
	inventorycontroller "wms-api/controller/inventory"
	mastercontroller "wms-api/controller/master"
	outboundcontroller "wms-api/controller/outbound"
	stockcontrolcontroller "wms-api/controller/stock_control"
	"wms-api/middleware"
	auditrepository "wms-api/repository/audit"
	authrepository "wms-api/repository/authentication"
	billingrepository "wms-api/repository/billing"
	inboundrepository "wms-api/repository/inbound"
	inventoryrepository "wms-api/repository/inventory"
	masterrepository "wms-api/repository/master"
	migrationrepository "wms-api/repository/migration"
	outboundrepository "wms-api/repository/outbound"
	stockcontrolrepository "wms-api/repository/stock_control"
	"wms-api/routes"
	authservice "wms-api/services/authentication"
	billingservice "wms-api/services/billing"
	inboundservice "wms-api/services/inbound"
	inventoryservice "wms-api/services/inventory"
	masterservice "wms-api/services/master"
	outboundservice "wms-api/services/outbound"
	stockcontrolservice "wms-api/services/stock_control"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}

	db, err := config.OpenDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("initialize database: %v", err)
	}

	if err := migrationrepository.Apply(context.Background(), db, []migrationrepository.Step{
		{Version: 1, Name: "authentication", Up: authrepository.Migrate},
		{Version: 2, Name: "warehouse master", Up: masterrepository.Migrate},
		{Version: 3, Name: "catalog master", Up: masterrepository.MigrateCatalog},
		{Version: 4, Name: "operational configuration", Up: masterrepository.MigrateOperational},
		{Version: 5, Name: "account permissions", Up: authrepository.MigratePermissions},
		{Version: 6, Name: "inventory", Up: inventoryrepository.Migrate},
		{Version: 7, Name: "stock control", Up: stockcontrolrepository.Migrate},
		{Version: 8, Name: "inbound", Up: inboundrepository.Migrate},
		{Version: 9, Name: "outbound", Up: outboundrepository.Migrate},
		{Version: 10, Name: "billing", Up: billingrepository.Migrate},
		{Version: 11, Name: "API audit", Up: auditrepository.Migrate},
		{Version: 12, Name: "API audit request index", Up: auditrepository.MigrateRequestIndex},
		{Version: 13, Name: "account and role administration", Up: authrepository.MigrateAdministration},
	}); err != nil {
		log.Fatalf("apply database migrations: %v", err)
	}

	statusRepository := authrepository.NewAccountStatusRepository(db)
	catalogService := masterservice.NewCatalogService(masterrepository.NewCatalogRepositories(db))
	if err := catalogService.SeedCatalog(context.Background()); err != nil {
		log.Fatalf("seed catalog references: %v", err)
	}
	catalogController := mastercontroller.NewCatalogController(catalogService)
	operationalService, err := masterservice.NewOperationalService(masterrepository.NewOperationalRepositories(db), cfg.Database.Timezone)
	if err != nil {
		log.Fatalf("configure operational timezone: %v", err)
	}
	if err := operationalService.SeedOperational(context.Background()); err != nil {
		log.Fatalf("seed operational configuration: %v", err)
	}
	operationalController := mastercontroller.NewOperationalController(operationalService)
	if err := inboundrepository.SeedReferenceData(db); err != nil {
		log.Fatalf("seed inbound references: %v", err)
	}
	if err := outboundrepository.SeedReferenceData(db); err != nil {
		log.Fatalf("seed outbound references: %v", err)
	}
	inventoryController := inventorycontroller.NewController(inventoryservice.NewService(inventoryrepository.NewRepositories(db)))
	stockControlController := stockcontrolcontroller.NewController(stockcontrolservice.NewService(stockcontrolrepository.NewRepositories(db)))
	inboundService, err := inboundservice.NewService(inboundrepository.NewRepositories(db), cfg.Database.Timezone)
	if err != nil {
		log.Fatalf("configure inbound timezone: %v", err)
	}
	inboundController := inboundcontroller.NewController(inboundService)
	outboundService, err := outboundservice.NewService(outboundrepository.NewRepositories(db), cfg.Database.Timezone)
	if err != nil {
		log.Fatalf("configure outbound timezone: %v", err)
	}
	outboundController := outboundcontroller.NewController(outboundService)
	billingService, err := billingservice.NewService(billingrepository.NewRepositories(db), cfg.Database.Timezone)
	if err != nil {
		log.Fatalf("configure billing timezone: %v", err)
	}
	billingController := billingcontroller.NewController(billingService)
	policyRepository := authrepository.NewAuthenticationPolicyRepository(db)
	reasonRepository := authrepository.NewSessionRevocationReasonRepository(db)
	accountRepository := authrepository.NewAppAccountRepository(db)
	accountPermissionRepository := authrepository.NewAccountPermissionRepository(db)
	auditRepository := auditrepository.NewAPIAuditLogRepository(db)

	bootstrapService := authservice.NewBootstrapService(
		statusRepository,
		policyRepository,
		reasonRepository,
		accountRepository,
		accountPermissionRepository,
	)
	if err := bootstrapService.Seed(context.Background(), cfg.Auth); err != nil {
		log.Fatalf("seed authentication data: %v", err)
	}

	administrationRepositories := authrepository.NewAdministrationRepositories(db)
	authenticationService := authservice.NewService(administrationRepositories)
	authenticationController := authcontroller.NewController(authenticationService)
	securityController := authcontroller.NewSecurityController(
		authservice.NewPermissionService(accountRepository, accountPermissionRepository),
		authservice.NewAccountAdminService(administrationRepositories),
		authservice.NewRoleService(administrationRepositories),
	)
	authenticationMiddleware := middleware.NewAuthenticationWithAuthorization(authenticationService, cfg.Security.AuthorizationEnforced)
	hardeningMiddleware := middleware.NewHardening(cfg.Security, auditRepository.Create)
	organizationRepository := masterrepository.NewOrganizationRepository(db)
	warehouseRepository := masterrepository.NewWarehouseRepository(db)
	warehouseOwnerRepository := masterrepository.NewWarehouseOwnerRepository(db)
	locationTypeRepository := masterrepository.NewLocationTypeRepository(db)
	warehouseZoneRepository := masterrepository.NewWarehouseZoneRepository(db)
	warehouseLocationRepository := masterrepository.NewWarehouseLocationRepository(db)
	accountOwnerAccessRepository := masterrepository.NewAccountOwnerAccessRepository(db)
	accountWarehouseAccessRepository := masterrepository.NewAccountWarehouseAccessRepository(db)
	organizationService := masterservice.NewOrganizationService(organizationRepository)
	warehouseService := masterservice.NewWarehouseService(warehouseRepository, organizationRepository)
	warehouseStructureService := masterservice.NewWarehouseStructureService(
		warehouseRepository,
		organizationRepository,
		warehouseOwnerRepository,
		locationTypeRepository,
		warehouseZoneRepository,
		warehouseLocationRepository,
	)
	if err := warehouseStructureService.SeedLocationTypes(context.Background()); err != nil {
		log.Fatalf("seed location types: %v", err)
	}
	accessScopeService := masterservice.NewAccessScopeService(
		accountRepository,
		organizationRepository,
		warehouseRepository,
		accountOwnerAccessRepository,
		accountWarehouseAccessRepository,
	)
	masterController := mastercontroller.NewController(
		organizationService,
		warehouseService,
		warehouseStructureService,
		accessScopeService,
	)

	router := routes.New(db, routes.Dependencies{
		AuthenticationController: authenticationController,
		SecurityController:       securityController,
		AuthenticationMiddleware: authenticationMiddleware,
		MasterController:         masterController,
		CatalogController:        catalogController,
		OperationalController:    operationalController,
		InventoryController:      inventoryController,
		InboundController:        inboundController,
		OutboundController:       outboundController,
		StockControlController:   stockControlController,
		BillingController:        billingController,
		Hardening:                hardeningMiddleware,
		TrustedProxies:           cfg.App.TrustedProxies,
	})
	log.Printf("WMS API listening on %s", cfg.App.Address())
	server := &http.Server{Addr: cfg.App.Address(), Handler: router, ReadHeaderTimeout: time.Duration(cfg.App.ReadHeaderTimeoutSeconds) * time.Second, ReadTimeout: time.Duration(cfg.App.ReadTimeoutSeconds) * time.Second, WriteTimeout: time.Duration(cfg.App.WriteTimeoutSeconds) * time.Second, IdleTimeout: time.Duration(cfg.App.IdleTimeoutSeconds) * time.Second}
	stopping, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	errorsFromServer := make(chan error, 1)
	go func() { errorsFromServer <- server.ListenAndServe() }()
	select {
	case err := <-errorsFromServer:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("run HTTP server: %v", err)
		}
	case <-stopping.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.App.ShutdownTimeoutSeconds)*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
		if err := <-errorsFromServer; !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server stopped: %v", err)
		}
	}
}
