package outbound

import (
	"context"
	"math/big"

	inventorydto "wms-api/dto/inventory"
	dto "wms-api/dto/outbound"
	repository "wms-api/repository/outbound"
	inventoryservice "wms-api/services/inventory"
)

func mapVendorReturn(row repository.VendorReturnRow) dto.VendorReturnResponse {
	return dto.VendorReturnResponse{
		ID: row.ID, QuarantineDispositionID: row.QuarantineDispositionID, QuarantineCaseID: row.QuarantineCaseID,
		OwnerID: row.OwnerID, OwnerCode: row.OwnerCode, OwnerName: row.OwnerName,
		VendorID: row.VendorID, VendorCode: row.VendorCode, VendorName: row.VendorName,
		WarehouseID: row.WarehouseID, WarehouseCode: row.WarehouseCode, WarehouseName: row.WarehouseName,
		BusinessDate: row.BusinessDate.Format("2006-01-02"), SourceBalanceID: row.SourceBalanceID,
		PlannedBalanceVersionNo: row.PlannedBalanceVersionNo, ItemID: row.ItemID, ItemCode: row.ItemCode, ItemName: row.ItemName,
		LotID: row.LotID, LotNumber: row.LotNumber, SerialID: row.SerialID, SerialNo: row.SerialNo,
		HandlingUnitID: row.HandlingUnitID, HandlingUnitBarcode: row.HandlingUnitBarcode,
		SourceLocationID: row.SourceLocationID, SourceLocationCode: row.SourceLocationCode,
		SourceInventoryStatusID: row.SourceInventoryStatusID, SourceInventoryStatusCode: row.SourceInventoryStatusCode,
		Quantity: row.Quantity, UOMID: row.UOMID, UOMCode: row.UOMCode, StatusCode: row.StatusCode,
		Notes: row.Notes, PlannedAt: row.PlannedAt, CompletedAt: row.CompletedAt, CompletedBy: row.CompletedBy,
		CompletedByDisplayName: row.CompletedByDisplayName, CancelledAt: row.CancelledAt, CancelledBy: row.CancelledBy,
		CancelledByDisplayName: row.CancelledByDisplayName, CancellationReason: row.CancellationReason,
		InventoryMovementID: row.InventoryMovementID, CreatedAt: row.CreatedAt, CreatedBy: row.CreatedBy,
		CreatedByDisplayName: row.CreatedByDisplayName, VersionNo: row.VersionNo,
	}
}

func (s *Service) vendorReturnResponse(ctx context.Context, row repository.VendorReturnRow) (dto.VendorReturnResponse, error) {
	result := mapVendorReturn(row)
	if row.StatusCode != "PLANNED" {
		return result, nil
	}
	balance, err := s.repositories.Inventory.Balance.Get(ctx, row.SourceBalanceID)
	if err != nil {
		return result, err
	}
	result.SourceBalanceVersionNo = &balance.VersionNo
	result.AvailableQty = balance.AvailableQty
	return result, nil
}

func (s *Service) GetVendorReturn(ctx context.Context, id string) (dto.VendorReturnResponse, error) {
	if !validID(id, 140) {
		return dto.VendorReturnResponse{}, invalid("invalid vendor_return_id")
	}
	row, err := s.repositories.VendorReturn.Get(ctx, id)
	if err != nil {
		return dto.VendorReturnResponse{}, err
	}
	return s.vendorReturnResponse(ctx, row)
}

func (s *Service) ListVendorReturns(ctx context.Context, filter repository.ListFilter) (dto.PageResponse[dto.VendorReturnResponse], error) {
	if !validUUID(filter.OwnerID) || !validUUID(filter.WarehouseID) || filter.Page < 1 || filter.PageSize < 1 || filter.PageSize > 100 {
		return dto.PageResponse[dto.VendorReturnResponse]{}, invalid("owner_id, warehouse_id, and valid pagination are required")
	}
	rows, total, err := s.repositories.VendorReturn.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.VendorReturnResponse]{}, err
	}
	items := make([]dto.VendorReturnResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapVendorReturn(row))
	}
	return page(items, filter.Page, filter.PageSize, total), nil
}

