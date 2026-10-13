package outbound

import (
	"errors"
	"testing"
	dto "wms-api/dto/outbound"
	inventorymodel "wms-api/models/inventory"
	model "wms-api/models/outbound"
	repository "wms-api/repository/outbound"
)

func shipmentLineForTest(packing, balance, qty string) model.ShipmentLine {
	return model.ShipmentLine{PackingLineID: packing, SourceBalanceID: balance, ShippedQty: qty}
}

func TestValidateHandlingUnitDispatch(t *testing.T) {
	hu := "HU-1"
	baseLines := []repository.ShipmentLineRow{{HandlingUnitID: &hu, ShipmentLine: shipmentLineForTest("P1", "B1", "2")}, {HandlingUnitID: &hu, ShipmentLine: shipmentLineForTest("P2", "B1", "3")}, {HandlingUnitID: &hu, ShipmentLine: shipmentLineForTest("P3", "B2", "1")}}
	baseBalances := []inventorymodel.InventoryBalance{{ID: "B1", OwnerID: "owner", WarehouseID: "warehouse", LocationID: "packing", HandlingUnitID: &hu, OnHandQty: "5", ReservedQty: "0", VersionNo: 4}, {ID: "B2", OwnerID: "owner", WarehouseID: "warehouse", LocationID: "packing", HandlingUnitID: &hu, OnHandQty: "1", ReservedQty: "0", VersionNo: 7}}
	baseVersions := []dto.DispatchBalanceVersion{{BalanceID: "B1", ExpectedVersion: 4}, {BalanceID: "B2", ExpectedVersion: 7}}
	for _, tc := range []struct {
		name   string
		change func([]repository.ShipmentLineRow, []inventorymodel.InventoryBalance, []dto.DispatchBalanceVersion)
		want   error
		valid  bool
	}{
		{name: "multiple identities and shared balance", valid: true},
		{name: "foreign stock not covered", change: func(_ []repository.ShipmentLineRow, b []inventorymodel.InventoryBalance, _ []dto.DispatchBalanceVersion) {
			b[1].OnHandQty = "2"
		}, want: ErrInvalidState},
		{name: "stale balance version", change: func(_ []repository.ShipmentLineRow, _ []inventorymodel.InventoryBalance, v []dto.DispatchBalanceVersion) {
			v[1].ExpectedVersion = 6
		}, want: repository.ErrConcurrentWrite},
		{name: "duplicate balance snapshot", change: func(_ []repository.ShipmentLineRow, _ []inventorymodel.InventoryBalance, v []dto.DispatchBalanceVersion) {
			v[1] = v[0]
		}, want: ErrInvalidInput},
		{name: "wrong owner", change: func(_ []repository.ShipmentLineRow, b []inventorymodel.InventoryBalance, _ []dto.DispatchBalanceVersion) {
			b[1].OwnerID = "other"
		}, want: ErrInvalidState},
		{name: "wrong location", change: func(_ []repository.ShipmentLineRow, b []inventorymodel.InventoryBalance, _ []dto.DispatchBalanceVersion) {
			b[1].LocationID = "receiving"
		}, want: ErrInvalidState},
		{name: "already dispatched", change: func(l []repository.ShipmentLineRow, _ []inventorymodel.InventoryBalance, _ []dto.DispatchBalanceVersion) {
			l[0].ID = "SHIP-L1"
		}, want: ErrInvalidState},
		{name: "reserved stock", change: func(_ []repository.ShipmentLineRow, b []inventorymodel.InventoryBalance, _ []dto.DispatchBalanceVersion) {
			b[0].ReservedQty = "1"
		}, want: ErrInvalidState},
		{name: "zero quantity", change: func(l []repository.ShipmentLineRow, _ []inventorymodel.InventoryBalance, _ []dto.DispatchBalanceVersion) {
			l[0].ShippedQty = "0"
		}, want: ErrInvalidInput},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lines := append([]repository.ShipmentLineRow(nil), baseLines...)
			balances := append([]inventorymodel.InventoryBalance(nil), baseBalances...)
			versions := append([]dto.DispatchBalanceVersion(nil), baseVersions...)
			if tc.change != nil {
				tc.change(lines, balances, versions)
			}
			err := validateHandlingUnitDispatch(lines, balances, versions, "owner", "warehouse", hu, "packing")
			if tc.valid {
				if err != nil {
					t.Fatal(err)
				}
			} else if !errors.Is(err, tc.want) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
	if err := validateHandlingUnitDispatch(nil, baseBalances, baseVersions, "owner", "warehouse", hu, "packing"); err == nil {
		t.Fatal("empty shipment HU accepted")
	}
	if err := validateHandlingUnitDispatch(baseLines, baseBalances[:1], baseVersions[:1], "owner", "warehouse", hu, "packing"); err == nil {
		t.Fatal("missing HU balance accepted")
	}
}
