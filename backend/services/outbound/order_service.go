package outbound

import (
	"context"
	"fmt"
	"strings"
	"time"

	dto "wms-api/dto/outbound"
	model "wms-api/models/outbound"
	repository "wms-api/repository/outbound"
)

func (s *Service) CreateOutboundOrder(ctx context.Context, request dto.CreateOutboundOrderRequest, actor string) (dto.OutboundOrderResponse, error) {
	if !validUUID(request.OwnerID) || !validUUID(request.CustomerID) || !validUUID(request.ShipToPartnerID) || !validUUID(request.WarehouseID) || !validUUID(actor) {
		return dto.OutboundOrderResponse{}, invalid("owner_id, customer_id, ship_to_partner_id, warehouse_id and actor must be UUIDs")
	}
	request.OwnerID = strings.ToLower(request.OwnerID)
	request.CustomerID = strings.ToLower(request.CustomerID)
	request.ShipToPartnerID = strings.ToLower(request.ShipToPartnerID)
	request.WarehouseID = strings.ToLower(request.WarehouseID)
	doNo, err := clean(request.ClientDeliveryOrderNo, 120, "client_delivery_order_no")
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	businessDate, err := date(request.BusinessDate, "business_date")
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	shipAt, err := timestamp(request.RequestedShipAt, "requested_ship_at")
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	customerOrderNo, err := optional(request.CustomerOrderNo, 120, "customer_order_no")
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	external, err := optional(request.ExternalReference, 120, "external_reference")
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	notes, err := optional(request.Notes, 4000, "notes")
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	if len(request.Lines) == 0 || len(request.Lines) > 500 {
		return dto.OutboundOrderResponse{}, invalid("lines must contain 1..500 entries")
	}
	var id string
	err = s.transaction(ctx, func(local *Service) error {
		owner, err := local.repositories.Master.Organization.FindByID(ctx, request.OwnerID)
		if err != nil || !owner.IsActive {
			return invalid("owner does not exist or is inactive")
		}
		customer, err := local.repositories.Master.Catalog.BusinessPartner.Get(ctx, request.CustomerID)
		if err != nil || !customer.IsActive || customer.OwnerID != request.OwnerID {
			return invalid("customer does not belong to the active owner")
		}
		customerTypes, err := local.repositories.Master.Catalog.BusinessPartnerType.List(ctx, customer.ID)
		if err != nil {
			return err
		}
		customerAllowed := false
		for _, partnerType := range customerTypes {
			if partnerType.IsActive && partnerType.Code == "CUSTOMER" {
				customerAllowed = true
			}
		}
		if !customerAllowed {
			return invalid("customer requires an active CUSTOMER partner type")
		}
		shipTo, err := local.repositories.Master.Catalog.BusinessPartner.Get(ctx, request.ShipToPartnerID)
		if err != nil || !shipTo.IsActive || shipTo.OwnerID != request.OwnerID {
			return invalid("ship-to partner does not belong to the active owner")
		}
		if shipTo.AddressLine1 == nil || strings.TrimSpace(*shipTo.AddressLine1) == "" {
			return invalid("ship-to partner requires address_line_1")
		}
		shipTypes, err := local.repositories.Master.Catalog.BusinessPartnerType.List(ctx, shipTo.ID)
		if err != nil {
			return err
		}
		shipAllowed := false
		for _, partnerType := range shipTypes {
			if partnerType.IsActive && (partnerType.Code == "CUSTOMER" || partnerType.Code == "STORE") {
				shipAllowed = true
			}
		}
		if !shipAllowed {
			return invalid("ship-to partner requires an active CUSTOMER or STORE partner type")
		}
		warehouse, err := local.repositories.Master.Warehouse.FindByID(ctx, request.WarehouseID)
		if err != nil || !warehouse.IsActive {
			return invalid("warehouse does not exist or is inactive")
		}
		assigned, err := local.repositories.Master.WarehouseOwner.IsActive(ctx, request.WarehouseID, request.OwnerID)
		if err != nil || !assigned {
			return invalid("owner is not active for the warehouse")
		}
		kind, err := local.documentType(ctx, "OUTBOUND")
		if err != nil {
			return err
		}
		initial, err := local.initialStatus(ctx, kind.ID)
		if err != nil {
			return err
		}
		id, err = local.generateID(ctx, "OUTBOUND", businessDate, &customer.ID, &warehouse.ID)
		if err != nil {
			return err
		}
		header := model.OutboundOrder{ID: id, DocumentTypeID: kind.ID, StatusID: initial.ID, OwnerID: request.OwnerID, CustomerID: request.CustomerID, WarehouseID: request.WarehouseID, BusinessDate: businessDate, RequestedShipAt: &shipAt, ExternalReference: external, CustomerOrderNo: customerOrderNo, ClientDeliveryOrderNo: doNo, ShipToPartnerID: &request.ShipToPartnerID, ShipToName: shipTo.Name, ShipToAddress1: *shipTo.AddressLine1, ShipToAddress2: shipTo.AddressLine2, ShipToCity: shipTo.City, ShipToProvince: shipTo.Province, ShipToPostalCode: shipTo.PostalCode, ShipToCountryCode: shipTo.CountryCode, Notes: notes, CreatedBy: actor, UpdatedBy: &actor, VersionNo: 1}
		if err := local.repositories.Order.Create(ctx, &header); err != nil {
			return err
		}
		lines := make([]model.OutboundOrderLine, 0, len(request.Lines))
		for i, x := range request.Lines {
			if !validUUID(x.ItemID) {
				return invalid("line item_id must be a UUID")
			}
			itemID := strings.ToLower(x.ItemID)
			item, err := local.repositories.Master.Catalog.Item.Get(ctx, itemID)
			if err != nil || !item.IsActive || item.OwnerID != request.OwnerID {
				return invalid("line item does not belong to the active owner")
			}
			qty, _, err := quantity(x.OrderedQty, "ordered_qty", false)
			if err != nil {
				return err
			}
			lot, err := optional(x.RequestedLotNo, 100, "requested_lot_no")
			if err != nil {
				return err
			}
			reference, err := optional(x.CustomerLineReference, 100, "customer_line_reference")
			if err != nil {
				return err
			}
			lineNotes, err := optional(x.Notes, 4000, "line notes")
			if err != nil {
				return err
			}
			no := i + 1
			lines = append(lines, model.OutboundOrderLine{ID: fmt.Sprintf("%s-L%04d", id, no), OutboundID: id, LineNo: no, ItemID: item.ID, OrderedQty: qty, AllocatedQty: "0", PickedQty: "0", CheckedQty: "0", PackedQty: "0", ShippedQty: "0", DeliveredQty: "0", RejectedQty: "0", ShortAcceptedQty: "0", UOMID: item.BaseUOMID, RequestedLotNo: lot, CustomerLineReference: reference, Notes: lineNotes, CreatedBy: actor})
		}
		return local.repositories.Line.CreateBatch(ctx, lines)
	})
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	return s.GetOutboundOrder(ctx, id)
}

