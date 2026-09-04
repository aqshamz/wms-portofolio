package master

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"sync"
	"testing"
	"time"
	dto "wms-api/dto/master"
	repository "wms-api/repository/master"
)

// The committed schema is test-owned so independent connections can contend on
// the same rows. Cleanup drops only this uniquely named temporary schema.
func TestOperationalConcurrentNumbering(t *testing.T) {
	db := operationalDatabase(t)
	schema := fmt.Sprintf("ops_concurrency_%d", time.Now().UnixNano())
	catalogOK(t, db.Exec("CREATE SCHEMA "+schema).Error)
	t.Cleanup(func() {
		if err := db.Exec("DROP SCHEMA " + schema + " CASCADE").Error; err != nil {
			t.Errorf("cleanup test schema: %v", err)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	typeID, actor := "", ""
	date := "2026-09-03"
	withSchema := func(work func(*OperationalService, *gorm.DB) error) error {
		return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("SET LOCAL search_path TO " + schema).Error; err != nil {
				return err
			}
			service, err := NewOperationalService(repository.NewOperationalRepositories(tx), "Asia/Jakarta")
			if err != nil {
				return err
			}
			return work(service, tx)
		})
	}
	catalogOK(t, withSchema(func(s *OperationalService, tx *gorm.DB) error {
		migrateOperationalTest(t, tx)
		actor = operationalActor(t, tx)
		_, err := s.CreateAppModule(ctx, dto.CreateAppModuleRequest{Code: "MASTER", Name: "Master"})
		if err != nil {
			return err
		}
		kind, err := s.CreateDocumentType(ctx, dto.CreateDocumentTypeRequest{Code: "CONCURRENT", Name: "Concurrent", ModuleCode: "MASTER"})
		if err != nil {
			return err
		}
		typeID = kind.ID
		_, err = s.ReplaceDocumentNumberRule(ctx, typeID, dto.ReplaceDocumentNumberRuleRequest{Prefix: "SEQ", Separator: catalogPointer("-"), SequenceLength: 3, IncludePartnerCode: catalogPointer(false), IncludeWarehouseCode: catalogPointer(false), EffectiveFrom: "2020-01-01"}, actor)
		return err
	}))
	const clients = 24
	results := make(chan dto.GeneratedDocumentIDResponse, clients)
	failures := make(chan error, clients)
	var workers sync.WaitGroup
	for i := 0; i < clients; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			err := withSchema(func(s *OperationalService, _ *gorm.DB) error {
				response, err := s.GenerateDocumentID(ctx, typeID, dto.GenerateDocumentIDRequest{BusinessDate: date})
				if err == nil {
					results <- response
				}
				return err
			})
			if err != nil {
				failures <- err
			}
		}()
	}
	workers.Wait()
	close(results)
	close(failures)
	for err := range failures {
		t.Error(err)
	}
	seen := map[string]bool{}
	for response := range results {
		if seen[response.DocumentID] {
			t.Errorf("duplicate ID %s", response.DocumentID)
		}
		seen[response.DocumentID] = true
	}
	if len(seen) != clients {
		t.Fatalf("allocated %d unique IDs, want %d", len(seen), clients)
	}
	catalogOK(t, withSchema(func(s *OperationalService, tx *gorm.DB) error {
		list := dto.OperationalListRequest{Page: 1, PageSize: 100}
		counters, err := s.ListDocumentDailyCounter(ctx, typeID, date, date, list)
		if err != nil {
			return err
		}
		if len(counters.Items) != 1 || counters.Items[0].LastNumber != "24" {
			return fmt.Errorf("unexpected concurrent counter: %+v", counters)
		}
		nextDay, err := s.GenerateDocumentID(ctx, typeID, dto.GenerateDocumentIDRequest{BusinessDate: "2026-09-04"})
		if err != nil {
			return err
		}
		if nextDay.SequenceNumber != "1" {
			return fmt.Errorf("next day did not restart at one")
		}
		if err := tx.Exec("UPDATE document_daily_counter SET last_number=999 WHERE document_type_id=? AND business_date=?::date", typeID, date).Error; err != nil {
			return err
		}
		_, err = s.GenerateDocumentID(ctx, typeID, dto.GenerateDocumentIDRequest{BusinessDate: date})
		catalogWantError(t, err, ErrCounterExhausted)
		counters, err = s.ListDocumentDailyCounter(ctx, typeID, date, date, list)
		if err != nil {
			return err
		}
		if counters.Items[0].LastNumber != "999" {
			return fmt.Errorf("overflow changed counter")
		}
		_, err = s.ReplaceDocumentNumberRule(ctx, typeID, dto.ReplaceDocumentNumberRuleRequest{Prefix: "SEQ", Separator: catalogPointer("-"), SequenceLength: 18, IncludePartnerCode: catalogPointer(false), IncludeWarehouseCode: catalogPointer(false), EffectiveFrom: "2020-01-01"}, actor)
		if err != nil {
			return err
		}
		if err := tx.Exec("UPDATE document_daily_counter SET last_number=9007199254740992 WHERE document_type_id=? AND business_date=?::date", typeID, date).Error; err != nil {
			return err
		}
		precise, err := s.GenerateDocumentID(ctx, typeID, dto.GenerateDocumentIDRequest{BusinessDate: date})
		if err != nil {
			return err
		}
		if precise.SequenceNumber != "9007199254740993" {
			return fmt.Errorf("large sequence lost precision")
		}
		return nil
	}))
}
