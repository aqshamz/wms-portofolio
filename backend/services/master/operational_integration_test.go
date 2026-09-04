package master

import (
	"context"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"strings"
	"testing"
	"time"
	"wms-api/config"
	dto "wms-api/dto/master"
	authmodel "wms-api/models/authentication"
	model "wms-api/models/master"
	authrepo "wms-api/repository/authentication"
	repository "wms-api/repository/master"
)

func operationalDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	if os.Getenv("WMS_INTEGRATION_TEST") != "1" {
		t.Skip("set WMS_INTEGRATION_TEST=1")
	}
	_ = godotenv.Load("../../.env")
	cfg, err := config.Load()
	catalogOK(t, err)
	cfg.Database.AutoCreate = false
	db, err := config.OpenDatabase(cfg.Database)
	catalogOK(t, err)
	pool, err := db.DB()
	catalogOK(t, err)
	t.Cleanup(func() { pool.Close() })
	return db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
}
func migrateOperationalTest(t *testing.T, db *gorm.DB) {
	t.Helper()
	catalogOK(t, authrepo.Migrate(db))
	catalogOK(t, repository.Migrate(db))
	catalogOK(t, repository.MigrateCatalog(db))
	catalogOK(t, repository.MigrateOperational(db))
}
func operationalActor(t *testing.T, db *gorm.DB) string {
	t.Helper()
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	status := authmodel.AccountStatus{Code: "OPS_" + suffix, Name: "Test", AllowsLogin: true, IsActive: true}
	catalogOK(t, db.Create(&status).Error)
	account := authmodel.AppAccount{Username: "ops_" + suffix, DisplayName: "Test", AccountStatusID: status.ID, ExternalSubject: catalogPointer("ops-" + suffix)}
	catalogOK(t, db.Create(&account).Error)
	return account.ID
}
func TestOperationalPostgreSQL(t *testing.T) {
	for _, fresh := range []bool{false, true} {
		t.Run(fmt.Sprintf("fresh=%v", fresh), func(t *testing.T) {
			db := operationalDatabase(t)
			tx := db.Begin()
			catalogOK(t, tx.Error)
			defer tx.Rollback()
			if fresh {
				schema := fmt.Sprintf("ops_test_%d", time.Now().UnixNano())
				catalogOK(t, tx.Exec("CREATE SCHEMA "+schema).Error)
				catalogOK(t, tx.Exec("SET LOCAL search_path TO "+schema).Error)
			}
			migrateOperationalTest(t, tx)
			catalogOK(t, repository.MigrateOperational(tx))
			ctx := context.Background()
			repos := repository.NewOperationalRepositories(tx)
			s, err := NewOperationalService(repos, "Asia/Jakarta")
			catalogOK(t, err)
			catalogOK(t, NewCatalogService(repos.Catalog).SeedCatalog(ctx))
			catalogOK(t, s.SeedOperational(ctx))
			catalogOK(t, s.SeedOperational(ctx))
			actor := operationalActor(t, tx)
			suffix := fmt.Sprintf("%d", time.Now().UnixNano())
			yes, no := true, false
			list := dto.OperationalListRequest{Page: 1, PageSize: 100}
			docs, err := s.ListDocumentType(ctx, list)
			catalogOK(t, err)
			if docs.TotalItems < 37 {
				t.Fatal("missing document type seeds")
			}
			po, err := repos.DocumentType.ByCode(ctx, "PURCHASE_ORDER")
			catalogOK(t, err)
			formats, err := s.ListDocumentNumberRule(ctx, po.ID, list)
			catalogOK(t, err)
			if len(formats.Items) != 1 || formats.Items[0].Prefix != "PO" {
				t.Fatal("missing PO number rule")
			}
			_, err = s.DeactivateDocumentNumberRule(ctx, po.ID, formats.Items[0].ID)
			catalogOK(t, err)
			catalogOK(t, s.SeedOperational(ctx))
			formats, err = s.ListDocumentNumberRule(ctx, po.ID, list)
			catalogOK(t, err)
			if len(formats.Items) != 1 || formats.Items[0].IsActive {
				t.Fatal("seed re-enabled disabled numbering rule")
			}

			kind, err := s.CreateDocumentType(ctx, dto.CreateDocumentTypeRequest{Code: "OPS_" + suffix, Name: "Operational test", ModuleCode: "MASTER"})
			catalogOK(t, err)
			other, err := s.CreateDocumentType(ctx, dto.CreateDocumentTypeRequest{Code: "OTHER_" + suffix, Name: "Other", ModuleCode: "MASTER"})
			catalogOK(t, err)
			first, err := s.CreateDocumentStatus(ctx, kind.ID, dto.CreateDocumentStatusRequest{Code: "DRAFT", Name: "Draft", IsInitial: true, DisplayOrder: 10})
			catalogOK(t, err)
			ready, err := s.CreateDocumentStatus(ctx, kind.ID, dto.CreateDocumentStatusRequest{Code: "READY", Name: "Ready", IsInitial: true, DisplayOrder: 20})
			catalogOK(t, err)
			first, err = s.GetDocumentStatus(ctx, kind.ID, first.ID)
			catalogOK(t, err)
			if first.IsInitial {
				t.Fatal("old initial status not cleared")
			}
			_, err = s.CreateDocumentStatus(ctx, kind.ID, dto.CreateDocumentStatusRequest{Code: "READY", Name: "Duplicate", IsInitial: true})
			catalogWantError(t, err, ErrConflict)
			ready, err = s.GetDocumentStatus(ctx, kind.ID, ready.ID)
			catalogOK(t, err)
			if !ready.IsInitial {
				t.Fatal("failed status insert cleared existing initial")
			}
			done, err := s.CreateDocumentStatus(ctx, kind.ID, dto.CreateDocumentStatusRequest{Code: "DONE", Name: "Done", IsFinal: true})
			catalogOK(t, err)
			foreign, err := s.CreateDocumentStatus(ctx, other.ID, dto.CreateDocumentStatusRequest{Code: "OPEN", Name: "Open"})
			catalogOK(t, err)
			_, err = s.CreateDocumentStatusTransition(ctx, kind.ID, dto.CreateDocumentStatusTransitionRequest{FromStatusID: first.ID, ToStatusID: foreign.ID})
			catalogWantError(t, err, ErrInvalidInput)
			_, err = s.CreateDocumentStatusTransition(ctx, kind.ID, dto.CreateDocumentStatusTransitionRequest{FromStatusID: first.ID, ToStatusID: first.ID})
			catalogWantError(t, err, ErrInvalidInput)
			_, err = s.CreateDocumentStatusTransition(ctx, kind.ID, dto.CreateDocumentStatusTransitionRequest{FromStatusID: done.ID, ToStatusID: ready.ID})
			catalogWantError(t, err, ErrInvalidInput)
			transition, err := s.CreateDocumentStatusTransition(ctx, kind.ID, dto.CreateDocumentStatusTransitionRequest{FromStatusID: first.ID, ToStatusID: ready.ID})
			catalogOK(t, err)
			_, err = s.UpdateDocumentStatus(ctx, kind.ID, first.ID, dto.UpdateDocumentStatusRequest{Name: "Draft", IsFinal: true, IsActive: &yes})
			catalogWantError(t, err, ErrInvalidInput)
			_, err = s.GetDocumentStatusTransition(ctx, other.ID, transition.ID)
			catalogWantError(t, err, ErrNotFound)
			permission := model.AppPermission{Code: "OPS." + suffix, Name: "Workflow permission", ModuleCode: "MASTER", IsActive: true}
			catalogOK(t, tx.Create(&permission).Error)
			transition, err = s.UpdateDocumentStatusTransition(ctx, kind.ID, transition.ID, dto.UpdateDocumentStatusTransitionRequest{RequiredPermissionID: &permission.ID, IsActive: &yes})
			catalogOK(t, err)
			if transition.RequiredPermissionID == nil || *transition.RequiredPermissionID != permission.ID {
				t.Fatal("permission link not saved")
			}
			_, err = s.DeactivateDocumentStatusTransition(ctx, kind.ID, transition.ID)
			catalogOK(t, err)

			taskInitial, err := s.CreateTaskStatus(ctx, dto.CreateTaskStatusRequest{Code: "START_" + suffix, Name: "Custom start", IsInitial: true})
			catalogOK(t, err)
			taskFinal, err := s.CreateTaskStatus(ctx, dto.CreateTaskStatusRequest{Code: "END_" + suffix, Name: "Custom end", IsFinal: true})
			catalogOK(t, err)
			_, err = s.CreateTaskStatusTransition(ctx, dto.CreateTaskStatusTransitionRequest{FromStatusID: taskInitial.ID, ToStatusID: taskFinal.ID})
			catalogOK(t, err)
			_, err = s.CreateTaskStatusTransition(ctx, dto.CreateTaskStatusTransitionRequest{FromStatusID: taskFinal.ID, ToStatusID: taskInitial.ID})
			catalogWantError(t, err, ErrInvalidInput)
			catalogOK(t, s.SeedOperational(ctx))
			states, err := s.ListTaskStatus(ctx, list)
			catalogOK(t, err)
			initialCount := 0
			for _, state := range states.Items {
				if state.IsInitial {
					initialCount++
					if state.ID != taskInitial.ID {
						t.Fatal("seed changed initial task status")
					}
				}
			}
			if initialCount != 1 {
				t.Fatal("expected one initial task status")
			}
			_, err = s.CreateTaskPriority(ctx, dto.CreateTaskPriorityRequest{Code: "P_" + suffix, Name: "Duplicate priority", PriorityValue: 200})
			catalogWantError(t, err, ErrConflict)

			owner := model.Organization{Code: "OPS_" + suffix, Name: "Owner", TimezoneName: "Asia/Jakarta", IsActive: true}
			catalogOK(t, repos.Organization.Create(ctx, &owner))
			otherOwner := model.Organization{Code: "OPS2_" + suffix, Name: "Other", TimezoneName: "Asia/Jakarta", IsActive: true}
			catalogOK(t, repos.Organization.Create(ctx, &otherOwner))
			warehouse := model.Warehouse{OperatorID: owner.ID, Code: "WH_" + suffix, Name: "Warehouse", TimezoneName: "Asia/Jakarta", IsActive: true}
			catalogOK(t, repos.Warehouse.Create(ctx, &warehouse))
			otherWarehouse := model.Warehouse{OperatorID: owner.ID, Code: "WH2_" + suffix, Name: "Other warehouse", TimezoneName: "Asia/Jakarta", IsActive: true}
			catalogOK(t, repos.Warehouse.Create(ctx, &otherWarehouse))
			catalogOK(t, repos.WarehouseOwner.Assign(ctx, &model.WarehouseOwner{WarehouseID: warehouse.ID, OwnerID: owner.ID, IsActive: true}))
			zone := model.WarehouseZone{WarehouseID: warehouse.ID, Code: "Z", Name: "Zone", IsActive: true}
			catalogOK(t, repos.Zone.Create(ctx, &zone))
			wrongZone := model.WarehouseZone{WarehouseID: otherWarehouse.ID, Code: "Z", Name: "Other", IsActive: true}
			catalogOK(t, repos.Zone.Create(ctx, &wrongZone))
			category := model.ItemCategory{OwnerID: owner.ID, Code: "CAT", Name: "Category", IsActive: true}
			catalogOK(t, repos.Catalog.ItemCategory.Create(ctx, &category))
			wrongCategory := model.ItemCategory{OwnerID: otherOwner.ID, Code: "CAT", Name: "Category", IsActive: true}
			catalogOK(t, repos.Catalog.ItemCategory.Create(ctx, &wrongCategory))
			pick, err := s.CreatePickingStrategy(ctx, dto.CreatePickingStrategyRequest{OwnerID: &owner.ID, WarehouseID: &warehouse.ID, Code: "PICK", Name: "Picking"})
			catalogOK(t, err)
			_, err = s.CreatePickingStrategy(ctx, dto.CreatePickingStrategyRequest{OwnerID: &otherOwner.ID, WarehouseID: &warehouse.ID, Code: "BAD", Name: "Bad"})
			catalogWantError(t, err, ErrInvalidInput)
			global, err := s.CreatePickingStrategy(ctx, dto.CreatePickingStrategyRequest{Code: "GLOBAL_" + suffix, Name: "Global"})
			catalogOK(t, err)
			_, err = s.CreatePickingStrategy(ctx, dto.CreatePickingStrategyRequest{Code: global.Code, Name: "Duplicate"})
			catalogWantError(t, err, ErrConflict)
			method, err := repos.PickingSortMethod.ByCode(ctx, "FEFO")
			catalogOK(t, err)
			_, err = s.CreatePickingStrategyRule(ctx, pick.ID, dto.CreatePickingStrategyRuleRequest{SequenceNo: 10, ZoneID: &wrongZone.ID, PickingSortMethodID: method.ID})
			catalogWantError(t, err, ErrInvalidInput)
			_, err = s.CreatePickingStrategyRule(ctx, global.ID, dto.CreatePickingStrategyRuleRequest{SequenceNo: 10, ZoneID: &zone.ID, PickingSortMethodID: method.ID})
			catalogWantError(t, err, ErrInvalidInput)
			rule, err := s.CreatePickingStrategyRule(ctx, pick.ID, dto.CreatePickingStrategyRuleRequest{SequenceNo: 20, ZoneID: &zone.ID, PickingSortMethodID: method.ID})
			catalogOK(t, err)
			_, err = s.CreatePickingStrategyRule(ctx, pick.ID, dto.CreatePickingStrategyRuleRequest{SequenceNo: 10, PickingSortMethodID: method.ID})
			catalogOK(t, err)
			rules, err := s.ListPickingStrategyRule(ctx, pick.ID, list)
			catalogOK(t, err)
			if len(rules.Items) != 2 || rules.Items[0].SequenceNo != 10 {
				t.Fatal("picking rules not ordered")
			}
			_, err = s.GetPickingStrategyRule(ctx, global.ID, rule.ID)
			catalogWantError(t, err, ErrNotFound)
			put, err := s.CreatePutawayStrategy(ctx, dto.CreatePutawayStrategyRequest{OwnerID: &owner.ID, WarehouseID: &warehouse.ID, Code: "PUT", Name: "Putaway"})
			catalogOK(t, err)
			_, err = s.CreatePutawayStrategyRule(ctx, put.ID, dto.CreatePutawayStrategyRuleRequest{SequenceNo: 10, CategoryID: &wrongCategory.ID})
			catalogWantError(t, err, ErrInvalidInput)
			_, err = s.CreatePutawayStrategyRule(ctx, put.ID, dto.CreatePutawayStrategyRuleRequest{SequenceNo: 10, MinimumEmptyPercent: catalogPointer("100.0001")})
			catalogWantError(t, err, ErrInvalidInput)
			putRule, err := s.CreatePutawayStrategyRule(ctx, put.ID, dto.CreatePutawayStrategyRuleRequest{SequenceNo: 10, CategoryID: &category.ID, ZoneID: &zone.ID, MinimumEmptyPercent: catalogPointer("25.1234")})
			catalogOK(t, err)
			if putRule.MinimumEmptyPercent == nil || *putRule.MinimumEmptyPercent != "25.1234" {
				t.Fatal("putaway percentage lost precision")
			}
			_, err = s.DeactivatePutawayStrategy(ctx, put.ID)
			catalogOK(t, err)
			_, err = s.CreatePutawayStrategyRule(ctx, put.ID, dto.CreatePutawayStrategyRuleRequest{SequenceNo: 20})
			catalogWantError(t, err, ErrInvalidInput)
			_, err = s.DeactivatePutawayStrategyRule(ctx, put.ID, putRule.ID)
			catalogOK(t, err)

			today := time.Now().In(s.location).Format("2006-01-02")
			numbering := dto.ReplaceDocumentNumberRuleRequest{Prefix: "TST", Separator: catalogPointer(""), SequenceLength: 3, IncludePartnerCode: &no, IncludeWarehouseCode: &no, EffectiveFrom: today}
			numberRule, err := s.ReplaceDocumentNumberRule(ctx, kind.ID, numbering, actor)
			catalogOK(t, err)
			if numberRule.IncludePartnerCode || numberRule.IncludeWarehouseCode || numberRule.Separator != "" {
				t.Fatal("explicit false/empty format lost")
			}
			generated, err := s.GenerateDocumentID(ctx, kind.ID, dto.GenerateDocumentIDRequest{BusinessDate: today})
			catalogOK(t, err)
			if generated.SequenceNumber != "1" || !strings.HasSuffix(generated.DocumentID, "001") {
				t.Fatal("incorrect first document ID")
			}
			numbering.Prefix = "NEW"
			replacement, err := s.ReplaceDocumentNumberRule(ctx, kind.ID, numbering, actor)
			catalogOK(t, err)
			history, err := s.ListDocumentNumberRule(ctx, kind.ID, list)
			catalogOK(t, err)
			if len(history.Items) != 2 || replacement.ID == numberRule.ID {
				t.Fatal("same-day rule history lost")
			}
			generated, err = s.GenerateDocumentID(ctx, kind.ID, dto.GenerateDocumentIDRequest{BusinessDate: today})
			catalogOK(t, err)
			if generated.SequenceNumber != "2" || !strings.HasPrefix(generated.DocumentID, "NEW") {
				t.Fatal("rule replacement reset counter")
			}
			rollback := errors.New("rollback allocation")
			err = s.transaction(ctx, func(local *OperationalService) error {
				_, err := local.GenerateDocumentID(ctx, kind.ID, dto.GenerateDocumentIDRequest{BusinessDate: today})
				if err != nil {
					return err
				}
				return rollback
			})
			catalogWantError(t, err, rollback)
			counters, err := s.ListDocumentDailyCounter(ctx, kind.ID, today, today, list)
			catalogOK(t, err)
			if counters.TotalItems != 1 || counters.Items[0].LastNumber != "2" {
				t.Fatal("failed transaction consumed counter")
			}
			_, err = s.GenerateDocumentID(ctx, kind.ID, dto.GenerateDocumentIDRequest{BusinessDate: "invalid"})
			catalogWantError(t, err, ErrInvalidInput)
			numbering.IncludePartnerCode = &yes
			_, err = s.ReplaceDocumentNumberRule(ctx, kind.ID, numbering, actor)
			catalogOK(t, err)
			_, err = s.GenerateDocumentID(ctx, kind.ID, dto.GenerateDocumentIDRequest{BusinessDate: today})
			catalogWantError(t, err, ErrInvalidInput)
			counters, err = s.ListDocumentDailyCounter(ctx, kind.ID, today, today, list)
			catalogOK(t, err)
			if counters.Items[0].LastNumber != "2" {
				t.Fatal("invalid allocation consumed counter")
			}
			partner, err := NewCatalogService(repos.Catalog).CreateBusinessPartner(ctx, dto.CreateBusinessPartnerRequest{OwnerID: owner.ID, Code: "VENDOR", Name: "Vendor"}, actor)
			catalogOK(t, err)
			numbering.IncludeWarehouseCode = &yes
			numbering.Separator = catalogPointer("-")
			activeRule, err := s.ReplaceDocumentNumberRule(ctx, kind.ID, numbering, actor)
			catalogOK(t, err)
			generated, err = s.GenerateDocumentID(ctx, kind.ID, dto.GenerateDocumentIDRequest{BusinessDate: today, PartnerID: &partner.ID, WarehouseID: &warehouse.ID})
			catalogOK(t, err)
			if generated.SequenceNumber != "3" || !strings.Contains(generated.DocumentID, "-VENDOR-"+warehouse.Code+"-") {
				t.Fatal("partner/warehouse format incorrect")
			}
			catalogOK(t, repos.WarehouseOwner.Assign(ctx, &model.WarehouseOwner{WarehouseID: otherWarehouse.ID, OwnerID: owner.ID, IsActive: true}))
			generated, err = s.GenerateDocumentID(ctx, kind.ID, dto.GenerateDocumentIDRequest{BusinessDate: today, PartnerID: &partner.ID, WarehouseID: &otherWarehouse.ID})
			catalogOK(t, err)
			if generated.SequenceNumber != "4" {
				t.Fatal("warehouse incorrectly partitioned daily counter")
			}
			_, err = s.ReplaceDocumentNumberRule(ctx, kind.ID, numbering, "00000000-0000-4000-8000-000000000099")
			catalogWantError(t, err, ErrInvalidInput)
			keptRule, err := repos.DocumentNumberRule.Get(ctx, activeRule.ID)
			catalogOK(t, err)
			if !keptRule.IsActive {
				t.Fatal("failed replacement retired active rule")
			}
		})
	}
}