func (s *Service) CompleteVendorReturn(ctx context.Context, id string, request dto.CompleteVendorReturnRequest, actor string) (dto.VendorReturnResponse, error) {
	if !validID(id, 140) || !validUUID(actor) || request.ExpectedVersion < 1 || request.ExpectedBalanceVersion < 1 {
		return dto.VendorReturnResponse{}, invalid("invalid vendor return completion")
	}
	completedAt, err := timestamp(request.CompletedAt, "completed_at")
	if err != nil {
		return dto.VendorReturnResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		vendorReturn, err := local.repositories.VendorReturn.Lock(ctx, id)
		if err != nil {
			return err
		}
		if vendorReturn.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		current, err := local.repositories.VendorReturn.Get(ctx, id)
		if err != nil {
			return err
		}
		if current.StatusCode != "PLANNED" {
			return state("return to vendor transaction is not planned")
		}
		contextRow, err := local.repositories.Disposal.LockDispositionContext(ctx, vendorReturn.QuarantineDispositionID)
		if err != nil {
			return err
		}
		if contextRow.DispositionStatusCode != "DECIDED" {
			return state("linked quarantine disposition is not pending")
		}
		balance, err := local.repositories.Inventory.Balance.Get(ctx, vendorReturn.SourceBalanceID)
		if err != nil || balance.VersionNo != request.ExpectedBalanceVersion {
			if err != nil {
				return err
			}
			return repository.ErrConcurrentWrite
		}
		if balance.OwnerID != vendorReturn.OwnerID || balance.WarehouseID != vendorReturn.WarehouseID || balance.ItemID != vendorReturn.ItemID || balance.LocationID != vendorReturn.SourceLocationID || balance.InventoryStatusID != vendorReturn.SourceInventoryStatusID || balance.InventoryStatusCode != "QUARANTINE" {
			return state("return source stock no longer matches the planned quarantine balance")
		}
		plannedQty, ok := new(big.Rat).SetString(vendorReturn.Quantity)
		availableQty, availableOK := new(big.Rat).SetString(balance.AvailableQty)
		if !ok || !availableOK || availableQty.Cmp(plannedQty) < 0 {
			return state("insufficient quarantine stock for return to vendor")
		}
		if vendorReturn.HandlingUnitID != nil && availableQty.Cmp(plannedQty) != 0 {
			return state("handling-unit return must process its entire source balance")
		}
		serialIDs := make([]string, 0, 1)
		if vendorReturn.SerialID != nil {
			serialIDs = append(serialIDs, *vendorReturn.SerialID)
		}
		posted, err := inventoryservice.NewService(local.repositories.Inventory).PostMovement(ctx, inventorydto.PostingRequest{
			OperationKey: "outbound.vendor-return." + vendorReturn.ID, MovementTypeCode: "RETURN_TO_VENDOR",
			OwnerID: vendorReturn.OwnerID, WarehouseID: vendorReturn.WarehouseID, BusinessDate: vendorReturn.BusinessDate.Format("2006-01-02"),
			ItemID: vendorReturn.ItemID, LotID: vendorReturn.LotID, HandlingUnitID: vendorReturn.HandlingUnitID,
			From:     &inventorydto.BalanceDimension{LocationID: vendorReturn.SourceLocationID, InventoryStatusID: vendorReturn.SourceInventoryStatusID},
			Quantity: vendorReturn.Quantity, SerialIDs: serialIDs, ExpectedSourceVersion: &request.ExpectedBalanceVersion,
			SourceDocumentID: vendorReturn.ID, SourceLineID: &vendorReturn.QuarantineDispositionID, Notes: vendorReturn.Notes,
		}, actor)
		if err != nil {
			return err
		}
		completedStatus, err := local.transition(ctx, vendorReturn.DocumentTypeID, vendorReturn.StatusID, "COMPLETED")
		if err != nil {
			return err
		}
		if err := local.repositories.VendorReturn.Complete(ctx, id, completedStatus.ID, posted.Movement.ID, actor, completedAt, request.ExpectedVersion); err != nil {
			return err
		}
		processedStatus, err := local.transition(ctx, contextRow.DispositionDocumentTypeID, contextRow.DispositionStatusID, "PROCESSED")
		if err != nil {
			return err
		}
		if err := local.repositories.Disposal.ProcessDisposition(ctx, vendorReturn.QuarantineDispositionID, processedStatus.ID, posted.Movement.ID); err != nil {
			return err
		}
		updated, err := local.repositories.Disposal.LockDispositionContext(ctx, vendorReturn.QuarantineDispositionID)
		if err != nil {
			return err
		}
		processed, processedOK := new(big.Rat).SetString(updated.ProcessedQty)
		total, totalOK := new(big.Rat).SetString(updated.QuarantineQty)
		if !processedOK || !totalOK {
			return state("linked quarantine quantities are invalid")
		}
		targetCode, closeCase := "PARTIALLY_DECIDED", false
		if processed.Cmp(total) == 0 {
			targetCode, closeCase = "CLOSED", true
		}
		targetStatusID := updated.CaseStatusID
		if updated.CaseStatusCode != targetCode {
			target, err := local.transition(ctx, updated.CaseDocumentTypeID, updated.CaseStatusID, targetCode)
			if err != nil {
				return err
			}
			targetStatusID = target.ID
		}
		return local.repositories.Disposal.UpdateCaseStatus(ctx, vendorReturn.QuarantineDispositionID, targetStatusID, actor, closeCase, updated.CaseVersionNo)
	})
	if err != nil {
		return dto.VendorReturnResponse{}, err
	}
	return s.GetVendorReturn(ctx, id)
}

