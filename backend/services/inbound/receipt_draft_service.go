package inbound

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	dto "wms-api/dto/inbound"
	model "wms-api/models/inbound"
	repository "wms-api/repository/inbound"
)

func (s *Service) UpdateReceipt(ctx context.Context, id string, request dto.UpdateReceiptRequest, actor string) (dto.ReceiptResponse, error) {
	if !inboundID(id, 120) || !inboundUUID(request.DockLocationID) || !inboundUUID(actor) || request.ExpectedVersion < 1 {
		return dto.ReceiptResponse{}, invalid("invalid open receipt update")
	}
	if len(request.Lines) == 0 || len(request.Lines) > 500 {
		return dto.ReceiptResponse{}, invalid("lines must contain 1..500 entries")
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

	err = s.transaction(ctx, func(local *Service) error {
		header, err := local.repositories.Receipt.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.Receipt.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "OPEN" {
			return state("only an OPEN unposted receipt can be edited")
		}
		if header.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		posted, err := local.repositories.ReceiptInventory.CountPostedByReceipt(ctx, id)
		if err != nil {
			return err
		}
		if posted != 0 {
			return state("receipt inventory is already posted")
		}
		if header.InboundID == nil {
			return state("receipt has no inbound order")
		}
		inbound, err := local.repositories.InboundOrder.Lock(ctx, *header.InboundID)
		if err != nil {
			return err
		}
		dock, err := local.repositories.Inventory.Location.GetShared(ctx, strings.ToLower(request.DockLocationID))
		if err != nil || dock.WarehouseID != header.WarehouseID || !dock.IsActive || dock.IsLocked {
			return invalid("dock location is unavailable or belongs to another warehouse")
		}
		dockType, err := local.repositories.Master.LocationType.GetShared(ctx, dock.LocationTypeID)
		if err != nil || !dockType.IsActive || dockType.Code != "DOCK" || !dockType.AllowsReceiving {
			return invalid("receiving dock must use an active DOCK location type that allows receiving")
		}
		existingLines, err := local.repositories.ReceiptLine.List(ctx, id)
		if err != nil {
			return err
		}
		existingByInboundLine := make(map[string]model.ReceiptLine, len(existingLines))
		allowedSerials := make(map[string]string)
		for _, line := range existingLines {
			if line.InboundLineID != nil {
				existingByInboundLine[*line.InboundLineID] = line.ReceiptLine
			}
			batches, lookupErr := local.repositories.ReceiptInventory.ListByLine(ctx, line.ID)
			if lookupErr != nil {
				return lookupErr
			}
			for _, batch := range batches {
				if batch.SerialID != nil && batch.SerialNo != nil {
					allowedSerials[receiptSerialKey(line.ItemID, *batch.SerialNo)] = *batch.SerialID
				}
			}
		}
		if err := local.repositories.ReceiptLine.DeleteByReceipt(ctx, id); err != nil {
			return err
		}
		if err := local.writeReceiptLines(ctx, header, inbound, request.Lines, existingByInboundLine, allowedSerials, actor); err != nil {
			return err
		}
		return local.repositories.Receipt.UpdateOpen(ctx, id, actor, request.ExpectedVersion, map[string]interface{}{
			"received_at": receivedAt, "dock_location_id": dock.ID, "vehicle_number": vehicle, "seal_number": seal, "delivery_note_no": deliveryNote, "notes": notes,
		})
	})
	if err != nil {
		return dto.ReceiptResponse{}, err
	}
	return s.GetReceipt(ctx, id)
}

func (s *Service) writeReceiptLines(ctx context.Context, receipt model.Receipt, inbound model.InboundOrder, lines []dto.ReceiptLineRequest, existingByInboundLine map[string]model.ReceiptLine, allowedSerials map[string]string, actor string) error {
	qcPending, err := s.inventoryStatus(ctx, "QC_PENDING")
	if err != nil {
		return err
	}
	seenLines := make(map[string]bool, len(lines))
	seenSerials := make(map[string]bool)
	seenHandlingUnits := make(map[string]bool)
	for lineIndex, requestLine := range lines {
		lineID := strings.TrimSpace(requestLine.InboundLineID)
		if !inboundID(lineID, 150) || seenLines[lineID] {
			return invalid("inbound_line_id is invalid or repeated")
		}
		seenLines[lineID] = true
		inboundLine, err := s.repositories.InboundOrderLine.Get(ctx, lineID)
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
		item, err := s.repositories.Master.Catalog.Item.Get(ctx, inboundLine.ItemID)
		if err != nil || !item.IsActive || item.OwnerID != inbound.OwnerID {
			return invalid("receipt item is unavailable")
		}
		var existing *model.ReceiptLine
		if value, ok := existingByInboundLine[inboundLine.ID]; ok {
			existing = &value
		}
		receiptUOMID, conversion, conversionNumber, err := s.receiptUnit(ctx, item, inboundLine, requestLine.UOMID, existing)
		if err != nil {
			return err
		}
		_, receivedBaseQty, err := inboundQuantitySnapshot(receivedNumber, conversion)
		if err != nil {
			return err
		}
		_, rejectedBaseQty, err := inboundQuantitySnapshot(rejectedNumber, conversion)
		if err != nil {
			return err
		}
		receivedBaseNumber := new(big.Rat).Mul(receivedNumber, conversionNumber)
		alreadyReceivedText, err := s.repositories.ReceiptLine.ReceivedBaseForInboundLine(ctx, inboundLine.ID)
		if err != nil {
			return err
		}
		alreadyReceived, ok := new(big.Rat).SetString(alreadyReceivedText)
		expected, expectedOK := new(big.Rat).SetString(inboundLine.ExpectedBaseQty)
		if !ok || !expectedOK {
			return state("stored inbound base quantity is invalid")
		}
		tolerance := new(big.Rat)
		if inboundLine.PurchaseOrderLineID != nil {
			poLine, lookupErr := s.repositories.PurchaseOrderLine.Get(ctx, *inboundLine.PurchaseOrderLineID)
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
		newTotal := new(big.Rat).Add(alreadyReceived, receivedBaseNumber)
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
		for index, batch := range requestLine.Batches {
			_, number, err := inboundQuantity(batch.SourceQty, "source_qty", false)
			if err != nil {
				return err
			}
			batchNumbers[index] = number
			batchTotal.Add(batchTotal, number)
		}
		if batchTotal.Cmp(accepted) != 0 {
			return invalid("batch source quantities must equal received_qty minus rejected_qty")
		}
		lineNo := lineIndex + 1
		receiptLineID := fmt.Sprintf("%s-L%04d", receipt.ID, lineNo)
		inboundLineID := inboundLine.ID
		receiptLine := model.ReceiptLine{ID: receiptLineID, ReceiptID: receipt.ID, InboundLineID: &inboundLineID, LineNo: lineNo, ItemID: item.ID, ReceivedQty: receivedQty, RejectedQty: rejectedQty, UOMConversionToBase: conversion, ReceivedBaseQty: receivedBaseQty, RejectedBaseQty: rejectedBaseQty, BaseUOMID: inboundLine.BaseUOMID, ExceptionNotes: exceptionNotes, ExceptionTypeCode: exceptionType, UOMID: receiptUOMID, CreatedBy: actor}
		if err := s.repositories.ReceiptLine.Create(ctx, &receiptLine); err != nil {
			return err
		}
		for batchIndex, requestBatch := range requestLine.Batches {
			sourceQty := batchNumbers[batchIndex].FloatString(6)
			baseQty, baseNumber, err := inboundMultiply(batchNumbers[batchIndex], conversionNumber)
			if err != nil {
				return err
			}
			location, err := s.repositories.Inventory.Location.GetShared(ctx, strings.ToLower(requestBatch.ReceivedLocationID))
			if err != nil || location.WarehouseID != inbound.WarehouseID || !location.IsActive || location.IsLocked {
				return invalid("received location is unavailable or belongs to another warehouse")
			}
			locationType, err := s.repositories.Master.LocationType.GetShared(ctx, location.LocationTypeID)
			if err != nil || !locationType.IsActive || !locationType.AllowsReceiving {
				return invalid("received location must use an active receiving location type")
			}
			lotID, err := s.ensureReceiptLot(ctx, inbound.OwnerID, item, requestBatch.Lot, actor)
			if err != nil {
				return err
			}
			if lotID != nil && item.MinimumReceiveDays != nil {
				lot, err := s.repositories.Inventory.Lot.Get(ctx, *lotID)
				if err != nil {
					return err
				}
				minimumExpiry := receipt.BusinessDate.AddDate(0, 0, *item.MinimumReceiveDays)
				if lot.ExpiryDate == nil || lot.ExpiryDate.Before(minimumExpiry) {
					return invalid("lot expiry does not meet the item's minimum receive days")
				}
			}
			if item.SerialControlled && baseNumber.Cmp(big.NewRat(1, 1)) != 0 {
				return invalid("each serialized receipt batch must convert to exactly 1 base unit")
			}
			serialID, err := s.ensureReceiptSerial(ctx, inbound.OwnerID, item, requestBatch.SerialNo, actor, seenSerials, allowedSerials)
			if err != nil {
				return err
			}
			var handlingUnitID *string
			if requestBatch.HandlingUnitID != nil {
				handlingUnitIDValue := strings.TrimSpace(*requestBatch.HandlingUnitID)
				if !inboundID(handlingUnitIDValue, 120) {
					return invalid("invalid handling_unit_id")
				}
				if seenHandlingUnits[handlingUnitIDValue] {
					return invalid("a handling unit can only be assigned to one receipt batch")
				}
				handlingUnit, err := s.repositories.Inventory.HandlingUnit.Get(ctx, handlingUnitIDValue)
				if err != nil || handlingUnit.OwnerID != inbound.OwnerID || handlingUnit.WarehouseID != inbound.WarehouseID || handlingUnit.CurrentLocationID == nil || *handlingUnit.CurrentLocationID != location.ID || handlingUnit.ParentHandlingUnitID != nil || handlingUnit.IsClosed {
					return invalid("handling unit is unavailable or is not at the received location")
				}
				positiveBalances, err := s.repositories.Inventory.Balance.CountPositiveForHandlingUnit(ctx, handlingUnitIDValue)
				if err != nil {
					return err
				}
				if positiveBalances != 0 {
					return invalid("handling unit must be empty before receiving stock into it")
				}
				children, err := s.repositories.Inventory.HandlingUnit.CountChildren(ctx, handlingUnitIDValue)
				if err != nil {
					return err
				}
				if children != 0 {
					return invalid("handling unit must not contain child handling units before receiving stock into it")
				}
				seenHandlingUnits[handlingUnitIDValue] = true
				handlingUnitID = &handlingUnitIDValue
			}
			batchNo := batchIndex + 1
			batch := model.ReceiptInventory{ID: fmt.Sprintf("%s-B%04d", receiptLineID, batchNo), ReceiptLineID: receiptLineID, ItemID: item.ID, SourceQty: sourceQty, SourceUOMID: receiptUOMID, BaseQty: baseQty, BaseUOMID: inboundLine.BaseUOMID, LotID: lotID, HandlingUnitID: handlingUnitID, ReceivedLocationID: location.ID, InitialInventoryStatusID: qcPending.ID, CreatedBy: actor}
			if err := s.repositories.ReceiptInventory.Create(ctx, &batch); err != nil {
				return err
			}
			if serialID != nil {
				if err := s.repositories.ReceiptLineSerial.Create(ctx, &model.ReceiptLineSerial{ReceiptInventoryID: batch.ID, SerialID: *serialID}); err != nil {
					if errors.Is(err, repository.ErrConflict) {
						return invalid("serial_no is already linked to another receipt")
					}
					return err
				}
			}
		}
	}
	return nil
}
