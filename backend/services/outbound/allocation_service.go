package outbound

import (
	"context"
	"math/big"

	dto "wms-api/dto/outbound"
	model "wms-api/models/outbound"
	inventoryrepository "wms-api/repository/inventory"
	repository "wms-api/repository/outbound"
)

func (s *Service) AllocateOutboundOrder(ctx context.Context, id string, request dto.AllocateOutboundRequest, actor string) (dto.AllocationResponse, error) {
	if !validID(id, 120) || !validUUID(actor) || request.ExpectedVersion < 1 {
		return dto.AllocationResponse{}, invalid("invalid allocation request")
	}
	if request.PickingStrategyID != nil && !validUUID(*request.PickingStrategyID) {
		return dto.AllocationResponse{}, invalid("picking_strategy_id must be a UUID")
	}
	requestedTotal := new(big.Rat)
	allocatedTotal := new(big.Rat)
	err := s.transaction(ctx, func(local *Service) error {
		header, err := local.repositories.Order.Lock(ctx, id)
		if err != nil {
			return err
		}
		if header.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		row, err := local.repositories.Order.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "RELEASED" && row.StatusCode != "PARTIALLY_ALLOCATED" {
			return state("only RELEASED or PARTIALLY_ALLOCATED orders can be allocated")
		}
		strategy, err := local.pickingStrategy(ctx, header.OwnerID, header.WarehouseID, request.PickingStrategyID)
		if err != nil {
			return err
		}
		kind, err := local.documentType(ctx, "RESERVATION")
		if err != nil {
			return err
		}
		active, err := local.initialStatus(ctx, kind.ID)
		if err != nil {
			return err
		}
		lines, err := local.repositories.Line.List(ctx, id)
		if err != nil {
			return err
		}
		warehouse, err := local.repositories.Master.Warehouse.FindByID(ctx, header.WarehouseID)
		if err != nil {
			return err
		}
		for _, line := range lines {
			_, ordered, err := quantity(line.OrderedQty, "ordered_qty", false)
			if err != nil {
				return err
			}
			_, already, err := quantity(line.AllocatedQty, "allocated_qty", true)
			if err != nil {
				return err
			}
			remaining := new(big.Rat).Sub(ordered, already)
			requestedTotal.Add(requestedTotal, remaining)
			if remaining.Sign() == 0 {
				continue
			}
			candidates, err := local.repositories.Reservation.Candidates(ctx, line.ID, &strategy.ID)
			if err != nil {
				return err
			}
			for _, candidate := range candidates {
				if remaining.Sign() == 0 {
					break
				}
				identity := inventoryrepository.BalanceIdentity{OwnerID: candidate.OwnerID, WarehouseID: candidate.WarehouseID, LocationID: candidate.LocationID, ItemID: candidate.ItemID, LotID: candidate.LotID, HandlingUnitID: candidate.HandlingUnitID, InventoryStatusID: candidate.InventoryStatusID}
				balance, err := local.repositories.Inventory.Balance.FindLocked(ctx, identity)
				if err != nil {
					return err
				}
				_, onHand, err := quantity(balance.OnHandQty, "on_hand_qty", true)
				if err != nil {
					return err
				}
				_, reserved, err := quantity(balance.ReservedQty, "reserved_qty", true)
				if err != nil {
					return err
				}
				available := new(big.Rat).Sub(onHand, reserved)
				if available.Sign() <= 0 {
					continue
				}
				take := new(big.Rat).Set(available)
				if take.Cmp(remaining) > 0 {
					if balance.HandlingUnitID != nil {
						continue
					}
					take.Set(remaining)
				}
				qty := decimal(take)
				newReserved := decimal(new(big.Rat).Add(reserved, take))
				if err := local.repositories.Inventory.Balance.SetQuantities(ctx, balance.ID, balance.OnHandQty, newReserved); err != nil {
					return err
				}
				reservationID, err := local.generateID(ctx, "RESERVATION", header.BusinessDate, nil, &warehouse.ID)
				if err != nil {
					return err
				}
				reservation := model.InventoryReservation{ID: reservationID, DocumentTypeID: kind.ID, StatusID: active.ID, PickingStrategyID: &strategy.ID, OutboundLineID: line.ID, BalanceID: balance.ID, ReservedQty: qty, PickedQty: "0", UOMID: line.UOMID, CreatedBy: actor}
				if err := local.repositories.Reservation.Create(ctx, &reservation); err != nil {
					return err
				}
				if err := local.repositories.Line.AddAllocated(ctx, line.ID, qty); err != nil {
					return err
				}
				remaining.Sub(remaining, take)
				allocatedTotal.Add(allocatedTotal, take)
			}
		}
		shortage := new(big.Rat).Sub(requestedTotal, allocatedTotal)
		if shortage.Sign() > 0 && !request.AllowPartial {
			return state("insufficient allocatable inventory; retry with allow_partial=true to reserve available stock")
		}
		all, err := local.repositories.Line.AllAllocated(ctx, id)
		if err != nil {
			return err
		}
		nextCode := "PARTIALLY_ALLOCATED"
		if all {
			nextCode = "ALLOCATED"
		}
		var targetID string
		if row.StatusCode == nextCode {
			targetID = header.StatusID
		} else {
			target, err := local.transition(ctx, header.DocumentTypeID, header.StatusID, nextCode)
			if err != nil {
				return err
			}
			targetID = target.ID
		}
		return local.repositories.Order.TouchAllocated(ctx, id, targetID, actor, all)
	})
	if err != nil {
		return dto.AllocationResponse{}, err
	}
	order, err := s.GetOutboundOrder(ctx, id)
	if err != nil {
		return dto.AllocationResponse{}, err
	}
	rows, err := s.repositories.Reservation.ListByOrder(ctx, id)
	if err != nil {
		return dto.AllocationResponse{}, err
	}
	reservations := make([]dto.ReservationResponse, 0, len(rows))
	for _, v := range rows {
		reservations = append(reservations, mapReservation(v))
	}
	shortage := new(big.Rat).Sub(requestedTotal, allocatedTotal)
	return dto.AllocationResponse{Order: order, Reservations: reservations, RequestedQty: decimal(requestedTotal), AllocatedQty: decimal(allocatedTotal), ShortageQty: decimal(shortage)}, nil
}

