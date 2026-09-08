package inbound

import (
	"context"
	"math/big"

	dto "wms-api/dto/inbound"
	inventorydto "wms-api/dto/inventory"
	repository "wms-api/repository/inbound"
	inventoryservice "wms-api/services/inventory"
)

func (s *Service) StartPutawayTask(ctx context.Context, id string, request dto.PutawayTransitionRequest, actor string) (dto.PutawayTaskResponse, error) {
	if !inboundID(id, 120) || !inboundUUID(actor) || request.ExpectedVersion < 1 {
		return dto.PutawayTaskResponse{}, invalid("invalid putaway start")
	}
	err := s.transaction(ctx, func(local *Service) error {
		task, err := local.repositories.PutawayTask.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.PutawayTask.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.TaskStatusCode == "IN_PROGRESS" && task.AssignedTo != nil && *task.AssignedTo == actor {
			return nil
		}
		if task.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		if row.TaskStatusCode != "OPEN" && row.TaskStatusCode != "ASSIGNED" {
			return state("only an OPEN or ASSIGNED putaway task can be started")
		}
		if row.TaskStatusCode == "ASSIGNED" && (task.AssignedTo == nil || *task.AssignedTo != actor) {
			return state("putaway task must be started by its assigned account")
		}
		target, err := local.taskTransitionTarget(ctx, task.TaskStatusID, "IN_PROGRESS")
		if err != nil {
			return err
		}
		return local.repositories.PutawayTask.Start(ctx, id, target.ID, actor, task.VersionNo)
	})
	if err != nil {
		return dto.PutawayTaskResponse{}, err
	}
	return s.GetPutawayTask(ctx, id)
}

func (s *Service) AssignPutawayTask(ctx context.Context, id string, request dto.AssignPutawayRequest, actor string) (dto.PutawayTaskResponse, error) {
	if !inboundID(id, 120) || !inboundUUID(actor) || !inboundUUID(request.AccountID) || request.ExpectedVersion < 1 {
		return dto.PutawayTaskResponse{}, invalid("invalid putaway assignment")
	}
	err := s.transaction(ctx, func(local *Service) error {
		task, err := local.repositories.PutawayTask.Lock(ctx, id)
		if err != nil {
			return err
		}
		if task.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		row, err := local.repositories.PutawayTask.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.TaskStatusCode != "OPEN" && row.TaskStatusCode != "ASSIGNED" {
			return state("only an OPEN or ASSIGNED putaway task can be assigned")
		}
		allowed, err := local.repositories.PutawayTask.AccountCanAccess(ctx, request.AccountID, task.OwnerID, task.WarehouseID)
		if err != nil {
			return err
		}
		if !allowed {
			return invalid("assigned account lacks active owner and warehouse access")
		}
		assigned, err := local.taskStatus(ctx, "ASSIGNED")
		if err != nil {
			return err
		}
		return local.repositories.PutawayTask.Assign(ctx, id, assigned.ID, request.AccountID, actor, task.VersionNo)
	})
	if err != nil {
		return dto.PutawayTaskResponse{}, err
	}
	return s.GetPutawayTask(ctx, id)
}

func (s *Service) RetargetPutawayTask(ctx context.Context, id string, request dto.RetargetPutawayRequest, actor string) (dto.PutawayTaskResponse, error) {
	if !inboundID(id, 120) || !inboundUUID(actor) || !inboundUUID(request.TargetLocationID) || request.ExpectedVersion < 1 {
		return dto.PutawayTaskResponse{}, invalid("invalid putaway retarget")
	}
	err := s.transaction(ctx, func(local *Service) error {
		task, err := local.repositories.PutawayTask.Lock(ctx, id)
		if err != nil {
			return err
		}
		if task.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		row, err := local.repositories.PutawayTask.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.TaskStatusCode != "OPEN" && row.TaskStatusCode != "ASSIGNED" {
			return state("only an OPEN or ASSIGNED putaway task can be retargeted")
		}
		item, err := local.repositories.Master.Catalog.Item.Get(ctx, task.ItemID)
		if err != nil || !item.IsActive {
			return invalid("putaway item is unavailable")
		}
		target, err := local.validatePutawayTarget(ctx, task.OwnerID, task.WarehouseID, item, request.TargetLocationID)
		if err != nil {
			return err
		}
		return local.repositories.PutawayTask.Retarget(ctx, id, target.ID, actor, task.VersionNo)
	})
	if err != nil {
		return dto.PutawayTaskResponse{}, err
	}
	return s.GetPutawayTask(ctx, id)
}

