package inbound

import (
	"context"
	"math/big"

	dto "wms-api/dto/inbound"
	inventorydto "wms-api/dto/inventory"
	model "wms-api/models/inbound"
	mastermodel "wms-api/models/master"
	repository "wms-api/repository/inbound"
	inventoryservice "wms-api/services/inventory"
)

func (s *Service) CreateQualityInspection(ctx context.Context, request dto.CreateQualityInspectionRequest, actor string) (dto.QualityInspectionResponse, error) {
	if !inboundID(request.ReceiptInventoryID, 160) || !inboundUUID(actor) {
		return dto.QualityInspectionResponse{}, invalid("invalid receipt_inventory_id or actor")
	}
	notes, err := optional(request.Notes, 4000, "notes")
	if err != nil {
		return dto.QualityInspectionResponse{}, err
	}
	var inspectionID string
	err = s.transaction(ctx, func(local *Service) error {
		batch, err := local.repositories.ReceiptInventory.GetContext(ctx, request.ReceiptInventoryID)
		if err != nil {
			return err
		}
		if batch.ReceiptStatusCode != "COMPLETED" || batch.InitialBalanceID == nil {
			return state("receipt batch must belong to a completed receipt with posted inventory")
		}
		balance, err := local.repositories.Inventory.Balance.Get(ctx, *batch.InitialBalanceID)
		available, availableErr := storedRatio(balance.AvailableQty)
		batchQty, batchErr := storedRatio(batch.BaseQty)
		if err != nil || availableErr != nil || batchErr != nil || balance.InventoryStatusCode != "QC_PENDING" || available.Cmp(batchQty) < 0 || ((batch.SerialID != nil || batch.HandlingUnitID != nil) && available.Cmp(batchQty) != 0) {
			return state("the full receipt batch is no longer available in QC_PENDING")
		}
		kind, err := local.documentType(ctx, "QUALITY_INSPECTION")
		if err != nil {
			return err
		}
		inspectionID, err = local.generateID(ctx, kind.ID, batch.BusinessDate, batch.VendorID, batch.WarehouseID)
		if err != nil {
			return err
		}
		pending, err := local.qualityStatus(ctx, "PENDING")
		if err != nil {
			return err
		}
		sourceBalanceID := *batch.InitialBalanceID
		return local.repositories.QualityInspection.Create(ctx, &model.QualityInspection{ID: inspectionID, ReceiptInventoryID: batch.ID, SourceBalanceID: &sourceBalanceID, QualityStatusID: pending.ID, InspectedQty: batch.BaseQty, PassedQty: "0.000000", FailedQty: "0.000000", Notes: notes, CreatedBy: actor, UpdatedBy: &actor, VersionNo: 1})
	})
	if err != nil {
		return dto.QualityInspectionResponse{}, err
	}
	return s.GetQualityInspection(ctx, inspectionID)
}

