package inbound

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"wms-api/config"
)

// A normalized, isolated fixture tests filtering before pagination. No business
// records are changed, and the temporary schema is rolled back with the test.
func TestInspectionEligibleReceiptsPostgreSQL(t *testing.T) {
	if os.Getenv("WMS_INTEGRATION_TEST") != "1" {
		t.Skip("set WMS_INTEGRATION_TEST=1")
	}
	check := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	_ = godotenv.Load("../../.env")
	cfg, err := config.Load()
	check(err)
	cfg.Database.AutoCreate = false
	db, err := config.OpenDatabase(cfg.Database)
	check(err)
	sqlDB, err := db.DB()
	check(err)
	defer sqlDB.Close()
	tx := db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Begin()
	check(tx.Error)
	defer tx.Rollback()
	schema := fmt.Sprintf("receipt_lookup_test_%d", time.Now().UnixNano())
	check(tx.Exec("CREATE SCHEMA " + schema).Error)
	check(tx.Exec("SET LOCAL search_path TO " + schema).Error)
	for _, statement := range []string{
		"CREATE TABLE organization (organization_id text PRIMARY KEY, code text)",
		"CREATE TABLE warehouse (warehouse_id text PRIMARY KEY, code text)",
		"CREATE TABLE document_status (status_id text PRIMARY KEY, code text)",
		"CREATE TABLE inventory_status (inventory_status_id text PRIMARY KEY, code text)",
		"CREATE TABLE receipt (receipt_id text PRIMARY KEY, owner_id text, warehouse_id text, status_id text, business_date date, supersedes_receipt_id text, delivery_note_no text, vehicle_number text)",
		"CREATE TABLE receipt_line (receipt_line_id text PRIMARY KEY, receipt_id text)",
		"CREATE TABLE receipt_inventory (receipt_inventory_id text PRIMARY KEY, receipt_line_id text, initial_balance_id text, base_qty numeric, handling_unit_id text)",
		"CREATE TABLE inventory_balance (balance_id text PRIMARY KEY, inventory_status_id text, on_hand_qty numeric, reserved_qty numeric)",
		"CREATE TABLE receipt_line_serial (receipt_inventory_id text, serial_id text)",
		"CREATE TABLE quality_inspection (receipt_inventory_id text, quality_status_id text)",
		"INSERT INTO organization VALUES ('owner','OWNER'),('other','OTHER')",
		"INSERT INTO warehouse VALUES ('wh','WH'),('other-wh','OTHER')",
		"INSERT INTO document_status VALUES ('completed','COMPLETED'),('open','OPEN'),('reversed','REVERSED')",
		"INSERT INTO inventory_status VALUES ('qc','QC_PENDING'),('available','AVAILABLE')",
	} {
		check(tx.Exec(statement).Error)
	}
	addReceipt := func(id, owner, warehouse, status string) {
		check(tx.Exec("INSERT INTO receipt VALUES (?,?,?,?,'2026-09-17',NULL,'delivery note',NULL)", id, owner, warehouse, status).Error)
		check(tx.Exec("INSERT INTO receipt_line VALUES (?,?)", id+"-line", id).Error)
	}
	addBatch := func(receipt, id, status string, onHand, reserved int, hu *string) {
		check(tx.Exec("INSERT INTO inventory_balance VALUES (?,?,?,?)", id+"-balance", status, onHand, reserved).Error)
		check(tx.Exec("INSERT INTO receipt_inventory VALUES (?,?,?,10,?)", id, receipt+"-line", id+"-balance", hu).Error)
	}
	// Newer ineligible receipts must not consume page slots.
	for _, id := range []string{"R10", "R20", "R90", "R91", "R92", "R93", "R94", "R95", "R96", "R97", "R98"} {
		addReceipt(id, "owner", "wh", "completed")
	}
	addBatch("R10", "a", "qc", 10, 0, nil)
	addBatch("R10", "b", "qc", 10, 0, nil)
	check(tx.Exec("INSERT INTO quality_inspection VALUES ('a','PENDING')").Error)
	addBatch("R20", "c", "qc", 10, 0, nil)
	for i, status := range []string{"PENDING", "COMPLETED", "CANCELLED"} {
		id := fmt.Sprintf("R%d", 90+i)
		addBatch(id, id, "qc", 10, 0, nil)
		check(tx.Exec("INSERT INTO quality_inspection VALUES (?,?)", id, status).Error)
	}
	addBatch("R93", "non-qc", "available", 10, 0, nil)
	addBatch("R94", "short", "qc", 9, 0, nil)
	addBatch("R95", "reserved", "qc", 10, 1, nil)
	hu := "HU-1"
	addBatch("R96", "hu", "qc", 20, 0, &hu)
	addBatch("R97", "serial", "qc", 20, 0, nil)
	check(tx.Exec("INSERT INTO receipt_line_serial VALUES ('serial','SER-1')").Error)
	// R98 has no posted batches.
	for _, f := range []struct{ id, owner, warehouse, status string }{
		{"R99", "other", "wh", "completed"},
		{"R99-WH", "owner", "other-wh", "completed"},
		{"R99-OPEN", "owner", "wh", "open"},
		{"R99-REVERSED", "owner", "wh", "reversed"},
	} {
		addReceipt(f.id, f.owner, f.warehouse, f.status)
		addBatch(f.id, f.id, "qc", 10, 0, nil)
	}
	r := NewReceiptRepository(tx)
	ctx := context.Background()
	filter := ListFilter{OwnerID: "owner", WarehouseID: "wh", Page: 1, PageSize: 1, InspectionEligible: true}
	assertPage := func(want string, total int64) {
		t.Helper()
		rows, count, err := r.List(ctx, filter)
		check(err)
		if count != total || len(rows) != 1 || rows[0].ID != want {
			t.Fatalf("rows=%+v total=%d, want %s total=%d", rows, count, want, total)
		}
	}
	assertPage("R20", 2)
	filter.Page = 2
	assertPage("R10", 2) // Partly inspected receipts remain eligible.
	filter.Page = 1
	filter.Search = "R10"
	assertPage("R10", 1)
	filter.Search = ""
	check(tx.Exec("INSERT INTO quality_inspection VALUES ('b','PENDING')").Error)
	assertPage("R20", 1) // Last batch opened: receipt disappears immediately.
	filter.InspectionEligible = false
	rows, count, err := r.List(ctx, filter)
	check(err)
	if count != 13 || len(rows) != 1 || rows[0].ID != "R99-REVERSED" {
		t.Fatalf("normal receipt history changed: rows=%+v total=%d", rows, count)
	}
}