func (s *Service) ListReservations(ctx context.Context, f repository.ListFilter) (dto.PageResponse[dto.ReservationResponse], error) {
	if !validUUID(f.OwnerID) {
		return dto.PageResponse[dto.ReservationResponse]{}, invalid("owner_id is required and must be a UUID")
	}
	rows, total, err := s.repositories.Reservation.List(ctx, f)
	if err != nil {
		return dto.PageResponse[dto.ReservationResponse]{}, err
	}
	items := make([]dto.ReservationResponse, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapReservation(v))
	}
	return page(items, f.Page, f.PageSize, total), nil
}

func (s *Service) ReleaseReservation(ctx context.Context, id string, request dto.ReleaseReservationRequest, actor string) (dto.OutboundOrderResponse, error) {
	if !validID(id, 140) || !validUUID(actor) {
		return dto.OutboundOrderResponse{}, invalid("invalid reservation release request")
	}
	if _, err := clean(request.Reason, 4000, "reason"); err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	var outboundID string
	err := s.transaction(ctx, func(local *Service) error {
		reservation, err := local.repositories.Reservation.Lock(ctx, id)
		if err != nil {
			return err
		}
		exists, err := local.repositories.PickTask.ExistsForReservation(ctx, id)
		if err != nil {
			return err
		}
		if exists {
			return state("reservation already belongs to a wave pick task")
		}
		line, err := local.repositories.Line.Get(ctx, reservation.OutboundLineID)
		if err != nil {
			return err
		}
		order, err := local.repositories.Order.Lock(ctx, line.OutboundID)
		if err != nil {
			return err
		}
		outboundID = order.ID
		row, err := local.repositories.Order.Get(ctx, order.ID)
		if err != nil {
			return err
		}
		if row.StatusCode != "PARTIALLY_ALLOCATED" && row.StatusCode != "ALLOCATED" {
			return state("reservation cannot be released in the current outbound state")
		}
		statusRow, err := local.status(ctx, reservation.DocumentTypeID, "ACTIVE")
		if err != nil {
			return err
		}
		if reservation.StatusID != statusRow.ID {
			return state("only an ACTIVE reservation can be released")
		}
		_, reservedQty, err := quantity(reservation.ReservedQty, "reserved_qty", false)
		if err != nil {
			return err
		}
		_, pickedQty, err := quantity(reservation.PickedQty, "picked_qty", true)
		if err != nil {
			return err
		}
		releaseQty := new(big.Rat).Sub(reservedQty, pickedQty)
		balanceRow, err := local.repositories.Inventory.Balance.Get(ctx, reservation.BalanceID)
		if err != nil {
			return err
		}
		identity := inventoryrepository.BalanceIdentity{OwnerID: balanceRow.OwnerID, WarehouseID: balanceRow.WarehouseID, LocationID: balanceRow.LocationID, ItemID: balanceRow.ItemID, LotID: balanceRow.LotID, HandlingUnitID: balanceRow.HandlingUnitID, InventoryStatusID: balanceRow.InventoryStatusID}
		balance, err := local.repositories.Inventory.Balance.FindLocked(ctx, identity)
		if err != nil {
			return err
		}
		_, balanceReserved, err := quantity(balance.ReservedQty, "balance reserved_qty", true)
		if err != nil || balanceReserved.Cmp(releaseQty) < 0 {
			return state("reserved inventory balance is inconsistent")
		}
		if err := local.repositories.Inventory.Balance.SetQuantities(ctx, balance.ID, balance.OnHandQty, decimal(new(big.Rat).Sub(balanceReserved, releaseQty))); err != nil {
			return err
		}
		released, err := local.transition(ctx, reservation.DocumentTypeID, reservation.StatusID, "RELEASED")
		if err != nil {
			return err
		}
		if err := local.repositories.Reservation.Release(ctx, id, released.ID); err != nil {
			return err
		}
		if err := local.repositories.Line.SubtractAllocated(ctx, line.ID, decimal(releaseQty)); err != nil {
			return err
		}
		has, err := local.repositories.Line.HasAllocation(ctx, order.ID)
		if err != nil {
			return err
		}
		nextCode := "RELEASED"
		if has {
			nextCode = "PARTIALLY_ALLOCATED"
		}
		target, err := local.status(ctx, order.DocumentTypeID, nextCode)
		if err != nil {
			return err
		}
		return local.repositories.Order.SetStatusLocked(ctx, order.ID, target.ID, actor, nil)
	})
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	return s.GetOutboundOrder(ctx, outboundID)
}
