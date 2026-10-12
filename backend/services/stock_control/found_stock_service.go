package stockcontrol

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	dto "wms-api/dto/stock_control"
	inventorymodel "wms-api/models/inventory"
	model "wms-api/models/stock_control"
	inventoryrepo "wms-api/repository/inventory"
	repository "wms-api/repository/stock_control"
)

func countLineIdentity(header model.CycleCount, line model.CycleCountLine) inventoryrepo.BalanceIdentity {
	return inventoryrepo.BalanceIdentity{OwnerID: header.OwnerID, WarehouseID: header.WarehouseID, ItemID: line.ItemID, LocationID: line.LocationID, InventoryStatusID: line.InventoryStatusID, LotID: line.LotID, HandlingUnitID: line.HandlingUnitID}
}

// Found lines have no inventory balance until approval. Existing zero balances
// are snapshotted so ordinary counts still detect stock changes before posting.
func loadCountBalance(ctx context.Context, r *repository.Repositories, ir *inventoryrepo.Repositories, header model.CycleCount, line model.CycleCountLine, posting bool) (inventoryrepo.BalanceRow, error) {
	if line.FoundStock {
		item, err := ir.Catalog.Item.GetShared(ctx, line.ItemID)
		if err != nil {
			return inventoryrepo.BalanceRow{}, err
		}
		if !item.IsActive || item.OwnerID != header.OwnerID || item.BaseUOMID != line.UOMID {
			return inventoryrepo.BalanceRow{}, invalid("the found-stock item or its base unit changed; reject this line and create a new count")
		}
		key := countLineIdentity(header, line)
		var current inventorymodel.InventoryBalance
		if posting {
			postingCtx := countPostingContext(ctx, header)
			if err = ir.Balance.LockStockKey(postingCtx, header.OwnerID, header.WarehouseID, line.ItemID); err != nil {
				return inventoryrepo.BalanceRow{}, err
			}
			current, err = ir.Balance.FindLocked(postingCtx, key)
		} else {
			current, err = ir.Balance.FindIdentity(ctx, key)
		}
		if err != nil && !errors.Is(err, inventoryrepo.ErrNotFound) {
			return inventoryrepo.BalanceRow{}, err
		}
		if line.BalanceID == nil {
			if err == nil {
				return inventoryrepo.BalanceRow{}, invalid("stock was registered after this found-stock line; cancel or reject the line and count the current balance")
			}
			return inventoryrepo.BalanceRow{InventoryBalance: inventorymodel.InventoryBalance{OwnerID: header.OwnerID, WarehouseID: header.WarehouseID, ItemID: line.ItemID, LocationID: line.LocationID, InventoryStatusID: line.InventoryStatusID, LotID: line.LotID, HandlingUnitID: line.HandlingUnitID, UOMID: line.UOMID, OnHandQty: "0.000000", ReservedQty: "0.000000"}, SerialControlled: item.SerialControlled}, nil
		}
		if err != nil || current.ID != *line.BalanceID || current.OnHandQty != "0.000000" || current.ReservedQty != "0.000000" || current.UOMID != line.UOMID {
			return inventoryrepo.BalanceRow{}, invalid("the found-stock balance changed; reject this line and count the current balance")
		}
	}
	if line.BalanceID == nil {
		return inventoryrepo.BalanceRow{}, fmt.Errorf("count line is missing its balance snapshot")
	}
	balance, err := r.Balances.Get(ctx, *line.BalanceID)
	if err != nil {
		return inventoryrepo.BalanceRow{}, err
	}
	if balance.VersionNo != line.SnapshotVersionNo {
		return inventoryrepo.BalanceRow{}, invalid("a balance changed after the count was created; cancel and create a fresh count")
	}
	return balance, nil
}

func countLineSelfReview(header model.CycleCount, line model.CycleCountLine, actor string) bool {
	return header.CreatedBy == actor || (line.FoundStock && (line.CreatedBy == actor || (line.CountedBy != nil && *line.CountedBy == actor)))
}

