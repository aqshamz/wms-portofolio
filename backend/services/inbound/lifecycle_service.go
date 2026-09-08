package inbound

import (
	"context"
	"math/big"

	dto "wms-api/dto/inbound"
	inventorydto "wms-api/dto/inventory"
	model "wms-api/models/inbound"
	repository "wms-api/repository/inbound"
	inventoryservice "wms-api/services/inventory"
)

func (s *Service) recordException(ctx context.Context, ownerID, warehouseID, documentID string, lineID *string, kind string, expected, actual, variance, notes *string, actor string) error {
	id, err := newInboundID("IEX")
	if err != nil {
		return err
	}
	return s.repositories.Exception.Create(ctx, &model.InboundException{ID: id, OwnerID: ownerID, WarehouseID: warehouseID, SourceDocumentID: documentID, SourceLineID: lineID, ExceptionTypeCode: kind, ExpectedQty: expected, ActualQty: actual, VarianceQty: variance, Notes: notes, CreatedBy: actor})
}

func (s *Service) GetInboundException(ctx context.Context, id string) (dto.InboundExceptionResponse, error) {
	if !inboundID(id, 140) {
		return dto.InboundExceptionResponse{}, invalid("invalid inbound_exception_id")
	}
	row, err := s.repositories.Exception.Get(ctx, id)
	return mapInboundException(row), err
}

func (s *Service) ListInboundExceptions(ctx context.Context, filter repository.ListFilter) (dto.PageResponse[dto.InboundExceptionResponse], error) {
	if err := validateList(&filter); err != nil {
		return dto.PageResponse[dto.InboundExceptionResponse]{}, err
	}
	rows, total, err := s.repositories.Exception.List(ctx, filter)
	items := make([]dto.InboundExceptionResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapInboundException(row))
	}
	return page(items, filter.Page, filter.PageSize, total), err
}

func validateExceptionTransition(id string, request dto.ExceptionTransitionRequest, actor string) (*string, error) {
	if !inboundID(id, 120) || !inboundUUID(actor) || request.ExpectedVersion < 1 {
		return nil, invalid("invalid document transition")
	}
	return optional(&request.Reason, 4000, "reason")
}

func (s *Service) CancelPurchaseOrder(ctx context.Context, id string, request dto.ExceptionTransitionRequest, actor string) (dto.PurchaseOrderResponse, error) {
	reason, err := validateExceptionTransition(id, request, actor)
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
		if row.StatusCode == "CANCELLED" {
			return nil
		}
		if header.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		if row.StatusCode != "DRAFT" && row.StatusCode != "APPROVED" {
			return state("only a DRAFT or unreceived APPROVED purchase order can be cancelled")
		}
		active, err := local.repositories.PurchaseOrder.CountNonFinalInboundOrders(ctx, id)
		if err != nil {
			return err
		}
		if active != 0 {
			return state("purchase order has an active inbound order")
		}
		target, err := local.transitionTarget(ctx, header.DocumentTypeID, header.StatusID, "CANCELLED")
		if err != nil {
			return err
		}
		if err := local.repositories.PurchaseOrder.SetStatus(ctx, id, target.ID, actor, header.VersionNo); err != nil {
			return err
		}
		return local.recordException(ctx, header.OwnerID, header.WarehouseID, id, nil, "CANCELLATION", nil, nil, nil, reason, actor)
	})
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	return s.GetPurchaseOrder(ctx, id)
}

func (s *Service) CancelInboundOrder(ctx context.Context, id string, request dto.ExceptionTransitionRequest, actor string) (dto.InboundOrderResponse, error) {
	reason, err := validateExceptionTransition(id, request, actor)
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
		if row.StatusCode == "CANCELLED" {
			return nil
		}
		if header.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		if row.StatusCode != "DRAFT" && row.StatusCode != "RELEASED" && row.StatusCode != "PARTIALLY_RECEIVED" {
			return state("inbound order cannot be cancelled from its current state")
		}
		openReceipts, err := local.repositories.InboundOrder.CountReceiptsWithStatus(ctx, id, "OPEN")
		if err != nil {
			return err
		}
		if openReceipts != 0 {
			return state("cancel or complete open receipts before cancelling the inbound order")
		}
		target, err := local.transitionTarget(ctx, header.DocumentTypeID, header.StatusID, "CANCELLED")
		if err != nil {
			return err
		}
		if err := local.repositories.InboundOrder.SetStatus(ctx, id, target.ID, actor, header.VersionNo); err != nil {
			return err
		}
		return local.recordException(ctx, header.OwnerID, header.WarehouseID, id, nil, "CANCELLATION", nil, nil, nil, reason, actor)
	})
	if err != nil {
		return dto.InboundOrderResponse{}, err
	}
	return s.GetInboundOrder(ctx, id)
}

