package outbound

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
	dto "wms-api/dto/outbound"
	inventorymodel "wms-api/models/inventory"
	model "wms-api/models/outbound"
	inventoryrepository "wms-api/repository/inventory"
	repository "wms-api/repository/outbound"
)

func (s *Service) GetDelivery(ctx context.Context, id string) (dto.DeliveryResponse, error) {
	if !validID(id, 120) {
		return dto.DeliveryResponse{}, invalid("invalid delivery_id")
	}
	row, err := s.repositories.Delivery.Get(ctx, id)
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	out := mapDelivery(row)
	lines, err := s.repositories.DeliveryLine.List(ctx, id)
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	out.Lines = make([]dto.DeliveryLineResponse, 0, len(lines))
	for _, v := range lines {
		out.Lines = append(out.Lines, mapDeliveryLine(v))
	}
	events, err := s.repositories.DeliveryEvent.List(ctx, id)
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	out.Events = make([]dto.DeliveryEventResponse, 0, len(events))
	for _, v := range events {
		out.Events = append(out.Events, mapDeliveryEvent(v))
	}
	return out, nil
}
func (s *Service) ListDeliveries(ctx context.Context, f repository.ListFilter) (dto.PageResponse[dto.DeliveryResponse], error) {
	if !validUUID(f.OwnerID) {
		return dto.PageResponse[dto.DeliveryResponse]{}, invalid("owner_id is required and must be a UUID")
	}
	rows, total, err := s.repositories.Delivery.List(ctx, f)
	if err != nil {
		return dto.PageResponse[dto.DeliveryResponse]{}, err
	}
	items := make([]dto.DeliveryResponse, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapDelivery(v))
	}
	return page(items, f.Page, f.PageSize, total), nil
}
func (s *Service) CreateDelivery(ctx context.Context, q dto.CreateDeliveryRequest, actor string) (dto.DeliveryResponse, error) {
	if !validID(q.ShipmentID, 120) || !validID(q.OutboundID, 120) || !validUUID(actor) {
		return dto.DeliveryResponse{}, invalid("invalid delivery request")
	}
	businessDate, err := date(q.BusinessDate, "business_date")
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	planned, err := optionalTimestamp(q.PlannedDeliveryAt, "planned_delivery_at")
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	notes, err := optional(q.Notes, 4000, "notes")
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	var id string
	err = s.transaction(ctx, func(local *Service) error {
		shipment, err := local.repositories.Shipment.Get(ctx, q.ShipmentID)
		if err != nil {
			return err
		}
		if shipment.StatusCode != "SHIPPED" {
			return state("delivery requires a SHIPPED shipment")
		}
		members, err := local.repositories.ShipmentOrder.List(ctx, q.ShipmentID)
		if err != nil {
			return err
		}
		member := false
		for _, v := range members {
			if v.OutboundID == q.OutboundID {
				member = true
				break
			}
		}
		if !member {
			return invalid("outbound order is not an active shipment member")
		}
		sources, err := local.repositories.DeliveryLine.SourceLines(ctx, q.ShipmentID, q.OutboundID)
		if err != nil {
			return err
		}
		if len(sources) == 0 {
			return state("shipment has no dispatched lines for this order")
		}
		kind, initial, err := local.document(ctx, "DELIVERY")
		if err != nil {
			return err
		}
		id, err = local.generateID(ctx, "DELIVERY", businessDate, nil, &shipment.WarehouseID)
		if err != nil {
			return err
		}
		delivery := model.Delivery{ID: id, DocumentTypeID: kind.ID, StatusID: initial.ID, ShipmentID: q.ShipmentID, OutboundID: q.OutboundID, BusinessDate: businessDate, PlannedDeliveryAt: planned, Notes: notes, CreatedBy: actor, UpdatedBy: actor}
		if err := local.repositories.Delivery.Create(ctx, &delivery); err != nil {
			return err
		}
		lines := make([]model.DeliveryLine, 0, len(sources))
		for i, v := range sources {
			lines = append(lines, model.DeliveryLine{ID: id + "-L" + pad4(i+1), DeliveryID: id, ShipmentLineID: v.ShipmentLineID, PlannedQty: v.PlannedQty, DeliveredQty: "0.000000", ReturnedQty: "0.000000", UOMID: v.UOMID, CreatedBy: actor})
		}
		return local.repositories.DeliveryLine.CreateBatch(ctx, lines)
	})
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	return s.GetDelivery(ctx, id)
}
func pad4(v int) string { return fmt.Sprintf("%04d", v) }
func eventCoordinates(lat, lon *string) error {
	if (lat == nil) != (lon == nil) {
		return invalid("latitude and longitude must be provided together")
	}
	return nil
}
func (s *Service) recordDeliveryPosition(ctx context.Context, id, eventCode string, q dto.DeliveryEventRequest, actor string) (dto.DeliveryResponse, error) {
	if !validID(id, 120) || !validUUID(actor) {
		return dto.DeliveryResponse{}, invalid("invalid delivery event")
	}
	eventAt, err := timestamp(q.EventAt, "event_at")
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	if err := eventCoordinates(q.Latitude, q.Longitude); err != nil {
		return dto.DeliveryResponse{}, err
	}
	notes, err := optional(q.Notes, 4000, "notes")
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		delivery, err := local.repositories.Delivery.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.Delivery.Get(ctx, id)
		if err != nil {
			return err
		}
		target := ""
		if eventCode == "DEPARTED" && (row.StatusCode == "PLANNED" || row.StatusCode == "FAILED") {
			target = "IN_TRANSIT"
		} else if eventCode == "ARRIVED_STORE" && row.StatusCode == "IN_TRANSIT" {
			target = "ARRIVED"
		} else {
			return state("delivery event is not valid for the current status")
		}
		kind, err := local.repositories.DeliveryEventType.ByCode(ctx, eventCode)
		if err != nil {
			return err
		}
		eventID, err := randomID(id + "-E")
		if err != nil {
			return err
		}
		event := model.DeliveryEvent{ID: eventID, DeliveryID: id, DeliveryEventTypeID: kind.ID, EventAt: eventAt, Notes: notes, Latitude: q.Latitude, Longitude: q.Longitude, RecordedBy: actor}
		if err := local.repositories.DeliveryEvent.Create(ctx, &event); err != nil {
			return err
		}
		next, err := local.transition(ctx, delivery.DocumentTypeID, delivery.StatusID, target)
		if err != nil {
			return err
		}
		updates := map[string]interface{}{"status_id": next.ID, "updated_at": time.Now(), "updated_by": actor}
		if target == "ARRIVED" {
			updates["arrived_at"] = eventAt
		}
		return local.repositories.Delivery.Update(ctx, id, updates)
	})
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	return s.GetDelivery(ctx, id)
}
func (s *Service) DepartDelivery(ctx context.Context, id string, q dto.DeliveryEventRequest, actor string) (dto.DeliveryResponse, error) {
	return s.recordDeliveryPosition(ctx, id, "DEPARTED", q, actor)
}
func (s *Service) ArriveDelivery(ctx context.Context, id string, q dto.DeliveryEventRequest, actor string) (dto.DeliveryResponse, error) {
	return s.recordDeliveryPosition(ctx, id, "ARRIVED_STORE", q, actor)
}
func (s *Service) DeliverLine(ctx context.Context, deliveryID, lineID string, q dto.DeliverLineRequest, actor string) (dto.DeliveryResponse, error) {
	if !validID(deliveryID, 120) || !validID(lineID, 160) || !validUUID(actor) {
		return dto.DeliveryResponse{}, invalid("invalid proof-of-delivery request")
	}
	qty, n, err := quantity(q.DeliveredQty, "delivered_qty", false)
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	eventAt, err := timestamp(q.EventAt, "event_at")
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	recipient, err := clean(q.RecipientName, 150, "recipient_name")
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	proof, err := clean(q.ProofReference, 200, "proof_reference")
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	recipientRef, err := optional(q.RecipientReference, 100, "recipient_reference")
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	proofURI, err := optional(q.ProofURI, 2000, "proof_uri")
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	notes, err := optional(q.Notes, 4000, "notes")
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	if err := eventCoordinates(q.Latitude, q.Longitude); err != nil {
		return dto.DeliveryResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		delivery, err := local.repositories.Delivery.Lock(ctx, deliveryID)
		if err != nil {
			return err
		}
		row, err := local.repositories.Delivery.Get(ctx, deliveryID)
		if err != nil {
			return err
		}
		if row.StatusCode != "ARRIVED" {
			return state("proof of delivery requires ARRIVED status")
		}
		line, err := local.repositories.DeliveryLine.Lock(ctx, lineID)
		if err != nil {
			return err
		}
		if line.DeliveryID != delivery.ID {
			return invalid("delivery line does not belong to delivery")
		}
		remaining := new(big.Rat).Sub(mustRat(line.PlannedQty), new(big.Rat).Add(mustRat(line.DeliveredQty), mustRat(line.ReturnedQty)))
		if n.Cmp(remaining) > 0 {
			return invalid("delivered_qty exceeds line remainder")
		}
		kind, err := local.repositories.DeliveryEventType.ByCode(ctx, "DELIVERED")
		if err != nil {
			return err
		}
		eventID, err := randomID(deliveryID + "-E")
		if err != nil {
			return err
		}
		event := model.DeliveryEvent{ID: eventID, DeliveryID: deliveryID, DeliveryEventTypeID: kind.ID, EventAt: eventAt, RecipientName: &recipient, RecipientReference: recipientRef, ProofReference: &proof, ProofURI: proofURI, Notes: notes, Latitude: q.Latitude, Longitude: q.Longitude, RecordedBy: actor}
		if err := local.repositories.DeliveryEvent.Create(ctx, &event); err != nil {
			return err
		}
		if err := local.repositories.DeliveryEventLine.Create(ctx, &model.DeliveryEventLine{DeliveryEventID: eventID, DeliveryLineID: lineID, DeliveredQty: qty}); err != nil {
			return err
		}
		if err := local.repositories.DeliveryLine.AddDelivered(ctx, lineID, qty); err != nil {
			return err
		}
		lineRow, err := local.repositories.DeliveryLine.Get(ctx, lineID)
		if err != nil {
			return err
		}
		if err := local.repositories.Line.AddDelivered(ctx, lineRow.OutboundLineID, qty); err != nil {
			return err
		}
		return local.repositories.Delivery.Update(ctx, deliveryID, map[string]interface{}{"recipient_name": recipient, "recipient_reference": recipientRef, "proof_reference": proof, "proof_uri": proofURI, "latitude": q.Latitude, "longitude": q.Longitude, "updated_at": time.Now(), "updated_by": actor})
	})
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	return s.GetDelivery(ctx, deliveryID)
}
func (s *Service) CompleteDelivery(ctx context.Context, id string, q dto.CompleteDeliveryRequest, actor string) (dto.DeliveryResponse, error) {
	if !validID(id, 120) || !validUUID(actor) {
		return dto.DeliveryResponse{}, invalid("invalid delivery completion")
	}
	deliveredAt, err := timestamp(q.DeliveredAt, "delivered_at")
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		delivery, err := local.repositories.Delivery.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.Delivery.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "ARRIVED" && row.StatusCode != "FAILED" {
			return state("delivery must be ARRIVED or FAILED before completion")
		}
		if row.RecipientName == nil || row.ProofReference == nil {
			return state("recipient and proof are required")
		}
		remaining, err := local.repositories.DeliveryLine.IncompleteCount(ctx, id)
		if err != nil {
			return err
		}
		if remaining > 0 {
			return state("every delivery line must be fully delivered")
		}
		next, err := local.transition(ctx, delivery.DocumentTypeID, delivery.StatusID, "DELIVERED")
		if err != nil {
			return err
		}
		if err := local.repositories.Delivery.Update(ctx, id, map[string]interface{}{"status_id": next.ID, "delivered_at": deliveredAt, "updated_at": time.Now(), "updated_by": actor}); err != nil {
			return err
		}
		order, err := local.repositories.Order.Lock(ctx, delivery.OutboundID)
		if err != nil {
			return err
		}
		orderNext, err := local.transition(ctx, order.DocumentTypeID, order.StatusID, "DELIVERED")
		if err != nil {
			return err
		}
		return local.repositories.Order.SetStatusLocked(ctx, order.ID, orderNext.ID, actor, map[string]interface{}{"completed_at": deliveredAt})
	})
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	return s.GetDelivery(ctx, id)
}
func (s *Service) FailDelivery(ctx context.Context, id string, q dto.FailDeliveryRequest, actor string) (dto.DeliveryResponse, error) {
	if !validID(id, 120) || !validUUID(actor) {
		return dto.DeliveryResponse{}, invalid("invalid delivery failure")
	}
	eventAt, err := timestamp(q.EventAt, "event_at")
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	if err := eventCoordinates(q.Latitude, q.Longitude); err != nil {
		return dto.DeliveryResponse{}, err
	}
	notes, err := optional(q.Notes, 4000, "notes")
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	code := strings.ToUpper(strings.TrimSpace(q.ReasonCode))
	err = s.transaction(ctx, func(local *Service) error {
		delivery, err := local.repositories.Delivery.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.Delivery.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "IN_TRANSIT" && row.StatusCode != "ARRIVED" {
			return state("only an IN_TRANSIT or ARRIVED delivery can fail")
		}
		remaining, err := local.repositories.DeliveryLine.UnaccountedCount(ctx, id)
		if err != nil {
			return err
		}
		if remaining == 0 {
			return state("delivery has no remaining quantity to fail")
		}
		reason, err := local.repositories.DeliveryFailureReason.ByCode(ctx, code)
		if err != nil {
			return invalid("unknown or inactive failure reason")
		}
		kind, err := local.repositories.DeliveryEventType.ByCode(ctx, "DELIVERY_FAILED")
		if err != nil {
			return err
		}
		eventID, err := randomID(id + "-E")
		if err != nil {
			return err
		}
		if err := local.repositories.DeliveryEvent.Create(ctx, &model.DeliveryEvent{ID: eventID, DeliveryID: id, DeliveryEventTypeID: kind.ID, EventAt: eventAt, DeliveryFailureReasonID: &reason.ID, Notes: notes, Latitude: q.Latitude, Longitude: q.Longitude, RecordedBy: actor}); err != nil {
			return err
		}
		next, err := local.transition(ctx, delivery.DocumentTypeID, delivery.StatusID, "FAILED")
		if err != nil {
			return err
		}
		if err := local.repositories.Delivery.Update(ctx, id, map[string]interface{}{"status_id": next.ID, "updated_at": time.Now(), "updated_by": actor}); err != nil {
			return err
		}
		order, err := local.repositories.Order.Lock(ctx, delivery.OutboundID)
		if err != nil {
			return err
		}
		orderNext, err := local.transition(ctx, order.DocumentTypeID, order.StatusID, "DELIVERY_FAILED")
		if err != nil {
			return err
		}
		return local.repositories.Order.SetStatusLocked(ctx, order.ID, orderNext.ID, actor, nil)
	})
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	return s.GetDelivery(ctx, id)
}
func (s *Service) UpsertReturnPolicy(ctx context.Context, q dto.UpsertReturnPolicyRequest, actor string) (dto.ReturnPolicyResponse, error) {
	if !validUUID(q.OwnerID) || !validUUID(q.WarehouseID) || !validUUID(q.ReturnLocationID) || !validUUID(q.ReturnInventoryStatusID) || !validUUID(actor) {
		return dto.ReturnPolicyResponse{}, invalid("invalid return policy")
	}
	valid, err := s.repositories.ReturnPolicy.Validate(ctx, q.OwnerID, q.WarehouseID, q.ReturnLocationID, q.ReturnInventoryStatusID)
	if err != nil {
		return dto.ReturnPolicyResponse{}, err
	}
	if !valid {
		return dto.ReturnPolicyResponse{}, invalid("return policy requires an active warehouse-owner location and a non-allocatable status")
	}
	v := model.OutboundReturnPolicy{OwnerID: q.OwnerID, WarehouseID: q.WarehouseID, ReturnLocationID: q.ReturnLocationID, ReturnInventoryStatusID: q.ReturnInventoryStatusID, IsActive: q.IsActive, CreatedBy: actor, UpdatedBy: &actor}
	if err := s.repositories.ReturnPolicy.Upsert(ctx, &v); err != nil {
		return dto.ReturnPolicyResponse{}, err
	}
	return s.GetReturnPolicy(ctx, q.OwnerID, q.WarehouseID)
}
func (s *Service) GetReturnPolicy(ctx context.Context, owner, warehouse string) (dto.ReturnPolicyResponse, error) {
	if !validUUID(owner) || !validUUID(warehouse) {
		return dto.ReturnPolicyResponse{}, invalid("owner_id and warehouse_id are required UUIDs")
	}
	v, err := s.repositories.ReturnPolicy.Get(ctx, owner, warehouse)
	if err != nil {
		return dto.ReturnPolicyResponse{}, err
	}
	return dto.ReturnPolicyResponse{OwnerID: v.OwnerID, WarehouseID: v.WarehouseID, ReturnLocationID: v.ReturnLocationID, ReturnLocationCode: v.ReturnLocationCode, ReturnInventoryStatusID: v.ReturnInventoryStatusID, ReturnInventoryStatusCode: v.ReturnInventoryStatusCode, IsActive: v.IsActive}, nil
}
func (s *Service) ReturnDeliveryLine(ctx context.Context, deliveryID, lineID string, q dto.ReturnDeliveryLineRequest, actor string) (dto.ReturnLineResponse, error) {
	if !validID(deliveryID, 120) || !validID(lineID, 160) || !validUUID(actor) {
		return dto.ReturnLineResponse{}, invalid("invalid delivery return request")
	}
	qty, n, err := quantity(q.ReturnedQty, "returned_qty", false)
	if err != nil {
		return dto.ReturnLineResponse{}, err
	}
	returnedAt, err := timestamp(q.EventAt, "event_at")
	if err != nil {
		return dto.ReturnLineResponse{}, err
	}
	operationKey, err := clean(q.OperationKey, 160, "operation_key")
	if err != nil {
		return dto.ReturnLineResponse{}, err
	}
	notes, err := optional(q.Notes, 4000, "notes")
	if err != nil {
		return dto.ReturnLineResponse{}, err
	}
	fp := fingerprint(deliveryID, lineID, qty, q.EventAt)
	if existing, e := s.repositories.Inventory.Movement.GetByOperationKey(ctx, operationKey); e == nil {
		if existing.OperationFingerprint == nil || *existing.OperationFingerprint != fp {
			return dto.ReturnLineResponse{}, repository.ErrConflict
		}
		ret, err := s.repositories.DeliveryReturnLine.GetByMovement(ctx, existing.ID)
		if err != nil {
			return dto.ReturnLineResponse{}, err
		}
		return dto.ReturnLineResponse{ID: ret.ID, DeliveryLineID: ret.DeliveryLineID, ReturnedBalanceID: ret.ReturnedBalanceID, ReturnedQty: ret.ReturnedQty, MovementID: ret.MovementID}, nil
	} else if !errors.Is(e, inventoryrepository.ErrNotFound) {
		return dto.ReturnLineResponse{}, e
	}
	var response dto.ReturnLineResponse
	err = s.transaction(ctx, func(local *Service) error {
		delivery, err := local.repositories.Delivery.Lock(ctx, deliveryID)
		if err != nil {
			return err
		}
		row, err := local.repositories.Delivery.Get(ctx, deliveryID)
		if err != nil {
			return err
		}
		if row.StatusCode != "FAILED" {
			return state("return to depot requires FAILED delivery status")
		}
		line, err := local.repositories.DeliveryLine.Lock(ctx, lineID)
		if err != nil {
			return err
		}
		if line.DeliveryID != deliveryID {
			return invalid("delivery line does not belong to delivery")
		}
		remaining := new(big.Rat).Sub(mustRat(line.PlannedQty), new(big.Rat).Add(mustRat(line.DeliveredQty), mustRat(line.ReturnedQty)))
		if n.Cmp(remaining) > 0 {
			return invalid("returned_qty exceeds undelivered remainder")
		}
		policy, err := local.repositories.ReturnPolicy.Get(ctx, row.OwnerID, row.WarehouseID)
		if err != nil {
			return state("active outbound return policy is required")
		}
		if !policy.IsActive || policy.StatusAllocatable {
			return state("return policy must use a non-allocatable status")
		}
		lineRow, err := local.repositories.DeliveryLine.Get(ctx, lineID)
		if err != nil {
			return err
		}
		source, err := local.repositories.Inventory.Balance.Get(ctx, lineRow.SourceBalanceID)
		if err != nil {
			return err
		}
		if source.HandlingUnitID != nil && (n.Cmp(remaining) != 0 || mustRat(line.DeliveredQty).Sign() != 0 || mustRat(line.ReturnedQty).Sign() != 0) {
			return state("handling-unit delivery stock must return as one complete line")
		}
		identity := inventoryrepository.BalanceIdentity{OwnerID: row.OwnerID, WarehouseID: row.WarehouseID, LocationID: policy.ReturnLocationID, ItemID: source.ItemID, LotID: source.LotID, HandlingUnitID: source.HandlingUnitID, InventoryStatusID: policy.ReturnInventoryStatusID}
		target, err := local.repositories.Inventory.Balance.FindLocked(ctx, identity)
		balanceID := ""
		if err != nil {
			if !errors.Is(err, inventoryrepository.ErrNotFound) {
				return err
			}
			balanceID, err = local.generateID(ctx, "INVENTORY_BALANCE", delivery.BusinessDate, nil, &row.WarehouseID)
			if err != nil {
				return err
			}
			target = inventorymodel.InventoryBalance{ID: balanceID, OwnerID: row.OwnerID, WarehouseID: row.WarehouseID, LocationID: policy.ReturnLocationID, ItemID: source.ItemID, LotID: source.LotID, HandlingUnitID: source.HandlingUnitID, InventoryStatusID: policy.ReturnInventoryStatusID, OnHandQty: qty, ReservedQty: "0.000000", UOMID: line.UOMID, VersionNo: 1}
			if err := local.repositories.Inventory.Balance.Create(ctx, &target); err != nil {
				return err
			}
		} else {
			balanceID = target.ID
			if err := local.repositories.Inventory.Balance.SetQuantities(ctx, balanceID, decimal(new(big.Rat).Add(mustRat(target.OnHandQty), n)), target.ReservedQty); err != nil {
				return err
			}
		}
		shipmentMovement, err := local.repositories.Inventory.Movement.Get(ctx, lineRow.ShipmentMovementID)
		if err != nil {
			return err
		}
		if shipmentMovement.SerialID != nil {
			if n.Cmp(big.NewRat(1, 1)) != 0 {
				return state("serialized delivery returns require quantity 1")
			}
			if err := local.repositories.Inventory.SerialState.Create(ctx, &inventorymodel.SerialInventory{SerialID: *shipmentMovement.SerialID, BalanceID: balanceID, OwnerID: row.OwnerID, ItemID: source.ItemID}); err != nil {
				return err
			}
		}
		if source.HandlingUnitID != nil {
			hu, err := local.repositories.Inventory.HandlingUnit.GetShared(ctx, *source.HandlingUnitID)
			if err != nil {
				return err
			}
			if hu.OwnerID != row.OwnerID || hu.WarehouseID != row.WarehouseID || hu.IsClosed {
				return state("returned handling unit is invalid or closed")
			}
			if hu.CurrentLocationID == nil {
				if err := local.repositories.Inventory.HandlingUnit.Place(ctx, hu.ID, policy.ReturnLocationID); err != nil {
					return err
				}
			} else if *hu.CurrentLocationID != policy.ReturnLocationID {
				return state("returned handling unit is already located elsewhere")
			}
		}
		kind, err := local.repositories.Inventory.MovementType.ByCodeShared(ctx, "DELIVERY_RETURN")
		if err != nil || !kind.IsActive {
			return state("DELIVERY_RETURN movement type is not configured")
		}
		movementID, err := local.generateID(ctx, "MOVEMENT", delivery.BusinessDate, nil, &row.WarehouseID)
		if err != nil {
			return err
		}
		to := policy.ReturnLocationID
		toStatus := policy.ReturnInventoryStatusID
		reason, err := local.repositories.Master.ReasonCode.ByModuleCode(ctx, "OUTBOUND", "DELIVERY_RETURN")
		if err != nil {
			return err
		}
		movement := inventorymodel.InventoryMovement{ID: movementID, MovementTypeID: kind.ID, OwnerID: row.OwnerID, WarehouseID: row.WarehouseID, BusinessDate: delivery.BusinessDate, OccurredAt: returnedAt, ItemID: source.ItemID, LotID: source.LotID, SerialID: shipmentMovement.SerialID, HandlingUnitID: source.HandlingUnitID, ToLocationID: &to, ToStatusID: &toStatus, Quantity: qty, UOMID: line.UOMID, SourceDocumentID: deliveryID, SourceLineID: &lineID, ReasonCodeID: &reason.ID, Notes: notes, OperationKey: &operationKey, OperationFingerprint: &fp, CreatedBy: actor}
		if err := local.repositories.Inventory.Movement.Create(ctx, &movement); err != nil {
			return err
		}
		returnID, err := randomID(deliveryID + "-RET")
		if err != nil {
			return err
		}
		ret := model.DeliveryReturnLine{ID: returnID, DeliveryID: deliveryID, DeliveryLineID: lineID, ReturnLocationID: policy.ReturnLocationID, ReturnedBalanceID: balanceID, ReturnedQty: qty, UOMID: line.UOMID, MovementID: movementID, ReturnedAt: returnedAt, ReturnedBy: actor}
		if err := local.repositories.DeliveryReturnLine.Create(ctx, &ret); err != nil {
			return err
		}
		if err := local.repositories.DeliveryLine.AddReturned(ctx, lineID, qty); err != nil {
			return err
		}
		response = dto.ReturnLineResponse{ID: returnID, DeliveryLineID: lineID, ReturnedBalanceID: balanceID, ReturnedQty: qty, MovementID: movementID}
		return nil
	})
	return response, err
}
func (s *Service) CloseReturnedDelivery(ctx context.Context, id string, q dto.DeliveryEventRequest, actor string) (dto.DeliveryResponse, error) {
	if !validID(id, 120) || !validUUID(actor) {
		return dto.DeliveryResponse{}, invalid("invalid returned delivery closure")
	}
	eventAt, err := timestamp(q.EventAt, "event_at")
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	notes, err := optional(q.Notes, 4000, "notes")
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		delivery, err := local.repositories.Delivery.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.Delivery.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "FAILED" {
			return state("only a FAILED delivery can be closed as returned")
		}
		remaining, err := local.repositories.DeliveryLine.UnaccountedCount(ctx, id)
		if err != nil {
			return err
		}
		if remaining > 0 {
			return state("return or deliver every line remainder before closure")
		}
		total, err := local.repositories.DeliveryLine.DeliveredTotal(ctx, id)
		if err != nil {
			return err
		}
		target := "RETURNED"
		if mustRat(total).Sign() > 0 {
			target = "PARTIALLY_DELIVERED"
		}
		kind, err := local.repositories.DeliveryEventType.ByCode(ctx, "RETURNED_TO_DEPOT")
		if err != nil {
			return err
		}
		eventID, err := randomID(id + "-E")
		if err != nil {
			return err
		}
		if err := local.repositories.DeliveryEvent.Create(ctx, &model.DeliveryEvent{ID: eventID, DeliveryID: id, DeliveryEventTypeID: kind.ID, EventAt: eventAt, Notes: notes, RecordedBy: actor}); err != nil {
			return err
		}
		next, err := local.transition(ctx, delivery.DocumentTypeID, delivery.StatusID, target)
		if err != nil {
			return err
		}
		if err := local.repositories.Delivery.Update(ctx, id, map[string]interface{}{"status_id": next.ID, "updated_at": time.Now(), "updated_by": actor}); err != nil {
			return err
		}
		order, err := local.repositories.Order.Lock(ctx, delivery.OutboundID)
		if err != nil {
			return err
		}
		orderNext, err := local.transition(ctx, order.DocumentTypeID, order.StatusID, target)
		if err != nil {
			return err
		}
		return local.repositories.Order.SetStatusLocked(ctx, order.ID, orderNext.ID, actor, map[string]interface{}{"completed_at": eventAt})
	})
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	return s.GetDelivery(ctx, id)
}

func (s *Service) CancelDelivery(ctx context.Context, id, actor string) (dto.DeliveryResponse, error) {
	if !validID(id, 120) || !validUUID(actor) {
		return dto.DeliveryResponse{}, invalid("invalid delivery cancellation")
	}
	err := s.transaction(ctx, func(local *Service) error {
		delivery, err := local.repositories.Delivery.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.Delivery.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "PLANNED" {
			return state("only a PLANNED delivery can be cancelled")
		}
		count, err := local.repositories.DeliveryEvent.Count(ctx, id)
		if err != nil {
			return err
		}
		if count != 0 {
			return state("delivery with events cannot be cancelled")
		}
		cancelled, err := local.transition(ctx, delivery.DocumentTypeID, delivery.StatusID, "CANCELLED")
		if err != nil {
			return err
		}
		return local.repositories.Delivery.Update(ctx, id, map[string]interface{}{"status_id": cancelled.ID, "updated_at": time.Now(), "updated_by": actor})
	})
	if err != nil {
		return dto.DeliveryResponse{}, err
	}
	return s.GetDelivery(ctx, id)
}