func (s *Service) CancelPutawayTask(ctx context.Context, id string, request dto.CancelPutawayRequest, actor string) (dto.PutawayTaskResponse, error) {
	if !inboundID(id, 120) || !inboundUUID(actor) || request.ExpectedVersion < 1 || request.ExpectedBalanceVersion < 1 {
		return dto.PutawayTaskResponse{}, invalid("invalid putaway cancellation")
	}
	if _, err := inboundDate(request.BusinessDate, "business_date"); err != nil {
		return dto.PutawayTaskResponse{}, err
	}
	reason, err := optional(&request.Reason, 4000, "reason")
	if err != nil {
		return dto.PutawayTaskResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		task, err := local.repositories.PutawayTask.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.PutawayTask.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.TaskStatusCode == "CANCELLED" {
			return nil
		}
		if task.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		if row.TaskStatusCode != "OPEN" && row.TaskStatusCode != "ASSIGNED" && row.TaskStatusCode != "IN_PROGRESS" {
			return state("completed putaway task cannot be cancelled")
		}
		if row.TaskStatusCode == "IN_PROGRESS" && (task.AssignedTo == nil || *task.AssignedTo != actor) {
			return state("in-progress putaway task must be cancelled by its assignee")
		}
		balance, err := local.repositories.Inventory.Balance.Get(ctx, task.SourceBalanceID)
		if err != nil || balance.OwnerID != task.OwnerID || balance.WarehouseID != task.WarehouseID || balance.ItemID != task.ItemID || balance.LocationID != task.SourceLocationID || balance.InventoryStatusCode != "PUTAWAY_PENDING" || balance.VersionNo != request.ExpectedBalanceVersion {
			return repository.ErrConcurrentWrite
		}
		availableQty, err := storedRatio(balance.AvailableQty)
		if err != nil {
			return err
		}
		plannedQty, err := storedRatio(task.PlannedQty)
		if err != nil || availableQty.Cmp(plannedQty) < 0 {
			return invalid("putaway quantity is no longer fully available")
		}
		if task.HandlingUnitID != nil && availableQty.Cmp(plannedQty) != 0 {
			return invalid("handling-unit cancellation must move its entire source balance")
		}
		available, err := local.inventoryStatus(ctx, "QC_PENDING")
		if err != nil {
			return err
		}
		batch, err := local.repositories.ReceiptInventory.GetContext(ctx, task.ReceiptInventoryID)
		if err != nil {
			return err
		}
		serialIDs := make([]string, 0, 1)
		if batch.SerialID != nil {
			serialIDs = append(serialIDs, *batch.SerialID)
		}
		posted, err := inventoryservice.NewService(local.repositories.Inventory).PostMovement(ctx, inventorydto.PostingRequest{OperationKey: "inbound.putaway-cancel." + task.ID, MovementTypeCode: "STATUS_CHANGE", OwnerID: task.OwnerID, WarehouseID: task.WarehouseID, BusinessDate: request.BusinessDate, ItemID: task.ItemID, LotID: task.LotID, HandlingUnitID: task.HandlingUnitID, From: &inventorydto.BalanceDimension{LocationID: task.SourceLocationID, InventoryStatusID: balance.InventoryStatusID}, To: &inventorydto.BalanceDimension{LocationID: task.SourceLocationID, InventoryStatusID: available.ID}, Quantity: task.PlannedQty, SerialIDs: serialIDs, ExpectedSourceVersion: &request.ExpectedBalanceVersion, SourceDocumentID: task.ID, SourceLineID: &task.ReceiptInventoryID, Notes: reason}, actor)
		if err != nil {
			return err
		}
		if posted.ToBalance == nil {
			return state("putaway cancellation returned no QC balance")
		}
		parent, err := local.repositories.QualityInspection.Lock(ctx, task.InspectionID)
		if err != nil {
			return err
		}
		if _, err := local.createChildInspection(ctx, parent, posted.ToBalance.ID, task.PlannedQty, reason, actor); err != nil {
			return err
		}
		cancelled, err := local.taskTransitionTarget(ctx, task.TaskStatusID, "CANCELLED")
		if err != nil {
			return err
		}
		if err := local.repositories.PutawayTask.Cancel(ctx, id, cancelled.ID, actor, task.VersionNo); err != nil {
			return err
		}
		return local.recordException(ctx, task.OwnerID, task.WarehouseID, id, nil, "CANCELLATION", nil, nil, nil, reason, actor)
	})
	if err != nil {
		return dto.PutawayTaskResponse{}, err
	}
	return s.GetPutawayTask(ctx, id)
}