func (s *Service) CancelReceipt(ctx context.Context, id string, request dto.ExceptionTransitionRequest, actor string) (dto.ReceiptResponse, error) {
	reason, err := validateExceptionTransition(id, request, actor)
	if err != nil {
		return dto.ReceiptResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		header, err := local.repositories.Receipt.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.Receipt.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode == "CANCELLED" {
			return nil
		}
		if header.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		if row.StatusCode != "OPEN" {
			return state("only an OPEN receipt can be cancelled")
		}
		target, err := local.transitionTarget(ctx, header.DocumentTypeID, header.StatusID, "CANCELLED")
		if err != nil {
			return err
		}
		if err := local.repositories.Receipt.SetStatus(ctx, id, target.ID, actor, header.VersionNo); err != nil {
			return err
		}
		return local.recordException(ctx, header.OwnerID, header.WarehouseID, id, nil, "CANCELLATION", nil, nil, nil, reason, actor)
	})
	if err != nil {
		return dto.ReceiptResponse{}, err
	}
	return s.GetReceipt(ctx, id)
}

func shortageExceedsTolerance(expected, actual, tolerance *big.Rat) bool {
	if actual.Cmp(expected) >= 0 {
		return false
	}
	allowedShort := new(big.Rat).Mul(expected, new(big.Rat).Quo(tolerance, big.NewRat(100, 1)))
	return new(big.Rat).Sub(expected, actual).Cmp(allowedShort) > 0
}

