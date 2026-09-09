package inbound

import (
	"context"
	"fmt"
	"strings"

	dto "wms-api/dto/inbound"
	model "wms-api/models/inbound"
	repository "wms-api/repository/inbound"
)

func (s *Service) CreatePurchaseOrder(ctx context.Context, request dto.CreatePurchaseOrderRequest, actor string) (dto.PurchaseOrderResponse, error) {
	if !inboundUUID(request.OwnerID) || !inboundUUID(request.VendorID) || !inboundUUID(request.WarehouseID) || !inboundUUID(actor) {
		return dto.PurchaseOrderResponse{}, invalid("owner_id, vendor_id, warehouse_id and actor must be UUIDs")
	}
	request.OwnerID = strings.ToLower(request.OwnerID)
	request.VendorID = strings.ToLower(request.VendorID)
	request.WarehouseID = strings.ToLower(request.WarehouseID)
	poNumber, err := clean(request.PurchaseOrderNo, 120, "purchase_order_no")
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	businessDate, err := inboundDate(request.BusinessDate, "business_date")
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
	supersedes, err := optional(request.SupersedesPurchaseOrderID, 120, "supersedes_purchase_order_id")
	if err != nil || (supersedes != nil && !inboundID(*supersedes, 120)) {
		return dto.PurchaseOrderResponse{}, invalid("invalid supersedes_purchase_order_id")
	}
	if len(request.Lines) == 0 || len(request.Lines) > 500 {
		return dto.PurchaseOrderResponse{}, invalid("lines must contain 1..500 entries")
	}
	var documentID string
	err = s.transaction(ctx, func(local *Service) error {
		owner, err := local.repositories.Master.Organization.FindByID(ctx, request.OwnerID)
		if err != nil || !owner.IsActive {
			return invalid("owner does not exist or is inactive")
		}
		vendor, err := local.repositories.Master.Catalog.BusinessPartner.Get(ctx, request.VendorID)
		if err != nil || !vendor.IsActive || vendor.OwnerID != request.OwnerID {
			return invalid("vendor does not belong to the active owner")
		}
		types, err := local.repositories.Master.Catalog.BusinessPartnerType.List(ctx, vendor.ID)
		if err != nil {
			return err
		}
		validVendor := false
		for _, partnerType := range types {
			if partnerType.IsActive && (partnerType.Code == "SUPPLIER" || partnerType.Code == "FACTORY") {
				validVendor = true
			}
		}
		if !validVendor {
			return invalid("vendor requires an active SUPPLIER or FACTORY type")
		}
		warehouse, err := local.repositories.Master.Warehouse.FindByID(ctx, request.WarehouseID)
		if err != nil || !warehouse.IsActive {
			return invalid("warehouse does not exist or is inactive")
		}
		assignment, err := local.repositories.Master.WarehouseOwner.GetShared(ctx, request.WarehouseID, request.OwnerID)
		if err != nil || !assignment.IsActive {
			return invalid("owner is not active for the warehouse")
		}
		kind, err := local.documentType(ctx, "PURCHASE_ORDER")
		if err != nil {
			return err
		}
		initial, err := local.initialStatus(ctx, kind.ID)
		if err != nil {
			return err
		}
		documentID, err = local.generateID(ctx, kind.ID, request.BusinessDate, request.VendorID, request.WarehouseID)
		if err != nil {
			return err
		}
		if supersedes != nil {
			prior, lookupErr := local.repositories.PurchaseOrder.Get(ctx, *supersedes)
			if lookupErr != nil || prior.StatusCode != "CANCELLED" || prior.OwnerID != request.OwnerID || prior.VendorID != request.VendorID || prior.WarehouseID != request.WarehouseID {
				return invalid("superseded purchase order must be a cancelled document for the same owner, vendor, and warehouse")
			}
			hasSuccessor, lookupErr := local.repositories.PurchaseOrder.HasSuccessor(ctx, *supersedes)
			if lookupErr != nil {
				return lookupErr
			}
			if hasSuccessor {
				return state("cancelled purchase order already has a successor")
			}
		}
		header := model.PurchaseOrder{ID: documentID, DocumentTypeID: kind.ID, StatusID: initial.ID, OwnerID: request.OwnerID, VendorID: request.VendorID, WarehouseID: request.WarehouseID, BusinessDate: businessDate, PurchaseOrderNo: poNumber, OrderedAt: orderedAt, ExpectedArrivalAt: expectedArrival, Notes: notes, SupersedesPurchaseOrderID: supersedes, CreatedBy: actor, UpdatedBy: &actor, VersionNo: 1}
		if err := local.repositories.PurchaseOrder.Create(ctx, &header); err != nil {
			return err
		}
		for index, requestLine := range request.Lines {
			if !inboundUUID(requestLine.ItemID) || !inboundUUID(requestLine.UOMID) {
				return invalid("line item_id and uom_id must be UUIDs")
			}
			quantity, _, err := inboundQuantity(requestLine.OrderedQty, "ordered_qty", false)
			if err != nil {
				return err
			}
			overTolerance, _, err := inboundPercentage(requestLine.OverReceiptTolerancePct, "over_receipt_tolerance_pct")
			if err != nil {
				return err
			}
			underTolerance, _, err := inboundPercentage(requestLine.UnderReceiptTolerancePct, "under_receipt_tolerance_pct")
			if err != nil {
				return err
			}
			item, err := local.repositories.Master.Catalog.Item.Get(ctx, strings.ToLower(requestLine.ItemID))
			if err != nil || !item.IsActive || item.OwnerID != request.OwnerID {
				return invalid("line item does not belong to the active owner")
			}
			unit, err := local.repositories.Master.Catalog.ItemUOM.ForUnit(ctx, item.ID, strings.ToLower(requestLine.UOMID))
			if err != nil || !unit.IsActive || !unit.IsReceivingUOM {
				return invalid("line UOM is not an active receiving UOM for the item")
			}
			expiry, err := inboundOptionalDate(requestLine.ExpectedExpiryDate, "expected_expiry_date")
			if err != nil {
				return err
			}
			vendorCode, err := optional(requestLine.VendorItemCode, 100, "vendor_item_code")
			if err != nil {
				return err
			}
			expectedLot, err := optional(requestLine.ExpectedLotNo, 100, "expected_lot_no")
			if err != nil {
				return err
			}
			lineNotes, err := optional(requestLine.Notes, 4000, "line notes")
			if err != nil {
				return err
			}
			lineNo := index + 1
			line := model.PurchaseOrderLine{ID: fmt.Sprintf("%s-L%04d", documentID, lineNo), PurchaseOrderID: documentID, OwnerID: request.OwnerID, LineNo: lineNo, ItemID: item.ID, OrderedQty: quantity, OverReceiptTolerancePct: overTolerance, UnderReceiptTolerancePct: underTolerance, UOMID: unit.UOMID, VendorItemCode: vendorCode, ExpectedLotNo: expectedLot, ExpectedExpiryDate: expiry, Notes: lineNotes, CreatedBy: actor}
			if err := local.repositories.PurchaseOrderLine.Create(ctx, &line); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	return s.GetPurchaseOrder(ctx, documentID)
}

func (s *Service) ApprovePurchaseOrder(ctx context.Context, id string, request dto.TransitionRequest, actor string) (dto.PurchaseOrderResponse, error) {
	if !inboundID(id, 120) || !inboundUUID(actor) || request.ExpectedVersion < 1 {
		return dto.PurchaseOrderResponse{}, invalid("invalid purchase order transition")
	}
	err := s.transaction(ctx, func(local *Service) error {
		header, err := local.repositories.PurchaseOrder.Lock(ctx, id)
		if err != nil {
			return err
		}
		if header.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		row, err := local.repositories.PurchaseOrder.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "DRAFT" {
			return state("only a DRAFT purchase order can be approved")
		}
		count, err := local.repositories.PurchaseOrderLine.Count(ctx, id)
		if err != nil {
			return err
		}
		if count == 0 {
			return state("purchase order has no lines")
		}
		target, err := local.transitionTarget(ctx, header.DocumentTypeID, header.StatusID, "APPROVED")
		if err != nil {
			return err
		}
		return local.repositories.PurchaseOrder.SetStatus(ctx, id, target.ID, actor, header.VersionNo)
	})
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	return s.GetPurchaseOrder(ctx, id)
}
