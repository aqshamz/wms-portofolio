package main

import (
	"context"
	"log"

	"wms-api/config"
	authcontroller "wms-api/controller/authentication"
	inboundcontroller "wms-api/controller/inbound"
	inventorycontroller "wms-api/controller/inventory"
	mastercontroller "wms-api/controller/master"
	outboundcontroller "wms-api/controller/outbound"
	stockcontrolcontroller "wms-api/controller/stock_control"
	"wms-api/middleware"
	authrepository "wms-api/repository/authentication"
	inboundrepository "wms-api/repository/inbound"
	inventoryrepository "wms-api/repository/inventory"
	masterrepository "wms-api/repository/master"
	outboundrepository "wms-api/repository/outbound"
	stockcontrolrepository "wms-api/repository/stock_control"
	"wms-api/routes"
	authservice "wms-api/services/authentication"
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

	if err := authrepository.Migrate(db); err != nil {
		log.Fatalf("migrate authentication tables: %v", err)
	}
	if err := masterrepository.Migrate(db); err != nil {
		log.Fatalf("migrate master tables: %v", err)
	}

	statusRepository := authrepository.NewAccountStatusRepository(db)
	if err := masterrepository.MigrateCatalog(db); err != nil {
		log.Fatalf("migrate catalog tables: %v", err)
	}
	catalogService := masterservice.NewCatalogService(masterrepository.NewCatalogRepositories(db))
	if err := catalogService.SeedCatalog(context.Background()); err != nil {
		log.Fatalf("seed catalog references: %v", err)
	}
	catalogController := mastercontroller.NewCatalogController(catalogService)
	if err := masterrepository.MigrateOperational(db); err != nil {
		log.Fatalf("migrate operational configuration: %v", err)
	}
	operationalService, err := masterservice.NewOperationalService(masterrepository.NewOperationalRepositories(db), cfg.Database.Timezone)
	if err != nil {
		log.Fatalf("configure operational timezone: %v", err)
	}
	if err := operationalService.SeedOperational(context.Background()); err != nil {
		log.Fatalf("seed operational configuration: %v", err)
	}
	operationalController := mastercontroller.NewOperationalController(operationalService)
	if err := inventoryrepository.Migrate(db); err != nil {
		log.Fatalf("migrate inventory identity: %v", err)
	}
	if err := stockcontrolrepository.Migrate(db); err != nil {
		log.Fatalf("migrate stock control: %v", err)
	}
	if err := inboundrepository.Migrate(db); err != nil {
		log.Fatalf("migrate inbound tables: %v", err)
	}
	if err := inboundrepository.SeedReferenceData(db); err != nil {
		log.Fatalf("seed inbound references: %v", err)
	}
	if err := outboundrepository.Migrate(db); err != nil {
		log.Fatalf("migrate outbound tables: %v", err)
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
	policyRepository := authrepository.NewAuthenticationPolicyRepository(db)
	reasonRepository := authrepository.NewSessionRevocationReasonRepository(db)
	accountRepository := authrepository.NewAppAccountRepository(db)
	sessionRepository := authrepository.NewAppSessionRepository(db)

	bootstrapService := authservice.NewBootstrapService(
		statusRepository,
		policyRepository,
		reasonRepository,
		accountRepository,
	)
	if err := bootstrapService.Seed(context.Background(), cfg.Auth); err != nil {
		log.Fatalf("seed authentication data: %v", err)
	}

	authenticationService := authservice.NewService(accountRepository, sessionRepository)
	authenticationController := authcontroller.NewController(authenticationService)
	authenticationMiddleware := middleware.NewAuthentication(authenticationService)
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
		AuthenticationMiddleware: authenticationMiddleware,
		MasterController:         masterController,
		CatalogController:        catalogController,
		OperationalController:    operationalController,
		InventoryController:      inventoryController,
		InboundController:        inboundController,
		OutboundController:       outboundController,
		StockControlController:   stockControlController,
	})
	log.Printf("WMS API listening on %s", cfg.App.Address())
	if err := router.Run(cfg.App.Address()); err != nil {
		log.Fatalf("run HTTP server: %v", err)
	}
}
