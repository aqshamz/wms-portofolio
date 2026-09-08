package inbound

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	dto "wms-api/dto/inbound"
	inventorydto "wms-api/dto/inventory"
	model "wms-api/models/inbound"
	mastermodel "wms-api/models/master"
	repository "wms-api/repository/inbound"
	inventoryrepository "wms-api/repository/inventory"
	masterrepository "wms-api/repository/master"
	inventoryservice "wms-api/services/inventory"
)

func (s *Service) inventoryStatus(ctx context.Context, code string) (mastermodel.InventoryStatus, error) {
	active := true
	rows, _, err := s.repositories.Master.Catalog.InventoryStatus.List(ctx, masterrepository.CatalogFilter{Search: code, Active: &active, Page: 1, PageSize: 100})
	if err != nil {
		return mastermodel.InventoryStatus{}, err
	}
	for _, row := range rows {
		if row.Code == code {
			return row, nil
		}
	}
	return mastermodel.InventoryStatus{}, state("inventory status " + code + " is not configured")
}

func (s *Service) qualityStatus(ctx context.Context, code string) (mastermodel.QualityStatus, error) {
	active := true
	rows, _, err := s.repositories.Master.Catalog.QualityStatus.List(ctx, masterrepository.CatalogFilter{Search: code, Active: &active, Page: 1, PageSize: 100})
	if err != nil {
		return mastermodel.QualityStatus{}, err
	}
	for _, row := range rows {
		if row.Code == code {
			return row, nil
		}
	}
	return mastermodel.QualityStatus{}, state("quality status " + code + " is not configured")
}

func (s *Service) ensureReceiptLot(ctx context.Context, ownerID string, item mastermodel.Item, request *dto.ReceiptLotRequest, actor string) (*string, error) {
	if !item.LotControlled {
		if request != nil {
			return nil, invalid("lot data is not allowed for a non-lot-controlled item")
		}
		return nil, nil
	}
	if request == nil {
		return nil, invalid("lot data is required for a lot-controlled item")
	}
	lotNumber, err := clean(request.LotNumber, 100, "lot_number")
	if err != nil {
		return nil, err
	}
	rows, total, err := s.repositories.Inventory.Lot.List(ctx, inventoryrepository.Filter{OwnerID: ownerID, ItemID: item.ID, Number: lotNumber, Page: 1, PageSize: 2})
	if err != nil {
		return nil, err
	}
	if total > 1 {
		return nil, repository.ErrConflict
	}
	if len(rows) == 1 {
		return &rows[0].ID, nil
	}
	pending, err := s.qualityStatus(ctx, "PENDING")
	if err != nil {
		return nil, err
	}
	created, err := inventoryservice.NewService(s.repositories.Inventory).CreateLot(ctx, inventorydto.CreateLotRequest{OwnerID: ownerID, ItemID: item.ID, LotNumber: lotNumber, ManufactureDate: request.ManufactureDate, ExpiryDate: request.ExpiryDate, QualityStatusID: &pending.ID}, actor)
	if err != nil {
		return nil, err
	}
	return &created.ID, nil
}

func (s *Service) ensureReceiptSerial(ctx context.Context, ownerID string, item mastermodel.Item, serialNo *string, actor string) (*string, error) {
	if !item.SerialControlled {
		if serialNo != nil {
			return nil, invalid("serial_no is not allowed for a non-serialized item")
		}
		return nil, nil
	}
	if serialNo == nil {
		return nil, invalid("serial_no is required for a serialized item")
	}
	number, err := clean(*serialNo, 120, "serial_no")
	if err != nil {
		return nil, err
	}
	rows, total, err := s.repositories.Inventory.Serial.List(ctx, inventoryrepository.Filter{OwnerID: ownerID, ItemID: item.ID, Number: number, Page: 1, PageSize: 2})
	if err != nil {
		return nil, err
	}
	if total > 1 {
		return nil, repository.ErrConflict
	}
	if len(rows) == 1 {
		return &rows[0].ID, nil
	}
	created, err := inventoryservice.NewService(s.repositories.Inventory).CreateSerial(ctx, inventorydto.CreateSerialRequest{OwnerID: ownerID, ItemID: item.ID, SerialNo: number}, actor)
	if err != nil {
		return nil, err
	}
	return &created.ID, nil
}