func (s *Service) CancelVendorReturn(ctx context.Context, id string, request dto.CancelVendorReturnRequest, actor string) (dto.VendorReturnResponse, error) {
	if !validID(id, 140) || !validUUID(actor) || request.ExpectedVersion < 1 {
		return dto.VendorReturnResponse{}, invalid("invalid vendor return cancellation")
	}
	reason, err := clean(request.Reason, 4000, "reason")
	if err != nil {
		return dto.VendorReturnResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		vendorReturn, err := local.repositories.VendorReturn.Lock(ctx, id)
		if err != nil {
			return err
		}
		if vendorReturn.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		current, err := local.repositories.VendorReturn.Get(ctx, id)
		if err != nil {
			return err
		}
		if current.StatusCode != "PLANNED" {
			return state("only a planned return to vendor can be cancelled")
		}
		contextRow, err := local.repositories.Disposal.LockDispositionContext(ctx, vendorReturn.QuarantineDispositionID)
		if err != nil {
			return err
		}
		if contextRow.DispositionStatusCode != "DECIDED" {
			return state("linked quarantine disposition is not pending")
		}
		cancelled, err := local.transition(ctx, vendorReturn.DocumentTypeID, vendorReturn.StatusID, "CANCELLED")
		if err != nil {
			return err
		}
		if err := local.repositories.VendorReturn.Cancel(ctx, id, cancelled.ID, reason, actor, request.ExpectedVersion); err != nil {
			return err
		}
		dispositionCancelled, err := local.transition(ctx, contextRow.DispositionDocumentTypeID, contextRow.DispositionStatusID, "CANCELLED")
		if err != nil {
			return err
		}
		if err := local.repositories.Disposal.CancelDisposition(ctx, vendorReturn.QuarantineDispositionID, dispositionCancelled.ID); err != nil {
			return err
		}
		updated, err := local.repositories.Disposal.LockDispositionContext(ctx, vendorReturn.QuarantineDispositionID)
		if err != nil {
			return err
		}
		committed, ok := new(big.Rat).SetString(updated.CommittedQty)
		if !ok {
			return state("linked quarantine quantity is invalid")
		}
		targetCode := "PARTIALLY_DECIDED"
		if committed.Sign() == 0 {
			targetCode = "OPEN"
		}
		targetStatusID := updated.CaseStatusID
		if updated.CaseStatusCode != targetCode {
			target, err := local.transition(ctx, updated.CaseDocumentTypeID, updated.CaseStatusID, targetCode)
			if err != nil {
				return err
			}
			targetStatusID = target.ID
		}
		return local.repositories.Disposal.UpdateCaseStatus(ctx, vendorReturn.QuarantineDispositionID, targetStatusID, actor, false, updated.CaseVersionNo)
	})
	if err != nil {
		return dto.VendorReturnResponse{}, err
	}
	return s.GetVendorReturn(ctx, id)
}
