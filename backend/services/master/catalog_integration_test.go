package master

import (
	"context"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"testing"
	"time"
	"wms-api/config"
	dto "wms-api/dto/master"
	authmodel "wms-api/models/authentication"
	model "wms-api/models/master"
	authrepo "wms-api/repository/authentication"
	repository "wms-api/repository/master"
)

func catalogMust[T any](t *testing.T, value T, err error) T {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	return value
}
func catalogOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func catalogWantError(t *testing.T, err, want error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want %v", err, want)
	}
}
func catalogPointer[T any](value T) *T { return &value }

// Opt-in: the entire migration and fixture suite is rolled back, even on failure.
// No persistent test partners/items/accounts are left in the development DB.
func TestCatalogPostgreSQL(t *testing.T) {
	if os.Getenv("WMS_INTEGRATION_TEST") != "1" {
		t.Skip("set WMS_INTEGRATION_TEST=1 to run against PostgreSQL")
	}
	t.Run("existing_schema", func(t *testing.T) { testCatalogPostgreSQL(t, false) })
	t.Run("fresh_schema", func(t *testing.T) { testCatalogPostgreSQL(t, true) })
}
func testCatalogPostgreSQL(t *testing.T, fresh bool) {
	_ = godotenv.Load("../../.env")
	cfg, err := config.Load()
	catalogOK(t, err)
	cfg.Database.AutoCreate = false
	db, err := config.OpenDatabase(cfg.Database)
	catalogOK(t, err)
	sqlDB, err := db.DB()
	catalogOK(t, err)
	defer sqlDB.Close()
	tx := db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Begin()
	catalogOK(t, tx.Error)
	defer tx.Rollback()
	if fresh {
		schema := fmt.Sprintf("catalog_test_%d", time.Now().UnixNano())
		catalogOK(t, tx.Exec("CREATE SCHEMA "+schema).Error)
		catalogOK(t, tx.Exec("SET LOCAL search_path TO "+schema).Error)
	}
	catalogOK(t, authrepo.Migrate(tx))
	catalogOK(t, repository.Migrate(tx))
	catalogOK(t, repository.MigrateCatalog(tx))
	catalogOK(t, repository.MigrateCatalog(tx))
	ctx := context.Background()
	repos := repository.NewCatalogRepositories(tx)
	service := NewCatalogService(repos)
	catalogOK(t, service.SeedCatalog(ctx))
	catalogOK(t, service.SeedCatalog(ctx))
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	status := authmodel.AccountStatus{Code: "TEST_" + suffix, Name: "Test", AllowsLogin: true, IsActive: true}
	catalogOK(t, tx.Create(&status).Error)
	account := authmodel.AppAccount{Username: "catalog_test_" + suffix, DisplayName: "Catalog test", AccountStatusID: status.ID, ExternalSubject: catalogPointer("catalog-test-" + suffix)}
	catalogOK(t, tx.Create(&account).Error)
	owner := model.Organization{Code: "CT_" + suffix, Name: "Test owner", TimezoneName: "Asia/Jakarta", IsActive: true}
	otherOwner := model.Organization{Code: "CT2_" + suffix, Name: "Other test owner", TimezoneName: "Asia/Jakarta", IsActive: true}
	catalogOK(t, repository.NewOrganizationRepository(tx).Create(ctx, &owner))
	catalogOK(t, repository.NewOrganizationRepository(tx).Create(ctx, &otherOwner))
	units, err := service.ListUOM(ctx, repository.CatalogFilter{Page: 1, PageSize: 100})
	catalogOK(t, err)
	var each, box, kilogram string
	for _, unit := range units.Items {
		switch unit.Code {
		case "EA":
			each = unit.ID
		case "BOX":
			box = unit.ID
		case "KG":
			kilogram = unit.ID
		}
	}
	if each == "" || box == "" || kilogram == "" {
		t.Fatal("missing UOM reference seeds")
	}
	yes, no := true, false

	// Reference seeds are repeatable and do not overwrite user edits.
	quality, err := service.CreateQualityStatus(ctx, dto.CreateQualityStatusRequest{Code: "T_" + suffix, Name: "Custom"})
	catalogOK(t, err)
	_, err = service.UpdateQualityStatus(ctx, quality.ID, dto.UpdateQualityStatusRequest{Name: "Edited", IsActive: &no})
	catalogOK(t, err)
	catalogOK(t, service.SeedCatalog(ctx))
	quality, err = service.GetQualityStatus(ctx, quality.ID)
	catalogOK(t, err)
	if quality.Name != "Edited" || quality.IsActive {
		t.Fatal("reference edits were lost")
	}
	seededQuality, err := service.ListQualityStatus(ctx, repository.CatalogFilter{Search: "PASSED", Page: 1, PageSize: 100})
	catalogOK(t, err)
	if len(seededQuality.Items) != 1 {
		t.Fatal("missing PASSED seed")
	}
	_, err = service.UpdateQualityStatus(ctx, seededQuality.Items[0].ID, dto.UpdateQualityStatusRequest{Name: "Custom passed name", IsActive: &no})
	catalogOK(t, err)
	catalogOK(t, service.SeedCatalog(ctx))
	preserved, err := service.GetQualityStatus(ctx, seededQuality.Items[0].ID)
	catalogOK(t, err)
	if preserved.Name != "Custom passed name" || preserved.IsActive {
		t.Fatal("seed overwrote existing reference")
	}
	inv, err := service.ListInventoryStatus(ctx, repository.CatalogFilter{Page: 1, PageSize: 100})
	catalogOK(t, err)
	for _, status := range inv.Items {
		if status.Code == "AVAILABLE" && (!status.IsAllocatable || !status.IsPickable) {
			t.Fatal("AVAILABLE flags incorrect")
		}
	}

	partner, err := service.CreateBusinessPartner(ctx, dto.CreateBusinessPartnerRequest{OwnerID: owner.ID, Code: "SUP_" + suffix, Name: "Supplier"}, account.ID)
	catalogOK(t, err)
	types, err := service.ListPartnerType(ctx, repository.CatalogFilter{Page: 1, PageSize: 100})
	catalogOK(t, err)
	if len(types.Items) < 2 {
		t.Fatal("partner types not seeded")
	}
	for _, kind := range types.Items[:2] {
		_, err = service.AssignBusinessPartnerType(ctx, partner.ID, dto.AssignPartnerTypeRequest{PartnerTypeID: kind.ID})
		catalogOK(t, err)
	}
	_, err = service.AssignBusinessPartnerType(ctx, partner.ID, dto.AssignPartnerTypeRequest{PartnerTypeID: types.Items[0].ID})
	catalogOK(t, err)
	partnerDetail, err := service.GetBusinessPartner(ctx, partner.ID)
	catalogOK(t, err)
	if len(partnerDetail.PartnerTypes) != 2 {
		t.Fatal("partner type assignments are not idempotent")
	}
	catalogOK(t, service.RemoveBusinessPartnerType(ctx, partner.ID, types.Items[0].ID))
	partnerUpdate := dto.UpdateBusinessPartnerRequest{Name: "Supplier updated", IsActive: &yes, ExpectedUpdatedAt: partner.UpdatedAt}
	updatedPartner, err := service.UpdateBusinessPartner(ctx, partner.ID, partnerUpdate, account.ID)
	catalogOK(t, err)
	_, err = service.UpdateBusinessPartner(ctx, partner.ID, partnerUpdate, account.ID)
	catalogWantError(t, err, ErrConcurrentUpdate)
	if updatedPartner.UpdatedBy == nil || *updatedPartner.UpdatedBy != account.ID {
		t.Fatal("audit actor missing")
	}
	_, err = service.CreateBusinessPartner(ctx, dto.CreateBusinessPartnerRequest{OwnerID: owner.ID, Code: partner.Code, Name: "Duplicate"}, account.ID)
	catalogWantError(t, err, ErrConflict)
	page, err := service.ListBusinessPartner(ctx, repository.CatalogFilter{OwnerID: owner.ID, Page: 999, PageSize: 10})
	catalogOK(t, err)
	if page.TotalItems != 1 || len(page.Items) != 0 {
		t.Fatal("out-of-range page lost total count")
	}

	parent, err := service.CreateItemCategory(ctx, dto.CreateItemCategoryRequest{OwnerID: owner.ID, Code: "PARENT", Name: "Parent"})
	catalogOK(t, err)
	child, err := service.CreateItemCategory(ctx, dto.CreateItemCategoryRequest{OwnerID: owner.ID, ParentCategoryID: &parent.ID, Code: "CHILD", Name: "Child"})
	catalogOK(t, err)
	_, err = service.UpdateItemCategory(ctx, parent.ID, dto.UpdateItemCategoryRequest{ParentCategoryID: &child.ID, Name: "Parent", IsActive: &yes})
	catalogWantError(t, err, ErrInvalidInput)
	_, err = service.CreateItemCategory(ctx, dto.CreateItemCategoryRequest{OwnerID: otherOwner.ID, ParentCategoryID: &parent.ID, Code: "CROSS", Name: "Cross owner"})
	catalogWantError(t, err, ErrInvalidInput)
	_, err = service.CreateItem(ctx, dto.CreateItemRequest{OwnerID: otherOwner.ID, CategoryID: &child.ID, Code: "CROSS", Name: "Invalid", BaseUOMID: each}, account.ID)
	catalogWantError(t, err, ErrInvalidInput)

	item, err := service.CreateItem(ctx, dto.CreateItemRequest{OwnerID: owner.ID, CategoryID: &child.ID, Code: "SKU_" + suffix, Name: "Test item", BaseUOMID: each, Weight: catalogPointer("99999999999999.999999")}, account.ID)
	catalogOK(t, err)
	detail, err := service.GetItem(ctx, item.ID)
	catalogOK(t, err)
	if len(detail.UOMs) != 1 || !decimalIsOne(detail.UOMs[0].ConversionToBase) || !detail.UOMs[0].IsActive {
		t.Fatal("missing base conversion")
	}
	if detail.Weight == nil || *detail.Weight != "99999999999999.999999" {
		t.Fatal("decimal precision lost")
	}
	baseID := detail.UOMs[0].ID
	_, err = service.DeactivateItemUOM(ctx, item.ID, baseID)
	catalogWantError(t, err, ErrInvalidInput)
	_, err = service.UpdateItemUOM(ctx, item.ID, baseID, dto.UpdateItemUOMRequest{ConversionToBase: "2", IsActive: &yes})
	catalogWantError(t, err, ErrInvalidInput)
	alt, err := service.CreateItemUOM(ctx, item.ID, dto.CreateItemUOMRequest{UOMID: box, ConversionToBase: "12.500001", IsReceivingUOM: &no, IsPickingUOM: &no})
	catalogOK(t, err)
	if alt.IsReceivingUOM || alt.IsPickingUOM || alt.ConversionToBase != "12.500001" {
		t.Fatal("explicit false or precise conversion lost")
	}
	storedAlt, err := repos.ItemUOM.Get(ctx, alt.ID)
	catalogOK(t, err)
	if storedAlt.IsReceivingUOM || storedAlt.IsPickingUOM {
		t.Fatal("stored false flags changed")
	}

	barcode1, err := service.CreateItemBarcode(ctx, item.ID, dto.CreateItemBarcodeRequest{UOMID: &each, Barcode: "B1_" + suffix, IsPrimary: true})
	catalogOK(t, err)
	barcode2, err := service.CreateItemBarcode(ctx, item.ID, dto.CreateItemBarcodeRequest{UOMID: &box, Barcode: "B2_" + suffix})
	catalogOK(t, err)
	_, err = service.CreateItemBarcode(ctx, item.ID, dto.CreateItemBarcodeRequest{UOMID: &kilogram, Barcode: "INVALID_" + suffix})
	catalogWantError(t, err, ErrInvalidInput)
	_, err = service.DeactivateItemUOM(ctx, item.ID, alt.ID)
	catalogWantError(t, err, ErrInvalidInput)
	// A duplicate barcode insert fails AFTER clearing primary; rollback must restore it.
	_, err = service.CreateItemBarcode(ctx, item.ID, dto.CreateItemBarcodeRequest{Barcode: barcode2.Barcode, IsPrimary: true})
	catalogWantError(t, err, ErrConflict)
	rows, err := service.ListItemBarcode(ctx, item.ID)
	catalogOK(t, err)
	for _, row := range rows {
		if row.ID == barcode1.ID && !row.IsPrimary {
			t.Fatal("failed insert cleared primary")
		}
	}
	_, err = service.SetPrimaryItemBarcode(ctx, item.ID, barcode2.ID)
	catalogOK(t, err)
	rows, err = service.ListItemBarcode(ctx, item.ID)
	catalogOK(t, err)
	primaryCount := 0
	for _, row := range rows {
		if row.IsPrimary && row.IsActive {
			primaryCount++
			if row.ID != barcode2.ID {
				t.Fatal("wrong primary barcode")
			}
		}
	}
	if primaryCount != 1 {
		t.Fatal("expected exactly one active primary")
	}
	another, err := service.CreateItem(ctx, dto.CreateItemRequest{OwnerID: owner.ID, Code: "SKU2_" + suffix, Name: "Second", BaseUOMID: each}, account.ID)
	catalogOK(t, err)
	_, err = service.SetPrimaryItemBarcode(ctx, another.ID, barcode2.ID)
	catalogWantError(t, err, ErrNotFound)
	_, err = service.DeactivateItemBarcode(ctx, item.ID, barcode2.ID)
	catalogOK(t, err)
	_, err = service.DeactivateItemUOM(ctx, item.ID, alt.ID)
	catalogOK(t, err)
	_, err = service.SetPrimaryItemBarcode(ctx, item.ID, barcode2.ID)
	catalogWantError(t, err, ErrInvalidInput)

	update := dto.UpdateItemRequest{Name: "Renamed", CategoryID: &child.ID, IsActive: &yes, ExpectedUpdatedAt: item.UpdatedAt}
	updatedItem, err := service.UpdateItem(ctx, item.ID, update, account.ID)
	catalogOK(t, err)
	_, err = service.UpdateItem(ctx, item.ID, update, account.ID)
	catalogWantError(t, err, ErrConcurrentUpdate)
	_, err = service.DeactivateItem(ctx, item.ID, dto.DeactivateRequest{ExpectedUpdatedAt: updatedItem.UpdatedAt}, account.ID)
	catalogOK(t, err)
	_, err = service.CreateItemBarcode(ctx, item.ID, dto.CreateItemBarcodeRequest{Barcode: "INACTIVE_" + suffix})
	catalogWantError(t, err, ErrInvalidInput)

	// A transaction failure rolls back the item and its automatically-created UOM.
	rollback := errors.New("intentional rollback")
	rolledID := ""
	err = repos.Transaction(ctx, func(nested *repository.CatalogRepositories) error {
		created, err := NewCatalogService(nested).CreateItem(ctx, dto.CreateItemRequest{OwnerID: owner.ID, Code: "ROLLBACK_" + suffix, Name: "Rollback", BaseUOMID: each}, account.ID)
		if err != nil {
			return err
		}
		rolledID = created.ID
		return rollback
	})
	catalogWantError(t, err, rollback)
	_, err = service.GetItem(ctx, rolledID)
	catalogWantError(t, err, ErrNotFound)
	orphanRows, _, err := repos.ItemUOM.List(ctx, repository.CatalogFilter{ItemID: rolledID})
	catalogOK(t, err)
	if len(orphanRows) != 0 {
		t.Fatal("orphan item UOM after rollback")
	}
}
