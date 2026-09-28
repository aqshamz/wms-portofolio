package outbound

import (
	"context"
	"math/big"

	inventorydto "wms-api/dto/inventory"
	dto "wms-api/dto/outbound"
	repository "wms-api/repository/outbound"
	inventoryservice "wms-api/services/inventory"
)

func mapDisposal(row repository.DisposalRow) dto.DisposalResponse {
	return dto.DisposalResponse{
		ID: row.ID, QuarantineDispositionID: row.QuarantineDispositionID, QuarantineCaseID: row.QuarantineCaseID,
		OwnerID: row.OwnerID, OwnerCode: row.OwnerCode, OwnerName: row.OwnerName,
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

func (s *Service) disposalResponse(ctx context.Context, row repository.DisposalRow) (dto.DisposalResponse, error) {
	result := mapDisposal(row)
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

func (s *Service) GetDisposal(ctx context.Context, id string) (dto.DisposalResponse, error) {
	if !validID(id, 140) {
		return dto.DisposalResponse{}, invalid("invalid disposal_id")
	}
	row, err := s.repositories.Disposal.Get(ctx, id)
	if err != nil {
		return dto.DisposalResponse{}, err
	}
	return s.disposalResponse(ctx, row)
}

func (s *Service) ListDisposals(ctx context.Context, filter repository.ListFilter) (dto.PageResponse[dto.DisposalResponse], error) {
	if !validUUID(filter.OwnerID) || !validUUID(filter.WarehouseID) || filter.Page < 1 || filter.PageSize < 1 || filter.PageSize > 100 {
		return dto.PageResponse[dto.DisposalResponse]{}, invalid("owner_id, warehouse_id, and valid pagination are required")
	}
	rows, total, err := s.repositories.Disposal.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.DisposalResponse]{}, err
	}
	items := make([]dto.DisposalResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapDisposal(row))
	}
	return page(items, filter.Page, filter.PageSize, total), nil
}

func (s *Service) CompleteDisposal(ctx context.Context, id string, request dto.CompleteDisposalRequest, actor string) (dto.DisposalResponse, error) {
	if !validID(id, 140) || !validUUID(actor) || request.ExpectedVersion < 1 || request.ExpectedBalanceVersion < 1 {
		return dto.DisposalResponse{}, invalid("invalid disposal completion")
	}
	completedAt, err := timestamp(request.CompletedAt, "completed_at")
	if err != nil {
		return dto.DisposalResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		disposal, err := local.repositories.Disposal.Lock(ctx, id)
		if err != nil {
			return err
		}
		if disposal.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		current, err := local.repositories.Disposal.Get(ctx, id)
		if err != nil {
			return err
		}
		if current.StatusCode != "PLANNED" {
			return state("disposal transaction is not planned")
		}
		contextRow, err := local.repositories.Disposal.LockDispositionContext(ctx, disposal.QuarantineDispositionID)
		if err != nil {
			return err
		}
		if contextRow.DispositionStatusCode != "DECIDED" {
			return state("linked quarantine disposition is not pending")
		}
		balance, err := local.repositories.Inventory.Balance.Get(ctx, disposal.SourceBalanceID)
		if err != nil || balance.VersionNo != request.ExpectedBalanceVersion {
			if err != nil {
				return err
			}
			return repository.ErrConcurrentWrite
		}
		if balance.OwnerID != disposal.OwnerID || balance.WarehouseID != disposal.WarehouseID || balance.ItemID != disposal.ItemID || balance.LocationID != disposal.SourceLocationID || balance.InventoryStatusID != disposal.SourceInventoryStatusID || balance.InventoryStatusCode != "QUARANTINE" {
			return state("disposal source stock no longer matches the planned quarantine balance")
		}
		plannedQty, ok := new(big.Rat).SetString(disposal.Quantity)
		availableQty, availableOK := new(big.Rat).SetString(balance.AvailableQty)
		if !ok || !availableOK || availableQty.Cmp(plannedQty) < 0 {
			return state("insufficient quarantine stock for disposal")
		}
		if disposal.HandlingUnitID != nil && availableQty.Cmp(plannedQty) != 0 {
			return state("handling-unit disposal must process its entire source balance")
		}
		serialIDs := make([]string, 0, 1)
		if disposal.SerialID != nil {
			serialIDs = append(serialIDs, *disposal.SerialID)
		}
		posted, err := inventoryservice.NewService(local.repositories.Inventory).PostMovement(ctx, inventorydto.PostingRequest{
			OperationKey: "outbound.disposal." + disposal.ID, MovementTypeCode: "DISPOSE",
			OwnerID: disposal.OwnerID, WarehouseID: disposal.WarehouseID, BusinessDate: disposal.BusinessDate.Format("2006-01-02"),
			ItemID: disposal.ItemID, LotID: disposal.LotID, HandlingUnitID: disposal.HandlingUnitID,
			From:     &inventorydto.BalanceDimension{LocationID: disposal.SourceLocationID, InventoryStatusID: disposal.SourceInventoryStatusID},
			Quantity: disposal.Quantity, SerialIDs: serialIDs, ExpectedSourceVersion: &request.ExpectedBalanceVersion,
			SourceDocumentID: disposal.ID, SourceLineID: &disposal.QuarantineDispositionID, Notes: disposal.Notes,
		}, actor)
		if err != nil {
			return err
		}
		completedStatus, err := local.transition(ctx, disposal.DocumentTypeID, disposal.StatusID, "COMPLETED")
		if err != nil {
			return err
		}
		if err := local.repositories.Disposal.Complete(ctx, id, completedStatus.ID, posted.Movement.ID, actor, completedAt, request.ExpectedVersion); err != nil {
			return err
		}
		processedStatus, err := local.transition(ctx, contextRow.DispositionDocumentTypeID, contextRow.DispositionStatusID, "PROCESSED")
		if err != nil {
			return err
		}
		if err := local.repositories.Disposal.ProcessDisposition(ctx, disposal.QuarantineDispositionID, processedStatus.ID, posted.Movement.ID); err != nil {
			return err
		}
		updated, err := local.repositories.Disposal.LockDispositionContext(ctx, disposal.QuarantineDispositionID)
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
		return local.repositories.Disposal.UpdateCaseStatus(ctx, disposal.QuarantineDispositionID, targetStatusID, actor, closeCase, updated.CaseVersionNo)
	})
	if err != nil {
		return dto.DisposalResponse{}, err
	}
	return s.GetDisposal(ctx, id)
}

