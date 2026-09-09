package inbound

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	dto "wms-api/dto/inbound"
	model "wms-api/models/inbound"
	repository "wms-api/repository/inbound"
)

func (s *Service) UpdatePurchaseOrder(ctx context.Context, id string, request dto.UpdatePurchaseOrderRequest, actor string) (dto.PurchaseOrderResponse, error) {
	if !inboundID(id, 120) || !inboundUUID(actor) || request.ExpectedVersion < 1 {
		return dto.PurchaseOrderResponse{}, invalid("invalid draft purchase-order update")
	}
	number, err := clean(request.PurchaseOrderNo, 120, "purchase_order_no")
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	orderedAt, err := inboundTimestamp(request.OrderedAt, "ordered_at")
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	expectedArrival, err := inboundOptionalTimestamp(request.ExpectedArrivalAt, "expected_arrival_at")
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	if expectedArrival != nil && expectedArrival.Before(orderedAt) {
		return dto.PurchaseOrderResponse{}, invalid("expected_arrival_at cannot precede ordered_at")
	}
	notes, err := optional(request.Notes, 4000, "notes")
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		header, err := local.repositories.PurchaseOrder.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.PurchaseOrder.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "DRAFT" {
			return state("only a DRAFT purchase order can be edited")
		}
		if header.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		return local.repositories.PurchaseOrder.UpdateDraft(ctx, id, actor, request.ExpectedVersion, map[string]interface{}{
			"purchase_order_no": number, "ordered_at": orderedAt, "expected_arrival_at": expectedArrival, "notes": notes,
		})
	})
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	return s.GetPurchaseOrder(ctx, id)
}

func (s *Service) validatePurchaseOrderLine(ctx context.Context, local *Service, ownerID string, request dto.PurchaseOrderLineRequest) (model.PurchaseOrderLine, error) {
	if !inboundUUID(request.ItemID) || !inboundUUID(request.UOMID) {
		return model.PurchaseOrderLine{}, invalid("line item_id and uom_id must be UUIDs")
	}
	quantity, _, err := inboundQuantity(request.OrderedQty, "ordered_qty", false)
	if err != nil {
		return model.PurchaseOrderLine{}, err
	}
	over, _, err := inboundPercentage(request.OverReceiptTolerancePct, "over_receipt_tolerance_pct")
	if err != nil {
		return model.PurchaseOrderLine{}, err
	}
	under, _, err := inboundPercentage(request.UnderReceiptTolerancePct, "under_receipt_tolerance_pct")
	if err != nil {
		return model.PurchaseOrderLine{}, err
	}
	item, err := local.repositories.Master.Catalog.Item.Get(ctx, strings.ToLower(request.ItemID))
	if err != nil || !item.IsActive || item.OwnerID != ownerID {
		return model.PurchaseOrderLine{}, invalid("line item does not belong to the active owner")
	}
	unit, err := local.repositories.Master.Catalog.ItemUOM.ForUnit(ctx, item.ID, strings.ToLower(request.UOMID))
	if err != nil || !unit.IsActive || !unit.IsReceivingUOM {
		return model.PurchaseOrderLine{}, invalid("line UOM is not an active receiving UOM for the item")
	}
	expiry, err := inboundOptionalDate(request.ExpectedExpiryDate, "expected_expiry_date")
	if err != nil {
		return model.PurchaseOrderLine{}, err
	}
	vendorCode, err := optional(request.VendorItemCode, 100, "vendor_item_code")
	if err != nil {
		return model.PurchaseOrderLine{}, err
	}
	lot, err := optional(request.ExpectedLotNo, 100, "expected_lot_no")
	if err != nil {
		return model.PurchaseOrderLine{}, err
	}
	notes, err := optional(request.Notes, 4000, "line notes")
	if err != nil {
		return model.PurchaseOrderLine{}, err
	}
	return model.PurchaseOrderLine{OwnerID: ownerID, ItemID: item.ID, OrderedQty: quantity, OverReceiptTolerancePct: over, UnderReceiptTolerancePct: under, UOMID: unit.UOMID, VendorItemCode: vendorCode, ExpectedLotNo: lot, ExpectedExpiryDate: expiry, Notes: notes}, nil
}