func (s *Service) UpdateOutboundOrder(ctx context.Context, id string, request dto.UpdateOutboundOrderRequest, actor string) (dto.OutboundOrderResponse, error) {
	if !validID(id, 120) || !validUUID(request.ShipToPartnerID) || !validUUID(actor) || request.ExpectedVersion < 1 {
		return dto.OutboundOrderResponse{}, invalid("invalid draft outbound update")
	}
	shipAt, err := timestamp(request.RequestedShipAt, "requested_ship_at")
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	customerOrderNo, err := optional(request.CustomerOrderNo, 120, "customer_order_no")
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	external, err := optional(request.ExternalReference, 120, "external_reference")
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	notes, err := optional(request.Notes, 4000, "notes")
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		order, err := local.repositories.Order.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.Order.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "DRAFT" {
			return state("only a DRAFT outbound order can be updated")
		}
		if order.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		shipTo, err := local.repositories.Master.Catalog.BusinessPartner.Get(ctx, strings.ToLower(request.ShipToPartnerID))
		if err != nil || !shipTo.IsActive || shipTo.OwnerID != order.OwnerID || shipTo.AddressLine1 == nil || strings.TrimSpace(*shipTo.AddressLine1) == "" {
			return invalid("ship-to partner must be active, addressed, and belong to the owner")
		}
		return local.repositories.Order.UpdateDraft(ctx, id, actor, request.ExpectedVersion, map[string]interface{}{
			"ship_to_partner_id": shipTo.ID, "requested_ship_at": shipAt,
			"customer_order_no": customerOrderNo, "external_reference": external, "notes": notes,
			"ship_to_name": shipTo.Name, "ship_to_address_1": *shipTo.AddressLine1,
			"ship_to_address_2": shipTo.AddressLine2, "ship_to_city": shipTo.City,
			"ship_to_province": shipTo.Province, "ship_to_postal_code": shipTo.PostalCode,
			"ship_to_country_code": shipTo.CountryCode,
		})
	})
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	return s.GetOutboundOrder(ctx, id)
}