func (s *Service) CreateReceipt(ctx context.Context, request dto.CreateReceiptRequest, actor string) (dto.ReceiptResponse, error) {
	if !inboundID(request.InboundID, 120) || !inboundUUID(request.DockLocationID) || !inboundUUID(actor) {
		return dto.ReceiptResponse{}, invalid("invalid inbound_id, dock_location_id or actor")
	}
	if len(request.Lines) == 0 || len(request.Lines) > 500 {
		return dto.ReceiptResponse{}, invalid("lines must contain 1..500 entries")
	}
	businessDate, err := inboundDate(request.BusinessDate, "business_date")
	if err != nil {
		return dto.ReceiptResponse{}, err
	}
	receivedAt, err := inboundTimestamp(request.ReceivedAt, "received_at")
	if err != nil {
		return dto.ReceiptResponse{}, err
	}
	vehicle, err := optional(request.VehicleNumber, 60, "vehicle_number")
	if err != nil {
		return dto.ReceiptResponse{}, err
	}
	seal, err := optional(request.SealNumber, 60, "seal_number")
	if err != nil {
		return dto.ReceiptResponse{}, err
	}
	deliveryNote, err := optional(request.DeliveryNoteNo, 100, "delivery_note_no")
	if err != nil {
		return dto.ReceiptResponse{}, err
	}
	notes, err := optional(request.Notes, 4000, "notes")
	if err != nil {
		return dto.ReceiptResponse{}, err
	}
	seenLines := make(map[string]bool, len(request.Lines))
	var receiptID string
	err = s.transaction(ctx, func(local *Service) error {
		inbound, err := local.repositories.InboundOrder.Lock(ctx, request.InboundID)
		if err != nil {
			return err
		}
		inboundRow, err := local.repositories.InboundOrder.Get(ctx, inbound.ID)
		if err != nil {
			return err
		}
		if inboundRow.StatusCode != "RELEASED" && inboundRow.StatusCode != "PARTIALLY_RECEIVED" {
			return state("inbound order must be RELEASED or PARTIALLY_RECEIVED")
		}
		dock, err := local.repositories.Inventory.Location.GetShared(ctx, strings.ToLower(request.DockLocationID))
		if err != nil || dock.WarehouseID != inbound.WarehouseID || !dock.IsActive || dock.IsLocked {
			return invalid("dock location is unavailable or belongs to another warehouse")
		}
		dockType, err := local.repositories.Master.LocationType.GetShared(ctx, dock.LocationTypeID)
		if err != nil || !dockType.IsActive || !dockType.AllowsReceiving {
			return invalid("dock location type does not allow receiving")
		}
		kind, err := local.documentType(ctx, "RECEIPT")
		if err != nil {
			return err
		}
		initial, err := local.initialStatus(ctx, kind.ID)
		if err != nil {
			return err
		}
		receiptID, err = local.generateID(ctx, kind.ID, request.BusinessDate, inbound.VendorID, inbound.WarehouseID)
		if err != nil {
			return err
		}
		headerInboundID := inbound.ID
		dockID := dock.ID
		header := model.Receipt{ID: receiptID, DocumentTypeID: kind.ID, StatusID: initial.ID, InboundID: &headerInboundID, OwnerID: inbound.OwnerID, WarehouseID: inbound.WarehouseID, BusinessDate: businessDate, ReceivedAt: receivedAt, DockLocationID: &dockID, VehicleNumber: vehicle, SealNumber: seal, DeliveryNoteNo: deliveryNote, Notes: notes, CreatedBy: actor, UpdatedBy: &actor, VersionNo: 1}
		if err := local.repositories.Receipt.Create(ctx, &header); err != nil {
			return err
		}
		qcPending, err := local.inventoryStatus(ctx, "QC_PENDING")
		if err != nil {
			return err
		}
		for lineIndex, requestLine := range request.Lines {
			lineID := strings.TrimSpace(requestLine.InboundLineID)
			if !inboundIDPattern.MatchString(lineID) || len(lineID) > 150 || seenLines[lineID] {
				return invalid("inbound_line_id is invalid or repeated")
			}
			seenLines[lineID] = true
			inboundLine, err := local.repositories.InboundOrderLine.Get(ctx, lineID)
			if err != nil || inboundLine.InboundID != inbound.ID {
				return invalid("receipt line does not belong to the inbound order")
			}
			receivedQty, receivedNumber, err := inboundQuantity(requestLine.ReceivedQty, "received_qty", false)
			if err != nil {
				return err
			}
			rejectedQty, rejectedNumber, err := inboundQuantity(requestLine.RejectedQty, "rejected_qty", true)
			if err != nil {
				return err
			}
			if rejectedNumber.Cmp(receivedNumber) > 0 {
				return invalid("rejected_qty cannot exceed received_qty")
			}
			exceptionNotes, err := optional(requestLine.ExceptionNotes, 4000, "exception_notes")
			if err != nil {
				return err
			}
			var exceptionType *string
			if requestLine.ExceptionTypeCode != nil {
				code := strings.ToUpper(strings.TrimSpace(*requestLine.ExceptionTypeCode))
				if code != "REJECTED_AT_DOCK" && code != "DAMAGED" && code != "WRONG_ITEM" {
					return invalid("exception_type_code must be REJECTED_AT_DOCK, DAMAGED, or WRONG_ITEM")
				}
				exceptionType = &code
			}
			if rejectedNumber.Sign() == 0 && exceptionType != nil {
				return invalid("exception_type_code requires rejected_qty")
			}
			if rejectedNumber.Sign() > 0 && exceptionType == nil {
				code := "REJECTED_AT_DOCK"
				exceptionType = &code
			}
			alreadyReceivedText, err := local.repositories.ReceiptLine.ReceivedForInboundLine(ctx, inboundLine.ID)
			if err != nil {
				return err
			}
			alreadyReceived, ok := new(big.Rat).SetString(alreadyReceivedText)
			expected, expectedOK := new(big.Rat).SetString(inboundLine.ExpectedQty)
			if !ok || !expectedOK {
				return state("stored inbound receipt quantity is invalid")
			}
			tolerance := new(big.Rat)
			if inboundLine.PurchaseOrderLineID != nil {
				poLine, lookupErr := local.repositories.PurchaseOrderLine.Get(ctx, *inboundLine.PurchaseOrderLineID)
				if lookupErr != nil {
					return lookupErr
				}
				if parsed, parsedOK := new(big.Rat).SetString(poLine.OverReceiptTolerancePct); parsedOK {
					tolerance = parsed
				} else {
					return state("stored over-receipt tolerance is invalid")
				}
			}
			allowed := new(big.Rat).Mul(expected, new(big.Rat).Add(big.NewRat(1, 1), new(big.Rat).Quo(tolerance, big.NewRat(100, 1))))
			newTotal := new(big.Rat).Add(alreadyReceived, receivedNumber)
			if newTotal.Cmp(allowed) > 0 {
				return invalid("received_qty exceeds the configured over-receipt tolerance")
			}
			if (newTotal.Cmp(expected) > 0 || rejectedNumber.Sign() > 0) && exceptionNotes == nil {
				return invalid("exception_notes is required for rejected or over-received quantity")
			}
			accepted := new(big.Rat).Sub(receivedNumber, rejectedNumber)
			if accepted.Sign() > 0 && len(requestLine.Batches) == 0 {
				return invalid("batches are required when the receipt line has accepted quantity")
			}
			if accepted.Sign() == 0 && len(requestLine.Batches) != 0 {
				return invalid("batches must be empty when the full receipt line is rejected")
			}
			if len(requestLine.Batches) > 1000 {
				return invalid("batches cannot contain more than 1000 entries")
			}
			batchNumbers := make([]*big.Rat, len(requestLine.Batches))
			batchTotal := new(big.Rat)
			for batchIndex, batch := range requestLine.Batches {
				_, number, err := inboundQuantity(batch.SourceQty, "source_qty", false)
				if err != nil {
					return err
				}
				batchNumbers[batchIndex] = number
				batchTotal.Add(batchTotal, number)
			}
			if batchTotal.Cmp(accepted) != 0 {
				return invalid("batch source quantities must equal received_qty minus rejected_qty")
			}
			item, err := local.repositories.Master.Catalog.Item.Get(ctx, inboundLine.ItemID)
			if err != nil || !item.IsActive || item.OwnerID != inbound.OwnerID {
				return invalid("receipt item is unavailable")
			}
			unit, err := local.repositories.Master.Catalog.ItemUOM.ForUnit(ctx, item.ID, inboundLine.UOMID)
			if err != nil || !unit.IsActive || !unit.IsReceivingUOM {
				return invalid("receipt UOM is unavailable")
			}
			conversion, ok := new(big.Rat).SetString(unit.ConversionToBase)
			if !ok || conversion.Sign() <= 0 {
				return state("stored item UOM conversion is invalid")
			}
			lineNo := lineIndex + 1
			receiptLineID := fmt.Sprintf("%s-L%04d", receiptID, lineNo)
			inboundLineID := inboundLine.ID
			receiptLine := model.ReceiptLine{ID: receiptLineID, ReceiptID: receiptID, InboundLineID: &inboundLineID, LineNo: lineNo, ItemID: item.ID, ReceivedQty: receivedQty, RejectedQty: rejectedQty, ExceptionNotes: exceptionNotes, ExceptionTypeCode: exceptionType, UOMID: inboundLine.UOMID, CreatedBy: actor}
			if err := local.repositories.ReceiptLine.Create(ctx, &receiptLine); err != nil {
				return err
			}
			for batchIndex, requestBatch := range requestLine.Batches {
				sourceQty := batchNumbers[batchIndex].FloatString(6)
				baseQty, baseNumber, err := inboundMultiply(batchNumbers[batchIndex], conversion)
				if err != nil {
					return err
				}
				location, err := local.repositories.Inventory.Location.GetShared(ctx, strings.ToLower(requestBatch.ReceivedLocationID))
				if err != nil || location.WarehouseID != inbound.WarehouseID || !location.IsActive || location.IsLocked {
					return invalid("received location is unavailable or belongs to another warehouse")
				}
				lotID, err := local.ensureReceiptLot(ctx, inbound.OwnerID, item, requestBatch.Lot, actor)
				if err != nil {
					return err
				}
				if lotID != nil && item.MinimumReceiveDays != nil {
					lot, err := local.repositories.Inventory.Lot.Get(ctx, *lotID)
					if err != nil {
						return err
					}
					minimumExpiry := businessDate.AddDate(0, 0, *item.MinimumReceiveDays)
					if lot.ExpiryDate == nil || lot.ExpiryDate.Before(minimumExpiry) {
						return invalid("lot expiry does not meet the item's minimum receive days")
					}
				}
				serialID, err := local.ensureReceiptSerial(ctx, inbound.OwnerID, item, requestBatch.SerialNo, actor)
				if err != nil {
					return err
				}
				if item.SerialControlled && baseNumber.Cmp(big.NewRat(1, 1)) != 0 {
					return invalid("each serialized receipt batch must convert to exactly 1 base unit")
				}
				var handlingUnitID *string
				if requestBatch.HandlingUnitID != nil {
					id := strings.TrimSpace(*requestBatch.HandlingUnitID)
					if !inboundID(id, 120) {
						return invalid("invalid handling_unit_id")
					}
					handlingUnit, err := local.repositories.Inventory.HandlingUnit.Get(ctx, id)
					if err != nil || handlingUnit.OwnerID != inbound.OwnerID || handlingUnit.WarehouseID != inbound.WarehouseID || handlingUnit.CurrentLocationID == nil || *handlingUnit.CurrentLocationID != location.ID || handlingUnit.IsClosed {
						return invalid("handling unit is unavailable or is not at the received location")
					}
					handlingUnitID = &id
				}
				batchNo := batchIndex + 1
				batch := model.ReceiptInventory{ID: fmt.Sprintf("%s-B%04d", receiptLineID, batchNo), ReceiptLineID: receiptLineID, ItemID: item.ID, SourceQty: sourceQty, SourceUOMID: inboundLine.UOMID, BaseQty: baseQty, BaseUOMID: item.BaseUOMID, LotID: lotID, HandlingUnitID: handlingUnitID, ReceivedLocationID: location.ID, InitialInventoryStatusID: qcPending.ID, CreatedBy: actor}
				if err := local.repositories.ReceiptInventory.Create(ctx, &batch); err != nil {
					return err
				}
				if serialID != nil {
					if err := local.repositories.ReceiptLineSerial.Create(ctx, &model.ReceiptLineSerial{ReceiptInventoryID: batch.ID, SerialID: *serialID}); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return dto.ReceiptResponse{}, err
	}
	return s.GetReceipt(ctx, receiptID)
}

func (s *Service) CompleteReceipt(ctx context.Context, id string, request dto.TransitionRequest, actor string) (dto.ReceiptResponse, error) {
	if !inboundID(id, 120) || !inboundUUID(actor) || request.ExpectedVersion < 1 {
		return dto.ReceiptResponse{}, invalid("invalid receipt completion")
	}
	err := s.transaction(ctx, func(local *Service) error {
		header, err := local.repositories.Receipt.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.Receipt.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode == "COMPLETED" {
			return nil
		}
		if row.StatusCode != "OPEN" {
			return state("only an OPEN receipt can be completed")
		}
		if header.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		lines, err := local.repositories.ReceiptLine.List(ctx, id)
		if err != nil {
			return err
		}
		if len(lines) == 0 {
			return state("receipt has no lines")
		}
		inventory := inventoryservice.NewService(local.repositories.Inventory)
		for _, line := range lines {
			accepted, acceptedOK := new(big.Rat).SetString(line.AcceptedQty)
			batched, batchedOK := new(big.Rat).SetString(line.BatchedQty)
			if !acceptedOK || !batchedOK || accepted.Cmp(batched) != 0 {
				return state("receipt batches do not equal accepted receipt quantity")
			}
			if rejected, parseErr := storedRatio(line.RejectedQty); parseErr != nil {
				return parseErr
			} else if rejected.Sign() > 0 {
				expectedQty, actualQty, varianceQty := line.ReceivedQty, line.AcceptedQty, new(big.Rat).Neg(rejected).FloatString(6)
				exceptionType := "REJECTED_AT_DOCK"
				if line.ExceptionTypeCode != nil {
					exceptionType = *line.ExceptionTypeCode
				}
				if err := local.recordException(ctx, header.OwnerID, header.WarehouseID, header.ID, &line.ID, exceptionType, &expectedQty, &actualQty, &varianceQty, line.ExceptionNotes, actor); err != nil {
					return err
				}
			}
			if line.InboundLineID != nil {
				inboundLine, lookupErr := local.repositories.InboundOrderLine.Get(ctx, *line.InboundLineID)
				if lookupErr != nil {
					return lookupErr
				}
				priorText, lookupErr := local.repositories.ReceiptLine.CompletedBeforeReceipt(ctx, inboundLine.ID, header.ID)
				if lookupErr != nil {
					return lookupErr
				}
				prior, parseOK := new(big.Rat).SetString(priorText)
				expected, expectedOK := new(big.Rat).SetString(inboundLine.ExpectedQty)
				current, currentOK := new(big.Rat).SetString(line.ReceivedQty)
				if !parseOK || !expectedOK || !currentOK {
					return state("stored receipt quantity is invalid")
				}
				priorOver := new(big.Rat).Sub(prior, expected)
				if priorOver.Sign() < 0 {
					priorOver.SetInt64(0)
				}
				newOver := new(big.Rat).Sub(new(big.Rat).Add(prior, current), expected)
				if newOver.Sign() > 0 {
					increment := new(big.Rat).Sub(newOver, priorOver)
					if increment.Sign() > 0 {
						expectedText, actualText, varianceText := inboundLine.ExpectedQty, new(big.Rat).Add(prior, current).FloatString(6), increment.FloatString(6)
						if err := local.recordException(ctx, header.OwnerID, header.WarehouseID, header.ID, &line.ID, "OVER_RECEIPT", &expectedText, &actualText, &varianceText, line.ExceptionNotes, actor); err != nil {
							return err
						}
					}
				}
			}
			batches, err := local.repositories.ReceiptInventory.ListByLine(ctx, line.ID)
			if err != nil {
				return err
			}
			for _, batch := range batches {
				if batch.InitialBalanceID != nil {
					return state("OPEN receipt batch is already linked to inventory")
				}
				serialIDs := make([]string, 0, 1)
				if batch.SerialID != nil {
					serialIDs = append(serialIDs, *batch.SerialID)
				}
				posted, err := inventory.PostMovement(ctx, inventorydto.PostingRequest{OperationKey: "inbound.receipt." + batch.ID, MovementTypeCode: "RECEIVE", OwnerID: header.OwnerID, WarehouseID: header.WarehouseID, BusinessDate: header.BusinessDate.Format("2006-01-02"), ItemID: batch.ItemID, LotID: batch.LotID, HandlingUnitID: batch.HandlingUnitID, To: &inventorydto.BalanceDimension{LocationID: batch.ReceivedLocationID, InventoryStatusID: batch.InitialInventoryStatusID}, Quantity: batch.BaseQty, SerialIDs: serialIDs, SourceDocumentID: header.ID, SourceLineID: &batch.ID}, actor)
				if err != nil {
					return err
				}
				if posted.ToBalance == nil {
					return state("inventory posting did not return a destination balance")
				}
				if err := local.repositories.ReceiptInventory.SetInitialBalance(ctx, batch.ID, posted.ToBalance.ID); err != nil {
					return err
				}
			}
		}
		target, err := local.transitionTarget(ctx, header.DocumentTypeID, header.StatusID, "COMPLETED")
		if err != nil {
			return err
		}
		if err := local.repositories.Receipt.SetStatus(ctx, id, target.ID, actor, header.VersionNo); err != nil {
			return err
		}
		return local.updateReceiptProgress(ctx, header, actor)
	})
	if err != nil {
		return dto.ReceiptResponse{}, err
	}
	return s.GetReceipt(ctx, id)
}

func storedRatio(value string) (*big.Rat, error) {
	result, ok := new(big.Rat).SetString(value)
	if !ok {
		return nil, state("stored quantity is invalid")
	}
	return result, nil
}

func (s *Service) updateReceiptProgress(ctx context.Context, receipt model.Receipt, actor string) error {
	if receipt.InboundID == nil {
		return state("receipt has no inbound order")
	}
	inbound, err := s.repositories.InboundOrder.Lock(ctx, *receipt.InboundID)
	if err != nil {
		return err
	}
	inboundRow, err := s.repositories.InboundOrder.Get(ctx, inbound.ID)
	if err != nil {
		return err
	}
	inboundLines, err := s.repositories.InboundOrderLine.List(ctx, inbound.ID)
	if err != nil {
		return err
	}
	expected := new(big.Rat)
	received := new(big.Rat)
	poIDs := make(map[string]bool)
	for _, line := range inboundLines {
		lineExpected, err := storedRatio(line.ExpectedQty)
		if err != nil {
			return err
		}
		lineReceived, err := storedRatio(line.CompletedReceiptQty)
		if err != nil {
			return err
		}
		expected.Add(expected, lineExpected)
		received.Add(received, lineReceived)
		if line.PurchaseOrderLineID != nil {
			poLine, err := s.repositories.PurchaseOrderLine.Get(ctx, *line.PurchaseOrderLineID)
			if err != nil {
				return err
			}
			poIDs[poLine.PurchaseOrderID] = true
		}
	}
	targetCode := "PARTIALLY_RECEIVED"
	if received.Sign() == 0 {
		targetCode = "RELEASED"
	} else if received.Cmp(expected) >= 0 {
		targetCode = "RECEIVED"
	}
	if inboundRow.StatusCode != targetCode {
		var target mastermodel.DocumentStatus
		if inboundRow.StatusCode == "RECEIVED" {
			target, err = s.documentStatus(ctx, inbound.DocumentTypeID, targetCode)
		} else {
			target, err = s.transitionTarget(ctx, inbound.DocumentTypeID, inbound.StatusID, targetCode)
		}
		if err != nil {
			return err
		}
		if err := s.repositories.InboundOrder.SetStatus(ctx, inbound.ID, target.ID, actor, inbound.VersionNo); err != nil {
			return err
		}
	}
	for poID := range poIDs {
		po, err := s.repositories.PurchaseOrder.Lock(ctx, poID)
		if err != nil {
			return err
		}
		poRow, err := s.repositories.PurchaseOrder.Get(ctx, poID)
		if err != nil {
			return err
		}
		lines, err := s.repositories.PurchaseOrderLine.List(ctx, poID)
		if err != nil {
			return err
		}
		ordered := new(big.Rat)
		completed := new(big.Rat)
		for _, line := range lines {
			lineOrdered, err := storedRatio(line.OrderedQty)
			if err != nil {
				return err
			}
			lineCompleted, err := storedRatio(line.CompletedReceiptQty)
			if err != nil {
				return err
			}
			ordered.Add(ordered, lineOrdered)
			completed.Add(completed, lineCompleted)
		}
		poTargetCode := "PARTIALLY_RECEIVED"
		if completed.Sign() == 0 {
			poTargetCode = "APPROVED"
		} else if completed.Cmp(ordered) >= 0 {
			poTargetCode = "RECEIVED"
		}
		if poRow.StatusCode != poTargetCode {
			var target mastermodel.DocumentStatus
			if poRow.StatusCode == "RECEIVED" {
				target, err = s.documentStatus(ctx, po.DocumentTypeID, poTargetCode)
			} else {
				target, err = s.transitionTarget(ctx, po.DocumentTypeID, po.StatusID, poTargetCode)
			}
			if err != nil {
				return err
			}
			if err := s.repositories.PurchaseOrder.SetStatus(ctx, po.ID, target.ID, actor, po.VersionNo); err != nil {
				return err
			}
		}
	}
	return nil
}
