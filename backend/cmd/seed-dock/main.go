// Development-only, additive dock seed. Existing records are never relabelled.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"wms-api/config"
	model "wms-api/models/master"
	authrepository "wms-api/repository/authentication"
	repository "wms-api/repository/master"
	service "wms-api/services/master"
)

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	warehouseCode := flag.String("warehouse", "STUDY_WH", "existing warehouse code")
	zoneCode := flag.String("zone", "STUDY_INBOUND", "existing zone code in that warehouse")
	locationCode := flag.String("location", "STUDY_DOCK_01", "shared dock location code")
	flag.Parse()
	hostIP := net.ParseIP(cfg.Database.Host)
	if cfg.App.Environment == "production" || (cfg.Database.Host != "localhost" && (hostIP == nil || !hostIP.IsLoopback())) {
		return fmt.Errorf("dock sample seed requires a non-production loopback database")
	}
	cfg.Database.AutoCreate = false
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
	actor, err := authrepository.NewAppAccountRepository(db).FindByUsername(ctx, cfg.Auth.BootstrapAdminUsername)
	if err != nil {
		return fmt.Errorf("find bootstrap seed actor: %w", err)
	}
	var dockType model.LocationType
	var location model.WarehouseLocation
	var created int64
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := repository.NewLocationTypeRepository(tx).Seed(ctx, []model.LocationType{service.DefaultDockLocationType()}); err != nil {
			return err
		}
		if err := tx.Where("code = ?", "DOCK").Take(&dockType).Error; err != nil {
			return err
		}
		if !dockType.IsActive || !dockType.AllowsReceiving || !dockType.AllowsShipping || dockType.AllowsStorage || dockType.AllowsPicking {
			return fmt.Errorf("existing DOCK capabilities differ; seed will not overwrite them")
		}
		var warehouse model.Warehouse
		if err := tx.Where("code = ? AND is_active = true", *warehouseCode).Take(&warehouse).Error; err != nil {
			return fmt.Errorf("find active warehouse %q: %w", *warehouseCode, err)
		}
		var zone model.WarehouseZone
		if err := tx.Where("warehouse_id = ? AND code = ? AND is_active = true", warehouse.ID, *zoneCode).Take(&zone).Error; err != nil {
			return fmt.Errorf("find active zone %q: %w", *zoneCode, err)
		}
		location = model.WarehouseLocation{WarehouseID: warehouse.ID, ZoneID: zone.ID, LocationTypeID: dockType.ID, Code: *locationCode, IsActive: true, CreatedBy: &actor.ID}
		result := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "warehouse_id"}, {Name: "code"}}, DoNothing: true}).Create(&location)
		if result.Error != nil {
			return result.Error
		}
		created = result.RowsAffected
		if err := tx.Where("warehouse_id = ? AND code = ?", warehouse.ID, *locationCode).Take(&location).Error; err != nil {
			return err
		}
		if location.LocationTypeID != dockType.ID || location.ZoneID != zone.ID || !location.IsActive || location.IsLocked {
			return fmt.Errorf("existing dock location differs or is unavailable; seed will not overwrite it")
		}
		return nil
	})
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
		"location_type": dockType, "location": location, "locations_created": created,
	})
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
