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

func (s *Service) CreateInboundOrder(ctx context.Context, request dto.CreateInboundOrderRequest, actor string) (dto.InboundOrderResponse, error) {
	if !inboundID(request.PurchaseOrderID, 120) || !inboundUUID(actor) {
		return dto.InboundOrderResponse{}, invalid("invalid purchase_order_id or actor")
	}
	if len(request.Lines) == 0 || len(request.Lines) > 500 {
		return dto.InboundOrderResponse{}, invalid("lines must contain 1..500 entries")
	}
	businessDate, err := inboundDate(request.BusinessDate, "business_date")
	if err != nil {
		return dto.InboundOrderResponse{}, err
	}
	expectedArrival, err := inboundOptionalTimestamp(request.ExpectedArrivalAt, "expected_arrival_at")
	if err != nil {
		return dto.InboundOrderResponse{}, err
	}
	externalReference, err := optional(request.ExternalReference, 120, "external_reference")
	if err != nil {
		return dto.InboundOrderResponse{}, err
	}
	supplierReference, err := optional(request.SupplierReference, 120, "supplier_reference")
	if err != nil {
		return dto.InboundOrderResponse{}, err
	}
	notes, err := optional(request.Notes, 4000, "notes")
	if err != nil {
		return dto.InboundOrderResponse{}, err
	}
	supersedes, err := optional(request.SupersedesInboundID, 120, "supersedes_inbound_id")
	if err != nil || (supersedes != nil && !inboundID(*supersedes, 120)) {
		return dto.InboundOrderResponse{}, invalid("invalid supersedes_inbound_id")
	}
	seen := make(map[string]bool, len(request.Lines))
	var documentID string
	err = s.transaction(ctx, func(local *Service) error {
		po, err := local.repositories.PurchaseOrder.Lock(ctx, request.PurchaseOrderID)
		if err != nil {
			return err
		}
		poRow, err := local.repositories.PurchaseOrder.Get(ctx, po.ID)
		if err != nil {
			return err
		}
		if poRow.StatusCode != "APPROVED" && poRow.StatusCode != "PARTIALLY_RECEIVED" {
			return state("purchase order must be APPROVED or PARTIALLY_RECEIVED")
		}
		kind, err := local.documentType(ctx, "INBOUND")
		if err != nil {
			return err
		}
		initial, err := local.initialStatus(ctx, kind.ID)
		if err != nil {
			return err
		}
		documentID, err = local.generateID(ctx, kind.ID, request.BusinessDate, po.VendorID, po.WarehouseID)
		if err != nil {
			return err
		}
		if supersedes != nil {
			prior, lookupErr := local.repositories.InboundOrder.Get(ctx, *supersedes)
			if lookupErr != nil || prior.StatusCode != "CANCELLED" || prior.OwnerID != po.OwnerID || prior.VendorID != po.VendorID || prior.WarehouseID != po.WarehouseID {
				return invalid("superseded inbound order must be a cancelled document for the same owner, vendor, and warehouse")
			}
			priorPOID, lookupErr := local.repositories.InboundOrder.SourcePurchaseOrderID(ctx, *supersedes)
			if lookupErr != nil || priorPOID != po.ID {
				return invalid("superseded inbound order must belong to the same purchase order")
			}
			hasSuccessor, lookupErr := local.repositories.InboundOrder.HasSuccessor(ctx, *supersedes)
			if lookupErr != nil {
				return lookupErr
			}
			if hasSuccessor {
				return state("cancelled inbound order already has a successor")
			}
		}
		header := model.InboundOrder{ID: documentID, DocumentTypeID: kind.ID, StatusID: initial.ID, OwnerID: po.OwnerID, VendorID: po.VendorID, WarehouseID: po.WarehouseID, BusinessDate: businessDate, ExpectedArrivalAt: expectedArrival, ExternalReference: externalReference, SupplierReference: supplierReference, Notes: notes, SupersedesInboundID: supersedes, CreatedBy: actor, UpdatedBy: &actor, VersionNo: 1}
		if err := local.repositories.InboundOrder.Create(ctx, &header); err != nil {
			return err
		}
		for index, requestLine := range request.Lines {
			lineID := strings.TrimSpace(requestLine.PurchaseOrderLineID)
			if !inboundID(lineID, 150) || seen[lineID] {
				return invalid("purchase_order_line_id is invalid or repeated")
			}
			seen[lineID] = true
			poLine, err := local.repositories.PurchaseOrderLine.Get(ctx, lineID)
			if err != nil || poLine.PurchaseOrderID != po.ID {
				return invalid("purchase order line does not belong to the purchase order")
			}
			expected, expectedNumber, err := inboundQuantity(requestLine.ExpectedQty, "expected_qty", false)
			if err != nil {
				return err
			}
			orderedNumber, _ := new(big.Rat).SetString(poLine.OrderedQty)
			scheduledText, err := local.repositories.InboundOrderLine.ScheduledForPurchaseOrderLine(ctx, poLine.ID)
			if err != nil {
				return err
			}
			scheduledNumber, ok := new(big.Rat).SetString(scheduledText)
			if !ok || new(big.Rat).Add(scheduledNumber, expectedNumber).Cmp(orderedNumber) > 0 {
				return invalid("inbound expected quantity exceeds unscheduled purchase order quantity")
			}
			customerReference, err := optional(requestLine.CustomerLineReference, 100, "customer_line_reference")
			if err != nil {
				return err
			}
			lineNotes, err := optional(requestLine.Notes, 4000, "line notes")
			if err != nil {
				return err
			}
			lineNo := index + 1
			poLineID := poLine.ID
			line := model.InboundOrderLine{ID: fmt.Sprintf("%s-L%04d", documentID, lineNo), InboundID: documentID, PurchaseOrderLineID: &poLineID, LineNo: lineNo, ItemID: poLine.ItemID, ExpectedQty: expected, UOMID: poLine.UOMID, ExpectedLotNo: poLine.ExpectedLotNo, ExpectedExpiryDate: poLine.ExpectedExpiryDate, CustomerLineReference: customerReference, Notes: lineNotes, CreatedBy: actor}
			if err := local.repositories.InboundOrderLine.Create(ctx, &line); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return dto.InboundOrderResponse{}, err
	}
	return s.GetInboundOrder(ctx, documentID)
}

func (s *Service) ReleaseInboundOrder(ctx context.Context, id string, request dto.TransitionRequest, actor string) (dto.InboundOrderResponse, error) {
	if !inboundID(id, 120) || !inboundUUID(actor) || request.ExpectedVersion < 1 {
		return dto.InboundOrderResponse{}, invalid("invalid inbound-order transition")
	}
	err := s.transaction(ctx, func(local *Service) error {
		header, err := local.repositories.InboundOrder.Lock(ctx, id)
		if err != nil {
			return err
		}
		if header.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		row, err := local.repositories.InboundOrder.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "DRAFT" {
			return state("only a DRAFT inbound order can be released")
		}
		count, err := local.repositories.InboundOrderLine.Count(ctx, id)
		if err != nil {
			return err
		}
		if count == 0 {
			return state("inbound order has no lines")
		}
		target, err := local.transitionTarget(ctx, header.DocumentTypeID, header.StatusID, "RELEASED")
		if err != nil {
			return err
		}
		return local.repositories.InboundOrder.SetStatus(ctx, id, target.ID, actor, header.VersionNo)
	})
	if err != nil {
		return dto.InboundOrderResponse{}, err
	}
	return s.GetInboundOrder(ctx, id)
}