func (s *Service) CancelDisposal(ctx context.Context, id string, request dto.CancelDisposalRequest, actor string) (dto.DisposalResponse, error) {
	if !validID(id, 140) || !validUUID(actor) || request.ExpectedVersion < 1 {
		return dto.DisposalResponse{}, invalid("invalid disposal cancellation")
	}
	reason, err := clean(request.Reason, 4000, "reason")
	if err != nil {
		return dto.DisposalResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		disposal, err := local.repositories.Disposal.Lock(ctx, id)
		if err != nil {
			return err
		}
		if disposal.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		current, err := local.repositories.Disposal.Get(ctx, id)
		if err != nil {
			return err
		}
		if current.StatusCode != "PLANNED" {
			return state("only a planned disposal can be cancelled")
		}
		contextRow, err := local.repositories.Disposal.LockDispositionContext(ctx, disposal.QuarantineDispositionID)
		if err != nil {
			return err
		}
		if contextRow.DispositionStatusCode != "DECIDED" {
			return state("linked quarantine disposition is not pending")
		}
		cancelled, err := local.transition(ctx, disposal.DocumentTypeID, disposal.StatusID, "CANCELLED")
		if err != nil {
			return err
		}
		if err := local.repositories.Disposal.Cancel(ctx, id, cancelled.ID, reason, actor, request.ExpectedVersion); err != nil {
			return err
		}
		dispositionCancelled, err := local.transition(ctx, contextRow.DispositionDocumentTypeID, contextRow.DispositionStatusID, "CANCELLED")
		if err != nil {
			return err
		}
		if err := local.repositories.Disposal.CancelDisposition(ctx, disposal.QuarantineDispositionID, dispositionCancelled.ID); err != nil {
			return err
		}
		updated, err := local.repositories.Disposal.LockDispositionContext(ctx, disposal.QuarantineDispositionID)
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
		return local.repositories.Disposal.UpdateCaseStatus(ctx, disposal.QuarantineDispositionID, targetStatusID, actor, false, updated.CaseVersionNo)
	})
	if err != nil {
		return dto.DisposalResponse{}, err
	}
	return s.GetDisposal(ctx, id)
}