func (s *Service) GetFoundStockOptions(ctx context.Context, id string, f repository.FoundStockOptionsFilter, actor string) (dto.FoundStockOptionsResponse, error) {
	row, err := s.repositories.CycleCounts.Get(ctx, id)
	if err != nil {
		return dto.FoundStockOptionsResponse{}, err
	}
	if err = authorize(ctx, s.repositories, actor, scopeBalance(row.OwnerID, row.WarehouseID)); err != nil {
		return dto.FoundStockOptionsResponse{}, err
	}
	options, err := s.repositories.CycleCounts.FoundStockOptions(ctx, row.CycleCount, f)
	if err != nil {
		return dto.FoundStockOptionsResponse{}, err
	}
	response := dto.FoundStockOptionsResponse{Items: []dto.FoundStockItemOption{}, Locations: []dto.FoundStockOption{}, Statuses: []dto.FoundStockOption{}, Lots: []dto.FoundStockOption{}, HandlingUnits: []dto.FoundStockOption{}}
	for _, item := range options.Items {
		response.Items = append(response.Items, dto.FoundStockItemOption(item))
	}
	for _, pair := range []struct {
		source []repository.FoundStockOption
		target *[]dto.FoundStockOption
	}{{options.Locations, &response.Locations}, {options.Statuses, &response.Statuses}, {options.Lots, &response.Lots}, {options.HandlingUnits, &response.HandlingUnits}} {
		for _, entry := range pair.source {
			*pair.target = append(*pair.target, dto.FoundStockOption(entry))
		}
	}
	return response, nil
}

