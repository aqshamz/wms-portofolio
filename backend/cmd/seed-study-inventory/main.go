package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"wms-api/config"
	authrepository "wms-api/repository/authentication"
	inventoryrepository "wms-api/repository/inventory"
	inventoryservice "wms-api/services/inventory"
)

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	actorUsername := flag.String("actor", cfg.Auth.BootstrapAdminUsername, "existing account username recorded as the seed actor")
	flag.Parse()
	db, err := config.OpenDatabase(cfg.Database)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	ctx := context.Background()
	actor, err := authrepository.NewAppAccountRepository(db).FindByUsername(ctx, *actorUsername)
	if err != nil {
		return fmt.Errorf("find seed actor %q: %w", *actorUsername, err)
	}
	manifest, err := inventoryservice.NewService(inventoryrepository.NewRepositories(db)).SeedStudyInventory(ctx, actor.ID)
	if err != nil {
		return fmt.Errorf("seed study inventory (run scripts/seed-study-data.ps1 first): %w", err)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(manifest)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
