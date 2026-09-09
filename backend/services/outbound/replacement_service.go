package outbound

import (
	"context"
	"errors"
	"math/big"
	"time"

	dto "wms-api/dto/outbound"
	model "wms-api/models/outbound"
	inventoryrepository "wms-api/repository/inventory"
	repository "wms-api/repository/outbound"
)

func (s *Service) CreateReplacementPick(ctx context.Context, exceptionID string, q dto.CreateReplacementPickRequest, actor string) (dto.ReplacementPickResponse, error) {
	if !validID(exceptionID, 170) || !validID(q.BalanceID, 160) || !validUUID(actor) || q.ExpectedBalanceVersion < 1 {
		return dto.ReplacementPickResponse{}, invalid("invalid replacement pick request")
	}
	qty, amount, err := quantity(q.ReplacementQty, "replacement_qty", false)
	if err != nil {
		return dto.ReplacementPickResponse{}, err
	}
	priorityCode, err := transportCode(q.PriorityCode, "priority_code")
	if err != nil {
		return dto.ReplacementPickResponse{}, err
	}
	notes, err := optional(q.Notes, 4000, "notes")
	if err != nil {
		return dto.ReplacementPickResponse{}, err
	}
	if q.PickingStrategyID != nil && !validUUID(*q.PickingStrategyID) {
		return dto.ReplacementPickResponse{}, invalid("invalid picking_strategy_id")
	}
	var response dto.ReplacementPickResponse
	err = s.transaction(ctx, func(local *Service) error {
		exception, err := local.repositories.CheckException.Lock(ctx, exceptionID)
		if err != nil {
			return err
		}
		row, err := local.repositories.CheckException.Get(ctx, exceptionID)
		if err != nil {
			return err
		}
		if row.StatusCode == "RESOLVED" {
			return state("check exception is already resolved")
		}
		if row.ResultCode == "OVER" {
			return state("OVER exceptions do not accept replacement stock")
		}
		totals, err := local.repositories.CheckResolution.Totals(ctx, exceptionID)
		if err != nil {
			return err
		}
		newFulfilled := new(big.Rat).Add(new(big.Rat).Add(mustRat(totals.AcceptedQty), mustRat(totals.ReplacementQty)), amount)
		if newFulfilled.Cmp(mustRat(exception.ExceptionQty)) > 0 || mustRat(totals.CorrectedQty).Cmp(newFulfilled) < 0 {
			return state("record enough STOCK_CORRECTION quantity before requesting replacement stock")
		}
		strategy, err := local.pickingStrategy(ctx, row.OwnerID, row.WarehouseID, q.PickingStrategyID)
		if err != nil {
			return err
		}
		candidates, err := local.repositories.Reservation.Candidates(ctx, row.OutboundLineID, &strategy.ID)
		if err != nil {
			return err
		}
		eligible := false
		for _, candidate := range candidates {
			if candidate.BalanceID == q.BalanceID {
				eligible = true
				break
			}
		}
		if !eligible {
			return state("balance is not currently eligible for this replacement")
		}
		balanceRow, err := local.repositories.Inventory.Balance.Get(ctx, q.BalanceID)
		if err != nil {
			return err
		}
		key := inventoryrepository.BalanceIdentity{OwnerID: balanceRow.OwnerID, WarehouseID: balanceRow.WarehouseID, LocationID: balanceRow.LocationID, ItemID: balanceRow.ItemID, LotID: balanceRow.LotID, HandlingUnitID: balanceRow.HandlingUnitID, InventoryStatusID: balanceRow.InventoryStatusID}
		balance, err := local.repositories.Inventory.Balance.FindLocked(ctx, key)
		if err != nil {
			return err
		}
		if balance.VersionNo != q.ExpectedBalanceVersion {
			return repository.ErrConcurrentWrite
		}
		available := new(big.Rat).Sub(mustRat(balance.OnHandQty), mustRat(balance.ReservedQty))
		if available.Cmp(amount) < 0 {
			return state("replacement balance has insufficient available stock")
		}
		if balance.HandlingUnitID != nil && available.Cmp(amount) != 0 {
			return state("handling-unit replacement stock must be reserved in full")
		}
		if err := local.repositories.Inventory.Balance.SetQuantities(ctx, balance.ID, balance.OnHandQty, decimal(new(big.Rat).Add(mustRat(balance.ReservedQty), amount))); err != nil {
			return err
		}
		line, err := local.repositories.Line.Get(ctx, row.OutboundLineID)
		if err != nil {
			return err
		}
		order, err := local.repositories.Order.Get(ctx, line.OutboundID)
		if err != nil {
			return err
		}
		reservationType, reservationStatus, err := local.document(ctx, "RESERVATION")
		if err != nil {
			return err
		}
		reservationID, err := local.generateID(ctx, "RESERVATION", order.BusinessDate, nil, &row.WarehouseID)
		if err != nil {
			return err
		}
		reservation := model.InventoryReservation{ID: reservationID, DocumentTypeID: reservationType.ID, StatusID: reservationStatus.ID, PickingStrategyID: &strategy.ID, OutboundLineID: row.OutboundLineID, BalanceID: balance.ID, ReservedQty: qty, PickedQty: "0", UOMID: row.UOMID, CreatedBy: actor}
		if err := local.repositories.Reservation.Create(ctx, &reservation); err != nil {
			return err
		}
		waveType, err := local.repositories.WaveType.ByCode(ctx, "SINGLE_ORDER")
		if err != nil {
			return state("SINGLE_ORDER wave type is not configured")
		}
		waveDoc, waveInitial, err := local.document(ctx, "OUTBOUND_WAVE")
		if err != nil {
			return err
		}
		waveID, err := local.generateID(ctx, "OUTBOUND_WAVE", order.BusinessDate, nil, &row.WarehouseID)
		if err != nil {
			return err
		}
		now := time.Now()
		wave := model.OutboundWave{ID: waveID, DocumentTypeID: waveDoc.ID, StatusID: waveInitial.ID, WaveTypeID: waveType.ID, OwnerID: row.OwnerID, WarehouseID: row.WarehouseID, PickingStrategyID: &strategy.ID, BusinessDate: order.BusinessDate, ReleasedAt: &now, Notes: notes, VersionNo: 1, CreatedBy: actor, UpdatedBy: actor}
		if err := local.repositories.Wave.Create(ctx, &wave); err != nil {
			return err
		}
		released, err := local.transition(ctx, waveDoc.ID, waveInitial.ID, "RELEASED")
		if err != nil {
			return err
		}
		if err := local.repositories.Wave.SetStatusLocked(ctx, waveID, released.ID, actor, map[string]interface{}{"released_at": now}); err != nil {
			return err
		}
		if err := local.repositories.WaveOrder.CreateBatch(ctx, []model.OutboundWaveOrder{{WaveID: waveID, OutboundID: order.ID, AddedBy: actor}}); err != nil {
			return err
		}
		staging, err := local.repositories.Staging.Get(ctx, row.StagingID)
		if err != nil {
			return err
		}
		taskType, err := local.repositories.Master.TaskType.ByCode(ctx, "PICK")
		if err != nil || !taskType.IsActive {
			return state("PICK task type is not configured")
		}
		open, err := local.repositories.Master.TaskStatus.ByCode(ctx, "OPEN")
		if err != nil || !open.IsActive {
			return state("OPEN task status is not configured")
		}
		priority, err := local.repositories.Master.TaskPriority.ByCode(ctx, priorityCode)
		if err != nil || !priority.IsActive {
			return invalid("priority_code is not configured")
		}
		taskID, err := local.generateID(ctx, "PICK_TASK", order.BusinessDate, nil, &row.WarehouseID)
		if err != nil {
			return err
		}
		task := model.PickTask{ID: taskID, TaskTypeID: taskType.ID, TaskStatusID: open.ID, TaskPriorityID: priority.ID, WaveID: waveID, ReservationID: reservationID, OutboundLineID: row.OutboundLineID, SourceLocationID: balance.LocationID, TargetLocationID: &staging.StagingLocationID, PlannedQty: qty, PickedQty: "0", UOMID: row.UOMID, ShortQty: "0", CreatedBy: actor}
		if err := local.repositories.PickTask.CreateBatch(ctx, []model.PickTask{task}); err != nil {
			return err
		}
		resolutionType, err := local.repositories.ResolutionType.ByCode(ctx, "REPLACEMENT")
		if err != nil {
			return err
		}
		resolutionID, err := randomID(exceptionID + "-R")
		if err != nil {
			return err
		}
		resolution := model.OutboundCheckResolution{ID: resolutionID, OutboundCheckExceptionID: exceptionID, OutboundCheckResolutionTypeID: resolutionType.ID, ResolvedQty: qty, Notes: notes, CreatedBy: actor}
		if err := local.repositories.CheckResolution.Create(ctx, &resolution); err != nil {
			return err
		}
		if err := local.repositories.ResolutionReservation.Create(ctx, &model.OutboundCheckResolutionReservation{OutboundCheckResolutionID: resolutionID, ReservationID: reservationID}); err != nil {
			return err
		}
		progress, err := local.repositories.ExceptionStatus.ByCode(ctx, "IN_PROGRESS")
		if err != nil {
			return err
		}
		if err := local.repositories.CheckException.SetStatus(ctx, exceptionID, progress.ID); err != nil {
			return err
		}
		response = dto.ReplacementPickResponse{ExceptionID: exceptionID, ResolutionID: resolutionID, WaveID: waveID, ReservationID: reservationID}
		return nil
	})
	if err != nil {
		return dto.ReplacementPickResponse{}, err
	}
	task, err := s.repositories.PickTask.GetByReservation(ctx, response.ReservationID)
	if err != nil {
		return dto.ReplacementPickResponse{}, err
	}
	response.Task = mapPick(task)
	return response, nil
}

func replacementWork(local *Service, ctx context.Context, reservationID string) (*repository.ReplacementWorkRow, error) {
	v, err := local.repositories.ResolutionReservation.ReplacementByReservation(ctx, reservationID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}