func (s *Service) AddPurchaseOrderLine(ctx context.Context, id string, request dto.AddPurchaseOrderLineRequest, actor string) (dto.PurchaseOrderResponse, error) {
	if !inboundID(id, 120) || !inboundUUID(actor) || request.ExpectedVersion < 1 {
		return dto.PurchaseOrderResponse{}, invalid("invalid draft purchase-order line")
	}
	err := s.transaction(ctx, func(local *Service) error {
		header, err := local.repositories.PurchaseOrder.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.PurchaseOrder.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "DRAFT" {
			return state("only a DRAFT purchase order can change lines")
		}
		if header.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		line, err := s.validatePurchaseOrderLine(ctx, local, header.OwnerID, request.PurchaseOrderLineRequest)
		if err != nil {
			return err
		}
		lineNo, err := local.repositories.PurchaseOrderLine.NextLineNo(ctx, id)
		if err != nil {
			return err
		}
		line.ID, line.PurchaseOrderID, line.LineNo, line.CreatedBy = fmt.Sprintf("%s-L%04d", id, lineNo), id, lineNo, actor
		if err := local.repositories.PurchaseOrderLine.Create(ctx, &line); err != nil {
			return err
		}
		return local.repositories.PurchaseOrder.UpdateDraft(ctx, id, actor, request.ExpectedVersion, map[string]interface{}{})
	})
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	return s.GetPurchaseOrder(ctx, id)
}

func (s *Service) UpdatePurchaseOrderLine(ctx context.Context, id, lineID string, request dto.UpdatePurchaseOrderLineRequest, actor string) (dto.PurchaseOrderResponse, error) {
	if !inboundID(id, 120) || !inboundID(lineID, 150) || !inboundUUID(actor) || request.ExpectedVersion < 1 {
		return dto.PurchaseOrderResponse{}, invalid("invalid draft purchase-order line update")
	}
	qty, _, err := inboundQuantity(request.OrderedQty, "ordered_qty", false)
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	over, _, err := inboundPercentage(request.OverReceiptTolerancePct, "over_receipt_tolerance_pct")
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	under, _, err := inboundPercentage(request.UnderReceiptTolerancePct, "under_receipt_tolerance_pct")
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	vendorCode, err := optional(request.VendorItemCode, 100, "vendor_item_code")
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	lot, err := optional(request.ExpectedLotNo, 100, "expected_lot_no")
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	expiry, err := inboundOptionalDate(request.ExpectedExpiryDate, "expected_expiry_date")
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	notes, err := optional(request.Notes, 4000, "line notes")
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		header, err := local.repositories.PurchaseOrder.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.PurchaseOrder.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "DRAFT" {
			return state("only a DRAFT purchase order can change lines")
		}
		if header.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		if err := local.repositories.PurchaseOrderLine.UpdateDraft(ctx, lineID, id, map[string]interface{}{"ordered_qty": qty, "over_receipt_tolerance_pct": over, "under_receipt_tolerance_pct": under, "vendor_item_code": vendorCode, "expected_lot_no": lot, "expected_expiry_date": expiry, "notes": notes}); err != nil {
			return err
		}
		return local.repositories.PurchaseOrder.UpdateDraft(ctx, id, actor, request.ExpectedVersion, map[string]interface{}{})
	})
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	return s.GetPurchaseOrder(ctx, id)
}

func (s *Service) DeletePurchaseOrderLine(ctx context.Context, id, lineID string, expectedVersion int64, actor string) (dto.PurchaseOrderResponse, error) {
	if !inboundID(id, 120) || !inboundID(lineID, 150) || !inboundUUID(actor) || expectedVersion < 1 {
		return dto.PurchaseOrderResponse{}, invalid("invalid draft purchase-order line deletion")
	}
	err := s.transaction(ctx, func(local *Service) error {
		header, err := local.repositories.PurchaseOrder.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.PurchaseOrder.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "DRAFT" {
			return state("only a DRAFT purchase order can change lines")
		}
		if header.VersionNo != expectedVersion {
			return repository.ErrConcurrentWrite
		}
		count, err := local.repositories.PurchaseOrderLine.Count(ctx, id)
		if err != nil {
			return err
		}
		if count <= 1 {
			return state("a purchase order must retain at least one line")
		}
		if err := local.repositories.PurchaseOrderLine.DeleteDraft(ctx, lineID, id); err != nil {
			return err
		}
		return local.repositories.PurchaseOrder.UpdateDraft(ctx, id, actor, expectedVersion, map[string]interface{}{})
	})
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	return s.GetPurchaseOrder(ctx, id)
}

