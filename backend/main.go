package main

import (
	"context"
	"log"

	"wms-api/config"
	authcontroller "wms-api/controller/authentication"
	"wms-api/middleware"
	authrepository "wms-api/repository/authentication"
	"wms-api/routes"
	authservice "wms-api/services/authentication"
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

	statusRepository := authrepository.NewAccountStatusRepository(db)
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

	router := routes.New(db, routes.Dependencies{
		AuthenticationController: authenticationController,
		AuthenticationMiddleware: authenticationMiddleware,
	})
	log.Printf("WMS API listening on %s", cfg.App.Address())
	if err := router.Run(cfg.App.Address()); err != nil {
		log.Fatalf("run HTTP server: %v", err)
	}
}
