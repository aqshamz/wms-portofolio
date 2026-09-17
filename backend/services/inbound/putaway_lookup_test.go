package inbound

import (
	"context"
	"errors"
	"testing"
	dto "wms-api/dto/inbound"
	model "wms-api/models/master"
)

func TestPutawayRecoveryRequiresReason(t *testing.T) {
	service := &Service{}
	request := dto.CancelPutawayRequest{ExpectedVersion: 1, ExpectedBalanceVersion: 1, BusinessDate: "2026-09-17", Reason: "  "}
	actor := "11111111-1111-4111-8111-111111111111"
	if _, err := service.CancelPutawayTask(context.Background(), "PUT-1", request, actor); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("cancel accepted blank reason: %v", err)
	}
	if _, err := service.ReversePutawayTask(context.Background(), "PUT-1", request, actor); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("reverse accepted blank reason: %v", err)
	}
}

func TestPutawayRuleMatches(t *testing.T) {
	category, kind, zone := "category", "storage", "zone"
	wrong := "other"
	item := model.Item{CategoryID: &category}
	location := model.WarehouseLocation{LocationTypeID: kind, ZoneID: zone}
	for _, test := range []struct {
		name string
		rule model.PutawayStrategyRule
		want bool
	}{
		{"wildcard", model.PutawayStrategyRule{}, true},
		{"all match", model.PutawayStrategyRule{CategoryID: &category, LocationTypeID: &kind, ZoneID: &zone}, true},
		{"wrong category", model.PutawayStrategyRule{CategoryID: &wrong}, false},
		{"wrong type", model.PutawayStrategyRule{LocationTypeID: &wrong}, false},
		{"wrong zone", model.PutawayStrategyRule{ZoneID: &wrong}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := putawayRuleMatches(item, location, test.rule); got != test.want {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
	if putawayRuleMatches(model.Item{}, location, model.PutawayStrategyRule{CategoryID: &category}) {
		t.Fatal("uncategorized item must not match a category-specific rule")
	}
}

func TestValidatePutawayLookup(t *testing.T) {
	search := "  storage  "
	if err := validatePutawayLookup("PUT-1", &search, 1, 20); err != nil || search != "storage" {
		t.Fatalf("valid lookup failed: %q %v", search, err)
	}
	for _, test := range []struct {
		id         string
		page, size int
	}{{"", 1, 20}, {"PUT-1", 0, 20}, {"PUT-1", 1, 101}, {"PUT-1", 1000001, 20}} {
		if err := validatePutawayLookup(test.id, &search, test.page, test.size); err == nil {
			t.Fatal("invalid lookup was accepted")
		}
	}
}