func (s *Service) UpdateInboundOrder(ctx context.Context, id string, request dto.UpdateInboundOrderRequest, actor string) (dto.InboundOrderResponse, error) {
	if !inboundID(id, 120) || !inboundUUID(actor) || request.ExpectedVersion < 1 {
		return dto.InboundOrderResponse{}, invalid("invalid draft inbound-order update")
	}
	arrival, err := inboundOptionalTimestamp(request.ExpectedArrivalAt, "expected_arrival_at")
	if err != nil {
		return dto.InboundOrderResponse{}, err
	}
	external, err := optional(request.ExternalReference, 120, "external_reference")
	if err != nil {
		return dto.InboundOrderResponse{}, err
	}
	supplier, err := optional(request.SupplierReference, 120, "supplier_reference")
	if err != nil {
		return dto.InboundOrderResponse{}, err
	}
	notes, err := optional(request.Notes, 4000, "notes")
	if err != nil {
		return dto.InboundOrderResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		header, err := local.repositories.InboundOrder.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.InboundOrder.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "DRAFT" {
			return state("only a DRAFT inbound order can be edited")
		}
		if header.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		return local.repositories.InboundOrder.UpdateDraft(ctx, id, actor, request.ExpectedVersion, map[string]interface{}{"expected_arrival_at": arrival, "external_reference": external, "supplier_reference": supplier, "notes": notes})
	})
	if err != nil {
		return dto.InboundOrderResponse{}, err
	}
	return s.GetInboundOrder(ctx, id)
}

func validateInboundExpectedQty(ctx context.Context, local *Service, poLine model.PurchaseOrderLine, excludedLineID, requested string) (string, error) {
	expected, expectedNumber, err := inboundQuantity(requested, "expected_qty", false)
	if err != nil {
		return "", err
	}
	ordered, ok := new(big.Rat).SetString(poLine.OrderedQty)
	if !ok {
		return "", state("stored purchase order quantity is invalid")
	}
	var scheduledText string
	if excludedLineID == "" {
		scheduledText, err = local.repositories.InboundOrderLine.ScheduledForPurchaseOrderLine(ctx, poLine.ID)
	} else {
		scheduledText, err = local.repositories.InboundOrderLine.ScheduledExcept(ctx, poLine.ID, excludedLineID)
	}
	if err != nil {
		return "", err
	}
	scheduled, ok := new(big.Rat).SetString(scheduledText)
	if !ok || new(big.Rat).Add(scheduled, expectedNumber).Cmp(ordered) > 0 {
		return "", invalid("inbound expected quantity exceeds unscheduled purchase order quantity")
	}
	return expected, nil
}

func (s *Service) AddInboundOrderLine(ctx context.Context, id string, request dto.AddInboundOrderLineRequest, actor string) (dto.InboundOrderResponse, error) {
	if !inboundID(id, 120) || !inboundUUID(actor) || request.ExpectedVersion < 1 || !inboundID(request.PurchaseOrderLineID, 150) {
		return dto.InboundOrderResponse{}, invalid("invalid draft inbound-order line")
	}
	err := s.transaction(ctx, func(local *Service) error {
		header, err := local.repositories.InboundOrder.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.InboundOrder.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "DRAFT" {
			return state("only a DRAFT inbound order can change lines")
		}
		if header.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		poLine, err := local.repositories.PurchaseOrderLine.Get(ctx, request.PurchaseOrderLineID)
		if err != nil {
			return invalid("purchase order line does not belong to the source purchase order")
		}
		sourcePOID, err := local.repositories.InboundOrder.SourcePurchaseOrderID(ctx, id)
		if err != nil || poLine.PurchaseOrderID != sourcePOID {
			return invalid("purchase order line does not belong to the source purchase order")
		}
		existing, err := local.repositories.InboundOrderLine.List(ctx, id)
		if err != nil {
			return err
		}
		for _, line := range existing {
			if line.PurchaseOrderLineID != nil && *line.PurchaseOrderLineID == poLine.ID {
				return invalid("purchase_order_line_id is repeated")
			}
		}
		expected, err := validateInboundExpectedQty(ctx, local, poLine, "", request.ExpectedQty)
		if err != nil {
			return err
		}
		reference, err := optional(request.CustomerLineReference, 100, "customer_line_reference")
		if err != nil {
			return err
		}
		notes, err := optional(request.Notes, 4000, "line notes")
		if err != nil {
			return err
		}
		lineNo, err := local.repositories.InboundOrderLine.NextLineNo(ctx, id)
		if err != nil {
			return err
		}
		poLineID := poLine.ID
		line := model.InboundOrderLine{ID: fmt.Sprintf("%s-L%04d", id, lineNo), InboundID: id, PurchaseOrderLineID: &poLineID, LineNo: lineNo, ItemID: poLine.ItemID, ExpectedQty: expected, UOMID: poLine.UOMID, ExpectedLotNo: poLine.ExpectedLotNo, ExpectedExpiryDate: poLine.ExpectedExpiryDate, CustomerLineReference: reference, Notes: notes, CreatedBy: actor}
		if err := local.repositories.InboundOrderLine.Create(ctx, &line); err != nil {
			return err
		}
		return local.repositories.InboundOrder.UpdateDraft(ctx, id, actor, request.ExpectedVersion, map[string]interface{}{})
	})
	if err != nil {
		return dto.InboundOrderResponse{}, err
	}
	return s.GetInboundOrder(ctx, id)
}