func (s *Service) ReversePutawayTask(ctx context.Context, id string, request dto.CancelPutawayRequest, actor string) (dto.PutawayTaskResponse, error) {
	if !inboundID(id, 120) || !inboundUUID(actor) || request.ExpectedVersion < 1 || request.ExpectedBalanceVersion < 1 {
		return dto.PutawayTaskResponse{}, invalid("invalid putaway reversal")
	}
	if _, err := inboundDate(request.BusinessDate, "business_date"); err != nil {
		return dto.PutawayTaskResponse{}, err
	}
	reason, err := optional(&request.Reason, 4000, "reason")
	if err != nil {
		return dto.PutawayTaskResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		task, err := local.repositories.PutawayTask.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.PutawayTask.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.TaskStatusCode == "REVERSED" {
			return nil
		}
		if task.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		if row.TaskStatusCode != "COMPLETED" || task.ResultingBalanceID == nil {
			return state("only a completed putaway task can be reversed")
		}
		balance, err := local.repositories.Inventory.Balance.Get(ctx, *task.ResultingBalanceID)
		if err != nil || balance.OwnerID != task.OwnerID || balance.WarehouseID != task.WarehouseID || balance.ItemID != task.ItemID || balance.LocationID != task.TargetLocationID || balance.InventoryStatusCode != "AVAILABLE" {
			return state("putaway result balance no longer matches the task")
		}
		if balance.VersionNo != request.ExpectedBalanceVersion {
			return repository.ErrConcurrentWrite
		}
		available, err := storedRatio(balance.AvailableQty)
		if err != nil {
			return err
		}
		planned, err := storedRatio(task.PlannedQty)
		if err != nil || available.Cmp(planned) < 0 {
			return invalid("putaway result quantity is no longer fully available")
		}
		if task.HandlingUnitID != nil && available.Cmp(planned) != 0 {
			return invalid("handling-unit reversal must move its entire result balance")
		}
		qcPending, err := local.inventoryStatus(ctx, "QC_PENDING")
		if err != nil {
			return err
		}
		batch, err := local.repositories.ReceiptInventory.GetContext(ctx, task.ReceiptInventoryID)
		if err != nil {
			return err
		}
		serialIDs := make([]string, 0, 1)
		if batch.SerialID != nil {
			serialIDs = append(serialIDs, *batch.SerialID)
		}
		posted, err := inventoryservice.NewService(local.repositories.Inventory).PostMovement(ctx, inventorydto.PostingRequest{OperationKey: "inbound.putaway-reversal." + task.ID, MovementTypeCode: "PUTAWAY_REVERSAL", OwnerID: task.OwnerID, WarehouseID: task.WarehouseID, BusinessDate: request.BusinessDate, ItemID: task.ItemID, LotID: task.LotID, HandlingUnitID: task.HandlingUnitID, From: &inventorydto.BalanceDimension{LocationID: task.TargetLocationID, InventoryStatusID: balance.InventoryStatusID}, To: &inventorydto.BalanceDimension{LocationID: task.SourceLocationID, InventoryStatusID: qcPending.ID}, Quantity: task.PlannedQty, SerialIDs: serialIDs, ExpectedSourceVersion: &request.ExpectedBalanceVersion, SourceDocumentID: task.ID, SourceLineID: &task.ReceiptInventoryID, Notes: reason, RelocateHandlingUnit: task.HandlingUnitID != nil}, actor)
		if err != nil {
			return err
		}
		if posted.ToBalance == nil {
			return state("putaway reversal returned no QC balance")
		}
		parent, err := local.repositories.QualityInspection.Lock(ctx, task.InspectionID)
		if err != nil {
			return err
		}
		inspectionID, err := local.createChildInspection(ctx, parent, posted.ToBalance.ID, task.PlannedQty, reason, actor)
		if err != nil {
			return err
		}
		reversed, err := local.taskStatus(ctx, "REVERSED")
		if err != nil {
			return err
		}
		if err := local.repositories.PutawayTask.Reverse(ctx, id, reversed.ID, posted.Movement.ID, inspectionID, actor, *reason, task.VersionNo); err != nil {
			return err
		}
		return local.recordException(ctx, task.OwnerID, task.WarehouseID, id, nil, "REVERSAL", nil, nil, nil, reason, actor)
	})
	if err != nil {
		return dto.PutawayTaskResponse{}, err
	}
	return s.GetPutawayTask(ctx, id)
}

