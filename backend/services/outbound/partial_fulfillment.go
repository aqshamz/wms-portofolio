package outbound

import (
	"context"
	"math/big"
	"time"
	dto "wms-api/dto/outbound"
	model "wms-api/models/outbound"
	repository "wms-api/repository/outbound"
)

// Demand is always the frozen base quantity. Accepted check shortages and an
// authorized demand closure are separate audit concepts; neither is a shipment.
func openDemand(line model.OutboundOrderLine) *big.Rat {
	remaining := progressQuantity(line.OrderedBaseQty)
	for _, qty := range []string{line.ShippedQty, line.ShortAcceptedQty, line.ClosedShortQty} {
		remaining.Sub(remaining, progressQuantity(qty))
	}
	return remaining
}
func unallocatedDemand(line model.OutboundOrderLine) *big.Rat {
	remaining := new(big.Rat).Sub(progressQuantity(line.OrderedBaseQty), progressQuantity(line.AllocatedQty))
	return remaining.Sub(remaining, progressQuantity(line.ClosedShortQty))
}
func progressQuantity(value string) *big.Rat {
	if value == "" {
		return new(big.Rat)
	}
	return mustRat(value)
}
func hasOpenDemand(lines []repository.OutboundOrderLineRow) bool {
	for _, line := range lines {
		if openDemand(line.OutboundOrderLine).Sign() > 0 {
			return true
		}
	}
	return false
}

func (s *Service) refreshDeliveryOutcome(ctx context.Context, id, actor string, at time.Time) error {
	order, err := s.repositories.Order.Lock(ctx, id)
	if err != nil {
		return err
	}
	row, err := s.repositories.Order.Get(ctx, id)
	if err != nil {
		return err
	}
	// Never allow an older delivery to replace a backorder or current warehouse
	// execution state. CLOSED_SHORT closes demand, not its shipment/POD records.
	if row.StatusCode != "SHIPPED" && row.StatusCode != "DELIVERY_FAILED" {
		return nil
	}
	lines, err := s.repositories.Line.List(ctx, id)
	if err != nil {
		return err
	}
	if hasOpenDemand(lines) {
		return nil
	}
	pending, failed, err := s.repositories.Delivery.OutstandingByOrder(ctx, id)
	if err != nil {
		return err
	}
	nextCode := "SHIPPED"
	extra := map[string]interface{}{}
	if failed > 0 {
		nextCode = "DELIVERY_FAILED"
	} else if pending == 0 {
		shipped, delivered := new(big.Rat), new(big.Rat)
		for _, line := range lines {
			shipped.Add(shipped, mustRat(line.ShippedQty))
			delivered.Add(delivered, mustRat(line.DeliveredQty))
		}
		nextCode = "RETURNED"
		if delivered.Cmp(shipped) == 0 {
			nextCode = "DELIVERED"
		} else if delivered.Sign() > 0 {
			nextCode = "PARTIALLY_DELIVERED"
		}
		extra["completed_at"] = at
	}
	if nextCode == row.StatusCode {
		return nil
	}
	next, err := s.transition(ctx, order.DocumentTypeID, order.StatusID, nextCode)
	if err != nil {
		return err
	}
	return s.repositories.Order.SetStatusLocked(ctx, id, next.ID, actor, extra)
}

func (s *Service) CloseShortOutboundOrder(ctx context.Context, id string, q dto.CloseShortOutboundRequest, actor string) (dto.OutboundOrderResponse, error) {
	if !validID(id, 120) || !validUUID(actor) || q.ExpectedVersion < 1 {
		return dto.OutboundOrderResponse{}, invalid("invalid close-short request")
	}
	reason, err := clean(q.Reason, 4000, "reason")
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		order, err := local.repositories.Order.Lock(ctx, id)
		if err != nil {
			return err
		}
		if order.VersionNo != q.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		row, err := local.repositories.Order.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "RELEASED" && row.StatusCode != "PARTIALLY_FULFILLED" {
			return state("close-short requires released or partially fulfilled demand; finish or release in-process stock first")
		}
		active, err := local.repositories.WaveOrder.HasActiveWave(ctx, id)
		if err != nil {
			return err
		}
		if active {
			return state("cancel the active draft wave before closing demand")
		}
		lines, err := local.repositories.Line.List(ctx, id)
		if err != nil {
			return err
		}
		remaining := new(big.Rat)
		for _, line := range lines {
			accounted := new(big.Rat).Add(mustRat(line.ShippedQty), mustRat(line.ShortAcceptedQty))
			if mustRat(line.AllocatedQty).Cmp(accounted) != 0 {
				return state("reserved, picked or in-process quantities cannot be closed short")
			}
			remaining.Add(remaining, unallocatedDemand(line.OutboundOrderLine))
		}
		if remaining.Sign() <= 0 {
			return state("order has no outstanding demand to close")
		}
		if err := local.repositories.Line.CloseRemaining(ctx, id); err != nil {
			return err
		}
		next, err := local.transition(ctx, order.DocumentTypeID, order.StatusID, "CLOSED_SHORT")
		if err != nil {
			return err
		}
		now := time.Now()
		return local.repositories.Order.SetStatusLocked(ctx, id, next.ID, actor, map[string]interface{}{"closed_short_at": now, "closed_short_by": actor, "closed_short_reason": reason, "completed_at": now})
	})
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	return s.GetOutboundOrder(ctx, id)
}
