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

func (s *Service) GetShipment(ctx context.Context, id string) (dto.ShipmentResponse, error) {
	if !validID(id, 120) {
		return dto.ShipmentResponse{}, invalid("invalid shipment_id")
	}
	row, err := s.repositories.Shipment.Get(ctx, id)
	if err != nil {
		return dto.ShipmentResponse{}, err
	}
	out := mapShipment(row)
	lines, err := s.repositories.ShipmentLine.Candidates(ctx, id)
	if err != nil {
		return dto.ShipmentResponse{}, err
	}
	out.Lines = make([]dto.ShipmentLineResponse, 0, len(lines))
	for _, v := range lines {
		out.Lines = append(out.Lines, mapShipmentLine(v))
	}
	drivers, err := s.repositories.ShipmentDriver.List(ctx, id)
	if err != nil {
		return dto.ShipmentResponse{}, err
	}
	out.Drivers = make([]dto.ShipmentDriverResponse, 0, len(drivers))
	for _, v := range drivers {
		out.Drivers = append(out.Drivers, mapShipmentDriver(v))
	}
	return out, nil
}
func (s *Service) ListShipments(ctx context.Context, f repository.ListFilter) (dto.PageResponse[dto.ShipmentResponse], error) {
	if !validUUID(f.OwnerID) {
		return dto.PageResponse[dto.ShipmentResponse]{}, invalid("owner_id is required and must be a UUID")
	}
	rows, total, err := s.repositories.Shipment.List(ctx, f)
	if err != nil {
		return dto.PageResponse[dto.ShipmentResponse]{}, err
	}
	items := make([]dto.ShipmentResponse, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapShipment(v))
	}
	return page(items, f.Page, f.PageSize, total), nil
}
func (s *Service) CreateShipment(ctx context.Context, q dto.CreateShipmentRequest, actor string) (dto.ShipmentResponse, error) {
	if !validUUID(q.OwnerID) || !validUUID(q.WarehouseID) || !validUUID(actor) || len(q.PackingIDs) == 0 {
		return dto.ShipmentResponse{}, invalid("invalid shipment request")
	}
	businessDate, err := date(q.BusinessDate, "business_date")
	if err != nil {
		return dto.ShipmentResponse{}, err
	}
	route, err := optional(q.RouteReference, 100, "route_reference")
	if err != nil {
		return dto.ShipmentResponse{}, err
	}
	tracking, err := optional(q.TrackingNumber, 150, "tracking_number")
	if err != nil {
		return dto.ShipmentResponse{}, err
	}
	vehicle, err := optional(q.VehicleNumber, 60, "vehicle_number")
	if err != nil {
		return dto.ShipmentResponse{}, err
	}
	seal, err := optional(q.SealNumber, 60, "seal_number")
	if err != nil {
		return dto.ShipmentResponse{}, err
	}
	notes, err := optional(q.Notes, 4000, "notes")
	if err != nil {
		return dto.ShipmentResponse{}, err
	}
	var id string
	err = s.transaction(ctx, func(local *Service) error {
		if q.CarrierServiceID != nil {
			carrierService, lookupErr := local.repositories.CarrierService.Get(ctx, *q.CarrierServiceID)
			if lookupErr != nil {
				return lookupErr
			}
			carrier, lookupErr := local.repositories.Carrier.Get(ctx, carrierService.CarrierID)
			if lookupErr != nil {
				return lookupErr
			}
			if !carrierService.IsActive || !carrier.IsActive {
				return state("carrier service is inactive")
			}
		}
		kind, initial, err := local.document(ctx, "SHIPMENT")
		if err != nil {
			return err
		}
		id, err = local.generateID(ctx, "SHIPMENT", businessDate, nil, &q.WarehouseID)
		if err != nil {
			return err
		}
		shipment := model.Shipment{ID: id, DocumentTypeID: kind.ID, StatusID: initial.ID, OwnerID: q.OwnerID, CarrierServiceID: q.CarrierServiceID, WarehouseID: q.WarehouseID, BusinessDate: businessDate, RouteReference: route, TrackingNumber: tracking, VehicleNumber: vehicle, SealNumber: seal, Notes: notes, CreatedBy: actor}
		if err := local.repositories.Shipment.Create(ctx, &shipment); err != nil {
			return err
		}
		orders := make([]model.ShipmentOrder, 0, len(q.PackingIDs))
		links := make([]model.ShipmentPacking, 0, len(q.PackingIDs))
		seen := map[string]bool{}
		for _, packingID := range q.PackingIDs {
			if !validID(packingID, 120) {
				return invalid("invalid packing_id")
			}
			packing, err := local.repositories.Packing.Get(ctx, packingID)
			if err != nil {
				return err
			}
			if packing.StatusCode != "COMPLETED" || packing.OwnerID != q.OwnerID || packing.WarehouseID != q.WarehouseID {
				return state("every packing must be COMPLETED for the shipment owner and warehouse")
			}
			if seen[packing.OutboundID] {
				return invalid("packing_ids contain the same outbound order more than once")
			}
			seen[packing.OutboundID] = true
			orders = append(orders, model.ShipmentOrder{ShipmentID: id, OutboundID: packing.OutboundID, AddedBy: actor})
			links = append(links, model.ShipmentPacking{ShipmentID: id, PackingID: packingID})
		}
		if err := local.repositories.ShipmentOrder.CreateBatch(ctx, orders); err != nil {
			return err
		}
		return local.repositories.ShipmentPacking.CreateBatch(ctx, links)
	})
	if err != nil {
		return dto.ShipmentResponse{}, err
	}
	return s.GetShipment(ctx, id)
}
func (s *Service) DispatchShipmentLine(ctx context.Context, shipmentID, packingLineID string, q dto.DispatchShipmentLineRequest, actor string) (dto.ShipmentResponse, error) {
	if !validID(shipmentID, 120) || !validID(packingLineID, 150) || !validUUID(actor) || q.ExpectedBalanceVersion < 1 {
		return dto.ShipmentResponse{}, invalid("invalid shipment dispatch request")
	}
	businessDate, err := date(q.BusinessDate, "business_date")
	if err != nil {
		return dto.ShipmentResponse{}, err
	}
	operationKey, err := clean(q.OperationKey, 160, "operation_key")
	if err != nil {
		return dto.ShipmentResponse{}, err
	}
	fp := fingerprint(shipmentID, packingLineID, q.BusinessDate)
	if existing, e := s.repositories.Inventory.Movement.GetByOperationKey(ctx, operationKey); e == nil {
		if existing.OperationFingerprint == nil || *existing.OperationFingerprint != fp {
			return dto.ShipmentResponse{}, repository.ErrConflict
		}
		return s.GetShipment(ctx, shipmentID)
	} else if !errors.Is(e, inventoryrepository.ErrNotFound) {
		return dto.ShipmentResponse{}, e
	}
	err = s.transaction(ctx, func(local *Service) error {
		shipment, err := local.repositories.Shipment.Lock(ctx, shipmentID)
		if err != nil {
			return err
		}
		row, err := local.repositories.Shipment.Get(ctx, shipmentID)
		if err != nil {
			return err
		}
		if row.StatusCode != "PLANNED" {
			return state("only a PLANNED shipment can dispatch stock")
		}
		candidate, err := local.repositories.ShipmentLine.GetCandidate(ctx, shipmentID, packingLineID)
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
		qty := candidate.ShippedQty
		if mustRat(source.OnHandQty).Cmp(mustRat(qty)) < 0 {
			return state("packing balance no longer covers shipped quantity")
		}
		if source.HandlingUnitID != nil {
			if mustRat(source.OnHandQty).Cmp(mustRat(qty)) != 0 {
				return state("handling-unit stock must be shipped in full")
			}
			positive, err := local.repositories.Inventory.Balance.CountPositiveForHandlingUnit(ctx, *source.HandlingUnitID)
			if err != nil {
				return err
			}
			if positive != 1 {
				return state("handling unit contains multiple positive stock identities and cannot be shipped by one line")
			}
		}
		if err := local.repositories.Inventory.Balance.SetQuantities(ctx, source.ID, decimal(new(big.Rat).Sub(mustRat(source.OnHandQty), mustRat(qty))), source.ReservedQty); err != nil {
			return err
		}
		serialID, err := local.moveSerialIdentity(ctx, source.ID, nil, mustRat(qty))
		if err != nil {
			return err
		}
		if source.HandlingUnitID != nil {
			if err := local.repositories.Inventory.HandlingUnit.ClearLocation(ctx, *source.HandlingUnitID, source.LocationID); err != nil {
				return err
			}
		}
		kind, err := local.repositories.Inventory.MovementType.ByCodeShared(ctx, "SHIP")
		if err != nil || !kind.IsActive {
			return state("SHIP movement type is not configured")
		}
		movementID, err := local.generateID(ctx, "MOVEMENT", businessDate, nil, &shipment.WarehouseID)
		if err != nil {
			return err
		}
		from := source.LocationID
		status := source.InventoryStatusID
		movement := inventorymodel.InventoryMovement{ID: movementID, MovementTypeID: kind.ID, OwnerID: source.OwnerID, WarehouseID: source.WarehouseID, BusinessDate: businessDate, OccurredAt: time.Now(), ItemID: source.ItemID, LotID: source.LotID, SerialID: serialID, HandlingUnitID: source.HandlingUnitID, FromLocationID: &from, FromStatusID: &status, Quantity: qty, UOMID: source.UOMID, SourceDocumentID: shipmentID, SourceLineID: &packingLineID, OperationKey: &operationKey, OperationFingerprint: &fp, CreatedBy: actor}
		if err := local.repositories.Inventory.Movement.Create(ctx, &movement); err != nil {
			return err
		}
		lineID, err := randomID(shipmentID + "-L")
		if err != nil {
			return err
		}
		line := model.ShipmentLine{ID: lineID, ShipmentID: shipmentID, PackingLineID: packingLineID, SourceBalanceID: source.ID, ShippedQty: qty, UOMID: source.UOMID, MovementID: movementID, CreatedBy: actor}
		if err := local.repositories.ShipmentLine.Create(ctx, &line); err != nil {
			return err
		}
		return local.repositories.Line.AddShipped(ctx, candidate.OutboundLineID, qty)
	})
	if err != nil {
		return dto.ShipmentResponse{}, err
	}
	return s.GetShipment(ctx, shipmentID)
}
func (s *Service) CompleteShipment(ctx context.Context, id, actor string) (dto.ShipmentResponse, error) {
	if !validID(id, 120) || !validUUID(actor) {
		return dto.ShipmentResponse{}, invalid("invalid shipment completion")
	}
	err := s.transaction(ctx, func(local *Service) error {
		shipment, err := local.repositories.Shipment.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.Shipment.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "PLANNED" {
			return state("only a PLANNED shipment can be completed")
		}
		missing, err := local.repositories.ShipmentLine.MissingCount(ctx, id)
		if err != nil {
			return err
		}
		if missing > 0 {
			return state("dispatch every packing line before completing the shipment")
		}
		if shipment.CarrierServiceID != nil {
			hasPrimary, err := local.repositories.ShipmentDriver.HasPrimary(ctx, id)
			if err != nil {
				return err
			}
			if !hasPrimary {
				return state("shipment requires a primary driver")
			}
		}
		next, err := local.transition(ctx, shipment.DocumentTypeID, shipment.StatusID, "SHIPPED")
		if err != nil {
			return err
		}
		if err := local.repositories.Shipment.Complete(ctx, id, next.ID, actor); err != nil {
			return err
		}
		orders, err := local.repositories.ShipmentOrder.List(ctx, id)
		if err != nil {
			return err
		}
		for _, membership := range orders {
			order, err := local.repositories.Order.Lock(ctx, membership.OutboundID)
			if err != nil {
				return err
			}
			orderNext, err := local.transition(ctx, order.DocumentTypeID, order.StatusID, "SHIPPED")
			if err != nil {
				return err
			}
			if err := local.repositories.Order.SetStatusLocked(ctx, order.ID, orderNext.ID, actor, nil); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return dto.ShipmentResponse{}, err
	}
	return s.GetShipment(ctx, id)
}

func (s *Service) CancelShipment(ctx context.Context, id, actor string) (dto.ShipmentResponse, error) {
	if !validID(id, 120) || !validUUID(actor) {
		return dto.ShipmentResponse{}, invalid("invalid shipment cancellation")
	}
	err := s.transaction(ctx, func(local *Service) error {
		shipment, err := local.repositories.Shipment.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.Shipment.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "PLANNED" {
			return state("only a PLANNED shipment can be cancelled")
		}
		count, err := local.repositories.ShipmentLine.Count(ctx, id)
		if err != nil {
			return err
		}
		if count != 0 {
			return state("shipment with dispatched lines cannot be cancelled")
		}
		cancelled, err := local.transition(ctx, shipment.DocumentTypeID, shipment.StatusID, "CANCELLED")
		if err != nil {
			return err
		}
		if err := local.repositories.ShipmentOrder.RemoveAll(ctx, id, actor); err != nil {
			return err
		}
		return local.repositories.Shipment.Cancel(ctx, id, cancelled.ID)
	})
	if err != nil {
		return dto.ShipmentResponse{}, err
	}
	return s.GetShipment(ctx, id)
}