func (s *Service) CompletePutawayTask(ctx context.Context, id string, request dto.CompletePutawayRequest, actor string) (dto.PutawayTaskResponse, error) {
	if !inboundID(id, 120) || !inboundUUID(actor) || request.ExpectedVersion < 1 || request.ExpectedBalanceVersion < 1 {
		return dto.PutawayTaskResponse{}, invalid("invalid putaway completion")
	}
	if _, err := inboundDate(request.BusinessDate, "business_date"); err != nil {
		return dto.PutawayTaskResponse{}, err
	}
	err := s.transaction(ctx, func(local *Service) error {
		task, err := local.repositories.PutawayTask.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.PutawayTask.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.TaskStatusCode == "COMPLETED" {
			return nil
		}
		if task.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		if row.TaskStatusCode != "IN_PROGRESS" {
			return state("only an IN_PROGRESS putaway task can be completed")
		}
		if task.AssignedTo == nil || *task.AssignedTo != actor {
			return state("putaway task must be completed by its assigned account")
		}
		balance, err := local.repositories.Inventory.Balance.Get(ctx, task.SourceBalanceID)
		if err != nil || balance.OwnerID != task.OwnerID || balance.WarehouseID != task.WarehouseID || balance.ItemID != task.ItemID || balance.LocationID != task.SourceLocationID || balance.InventoryStatusCode != "PUTAWAY_PENDING" {
			return state("putaway source balance no longer matches the task")
		}
		if balance.VersionNo != request.ExpectedBalanceVersion {
			return repository.ErrConcurrentWrite
		}
		availableQty, err := storedRatio(balance.AvailableQty)
		if err != nil {
			return err
		}
		plannedQty, err := storedRatio(task.PlannedQty)
		if err != nil || availableQty.Cmp(plannedQty) < 0 {
			return invalid("insufficient unreserved putaway quantity")
		}
		if task.HandlingUnitID != nil && availableQty.Cmp(plannedQty) != 0 {
			return invalid("handling-unit putaway must move its entire source balance")
		}
		item, err := local.repositories.Master.Catalog.Item.Get(ctx, task.ItemID)
		if err != nil || !item.IsActive {
			return invalid("putaway item is unavailable")
		}
		if _, err := local.validatePutawayTarget(ctx, task.OwnerID, task.WarehouseID, item, task.TargetLocationID); err != nil {
			return err
		}
		available, err := local.inventoryStatus(ctx, "AVAILABLE")
		if err != nil {
			return err
		}
		batch, err := local.repositories.ReceiptInventory.GetContext(ctx, task.ReceiptInventoryID)
		if err != nil {
			return err
		}
		serialIDs := make([]string, 0, 1)
		if batch.SerialID != nil {
			if plannedQty.Cmp(big.NewRat(1, 1)) != 0 {
				return invalid("serialized putaway must move exactly one unit")
			}
			serialIDs = append(serialIDs, *batch.SerialID)
		}
		posted, err := inventoryservice.NewService(local.repositories.Inventory).PostMovement(ctx, inventorydto.PostingRequest{OperationKey: "inbound.putaway." + task.ID, MovementTypeCode: "PUTAWAY", OwnerID: task.OwnerID, WarehouseID: task.WarehouseID, BusinessDate: request.BusinessDate, ItemID: task.ItemID, LotID: task.LotID, HandlingUnitID: task.HandlingUnitID, From: &inventorydto.BalanceDimension{LocationID: task.SourceLocationID, InventoryStatusID: balance.InventoryStatusID}, To: &inventorydto.BalanceDimension{LocationID: task.TargetLocationID, InventoryStatusID: available.ID}, Quantity: task.PlannedQty, SerialIDs: serialIDs, ExpectedSourceVersion: &request.ExpectedBalanceVersion, SourceDocumentID: task.ID, SourceLineID: &task.ReceiptInventoryID, RelocateHandlingUnit: task.HandlingUnitID != nil}, actor)
		if err != nil {
			return err
		}
		if posted.ToBalance == nil {
			return state("putaway posting returned no destination balance")
		}
		completed, err := local.taskTransitionTarget(ctx, task.TaskStatusID, "COMPLETED")
		if err != nil {
			return err
		}
		return local.repositories.PutawayTask.Complete(ctx, task.ID, completed.ID, posted.Movement.ID, posted.ToBalance.ID, task.PlannedQty, actor, task.VersionNo)
	})
	if err != nil {
		return dto.PutawayTaskResponse{}, err
	}
	return s.GetPutawayTask(ctx, id)
}
