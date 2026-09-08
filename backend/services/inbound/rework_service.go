package inbound

import (
	"context"

	dto "wms-api/dto/inbound"
	repository "wms-api/repository/inbound"
)

func (s *Service) StartReworkTask(ctx context.Context, id string, request dto.ReworkTransitionRequest, actor string) (dto.ReworkTaskResponse, error) {
	if !inboundID(id, 140) || !inboundUUID(actor) || request.ExpectedVersion < 1 {
		return dto.ReworkTaskResponse{}, invalid("invalid rework start")
	}
	err := s.transaction(ctx, func(local *Service) error {
		task, err := local.repositories.ReworkTask.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.ReworkTask.Get(ctx, id)
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
			return state("only an OPEN or ASSIGNED rework task can be started")
		}
		if row.TaskStatusCode == "ASSIGNED" && (task.AssignedTo == nil || *task.AssignedTo != actor) {
			return state("rework task must be started by its assigned account")
		}
		started, err := local.taskTransitionTarget(ctx, task.TaskStatusID, "IN_PROGRESS")
		if err != nil {
			return err
		}
		return local.repositories.ReworkTask.Start(ctx, id, started.ID, actor, task.VersionNo)
	})
	if err != nil {
		return dto.ReworkTaskResponse{}, err
	}
	return s.GetReworkTask(ctx, id)
}

func (s *Service) CompleteReworkTask(ctx context.Context, id string, request dto.CompleteReworkRequest, actor string) (dto.ReworkTaskResponse, error) {
	if !inboundID(id, 140) || !inboundUUID(actor) || request.ExpectedVersion < 1 {
		return dto.ReworkTaskResponse{}, invalid("invalid rework completion")
	}
	notes, err := optional(request.ResultNotes, 4000, "result_notes")
	if err != nil {
		return dto.ReworkTaskResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		task, err := local.repositories.ReworkTask.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.ReworkTask.Get(ctx, id)
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
			return state("only an IN_PROGRESS rework task can be completed")
		}
		if task.AssignedTo == nil || *task.AssignedTo != actor {
			return state("rework task must be completed by its assigned account")
		}
		balance, err := local.repositories.Inventory.Balance.Get(ctx, task.SourceBalanceID)
		if err != nil || balance.InventoryStatusCode != "QC_PENDING" {
			return state("rework source balance is unavailable")
		}
		available, err := storedRatio(balance.AvailableQty)
		if err != nil {
			return err
		}
		planned, err := storedRatio(task.PlannedQty)
		if err != nil || available.Cmp(planned) < 0 {
			return state("rework quantity is no longer available")
		}
		quarantine, err := local.repositories.QuarantineCase.Get(ctx, row.QuarantineCaseID)
		if err != nil {
			return err
		}
		parent, err := local.repositories.QualityInspection.Lock(ctx, quarantine.InspectionID)
		if err != nil {
			return err
		}
		inspectionID, err := local.createChildInspection(ctx, parent, task.SourceBalanceID, task.PlannedQty, notes, actor)
		if err != nil {
			return err
		}
		completed, err := local.taskTransitionTarget(ctx, task.TaskStatusID, "COMPLETED")
		if err != nil {
			return err
		}
		return local.repositories.ReworkTask.Complete(ctx, id, completed.ID, task.PlannedQty, inspectionID, actor, task.VersionNo, notes)
	})
	if err != nil {
		return dto.ReworkTaskResponse{}, err
	}
	return s.GetReworkTask(ctx, id)
}