func (s *Service) AddFoundStock(ctx context.Context, id string, q dto.AddFoundStockRequest, actor string) (dto.CycleCountResponse, error) {
	quantity, counted, err := number(q.CountedQty, true)
	if err != nil {
		return dto.CycleCountResponse{}, err
	}
	if q.Notes == nil || strings.TrimSpace(*q.Notes) == "" {
		return dto.CycleCountResponse{}, invalid("describe where and how this stock was found")
	}
	if q.LotID != nil {
		value := strings.TrimSpace(*q.LotID)
		if value == "" {
			q.LotID = nil
		} else {
			q.LotID = &value
		}
	}
	if q.HandlingUnitID != nil {
		value := strings.TrimSpace(*q.HandlingUnitID)
		if value == "" {
			q.HandlingUnitID = nil
		} else {
			q.HandlingUnitID = &value
		}
	}
	err = s.repositories.Transaction(ctx, func(r *repository.Repositories, ir *inventoryrepo.Repositories) error {
		header, err := r.CycleCounts.Lock(ctx, id)
		if err != nil {
			return err
		}
		if err = authorize(ctx, r, actor, scopeBalance(header.OwnerID, header.WarehouseID)); err != nil {
			return err
		}
		if header.VersionNo != q.ExpectedVersion {
			return inventoryrepo.ErrConflict
		}
		if header.CompletedAt != nil || header.CancelledAt != nil {
			return invalid("found stock can only be added to an open count")
		}
		if err = validateCountStorageLocation(ctx, ir, header.WarehouseID, q.LocationID, "found stock", header.CountTypeCode == "GRAND"); err != nil {
			return err
		}
		item, err := ir.Catalog.Item.GetShared(ctx, q.ItemID)
		if err != nil {
			return err
		}
		if !item.IsActive || item.OwnerID != header.OwnerID {
			return invalid("found stock item must be active and belong to this count's customer")
		}
		uom, err := ir.Catalog.UOM.GetShared(ctx, item.BaseUOMID)
		if err != nil {
			return err
		}
		if !uom.IsActive {
			return invalid("the item's base unit is inactive")
		}
		status, err := ir.Catalog.InventoryStatus.GetShared(ctx, q.InventoryStatusID)
		if err != nil {
			return err
		}
		if !status.IsActive {
			return invalid("found stock inventory status is inactive")
		}
		if item.SerialControlled && (!counted.IsInt() || counted.Cmp(big.NewRat(5000, 1)) > 0) {
			return invalid("serialized found stock must be a whole number of up to 5000 base units")
		}
		if item.LotControlled && q.LotID == nil {
			return invalid("a lot is required for this item; register the lot in Inventory lots first")
		}
		if !item.LotControlled && q.LotID != nil {
			return invalid("this item is not lot-controlled")
		}
		if q.LotID != nil {
			lot, err := ir.Lot.GetShared(ctx, *q.LotID)
			if err != nil {
				return err
			}
			if lot.OwnerID != header.OwnerID || lot.ItemID != item.ID {
				return invalid("the selected lot does not belong to this customer and item")
			}
		}
		if q.HandlingUnitID != nil {
			hu, err := ir.HandlingUnit.GetShared(ctx, *q.HandlingUnitID)
			if err != nil {
				return err
			}
			if hu.OwnerID != header.OwnerID || hu.WarehouseID != header.WarehouseID || hu.IsClosed || hu.ParentHandlingUnitID != nil || hu.CurrentLocationID == nil || *hu.CurrentLocationID != q.LocationID {
				return invalid("handling unit must be open, unnested, and at the found location in this customer's warehouse")
			}
		}
		line := model.CycleCountLine{CycleCountID: header.ID, FoundStock: true, ItemID: item.ID, LotID: q.LotID, HandlingUnitID: q.HandlingUnitID, LocationID: q.LocationID, InventoryStatusID: q.InventoryStatusID, UOMID: item.BaseUOMID, SystemQty: "0.000000", DecisionCode: "OPEN", CreatedBy: actor, UpdatedBy: &actor, VersionNo: 1}
		key := countLineIdentity(header, line)
		duplicate, err := r.CycleCounts.HasLineIdentity(ctx, id, key)
		if err != nil {
			return err
		}
		if duplicate {
			return invalid("this stock identity already has a line in the count; use that line")
		}
		balance, err := ir.Balance.FindIdentity(ctx, key)
		if err == nil {
			if balance.OnHandQty != "0.000000" || balance.ReservedQty != "0.000000" || balance.UOMID != line.UOMID {
				return invalid("stock already exists for this identity; include its balance in a count instead of adding found stock")
			}
			line.BalanceID, line.SnapshotVersionNo = &balance.ID, balance.VersionNo
		} else if !errors.Is(err, inventoryrepo.ErrNotFound) {
			return err
		}
		line.LineNo, err = r.CycleCounts.NextLineNo(ctx, id)
		if err != nil {
			return err
		}
		if line.LineNo > maxGrandStockOpnameLines {
			return invalid("the count cannot exceed 5000 lines")
		}
		line.ID = fmt.Sprintf("%s-L%04d", id, line.LineNo)
		tolerance, ok := new(big.Rat).SetString(header.ToleranceQty)
		if !ok {
			return fmt.Errorf("invalid stored count tolerance")
		}
		line.RequiresRecount = counted.Cmp(tolerance) > 0
		if err = r.CycleCounts.AddLine(ctx, &line); err != nil {
			return err
		}
		entry := model.CycleCountEntry{ID: line.ID + "-A01", CycleCountLineID: line.ID, AttemptNo: 1, SystemQty: line.SystemQty, SnapshotVersionNo: line.SnapshotVersionNo, CountedQty: quantity, VarianceQty: quantity, Notes: q.Notes, CountedBy: actor}
		if err = r.CycleCounts.Record(ctx, line, entry, actor); err != nil {
			return err
		}
		return updateCycleCountHeader(ctx, r, header, actor, nil)
	})
	if err != nil {
		return dto.CycleCountResponse{}, err
	}
	return s.GetCycleCount(ctx, id, actor)
}