func (s *Service) UpdateInboundOrderLine(ctx context.Context, id, lineID string, request dto.UpdateInboundOrderLineRequest, actor string) (dto.InboundOrderResponse, error) {
	if !inboundID(id, 120) || !inboundID(lineID, 150) || !inboundUUID(actor) || request.ExpectedVersion < 1 {
		return dto.InboundOrderResponse{}, invalid("invalid draft inbound-order line update")
	}
	err := s.transaction(ctx, func(local *Service) error {
		header, err := local.repositories.InboundOrder.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.InboundOrder.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "DRAFT" {
			return state("only a DRAFT inbound order can change lines")
		}
		if header.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		line, err := local.repositories.InboundOrderLine.Get(ctx, lineID)
		if err != nil || line.InboundID != id || line.PurchaseOrderLineID == nil {
			return repository.ErrNotFound
		}
		poLine, err := local.repositories.PurchaseOrderLine.Get(ctx, *line.PurchaseOrderLineID)
		if err != nil {
			return err
		}
		expected, err := validateInboundExpectedQty(ctx, local, poLine, lineID, request.ExpectedQty)
		if err != nil {
			return err
		}
		reference, err := optional(request.CustomerLineReference, 100, "customer_line_reference")
		if err != nil {
			return err
		}
		notes, err := optional(request.Notes, 4000, "line notes")
		if err != nil {
			return err
		}
		if err := local.repositories.InboundOrderLine.UpdateDraft(ctx, lineID, id, map[string]interface{}{"expected_qty": expected, "customer_line_reference": reference, "notes": notes}); err != nil {
			return err
		}
		return local.repositories.InboundOrder.UpdateDraft(ctx, id, actor, request.ExpectedVersion, map[string]interface{}{})
	})
	if err != nil {
		return dto.InboundOrderResponse{}, err
	}
	return s.GetInboundOrder(ctx, id)
}

func (s *Service) DeleteInboundOrderLine(ctx context.Context, id, lineID string, expectedVersion int64, actor string) (dto.InboundOrderResponse, error) {
	if !inboundID(id, 120) || !inboundID(lineID, 150) || !inboundUUID(actor) || expectedVersion < 1 {
		return dto.InboundOrderResponse{}, invalid("invalid draft inbound-order line deletion")
	}
	err := s.transaction(ctx, func(local *Service) error {
		header, err := local.repositories.InboundOrder.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.InboundOrder.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "DRAFT" {
			return state("only a DRAFT inbound order can change lines")
		}
		if header.VersionNo != expectedVersion {
			return repository.ErrConcurrentWrite
		}
		count, err := local.repositories.InboundOrderLine.Count(ctx, id)
		if err != nil {
			return err
		}
		if count <= 1 {
			return state("an inbound order must retain at least one line")
		}
		if err := local.repositories.InboundOrderLine.DeleteDraft(ctx, lineID, id); err != nil {
			return err
		}
		return local.repositories.InboundOrder.UpdateDraft(ctx, id, actor, expectedVersion, map[string]interface{}{})
	})
	if err != nil {
		return dto.InboundOrderResponse{}, err
	}
	return s.GetInboundOrder(ctx, id)
}