func (s *Service) CloseInboundOrder(ctx context.Context, id string, request dto.ExceptionTransitionRequest, actor string) (dto.InboundOrderResponse, error) {
	reason, err := validateExceptionTransition(id, request, actor)
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
		if row.StatusCode == "CLOSED" || row.StatusCode == "RECEIVED" {
			return nil
		}
		if header.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		if row.StatusCode != "RELEASED" && row.StatusCode != "PARTIALLY_RECEIVED" {
			return state("only a released inbound order can be closed short")
		}
		openReceipts, err := local.repositories.InboundOrder.CountReceiptsWithStatus(ctx, id, "OPEN")
		if err != nil {
			return err
		}
		if openReceipts != 0 {
			return state("complete or cancel open receipts before closing the inbound order")
		}
		lines, err := local.repositories.InboundOrderLine.List(ctx, id)
		if err != nil {
			return err
		}
		for _, line := range lines {
			expected, e1 := storedRatio(line.ExpectedQty)
			actual, e2 := storedRatio(line.CompletedReceiptQty)
			if e1 != nil || e2 != nil {
				return state("stored inbound quantity is invalid")
			}
			tolerance := new(big.Rat)
			if line.PurchaseOrderLineID != nil {
				poLine, e := local.repositories.PurchaseOrderLine.Get(ctx, *line.PurchaseOrderLineID)
				if e != nil {
					return e
				}
				parsed, ok := new(big.Rat).SetString(poLine.UnderReceiptTolerancePct)
				if !ok {
					return state("stored under-receipt tolerance is invalid")
				}
				tolerance = parsed
			}
			if shortageExceedsTolerance(expected, actual, tolerance) {
				variance := new(big.Rat).Sub(actual, expected).FloatString(6)
				expectedText, actualText := line.ExpectedQty, line.CompletedReceiptQty
				if err := local.recordException(ctx, header.OwnerID, header.WarehouseID, id, &line.ID, "UNDER_RECEIPT", &expectedText, &actualText, &variance, reason, actor); err != nil {
					return err
				}
			}
		}
		target, err := local.transitionTarget(ctx, header.DocumentTypeID, header.StatusID, "CLOSED")
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

func (s *Service) ClosePurchaseOrder(ctx context.Context, id string, request dto.ExceptionTransitionRequest, actor string) (dto.PurchaseOrderResponse, error) {
	reason, err := validateExceptionTransition(id, request, actor)
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
		if row.StatusCode == "CLOSED" || row.StatusCode == "RECEIVED" {
			return nil
		}
		if header.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		if row.StatusCode != "APPROVED" && row.StatusCode != "PARTIALLY_RECEIVED" {
			return state("only an approved purchase order can be closed short")
		}
		active, err := local.repositories.PurchaseOrder.CountNonFinalInboundOrders(ctx, id)
		if err != nil {
			return err
		}
		if active != 0 {
			return state("close or cancel all active inbound orders first")
		}
		lines, err := local.repositories.PurchaseOrderLine.List(ctx, id)
		if err != nil {
			return err
		}
		for _, line := range lines {
			expected, e1 := storedRatio(line.OrderedQty)
			actual, e2 := storedRatio(line.CompletedReceiptQty)
			tolerance, ok := new(big.Rat).SetString(line.UnderReceiptTolerancePct)
			if e1 != nil || e2 != nil || !ok {
				return state("stored purchase-order quantity is invalid")
			}
			if shortageExceedsTolerance(expected, actual, tolerance) {
				variance := new(big.Rat).Sub(actual, expected).FloatString(6)
				expectedText, actualText := line.OrderedQty, line.CompletedReceiptQty
				if err := local.recordException(ctx, header.OwnerID, header.WarehouseID, id, &line.ID, "UNDER_RECEIPT", &expectedText, &actualText, &variance, reason, actor); err != nil {
					return err
				}
			}
		}
		target, err := local.transitionTarget(ctx, header.DocumentTypeID, header.StatusID, "CLOSED")
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

func (s *Service) ReverseReceipt(ctx context.Context, id string, request dto.ReverseReceiptRequest, actor string) (dto.ReceiptResponse, error) {
	if !inboundID(id, 120) || !inboundUUID(actor) || request.ExpectedVersion < 1 {
		return dto.ReceiptResponse{}, invalid("invalid receipt reversal")
	}
	if _, err := inboundDate(request.BusinessDate, "business_date"); err != nil {
		return dto.ReceiptResponse{}, err
	}
	reason, err := optional(&request.Reason, 4000, "reason")
	if err != nil {
		return dto.ReceiptResponse{}, err
	}
	versions := make(map[string]int64, len(request.Balances))
	for _, balance := range request.Balances {
		if !inboundID(balance.BalanceID, 160) || balance.ExpectedVersion < 1 || versions[balance.BalanceID] != 0 {
			return dto.ReceiptResponse{}, invalid("balances must contain unique valid balance IDs and versions")
		}
		versions[balance.BalanceID] = balance.ExpectedVersion
	}
	err = s.transaction(ctx, func(local *Service) error {
		header, err := local.repositories.Receipt.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.Receipt.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode == "REVERSED" {
			return nil
		}
		if header.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		if row.StatusCode != "COMPLETED" {
			return state("only a COMPLETED receipt can be reversed")
		}
		if header.InboundID == nil {
			return state("receipt has no inbound order")
		}
		inboundRow, err := local.repositories.InboundOrder.Get(ctx, *header.InboundID)
		if err != nil {
			return err
		}
		if inboundRow.StatusCode == "CLOSED" || inboundRow.StatusCode == "CANCELLED" {
			return state("a receipt cannot be reversed after its inbound order is closed")
		}
		lines, err := local.repositories.ReceiptLine.List(ctx, id)
		if err != nil {
			return err
		}
		usedBalances := make(map[string]bool)
		inventory := inventoryservice.NewService(local.repositories.Inventory)
		for _, line := range lines {
			batches, err := local.repositories.ReceiptInventory.ListByLine(ctx, line.ID)
			if err != nil {
				return err
			}
			for _, batch := range batches {
				if batch.InitialBalanceID == nil {
					return state("receipt batch has no posted balance")
				}
				inspectionCount, err := local.repositories.QualityInspection.CountByBatch(ctx, batch.ID)
				if err != nil {
					return err
				}
				if inspectionCount != 0 {
					return state("receipt cannot be reversed after quality inspection has started")
				}
				expectedVersion, ok := versions[*batch.InitialBalanceID]
				if !ok {
					return invalid("an expected version is required for every affected balance")
				}
				balance, err := local.repositories.Inventory.Balance.Get(ctx, *batch.InitialBalanceID)
				if err != nil || balance.InventoryStatusCode != "QC_PENDING" || balance.LocationID != batch.ReceivedLocationID {
					return state("receipt stock is no longer untouched in QC_PENDING")
				}
				if balance.VersionNo != expectedVersion {
					return repository.ErrConcurrentWrite
				}
				serialIDs := make([]string, 0, 1)
				if batch.SerialID != nil {
					serialIDs = append(serialIDs, *batch.SerialID)
				}
				posted, err := inventory.PostMovement(ctx, inventorydto.PostingRequest{OperationKey: "inbound.receipt-reversal." + batch.ID, MovementTypeCode: "RECEIPT_REVERSAL", OwnerID: header.OwnerID, WarehouseID: header.WarehouseID, BusinessDate: request.BusinessDate, ItemID: batch.ItemID, LotID: batch.LotID, HandlingUnitID: batch.HandlingUnitID, From: &inventorydto.BalanceDimension{LocationID: balance.LocationID, InventoryStatusID: balance.InventoryStatusID}, Quantity: batch.BaseQty, SerialIDs: serialIDs, ExpectedSourceVersion: &expectedVersion, SourceDocumentID: header.ID, SourceLineID: &batch.ID, Notes: reason}, actor)
				if err != nil {
					return err
				}
				if posted.FromBalance == nil {
					return state("receipt reversal returned no source balance")
				}
				versions[*batch.InitialBalanceID] = posted.FromBalance.VersionNo
				usedBalances[*batch.InitialBalanceID] = true
			}
		}
		if len(usedBalances) != len(versions) {
			return invalid("balances contains an entry that is not used by this receipt")
		}
		reversed, err := local.documentStatus(ctx, header.DocumentTypeID, "REVERSED")
		if err != nil {
			return err
		}
		if err := local.repositories.Receipt.SetStatus(ctx, id, reversed.ID, actor, header.VersionNo); err != nil {
			return err
		}
		if err := local.recordException(ctx, header.OwnerID, header.WarehouseID, id, nil, "REVERSAL", nil, nil, nil, reason, actor); err != nil {
			return err
		}
		return local.updateReceiptProgress(ctx, header, actor)
	})
	if err != nil {
		return dto.ReceiptResponse{}, err
	}
	return s.GetReceipt(ctx, id)
}