func (s *Service) CompleteQualityInspection(ctx context.Context, id string, request dto.CompleteQualityInspectionRequest, actor string) (dto.QualityInspectionResponse, error) {
	if !inboundID(id, 120) || !inboundUUID(actor) || request.ExpectedVersion < 1 || request.ExpectedBalanceVersion < 1 {
		return dto.QualityInspectionResponse{}, invalid("invalid quality-inspection completion")
	}
	passedQty, passed, err := inboundQuantity(request.PassedQty, "passed_qty", true)
	if err != nil {
		return dto.QualityInspectionResponse{}, err
	}
	failedQty, failed, err := inboundQuantity(request.FailedQty, "failed_qty", true)
	if err != nil {
		return dto.QualityInspectionResponse{}, err
	}
	if passed.Sign() == 0 && failed.Sign() == 0 {
		return dto.QualityInspectionResponse{}, invalid("passed_qty or failed_qty must be greater than zero")
	}
	notes, err := optional(request.Notes, 4000, "notes")
	if err != nil {
		return dto.QualityInspectionResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		inspection, err := local.repositories.QualityInspection.Lock(ctx, id)
		if err != nil {
			return err
		}
		if inspection.InspectedAt != nil {
			return nil
		}
		if inspection.CancelledAt != nil {
			return state("cancelled quality inspection cannot be completed")
		}
		if inspection.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		inspected, err := storedRatio(inspection.InspectedQty)
		if err != nil {
			return err
		}
		if new(big.Rat).Add(passed, failed).Cmp(inspected) != 0 {
			return invalid("passed_qty plus failed_qty must equal inspected_qty")
		}
		batch, err := local.repositories.ReceiptInventory.GetContext(ctx, inspection.ReceiptInventoryID)
		if err != nil {
			return err
		}
		sourceBalanceID := inspection.SourceBalanceID
		if sourceBalanceID == nil {
			sourceBalanceID = batch.InitialBalanceID
		}
		if sourceBalanceID == nil {
			return state("inspection has no source inventory balance")
		}
		balance, err := local.repositories.Inventory.Balance.Get(ctx, *sourceBalanceID)
		available, availableErr := storedRatio(balance.AvailableQty)
		if err != nil || availableErr != nil || balance.InventoryStatusCode != "QC_PENDING" || available.Cmp(inspected) < 0 || ((batch.SerialID != nil || batch.HandlingUnitID != nil) && available.Cmp(inspected) != 0) {
			return state("inspection source is no longer the full QC_PENDING balance")
		}
		if balance.VersionNo != request.ExpectedBalanceVersion {
			return repository.ErrConcurrentWrite
		}
		item, err := local.repositories.Master.Catalog.Item.Get(ctx, batch.ItemID)
		if err != nil || !item.IsActive {
			return invalid("inspection item is unavailable")
		}
		if (item.SerialControlled || batch.HandlingUnitID != nil) && passed.Sign() > 0 && failed.Sign() > 0 {
			return invalid("a serialized or handling-unit receipt batch cannot be split between pass and fail")
		}
		serialIDs := make([]string, 0, 1)
		if batch.SerialID != nil {
			serialIDs = append(serialIDs, *batch.SerialID)
		}
		inventory := inventoryservice.NewService(local.repositories.Inventory)
		expectedBalanceVersion := request.ExpectedBalanceVersion
		if passed.Sign() > 0 {
			if request.PutawayTargetLocationID == nil {
				return invalid("putaway_target_location_id is required when passed_qty is positive")
			}
			target, err := local.validatePutawayTarget(ctx, batch.OwnerID, batch.WarehouseID, item, *request.PutawayTargetLocationID)
			if err != nil {
				return err
			}
			putawayPending, err := local.inventoryStatus(ctx, "PUTAWAY_PENDING")
			if err != nil {
				return err
			}
			posted, err := inventory.PostMovement(ctx, inventorydto.PostingRequest{OperationKey: "inbound.qc." + inspection.ID + ".pass", MovementTypeCode: "STATUS_CHANGE", OwnerID: batch.OwnerID, WarehouseID: batch.WarehouseID, BusinessDate: batch.BusinessDate, ItemID: batch.ItemID, LotID: batch.LotID, HandlingUnitID: batch.HandlingUnitID, From: &inventorydto.BalanceDimension{LocationID: balance.LocationID, InventoryStatusID: balance.InventoryStatusID}, To: &inventorydto.BalanceDimension{LocationID: balance.LocationID, InventoryStatusID: putawayPending.ID}, Quantity: passedQty, SerialIDs: serialIDs, ExpectedSourceVersion: &expectedBalanceVersion, SourceDocumentID: inspection.ID, SourceLineID: &batch.ID, Notes: notes}, actor)
			if err != nil {
				return err
			}
			if posted.FromBalance == nil || posted.ToBalance == nil {
				return state("quality pass posting returned incomplete balances")
			}
			expectedBalanceVersion = posted.FromBalance.VersionNo
			if err := local.createPutawayTask(ctx, inspection, batch, item, target, posted.ToBalance.ID, passedQty, actor); err != nil {
				return err
			}
		}
		if failed.Sign() > 0 {
			quarantine, err := local.inventoryStatus(ctx, "QUARANTINE")
			if err != nil {
				return err
			}
			failedSerials := serialIDs
			if passed.Sign() > 0 {
				failedSerials = nil
			}
			posted, err := inventory.PostMovement(ctx, inventorydto.PostingRequest{OperationKey: "inbound.qc." + inspection.ID + ".fail", MovementTypeCode: "STATUS_CHANGE", OwnerID: batch.OwnerID, WarehouseID: batch.WarehouseID, BusinessDate: batch.BusinessDate, ItemID: batch.ItemID, LotID: batch.LotID, HandlingUnitID: batch.HandlingUnitID, From: &inventorydto.BalanceDimension{LocationID: balance.LocationID, InventoryStatusID: balance.InventoryStatusID}, To: &inventorydto.BalanceDimension{LocationID: balance.LocationID, InventoryStatusID: quarantine.ID}, Quantity: failedQty, SerialIDs: failedSerials, ExpectedSourceVersion: &expectedBalanceVersion, SourceDocumentID: inspection.ID, SourceLineID: &batch.ID, Notes: notes}, actor)
			if err != nil {
				return err
			}
			if posted.ToBalance == nil {
				return state("quality failure posting returned no quarantine balance")
			}
			if err := local.createQuarantineCase(ctx, inspection, batch, posted.ToBalance.ID, failedQty, notes, actor); err != nil {
				return err
			}
		}
		resultCode, qualityCode := "PARTIAL", "FAILED"
		if failed.Sign() == 0 {
			resultCode, qualityCode = "ACCEPTED", "PASSED"
		}
		if passed.Sign() == 0 {
			resultCode = "REJECTED"
		}
		result, err := local.inspectionResult(ctx, resultCode)
		if err != nil {
			return err
		}
		quality, err := local.qualityStatus(ctx, qualityCode)
		if err != nil {
			return err
		}
		return local.repositories.QualityInspection.Complete(ctx, inspection.ID, quality.ID, result.ID, passedQty, failedQty, actor, inspection.VersionNo, notes)
	})
	if err != nil {
		return dto.QualityInspectionResponse{}, err
	}
	return s.GetQualityInspection(ctx, id)
}

