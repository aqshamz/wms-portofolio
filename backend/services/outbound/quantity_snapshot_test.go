package outbound

import (
	"testing"
	mastermodel "wms-api/models/master"
	model "wms-api/models/outbound"
)

func TestOrderQuantitySnapshot(t *testing.T) {
	for _, test := range []struct {
		qty, factor, base string
		serial, invalid   bool
	}{
		{"2", "12", "24.000000", false, false},
		{"0.5", "12", "6.000000", true, false},
		{"0.5", "3", "", true, true},
		{"0.000001", "0.000001", "", false, true},
		{"99999999999999", "2", "", false, true},
		{"0", "1", "", false, true},
		{"1", "0", "", false, true},
		{"bad", "1", "", false, true},
		{"0.000001", "1", "0.000001", false, false},
	} {
		t.Run(test.qty+"x"+test.factor, func(t *testing.T) {
			_, _, base, err := orderQuantitySnapshot(test.qty, test.factor, test.serial)
			if (err != nil) != test.invalid || base != test.base {
				t.Fatalf("base=%q error=%v", base, err)
			}
		})
	}
}

func TestValidateOrderSnapshot(t *testing.T) {
	item := mastermodel.Item{BaseUOMID: "base", IsActive: true}
	line := model.OutboundOrderLine{UOMID: "carton", BaseUOMID: "base", OrderedQty: "2", UOMConversionToBase: "12", OrderedBaseQty: "24"}
	if err := validateOrderSnapshot(line, item); err != nil {
		t.Fatal(err)
	}
	line.OrderedBaseQty = "25"
	if err := validateOrderSnapshot(line, item); err == nil {
		t.Fatal("inconsistent base quantity accepted")
	}
	line.OrderedBaseQty = "24"
	item.BaseUOMID = "changed"
	if err := validateOrderSnapshot(line, item); err == nil {
		t.Fatal("changed base UOM accepted")
	}
}