func (s *Service) AddOutboundOrderLine(ctx context.Context, id string, input dto.AddOutboundOrderLineRequest, actor string) (dto.OutboundOrderResponse, error) {
	request := input.OutboundOrderLineRequest
	if !validID(id, 120) || !validUUID(request.ItemID) || !validUUID(actor) || input.ExpectedVersion < 1 {
		return dto.OutboundOrderResponse{}, invalid("invalid draft outbound line")
	}
	qty, _, err := quantity(request.OrderedQty, "ordered_qty", false)
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	lot, err := optional(request.RequestedLotNo, 100, "requested_lot_no")
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	reference, err := optional(request.CustomerLineReference, 100, "customer_line_reference")
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	notes, err := optional(request.Notes, 4000, "line notes")
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		order, err := local.repositories.Order.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.Order.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "DRAFT" {
			return state("only a DRAFT outbound order accepts new lines")
		}
		if order.VersionNo != input.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		item, err := local.repositories.Master.Catalog.Item.Get(ctx, strings.ToLower(request.ItemID))
		if err != nil || !item.IsActive || item.OwnerID != order.OwnerID {
			return invalid("line item does not belong to the active owner")
		}
		lineNo, err := local.repositories.Line.NextLineNo(ctx, id)
		if err != nil {
			return err
		}
		line := model.OutboundOrderLine{ID: fmt.Sprintf("%s-L%04d", id, lineNo), OutboundID: id, LineNo: lineNo, ItemID: item.ID, OrderedQty: qty, AllocatedQty: "0", PickedQty: "0", CheckedQty: "0", PackedQty: "0", ShippedQty: "0", DeliveredQty: "0", RejectedQty: "0", ShortAcceptedQty: "0", UOMID: item.BaseUOMID, RequestedLotNo: lot, CustomerLineReference: reference, Notes: notes, CreatedBy: actor}
		if err := local.repositories.Line.Create(ctx, &line); err != nil {
			return err
		}
		return local.repositories.Order.UpdateDraft(ctx, id, actor, input.ExpectedVersion, map[string]interface{}{})
	})
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	return s.GetOutboundOrder(ctx, id)
}

func (s *Service) UpdateOutboundOrderLine(ctx context.Context, id, lineID string, request dto.UpdateOutboundOrderLineRequest, actor string) (dto.OutboundOrderResponse, error) {
	if !validID(id, 120) || !validID(lineID, 150) || !validUUID(actor) || request.ExpectedVersion < 1 {
		return dto.OutboundOrderResponse{}, invalid("invalid draft outbound-line update")
	}
	qty, _, err := quantity(request.OrderedQty, "ordered_qty", false)
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	lot, err := optional(request.RequestedLotNo, 100, "requested_lot_no")
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	reference, err := optional(request.CustomerLineReference, 100, "customer_line_reference")
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	notes, err := optional(request.Notes, 4000, "line notes")
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		order, err := local.repositories.Order.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.Order.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "DRAFT" {
			return state("only a DRAFT outbound order can change lines")
		}
		if order.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		if err := local.repositories.Line.UpdateDraft(ctx, lineID, id, map[string]interface{}{"ordered_qty": qty, "requested_lot_no": lot, "customer_line_reference": reference, "notes": notes}); err != nil {
			return err
		}
		return local.repositories.Order.UpdateDraft(ctx, id, actor, request.ExpectedVersion, map[string]interface{}{})
	})
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	return s.GetOutboundOrder(ctx, id)
}

func (s *Service) DeleteOutboundOrderLine(ctx context.Context, id, lineID string, expectedVersion int64, actor string) (dto.OutboundOrderResponse, error) {
	if !validID(id, 120) || !validID(lineID, 150) || !validUUID(actor) || expectedVersion < 1 {
		return dto.OutboundOrderResponse{}, invalid("invalid draft outbound-line deletion")
	}
	err := s.transaction(ctx, func(local *Service) error {
		order, err := local.repositories.Order.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.Order.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "DRAFT" {
			return state("only a DRAFT outbound order can delete lines")
		}
		if order.VersionNo != expectedVersion {
			return repository.ErrConcurrentWrite
		}
		lines, err := local.repositories.Line.List(ctx, id)
		if err != nil {
			return err
		}
		if len(lines) <= 1 {
			return state("an outbound order must retain at least one line")
		}
		if err := local.repositories.Line.DeleteDraft(ctx, lineID, id); err != nil {
			return err
		}
		return local.repositories.Order.UpdateDraft(ctx, id, actor, expectedVersion, map[string]interface{}{})
	})
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	return s.GetOutboundOrder(ctx, id)
}

