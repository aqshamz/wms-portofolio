package billing

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
	dto "wms-api/dto/billing"
	authmodel "wms-api/models/authentication"
	inventorymodel "wms-api/models/inventory"
	mastermodel "wms-api/models/master"
	authrepository "wms-api/repository/authentication"
	repository "wms-api/repository/billing"
	inventoryrepository "wms-api/repository/inventory"
	masterrepository "wms-api/repository/master"
	masterservice "wms-api/services/master"
)

func billingOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func billingPointer[T any](v T) *T { return &v }

func TestBillingWorkflowPostgreSQL(t *testing.T) {
	if os.Getenv("WMS_INTEGRATION_TEST") != "1" {
		t.Skip("set WMS_INTEGRATION_TEST=1")
	}
	_ = godotenv.Load("../../.env")
	cfg, err := config.Load()
	billingOK(t, err)
	cfg.Database.AutoCreate = false
	db, err := config.OpenDatabase(cfg.Database)
	billingOK(t, err)
	sqlDB, err := db.DB()
	billingOK(t, err)
	defer sqlDB.Close()
	tx := db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Begin()
	billingOK(t, tx.Error)
	defer tx.Rollback()
	suffix := fmt.Sprint(time.Now().UnixNano())
	schema := "billing_test_" + suffix
	billingOK(t, tx.Exec("CREATE SCHEMA "+schema).Error)
	billingOK(t, tx.Exec("SET LOCAL search_path TO "+schema).Error)
	billingOK(t, authrepository.Migrate(tx))
	billingOK(t, masterrepository.Migrate(tx))
	billingOK(t, masterrepository.MigrateCatalog(tx))
	billingOK(t, masterrepository.MigrateOperational(tx))
	billingOK(t, inventoryrepository.Migrate(tx))
	billingOK(t, repository.Migrate(tx))
	ctx := context.Background()
	billingOK(t, masterservice.NewCatalogService(masterrepository.NewCatalogRepositories(tx)).SeedCatalog(ctx))
	operational, err := masterservice.NewOperationalService(masterrepository.NewOperationalRepositories(tx), "Asia/Jakarta")
	billingOK(t, err)
	billingOK(t, operational.SeedOperational(ctx))
	accountStatus := authmodel.AccountStatus{Code: "BAS_" + suffix, Name: "Active", AllowsLogin: true}
	billingOK(t, tx.Create(&accountStatus).Error)
	account := authmodel.AppAccount{Username: "billing_" + suffix, DisplayName: "Billing test", AccountStatusID: accountStatus.ID, ExternalSubject: billingPointer("billing_" + suffix)}
	billingOK(t, tx.Create(&account).Error)
	owner := mastermodel.Organization{Code: "BIO_" + suffix, Name: "Billing owner", TimezoneName: "Asia/Jakarta"}
	billingOK(t, tx.Create(&owner).Error)
	warehouse := mastermodel.Warehouse{OperatorID: owner.ID, Code: "BIW_" + suffix, Name: "Billing warehouse", TimezoneName: "Asia/Jakarta"}
	billingOK(t, tx.Create(&warehouse).Error)
	billingOK(t, tx.Create(&mastermodel.WarehouseOwner{WarehouseID: warehouse.ID, OwnerID: owner.ID}).Error)
	billingOK(t, tx.Create(&mastermodel.AccountOwnerAccess{AccountID: account.ID, OwnerID: owner.ID}).Error)
	billingOK(t, tx.Create(&mastermodel.AccountWarehouseAccess{AccountID: account.ID, WarehouseID: warehouse.ID}).Error)
	zone := mastermodel.WarehouseZone{WarehouseID: warehouse.ID, Code: "BILL", Name: "Billing zone"}
	billingOK(t, tx.Create(&zone).Error)
	locationType := mastermodel.LocationType{Code: "BIL_" + suffix, Name: "Billing storage", AllowsStorage: true, AllowsReceiving: true}
	billingOK(t, tx.Create(&locationType).Error)
	location := mastermodel.WarehouseLocation{WarehouseID: warehouse.ID, ZoneID: zone.ID, LocationTypeID: locationType.ID, Code: "BILL-01"}
	billingOK(t, tx.Create(&location).Error)
	var each mastermodel.UOM
	billingOK(t, tx.Where("code='EA'").Take(&each).Error)
	var available mastermodel.InventoryStatus
	billingOK(t, tx.Where("code='AVAILABLE'").Take(&available).Error)
	item := mastermodel.Item{OwnerID: owner.ID, Code: "BII_" + suffix, Name: "Billing item", BaseUOMID: each.ID}
	billingOK(t, tx.Create(&item).Error)
	billingOK(t, tx.Create(&mastermodel.ItemUOM{ItemID: item.ID, UOMID: each.ID, ConversionToBase: "1"}).Error)
	balance := inventorymodel.InventoryBalance{ID: "BAL-BILL-" + suffix, OwnerID: owner.ID, WarehouseID: warehouse.ID, LocationID: location.ID, ItemID: item.ID, InventoryStatusID: available.ID, OnHandQty: "2", ReservedQty: "0", UOMID: each.ID, VersionNo: 1}
	billingOK(t, tx.Create(&balance).Error)
	var receive inventorymodel.MovementType
	billingOK(t, tx.Where("code='RECEIVE'").Take(&receive).Error)
	jakarta, _ := time.LoadLocation("Asia/Jakarta")
	today := time.Now().In(jakarta).Format("2006-01-02")
	movement := inventorymodel.InventoryMovement{ID: "MOV-BILL-" + suffix, MovementTypeID: receive.ID, OwnerID: owner.ID, WarehouseID: warehouse.ID, BusinessDate: time.Now(), ItemID: item.ID, ToLocationID: &location.ID, ToStatusID: &available.ID, Quantity: "2", UOMID: each.ID, SourceDocumentID: "RCV-BILL-" + suffix, CreatedBy: account.ID}
	billingOK(t, tx.Create(&movement).Error)
	svc, err := NewService(repository.NewRepositories(tx), "Asia/Jakarta")
	billingOK(t, err)
	allowed, err := svc.CanAccess(ctx, account.ID, owner.ID, warehouse.ID)
	billingOK(t, err)
	if !allowed {
		t.Fatal("configured billing scope was denied")
	}
	contract, err := svc.CreateContract(ctx, dto.CreateContractRequest{OwnerID: owner.ID, WarehouseID: warehouse.ID, BusinessDate: today, CurrencyCode: "IDR", BillingCycle: "MONTHLY", PaymentTermDays: 30, EffectiveFrom: today}, account.ID)
	billingOK(t, err)
	contract, err = svc.TransitionContract(ctx, contract.ID, "ACTIVE", contract.VersionNo, account.ID)
	billingOK(t, err)
	card, err := svc.CreateRateCard(ctx, dto.CreateRateCardRequest{BillingContractID: contract.ID, BusinessDate: today, Name: "Study rates", EffectiveFrom: today, Lines: []dto.CreateRateCardLineRequest{{ServiceCode: "RECEIVING", Description: "Receiving per unit", SourceKind: "MOVEMENT", MovementTypeID: &receive.ID, BillingBasis: "QUANTITY", UnitRate: "10", TaxPercent: "10"}, {ServiceCode: "STORAGE", Description: "Daily storage", SourceKind: "STORAGE", BillingBasis: "QUANTITY", UnitRate: "2"}, {ServiceCode: "ADMIN", Description: "Manual administration", SourceKind: "MANUAL", BillingBasis: "EVENT", UnitRate: "5"}}}, account.ID)
	billingOK(t, err)
	card, err = svc.TransitionRateCard(ctx, card.ID, "APPROVED", card.VersionNo, account.ID)
	billingOK(t, err)
	card, err = svc.TransitionRateCard(ctx, card.ID, "ACTIVE", card.VersionNo, account.ID)
	billingOK(t, err)
	collected, err := svc.CollectEvents(ctx, dto.CollectEventsRequest{OwnerID: owner.ID, WarehouseID: warehouse.ID, DateFrom: today, DateUntil: today}, account.ID)
	billingOK(t, err)
	if collected.Created != 1 {
		t.Fatalf("created %d movement events, want 1", collected.Created)
	}
	again, err := svc.CollectEvents(ctx, dto.CollectEventsRequest{OwnerID: owner.ID, WarehouseID: warehouse.ID, DateFrom: today, DateUntil: today}, account.ID)
	billingOK(t, err)
	if again.Existing != 1 {
		t.Fatalf("existing %d events, want 1", again.Existing)
	}
	storage, err := svc.SnapshotStorage(ctx, dto.StorageSnapshotRequest{OwnerID: owner.ID, WarehouseID: warehouse.ID, BusinessDate: today}, account.ID)
	billingOK(t, err)
	if storage.Created != 1 {
		t.Fatalf("storage events %d, want 1", storage.Created)
	}
	var manualID string
	for _, line := range card.Lines {
		if line.SourceKind == "MANUAL" {
			manualID = line.ID
		}
	}
	if manualID == "" {
		t.Fatal("manual rate line missing")
	}
	_, err = svc.CreateManualEvent(ctx, contract.ID, dto.CreateManualEventRequest{RateCardLineID: manualID, BusinessDate: today, SourceDocumentID: "MANUAL-BILL", Quantity: "1"}, account.ID)
	billingOK(t, err)
	run, err := svc.CreateRun(ctx, dto.CreateBillingRunRequest{BillingContractID: contract.ID, BusinessDate: today, PeriodFrom: today, PeriodUntil: today}, account.ID)
	billingOK(t, err)
	run, err = svc.CalculateRun(ctx, run.ID, run.VersionNo, account.ID)
	billingOK(t, err)
	if run.TotalAmount != "31.000000" {
		t.Fatalf("run total %s, want 31.000000", run.TotalAmount)
	}
	run, err = svc.TransitionRun(ctx, run.ID, "REVIEWED", run.VersionNo, account.ID)
	billingOK(t, err)
	invoice, err := svc.CreateInvoice(ctx, run.ID, dto.CreateInvoiceRequest{ExpectedVersion: run.VersionNo, IssueDate: today}, account.ID)
	billingOK(t, err)
	invoice, err = svc.TransitionInvoice(ctx, invoice.ID, "REVIEWED", invoice.VersionNo, account.ID)
	billingOK(t, err)
	invoice, err = svc.TransitionInvoice(ctx, invoice.ID, "ISSUED", invoice.VersionNo, account.ID)
	billingOK(t, err)
	_, err = svc.RecordPayment(ctx, invoice.ID, dto.PaymentRequest{ExpectedVersion: invoice.VersionNo, BusinessDate: today, Amount: "3", Reference: "PAY-A-" + suffix}, account.ID)
	billingOK(t, err)
	invoice, err = svc.GetInvoice(ctx, invoice.ID)
	billingOK(t, err)
	if invoice.StatusCode != "PARTIALLY_PAID" {
		t.Fatalf("invoice status %s", invoice.StatusCode)
	}
	_, err = svc.RecordPayment(ctx, invoice.ID, dto.PaymentRequest{ExpectedVersion: invoice.VersionNo, BusinessDate: today, Amount: "4", Reference: "PAY-B-" + suffix}, account.ID)
	billingOK(t, err)
	invoice, err = svc.GetInvoice(ctx, invoice.ID)
	billingOK(t, err)
	if invoice.StatusCode != "PARTIALLY_PAID" || invoice.PaidAmount != "7.000000" {
		t.Fatalf("second partial payment status=%s paid=%s", invoice.StatusCode, invoice.PaidAmount)
	}
	_, err = svc.IssueCreditNote(ctx, invoice.ID, dto.CreditNoteRequest{ExpectedVersion: invoice.VersionNo, BusinessDate: today, Amount: "24", Reason: "Study settlement"}, account.ID)
	billingOK(t, err)
	invoice, err = svc.GetInvoice(ctx, invoice.ID)
	billingOK(t, err)
	if invoice.StatusCode != "SETTLED" || invoice.OutstandingAmount != "0.000000" {
		t.Fatalf("invoice settlement status=%s outstanding=%s", invoice.StatusCode, invoice.OutstandingAmount)
	}
}
