package outbound

import (
	"context"
	"errors"
	"math/big"
	"time"
	dto "wms-api/dto/outbound"
	inventorymodel "wms-api/models/inventory"
	model "wms-api/models/outbound"
	inventoryrepository "wms-api/repository/inventory"
	repository "wms-api/repository/outbound"
)

func (s *Service) GetPacking(ctx context.Context, id string) (dto.PackingResponse, error) {
	if !validID(id, 120) {
		return dto.PackingResponse{}, invalid("invalid packing_id")
	}
	row, err := s.repositories.Packing.Get(ctx, id)
	if err != nil {
		return dto.PackingResponse{}, err
	}
	out := mapPacking(row)
	lines, err := s.repositories.PackingLine.Candidates(ctx, id)
	if err != nil {
		return dto.PackingResponse{}, err
	}
	out.Lines = make([]dto.PackingLineResponse, 0, len(lines))
	for _, v := range lines {
		out.Lines = append(out.Lines, mapPackingLine(v))
	}
	return out, nil
}
func (s *Service) ListPackings(ctx context.Context, f repository.ListFilter) (dto.PageResponse[dto.PackingResponse], error) {
	if !validUUID(f.OwnerID) {
		return dto.PageResponse[dto.PackingResponse]{}, invalid("owner_id is required and must be a UUID")
	}
	rows, total, err := s.repositories.Packing.List(ctx, f)
	if err != nil {
		return dto.PageResponse[dto.PackingResponse]{}, err
	}
	items := make([]dto.PackingResponse, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapPacking(v))
	}
	return page(items, f.Page, f.PageSize, total), nil
}
func (s *Service) CreatePacking(ctx context.Context, q dto.CreatePackingRequest, actor string) (dto.PackingResponse, error) {
	if !validID(q.OutboundCheckID, 120) || !validUUID(q.PackingLocationID) || !validUUID(actor) {
		return dto.PackingResponse{}, invalid("invalid packing request")
	}
	var id string
	err := s.transaction(ctx, func(local *Service) error {
		check, err := local.repositories.Check.Get(ctx, q.OutboundCheckID)
		if err != nil {
			return err
		}
		if check.StatusCode != "PASSED" {
			return state("packing requires a PASSED outbound check")
		}
		order, err := local.repositories.Order.Lock(ctx, check.OutboundID)
		if err != nil {
			return err
		}
		orderRow, err := local.repositories.Order.Get(ctx, order.ID)
		if err != nil {
			return err
		}
		if orderRow.StatusCode != "CHECKED" {
			return state("only a CHECKED order can enter packing")
		}
		active, err := local.repositories.Packing.HasActiveByOutbound(ctx, order.ID)
		if err != nil {
			return err
		}
		if active {
			return state("outbound order already has an active packing document")
		}
		location, err := local.repositories.Inventory.Location.GetShared(ctx, q.PackingLocationID)
		if err != nil {
			return err
		}
		if location.WarehouseID != order.WarehouseID || !location.IsActive || location.IsLocked {
			return invalid("packing location is not an active unlocked location in the order warehouse")
		}
		kind, initial, err := local.document(ctx, "PACKING")
		if err != nil {
			return err
		}
		id, err = local.generateID(ctx, "PACKING", order.BusinessDate, nil, &order.WarehouseID)
		if err != nil {
			return err
		}
		v := model.Packing{ID: id, DocumentTypeID: kind.ID, StatusID: initial.ID, OutboundID: order.ID, WarehouseID: order.WarehouseID, PackingLocationID: &q.PackingLocationID, CreatedBy: actor}
		if err := local.repositories.Packing.Create(ctx, &v); err != nil {
			return err
		}
		next, err := local.transition(ctx, order.DocumentTypeID, order.StatusID, "PACKING")
		if err != nil {
			return err
		}
		return local.repositories.Order.SetStatusLocked(ctx, order.ID, next.ID, actor, nil)
	})
	if err != nil {
		return dto.PackingResponse{}, err
	}
	return s.GetPacking(ctx, id)
}
func (s *Service) PackLine(ctx context.Context, packingID, checkLineID string, q dto.PackLineRequest, actor string) (dto.PackingResponse, error) {
	if !validID(packingID, 120) || !validID(checkLineID, 160) || !validUUID(actor) || q.ExpectedBalanceVersion < 1 {
		return dto.PackingResponse{}, invalid("invalid packing-line request")
	}
	businessDate, err := date(q.BusinessDate, "business_date")
	if err != nil {
		return dto.PackingResponse{}, err
	}
	operationKey, err := clean(q.OperationKey, 160, "operation_key")
	if err != nil {
		return dto.PackingResponse{}, err
	}
	handlingUnitID := ""
	if q.HandlingUnitID != nil {
		handlingUnitID = *q.HandlingUnitID
	}
	fp := fingerprint(packingID, checkLineID, q.BusinessDate, handlingUnitID)
	if existing, e := s.repositories.Inventory.Movement.GetByOperationKey(ctx, operationKey); e == nil {
		if existing.OperationFingerprint == nil || *existing.OperationFingerprint != fp {
			return dto.PackingResponse{}, repository.ErrConflict
		}
		return s.GetPacking(ctx, packingID)
	} else if !errors.Is(e, inventoryrepository.ErrNotFound) {
		return dto.PackingResponse{}, e
	}
	err = s.transaction(ctx, func(local *Service) error {
		packing, err := local.repositories.Packing.Lock(ctx, packingID)
		if err != nil {
			return err
		}
		packingRow, err := local.repositories.Packing.Get(ctx, packingID)
		if err != nil {
			return err
		}
		if packingRow.StatusCode != "OPEN" {
			return state("only an OPEN packing document accepts lines")
		}
		candidate, err := local.repositories.PackingLine.GetCandidate(ctx, packingID, checkLineID)
		if err != nil {
			return err
		}
		if candidate.ID != "" {
			return repository.ErrConflict
		}
		sourceRow, err := local.repositories.Inventory.Balance.Get(ctx, candidate.SourceBalanceID)
		if err != nil {
			return err
		}
		identity := inventoryrepository.BalanceIdentity{OwnerID: sourceRow.OwnerID, WarehouseID: sourceRow.WarehouseID, LocationID: sourceRow.LocationID, ItemID: sourceRow.ItemID, LotID: sourceRow.LotID, HandlingUnitID: sourceRow.HandlingUnitID, InventoryStatusID: sourceRow.InventoryStatusID}
		source, err := local.repositories.Inventory.Balance.FindLocked(ctx, identity)
		if err != nil {
			return err
		}
		if source.VersionNo != q.ExpectedBalanceVersion {
			return repository.ErrConcurrentWrite
		}
		qty := candidate.ExpectedQty
		qtyRat := mustRat(qty)
		if mustRat(source.OnHandQty).Cmp(qtyRat) < 0 {
			return state("staging balance no longer covers checked quantity")
		}
		if packing.PackingLocationID == nil {
			return state("packing document has no packing location")
		}
		targetHU := q.HandlingUnitID
		var targetHandlingUnit *inventorymodel.HandlingUnit
		if targetHU != nil {
			hu, err := local.repositories.Inventory.HandlingUnit.GetShared(ctx, *targetHU)
			if err != nil {
				return err
			}
			if hu.OwnerID != source.OwnerID || hu.WarehouseID != source.WarehouseID || hu.IsClosed {
				return invalid("handling unit is not open in the same owner and warehouse")
			}
			targetHandlingUnit = &hu
		}
		if source.HandlingUnitID != nil {
			if targetHU == nil || *targetHU != *source.HandlingUnitID {
				return state("stock already in a handling unit must remain in that handling unit while packing")
			}
			if mustRat(source.OnHandQty).Cmp(qtyRat) != 0 {
				return state("handling-unit stock must be packed in full")
			}
			positive, err := local.repositories.Inventory.Balance.CountPositiveForHandlingUnit(ctx, *source.HandlingUnitID)
			if err != nil {
				return err
			}
			if positive != 1 {
				return state("handling unit contains multiple positive stock identities and cannot be moved by one packing line")
			}
		}
		newSource := decimal(new(big.Rat).Sub(mustRat(source.OnHandQty), qtyRat))
		if err := local.repositories.Inventory.Balance.SetQuantities(ctx, source.ID, newSource, source.ReservedQty); err != nil {
			return err
		}
		targetIdentity := inventoryrepository.BalanceIdentity{OwnerID: source.OwnerID, WarehouseID: source.WarehouseID, LocationID: *packing.PackingLocationID, ItemID: source.ItemID, LotID: source.LotID, HandlingUnitID: targetHU, InventoryStatusID: source.InventoryStatusID}
		target, err := local.repositories.Inventory.Balance.FindLocked(ctx, targetIdentity)
		packingBalanceID := ""
		if err != nil {
			if !errors.Is(err, inventoryrepository.ErrNotFound) {
				return err
			}
			packingBalanceID, err = local.generateID(ctx, "INVENTORY_BALANCE", businessDate, nil, &source.WarehouseID)
			if err != nil {
				return err
			}
			target = inventorymodel.InventoryBalance{ID: packingBalanceID, OwnerID: source.OwnerID, WarehouseID: source.WarehouseID, LocationID: *packing.PackingLocationID, ItemID: source.ItemID, LotID: source.LotID, HandlingUnitID: targetHU, InventoryStatusID: source.InventoryStatusID, OnHandQty: qty, ReservedQty: "0.000000", UOMID: source.UOMID, VersionNo: 1}
			if err := local.repositories.Inventory.Balance.Create(ctx, &target); err != nil {
				return err
			}
		} else {
			packingBalanceID = target.ID
			if err := local.repositories.Inventory.Balance.SetQuantities(ctx, target.ID, decimal(new(big.Rat).Add(mustRat(target.OnHandQty), qtyRat)), target.ReservedQty); err != nil {
				return err
			}
		}
		serialID, err := local.moveSerialIdentity(ctx, source.ID, &packingBalanceID, qtyRat)
		if err != nil {
			return err
		}
		if targetHandlingUnit != nil {
			switch {
			case targetHandlingUnit.CurrentLocationID == nil:
				if err := local.repositories.Inventory.HandlingUnit.Place(ctx, targetHandlingUnit.ID, *packing.PackingLocationID); err != nil {
					return err
				}
			case *targetHandlingUnit.CurrentLocationID == *packing.PackingLocationID:
				if source.HandlingUnitID != nil && source.LocationID != *packing.PackingLocationID {
					return state("handling-unit location does not match its source balance")
				}
			case *targetHandlingUnit.CurrentLocationID == source.LocationID:
				if err := local.repositories.Inventory.HandlingUnit.Relocate(ctx, targetHandlingUnit.ID, source.LocationID, *packing.PackingLocationID); err != nil {
					return err
				}
			default:
				return state("handling unit is currently at another location")
			}
		}
		kind, err := local.repositories.Inventory.MovementType.ByCodeShared(ctx, "PACK")
		if err != nil || !kind.IsActive {
			return state("PACK movement type is not configured")
		}
		movementID, err := local.generateID(ctx, "MOVEMENT", businessDate, nil, &source.WarehouseID)
		if err != nil {
			return err
		}
		from := source.LocationID
		to := *packing.PackingLocationID
		status := source.InventoryStatusID
		movement := inventorymodel.InventoryMovement{ID: movementID, MovementTypeID: kind.ID, OwnerID: source.OwnerID, WarehouseID: source.WarehouseID, BusinessDate: businessDate, OccurredAt: time.Now(), ItemID: source.ItemID, LotID: source.LotID, SerialID: serialID, HandlingUnitID: targetHU, FromLocationID: &from, ToLocationID: &to, FromStatusID: &status, ToStatusID: &status, Quantity: qty, UOMID: source.UOMID, SourceDocumentID: packingID, SourceLineID: &checkLineID, OperationKey: &operationKey, OperationFingerprint: &fp, CreatedBy: actor}
		if err := local.repositories.Inventory.Movement.Create(ctx, &movement); err != nil {
			return err
		}
		lineID, err := randomID(packingID + "-L")
		if err != nil {
			return err
		}
		line := model.PackingLine{ID: lineID, PackingID: packingID, PickTaskID: candidate.PickTaskID, OutboundCheckLineID: checkLineID, SourceBalanceID: source.ID, PackingBalanceID: packingBalanceID, HandlingUnitID: targetHU, PackedQty: qty, UOMID: source.UOMID, MovementID: movementID, CreatedBy: actor}
		if err := local.repositories.PackingLine.Create(ctx, &line); err != nil {
			return err
		}
		if err := local.repositories.Line.AddPacked(ctx, candidate.OutboundLineID, qty); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return dto.PackingResponse{}, err
	}
	return s.GetPacking(ctx, packingID)
}
func (s *Service) CompletePacking(ctx context.Context, id, actor string) (dto.PackingResponse, error) {
	if !validID(id, 120) || !validUUID(actor) {
		return dto.PackingResponse{}, invalid("invalid packing completion")
	}
	err := s.transaction(ctx, func(local *Service) error {
		packing, err := local.repositories.Packing.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.Packing.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "OPEN" {
			return state("only an OPEN packing document can be completed")
		}
		missing, err := local.repositories.PackingLine.MissingCount(ctx, id)
		if err != nil {
			return err
		}
		if missing > 0 {
			return state("pack every passed check line before completion")
		}
		next, err := local.transition(ctx, packing.DocumentTypeID, packing.StatusID, "COMPLETED")
		if err != nil {
			return err
		}
		if err := local.repositories.Packing.Complete(ctx, id, next.ID, actor); err != nil {
			return err
		}
		order, err := local.repositories.Order.Lock(ctx, packing.OutboundID)
		if err != nil {
			return err
		}
		orderNext, err := local.transition(ctx, order.DocumentTypeID, order.StatusID, "PACKED")
		if err != nil {
			return err
		}
		return local.repositories.Order.SetStatusLocked(ctx, order.ID, orderNext.ID, actor, nil)
	})
	if err != nil {
		return dto.PackingResponse{}, err
	}
	return s.GetPacking(ctx, id)
}

func (s *Service) CancelPacking(ctx context.Context, id, actor string) (dto.PackingResponse, error) {
	if !validID(id, 120) || !validUUID(actor) {
		return dto.PackingResponse{}, invalid("invalid packing cancellation")
	}
	err := s.transaction(ctx, func(local *Service) error {
		packing, err := local.repositories.Packing.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.Packing.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "OPEN" {
			return state("only an OPEN packing document can be cancelled")
		}
		count, err := local.repositories.PackingLine.Count(ctx, id)
		if err != nil {
			return err
		}
		if count != 0 {
			return state("packing with posted lines cannot be cancelled")
		}
		cancelled, err := local.transition(ctx, packing.DocumentTypeID, packing.StatusID, "CANCELLED")
		if err != nil {
			return err
		}
		if err := local.repositories.Packing.Cancel(ctx, id, cancelled.ID); err != nil {
			return err
		}
		order, err := local.repositories.Order.Lock(ctx, packing.OutboundID)
		if err != nil {
			return err
		}
		checked, err := local.status(ctx, order.DocumentTypeID, "CHECKED")
		if err != nil {
			return err
		}
		return local.repositories.Order.SetStatusLocked(ctx, order.ID, checked.ID, actor, nil)
	})
	if err != nil {
		return dto.PackingResponse{}, err
	}
	return s.GetPacking(ctx, id)
}
