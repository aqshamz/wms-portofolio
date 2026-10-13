package outbound

import (
	"context"
	"errors"
	"math/big"
	"sort"

	dto "wms-api/dto/outbound"
	inventorymodel "wms-api/models/inventory"
	inventoryrepository "wms-api/repository/inventory"
	repository "wms-api/repository/outbound"
)

// Validate the entire manifest against positive HU stock, not just one SKU.
func validateHandlingUnitDispatch(lines []repository.ShipmentLineRow, balances []inventorymodel.InventoryBalance, versions []dto.DispatchBalanceVersion, owner, warehouse, hu, location string) error {
	if len(lines) == 0 || len(balances) == 0 {
		return state("handling unit has no dispatchable shipment stock")
	}
	expected := map[string]int64{}
	for _, v := range versions {
		if !validID(v.BalanceID, 160) || v.ExpectedVersion < 1 || expected[v.BalanceID] != 0 {
			return invalid("expected_balances contains an invalid or duplicate balance")
		}
		expected[v.BalanceID] = v.ExpectedVersion
	}
	if len(expected) != len(balances) {
		return state("refresh the entire handling unit before dispatch")
	}
	totals := map[string]*big.Rat{}
	for _, line := range lines {
		if line.ID != "" {
			return state("handling unit already contains dispatched shipment lines")
		}
		if line.HandlingUnitID == nil || *line.HandlingUnitID != hu {
			return state("shipment line does not belong to the handling unit")
		}
		_, qty, err := quantity(line.ShippedQty, "shipped_qty", false)
		if err != nil {
			return err
		}
		if totals[line.SourceBalanceID] == nil {
			totals[line.SourceBalanceID] = new(big.Rat)
		}
		totals[line.SourceBalanceID].Add(totals[line.SourceBalanceID], qty)
	}
	if len(totals) != len(balances) {
		return state("all handling-unit stock must be included in this shipment")
	}
	for _, b := range balances {
		if b.OwnerID != owner || b.WarehouseID != warehouse || b.LocationID != location || b.HandlingUnitID == nil || *b.HandlingUnitID != hu {
			return state("handling-unit stock has inconsistent ownership or location")
		}
		if expected[b.ID] != b.VersionNo {
			return repository.ErrConcurrentWrite
		}
		_, onHand, err := quantity(b.OnHandQty, "on_hand_qty", false)
		if err != nil {
			return err
		}
		_, reserved, err := quantity(b.ReservedQty, "reserved_qty", true)
		if err != nil {
			return err
		}
		if totals[b.ID] == nil || totals[b.ID].Cmp(onHand) != 0 || reserved.Sign() != 0 {
			return state("dispatch must cover the full unreserved quantity of every handling-unit balance")
		}
	}
	return nil
}

func (s *Service) DispatchShipmentHandlingUnit(ctx context.Context, shipmentID, huID string, q dto.DispatchHandlingUnitRequest, actor string) (dto.ShipmentResponse, error) {
	if !validID(shipmentID, 120) || !validID(huID, 120) || !validUUID(actor) || len(q.ExpectedBalances) == 0 || len(q.ExpectedBalances) > 1000 {
		return dto.ShipmentResponse{}, invalid("invalid handling-unit dispatch request")
	}
	if _, err := date(q.BusinessDate, "business_date"); err != nil {
		return dto.ShipmentResponse{}, err
	}
	key, err := clean(q.OperationKey, 160, "operation_key")
	if err != nil {
		return dto.ShipmentResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		shipment, err := local.repositories.Shipment.Lock(ctx, shipmentID)
		if err != nil {
			return err
		}
		rows, err := local.repositories.ShipmentLine.Candidates(ctx, shipmentID)
		if err != nil {
			return err
		}
		lines := make([]repository.ShipmentLineRow, 0)
		for _, line := range rows {
			if line.HandlingUnitID != nil && *line.HandlingUnitID == huID {
				lines = append(lines, line)
			}
		}
		if len(lines) == 0 {
			return invalid("handling unit is not part of this shipment")
		}
		sort.Slice(lines, func(i, j int) bool { return lines[i].PackingLineID < lines[j].PackingLineID })
		parts := []string{"HU_DISPATCH", shipmentID, huID, q.BusinessDate}
		for _, line := range lines {
			parts = append(parts, line.PackingLineID, line.ShippedQty)
		}
		fp := fingerprint(parts...)
		if movement, e := local.repositories.Inventory.Movement.GetByOperationKey(ctx, key); e == nil {
			if movement.OperationFingerprint == nil || *movement.OperationFingerprint != fp {
				return repository.ErrConflict
			}
			return nil
		} else if !errors.Is(e, inventoryrepository.ErrNotFound) {
			return e
		}
		row, err := local.repositories.Shipment.Get(ctx, shipmentID)
		if err != nil {
			return err
		}
		if row.StatusCode != "PLANNED" {
			return state("only a PLANNED shipment can dispatch a handling unit")
		}
		hu, err := local.repositories.Inventory.HandlingUnit.Lock(ctx, huID)
		if err != nil {
			return err
		}
		if hu.OwnerID != shipment.OwnerID || hu.WarehouseID != shipment.WarehouseID || hu.CurrentLocationID == nil {
			return state("handling unit is not located in the shipment scope")
		}
		children, err := local.repositories.Inventory.HandlingUnit.CountChildren(ctx, huID)
		if err != nil {
			return err
		}
		if children != 0 || hu.ParentHandlingUnitID != nil {
			return state("nested handling units must be unpacked before dispatch")
		}
		balances, err := local.repositories.Inventory.Balance.PositiveForHandlingUnit(ctx, huID)
		if err != nil {
			return err
		}
		for i, b := range balances {
			balances[i], err = local.repositories.Inventory.Balance.FindLocked(ctx, inventoryrepository.BalanceIdentity{OwnerID: b.OwnerID, WarehouseID: b.WarehouseID, LocationID: b.LocationID, ItemID: b.ItemID, LotID: b.LotID, HandlingUnitID: b.HandlingUnitID, InventoryStatusID: b.InventoryStatusID})
			if err != nil {
				return err
			}
		}
		if err := validateHandlingUnitDispatch(lines, balances, q.ExpectedBalances, shipment.OwnerID, shipment.WarehouseID, huID, *hu.CurrentLocationID); err != nil {
			return err
		}
		for i, line := range lines {
			balance, err := local.repositories.Inventory.Balance.Get(ctx, line.SourceBalanceID)
			if err != nil {
				return err
			}
			lineKey := key
			if i > 0 {
				lineKey = fingerprint(key, line.PackingLineID)
			}
			_, err = local.dispatchShipmentLine(ctx, shipmentID, line.PackingLineID, dto.DispatchShipmentLineRequest{ExpectedBalanceVersion: balance.VersionNo, BusinessDate: q.BusinessDate, OperationKey: lineKey}, actor, &fp)
			if err != nil {
				return err
			}
		}
		return local.repositories.Inventory.HandlingUnit.ClearLocation(ctx, huID, *hu.CurrentLocationID)
	})
	if err != nil {
		return dto.ShipmentResponse{}, err
	}
	return s.GetShipment(ctx, shipmentID)
}