func (s *Service) GetOutboundOrder(ctx context.Context, id string) (dto.OutboundOrderResponse, error) {
	if !validID(id, 120) {
		return dto.OutboundOrderResponse{}, invalid("invalid outbound_id")
	}
	row, err := s.repositories.Order.Get(ctx, id)
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	lines, err := s.repositories.Line.List(ctx, id)
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	return mapOrder(row, lines), nil
}
func (s *Service) ListOutboundOrders(ctx context.Context, f repository.ListFilter) (dto.PageResponse[dto.OutboundOrderResponse], error) {
	if !validUUID(f.OwnerID) {
		return dto.PageResponse[dto.OutboundOrderResponse]{}, invalid("owner_id is required and must be a UUID")
	}
	rows, total, err := s.repositories.Order.List(ctx, f)
	if err != nil {
		return dto.PageResponse[dto.OutboundOrderResponse]{}, err
	}
	items := make([]dto.OutboundOrderResponse, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapOrder(v, nil))
	}
	return page(items, f.Page, f.PageSize, total), nil
}

func (s *Service) ValidateOutboundOrder(ctx context.Context, id string, request dto.ValidateOutboundRequest, actor string) (dto.ValidationRunResponse, error) {
	if !validID(id, 120) || !validUUID(actor) || request.ExpectedVersion < 1 {
		return dto.ValidationRunResponse{}, invalid("invalid outbound validation request")
	}
	notes, err := optional(request.Notes, 4000, "notes")
	if err != nil {
		return dto.ValidationRunResponse{}, err
	}
	var runID string
	err = s.transaction(ctx, func(local *Service) error {
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
		if row.StatusCode != "DRAFT" {
			return state("only a DRAFT outbound order can be validated")
		}
		kind, err := local.documentType(ctx, "OUTBOUND_VALIDATION")
		if err != nil {
			return err
		}
		pending, err := local.initialStatus(ctx, kind.ID)
		if err != nil {
			return err
		}
		warehouse, err := local.repositories.Master.Warehouse.FindByID(ctx, header.WarehouseID)
		if err != nil {
			return err
		}
		runID, err = local.generateID(ctx, "OUTBOUND_VALIDATION", header.BusinessDate, nil, &warehouse.ID)
		if err != nil {
			return err
		}
		run := model.OutboundValidationRun{ID: runID, OutboundID: id, DocumentTypeID: kind.ID, StatusID: pending.ID, Notes: notes, CreatedBy: actor}
		if err := local.repositories.ValidationRun.Create(ctx, &run); err != nil {
			return err
		}
		rules, err := local.repositories.ValidationRule.ListActive(ctx)
		if err != nil {
			return err
		}
		lines, err := local.repositories.Line.List(ctx, id)
		if err != nil {
			return err
		}
		results := make([]model.OutboundValidationResultDetail, 0, len(rules))
		passed := true
		for i, rule := range rules {
			ok := false
			switch rule.HandlerCode {
			case "REQUIRE_DO_NUMBER":
				ok = strings.TrimSpace(header.ClientDeliveryOrderNo) != ""
			case "REQUIRE_SHIP_TO":
				ok = header.ShipToPartnerID != nil && strings.TrimSpace(header.ShipToName) != "" && strings.TrimSpace(header.ShipToAddress1) != ""
			case "REQUIRE_ORDER_LINES":
				ok = len(lines) > 0
			case "REQUIRE_BASE_UOM":
				ok = true
				for _, line := range lines {
					item, e := local.repositories.Master.Catalog.Item.Get(ctx, line.ItemID)
					if e != nil || line.UOMID != item.BaseUOMID {
						ok = false
						break
					}
				}
			case "REQUIRE_REQUESTED_SHIP":
				ok = header.RequestedShipAt != nil
			}
			if !ok && rule.BlocksProcessing {
				passed = false
			}
			message := "Passed"
			if !ok && rule.Description != nil {
				message = *rule.Description
			}
			results = append(results, model.OutboundValidationResultDetail{ID: fmt.Sprintf("%s-R%03d", runID, i+1), ValidationRunID: runID, RuleID: rule.ID, Passed: ok, ResultMessage: &message})
		}
		if len(results) == 0 {
			return state("no active outbound validation rules are configured")
		}
		if err := local.repositories.ValidationResult.CreateBatch(ctx, results); err != nil {
			return err
		}
		outcome := "FAILED"
		if passed {
			outcome = "PASSED"
		}
		final, err := local.transition(ctx, kind.ID, pending.ID, outcome)
		if err != nil {
			return err
		}
		if err := local.repositories.ValidationRun.Complete(ctx, runID, final.ID, actor); err != nil {
			return err
		}
		if passed {
			target, err := local.transition(ctx, header.DocumentTypeID, header.StatusID, "VALIDATED")
			if err != nil {
				return err
			}
			return local.repositories.Order.SetStatus(ctx, id, target.ID, actor, header.VersionNo, map[string]interface{}{"validated_at": time.Now(), "validated_by": actor})
		}
		return nil
	})
	if err != nil {
		return dto.ValidationRunResponse{}, err
	}
	return s.GetValidationRun(ctx, runID)
}

