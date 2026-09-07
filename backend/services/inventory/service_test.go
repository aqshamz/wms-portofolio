package inventory

import (
	"strings"
	"testing"
	repository "wms-api/repository/inventory"
)

func TestIdentityValidation(t *testing.T) {
	for _, input := range []string{"", "   ", "a\nb", strings.Repeat("a", 101)} {
		if _, err := normalizeNumber(input, 100); err == nil {
			t.Errorf("accepted %q", input)
		}
	}
	if v, err := normalizeNumber(" Batch-a ", 100); err != nil || v != "Batch-a" {
		t.Fatal(v, err)
	}
	for _, input := range []string{"2026-02-30", "2026-9-04", "", "0000-01-01", "2026-09-04T00:00:00Z"} {
		if _, err := date(&input); err == nil {
			t.Errorf("accepted date %q", input)
		}
	}
	value := "2028-02-29"
	if _, err := date(&value); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"../lot", "LOT/1", " LOT-1", "", strings.Repeat("a", 121)} {
		if identityID(id, 120) {
			t.Errorf("accepted ID %q", id)
		}
	}
	for _, prefix := range []string{"LOT", "SER", "HU"} {
		a, err := newID(prefix)
		if err != nil {
			t.Fatal(err)
		}
		b, err := newID(prefix)
		if err != nil || a == b || !identityID(a, 120) {
			t.Fatal(a, b, err)
		}
	}
	for _, f := range []repository.Filter{
		{Page: 0, PageSize: 20}, {Page: 1000001, PageSize: 20}, {Page: 1, PageSize: 101},
		{Page: 1, PageSize: 20, OwnerID: "STUDY_OWNER"},
		{Page: 1, PageSize: 20, WarehouseID: "11111111-1111-4111-8111-111111111111"},
	} {
		if validateFilter(&f, false) == nil {
			t.Errorf("accepted filter %+v", f)
		}
	}
}