func (s *Service) createChildInspection(ctx context.Context, parent model.QualityInspection, sourceBalanceID, quantity string, notes *string, actor string) (string, error) {
	batch, err := s.repositories.ReceiptInventory.GetContext(ctx, parent.ReceiptInventoryID)
	if err != nil {
		return "", err
	}
	kind, err := s.documentType(ctx, "QUALITY_INSPECTION")
	if err != nil {
		return "", err
	}
	id, err := s.generateID(ctx, kind.ID, batch.BusinessDate, batch.VendorID, batch.WarehouseID)
	if err != nil {
		return "", err
	}
	pending, err := s.qualityStatus(ctx, "PENDING")
	if err != nil {
		return "", err
	}
	parentID, balanceID := parent.ID, sourceBalanceID
	err = s.repositories.QualityInspection.Create(ctx, &model.QualityInspection{ID: id, ReceiptInventoryID: parent.ReceiptInventoryID, ParentInspectionID: &parentID, SourceBalanceID: &balanceID, QualityStatusID: pending.ID, InspectedQty: quantity, PassedQty: "0.000000", FailedQty: "0.000000", Notes: notes, CreatedBy: actor, UpdatedBy: &actor, VersionNo: 1})
	return id, err
}

func (s *Service) CancelQualityInspection(ctx context.Context, id string, request dto.ExceptionTransitionRequest, actor string) (dto.QualityInspectionResponse, error) {
	reason, err := validateExceptionTransition(id, request, actor)
	if err != nil {
		return dto.QualityInspectionResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		inspection, err := local.repositories.QualityInspection.Lock(ctx, id)
		if err != nil {
			return err
		}
		if inspection.CancelledAt != nil {
			return nil
		}
		if inspection.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		if inspection.InspectedAt != nil {
			return state("completed quality inspection cannot be cancelled")
		}
		source := inspection.SourceBalanceID
		if source == nil {
			batch, e := local.repositories.ReceiptInventory.GetContext(ctx, inspection.ReceiptInventoryID)
			if e != nil {
				return e
			}
			source = batch.InitialBalanceID
		}
		if source == nil {
			return state("inspection has no source balance")
		}
		balance, err := local.repositories.Inventory.Balance.Get(ctx, *source)
		available, parseErr := storedRatio(balance.AvailableQty)
		inspected, inspectedErr := storedRatio(inspection.InspectedQty)
		if err != nil || parseErr != nil || inspectedErr != nil || balance.InventoryStatusCode != "QC_PENDING" || available.Cmp(inspected) < 0 {
			return state("inspection stock is no longer intact in QC_PENDING")
		}
		waived, err := local.qualityStatus(ctx, "WAIVED")
		if err != nil {
			return err
		}
		if err := local.repositories.QualityInspection.Cancel(ctx, id, waived.ID, actor, *reason, inspection.VersionNo); err != nil {
			return err
		}
		_, err = local.createChildInspection(ctx, inspection, *source, inspection.InspectedQty, reason, actor)
		if err != nil {
			return err
		}
		batch, err := local.repositories.ReceiptInventory.GetContext(ctx, inspection.ReceiptInventoryID)
		if err != nil {
			return err
		}
		return local.recordException(ctx, batch.OwnerID, batch.WarehouseID, id, nil, "CANCELLATION", nil, nil, nil, reason, actor)
	})
	if err != nil {
		return dto.QualityInspectionResponse{}, err
	}
	return s.GetQualityInspection(ctx, id)
}

func (s *Service) createPutawayTask(ctx context.Context, inspection model.QualityInspection, batch repository.ReceiptInventoryContext, item mastermodel.Item, target mastermodel.WarehouseLocation, sourceBalanceID, quantity, actor string) error {
	kind, err := s.documentType(ctx, "PUTAWAY_TASK")
	if err != nil {
		return err
	}
	id, err := s.generateID(ctx, kind.ID, batch.BusinessDate, batch.VendorID, batch.WarehouseID)
	if err != nil {
		return err
	}
	taskType, err := s.taskType(ctx, "PUTAWAY")
	if err != nil {
		return err
	}
	status, err := s.initialTaskStatus(ctx)
	if err != nil {
		return err
	}
	priority, err := s.normalPriority(ctx)
	if err != nil {
		return err
	}
	return s.repositories.PutawayTask.Create(ctx, &model.PutawayTask{ID: id, TaskTypeID: taskType.ID, TaskStatusID: status.ID, TaskPriorityID: priority.ID, ReceiptInventoryID: batch.ID, InspectionID: inspection.ID, SourceBalanceID: sourceBalanceID, OwnerID: batch.OwnerID, WarehouseID: batch.WarehouseID, ItemID: item.ID, LotID: batch.LotID, HandlingUnitID: batch.HandlingUnitID, SourceLocationID: batch.ReceivedLocationID, TargetLocationID: target.ID, PlannedQty: quantity, CompletedQty: "0.000000", UOMID: batch.BaseUOMID, CreatedBy: actor, UpdatedBy: &actor, VersionNo: 1})
}
