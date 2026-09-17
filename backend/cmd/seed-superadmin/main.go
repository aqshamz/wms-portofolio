// Explicit, additive development seed for a named existing administrator.
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"

	"wms-api/config"
	repository "wms-api/repository/authentication"
)

func run() error {
	username := flag.String("account", "admin", "existing account username to assign SUPERADMIN")
	flag.Parse()
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	hostIP := net.ParseIP(cfg.Database.Host)
	if cfg.App.Environment == "production" || (cfg.Database.Host != "localhost" && (hostIP == nil || !hostIP.IsLoopback())) {
		return fmt.Errorf("superadmin seed requires a non-production loopback database")
	}
	cfg.Database.AutoCreate = false
	db, err := config.OpenDatabase(cfg.Database)
	if err != nil {
		return err
	}
	pool, err := db.DB()
	if err != nil {
		return err
	}
	defer pool.Close()
	ctx := context.Background()
	account, err := repository.NewAppAccountRepository(db).FindByUsername(ctx, *username)
	if err != nil {
		return fmt.Errorf("find existing administrator: %w", err)
	}
	if err := repository.NewAccountPermissionRepository(db).EnsureSuperadmin(ctx, account.ID); err != nil {
		return err
	}
	fmt.Printf("Assigned SUPERADMIN to %s (%s). Existing grants preserved.\n", account.Username, account.DisplayName)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
