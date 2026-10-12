package outbound

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"wms-api/config"
	repository "wms-api/repository/outbound"
)

func TestOutboundSnapshotMigrationPostgreSQL(t *testing.T) {
	if os.Getenv("WMS_INTEGRATION_TEST") != "1" {
		t.Skip("set WMS_INTEGRATION_TEST=1")
	}
	_ = godotenv.Load("../../.env")
	cfg, err := config.Load()
	outboundOK(t, err)
	cfg.Database.AutoCreate = false
	db, err := config.OpenDatabase(cfg.Database)
	outboundOK(t, err)
	sqlDB, err := db.DB()
	outboundOK(t, err)
	defer sqlDB.Close()
	tx := db.Begin()
	outboundOK(t, tx.Error)
	defer tx.Rollback()
	schema := fmt.Sprintf("outbound_snapshot_test_%d", time.Now().UnixNano())
	outboundOK(t, tx.Exec("CREATE SCHEMA "+schema).Error)
	outboundOK(t, tx.Exec("SET LOCAL search_path TO "+schema).Error)
	for _, statement := range []string{
		`CREATE TABLE uom(uom_id uuid PRIMARY KEY)`,
		`INSERT INTO uom VALUES ('11111111-1111-4111-8111-111111111111'), ('22222222-2222-4222-8222-222222222222')`,
		`CREATE TABLE outbound_validation_rule(handler_code text, name text, description text)`,
		`CREATE TABLE outbound_order_line(outbound_line_id text PRIMARY KEY, line_no int, uom_id uuid, ordered_qty numeric(20,6), allocated_qty numeric(20,6), picked_qty numeric(20,6), checked_qty numeric(20,6) DEFAULT 0, packed_qty numeric(20,6) DEFAULT 0, shipped_qty numeric(20,6) DEFAULT 0, delivered_qty numeric(20,6) DEFAULT 0, rejected_qty numeric(20,6) DEFAULT 0, short_accepted_qty numeric(20,6) DEFAULT 0, CONSTRAINT ck_outbound_line_qty CHECK(allocated_qty<=ordered_qty), CONSTRAINT ck_outbound_ordered_qty CHECK(allocated_qty<=ordered_qty))`,
		`INSERT INTO outbound_order_line(outbound_line_id,line_no,uom_id,ordered_qty,allocated_qty,picked_qty) VALUES ('legacy',1,'11111111-1111-4111-8111-111111111111',6,6,0)`,
	} {
		outboundOK(t, tx.Exec(statement).Error)
	}
	outboundOK(t, repository.MigrateQuantitySnapshots(tx))
	var snapshot struct{ UOMConversionToBase, OrderedBaseQty, BaseUOMID string }
	outboundOK(t, tx.Table("outbound_order_line").Where("outbound_line_id='legacy'").Take(&snapshot).Error)
	if snapshot.UOMConversionToBase != "1.000000" || snapshot.OrderedBaseQty != "6.000000" || snapshot.BaseUOMID != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("legacy history reinterpreted: %+v", snapshot)
	}
	// A carton order can allocate its full base demand, which exceeds the
	// document quantity. Re-running the upgrade must retain its snapshot.
	outboundOK(t, tx.Exec(`UPDATE outbound_order_line SET uom_id='22222222-2222-4222-8222-222222222222', ordered_qty=2, uom_conversion_to_base=3, ordered_base_qty=6 WHERE outbound_line_id='legacy'`).Error)
	outboundOK(t, repository.MigrateQuantitySnapshots(tx))
	outboundOK(t, tx.Table("outbound_order_line").Where("outbound_line_id='legacy'").Take(&snapshot).Error)
	if snapshot.UOMConversionToBase != "3.000000" || snapshot.OrderedBaseQty != "6.000000" {
		t.Fatalf("new snapshot overwritten: %+v", snapshot)
	}
}