func (s *Service) GetValidationRun(ctx context.Context, id string) (dto.ValidationRunResponse, error) {
	if !validID(id, 140) {
		return dto.ValidationRunResponse{}, invalid("invalid validation_run_id")
	}
	row, err := s.repositories.ValidationRun.Get(ctx, id)
	if err != nil {
		return dto.ValidationRunResponse{}, err
	}
	results, err := s.repositories.ValidationResult.List(ctx, id)
	if err != nil {
		return dto.ValidationRunResponse{}, err
	}
	r := dto.ValidationRunResponse{ID: row.ID, OutboundID: row.OutboundID, StatusCode: row.StatusCode, ValidatedAt: row.ValidatedAt, Notes: row.Notes, Results: make([]dto.ValidationResultResponse, 0, len(results))}
	for _, x := range results {
		r.Results = append(r.Results, dto.ValidationResultResponse{RuleCode: x.RuleCode, RuleName: x.RuleName, SeverityCode: x.SeverityCode, Passed: x.Passed, ResultMessage: x.ResultMessage})
	}
	return r, nil
}

func (s *Service) ReleaseOutboundOrder(ctx context.Context, id string, request dto.TransitionRequest, actor string) (dto.OutboundOrderResponse, error) {
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
		if row.StatusCode != "VALIDATED" {
			return state("only a VALIDATED outbound order can be released")
		}
		target, err := local.transition(ctx, header.DocumentTypeID, header.StatusID, "RELEASED")
		if err != nil {
			return err
		}
		return local.repositories.Order.SetStatus(ctx, id, target.ID, actor, header.VersionNo, nil)
	})
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	return s.GetOutboundOrder(ctx, id)
}

func (s *Service) CancelOutboundOrder(ctx context.Context, id string, request dto.CancelOrderRequest, actor string) (dto.OutboundOrderResponse, error) {
	if !validID(id, 120) || !validUUID(actor) || request.ExpectedVersion < 1 {
		return dto.OutboundOrderResponse{}, invalid("invalid outbound cancellation request")
	}
	if _, err := clean(request.Reason, 4000, "reason"); err != nil {
		return dto.OutboundOrderResponse{}, err
	}
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
		if row.StatusCode != "DRAFT" && row.StatusCode != "VALIDATED" && row.StatusCode != "RELEASED" {
			return state("release every reservation first; only DRAFT, VALIDATED, or unallocated RELEASED orders can be cancelled")
		}
		reservations, err := local.repositories.Reservation.ListActiveByOrder(ctx, id)
		if err != nil {
			return err
		}
		if len(reservations) != 0 {
			return state("active reservations must be released before cancellation")
		}
		activeWave, err := local.repositories.WaveOrder.HasActiveWave(ctx, id)
		if err != nil {
			return err
		}
		if activeWave {
			return state("outbound order belongs to an active wave")
		}
		target, err := local.transition(ctx, header.DocumentTypeID, header.StatusID, "CANCELLED")
		if err != nil {
			return err
		}
		return local.repositories.Order.SetStatus(ctx, id, target.ID, actor, header.VersionNo, map[string]interface{}{"completed_at": time.Now()})
	})
	if err != nil {
		return dto.OutboundOrderResponse{}, err
	}
	return s.GetOutboundOrder(ctx, id)
}
