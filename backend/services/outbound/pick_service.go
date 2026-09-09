package outbound

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
	"time"

	dto "wms-api/dto/outbound"
	inventorymodel "wms-api/models/inventory"
	model "wms-api/models/outbound"
	inventoryrepository "wms-api/repository/inventory"
	repository "wms-api/repository/outbound"
)

func (s *Service) GetPickTask(ctx context.Context, id string) (dto.PickTaskResponse, error) {
	if !validID(id, 140) {
		return dto.PickTaskResponse{}, invalid("invalid pick_task_id")
	}
	row, err := s.repositories.PickTask.Get(ctx, id)
	if err != nil {
		return dto.PickTaskResponse{}, err
	}
	return mapPick(row), nil
}
func (s *Service) ListPickTasks(ctx context.Context, f repository.ListFilter, assignee string) (dto.PageResponse[dto.PickTaskResponse], error) {
	if !validUUID(f.OwnerID) {
		return dto.PageResponse[dto.PickTaskResponse]{}, invalid("owner_id is required and must be a UUID")
	}
	if assignee != "" && !validUUID(assignee) {
		return dto.PageResponse[dto.PickTaskResponse]{}, invalid("assignee_id must be a UUID")
	}
	rows, total, err := s.repositories.PickTask.List(ctx, f, assignee)
	if err != nil {
		return dto.PageResponse[dto.PickTaskResponse]{}, err
	}
	items := make([]dto.PickTaskResponse, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapPick(v))
	}
	return page(items, f.Page, f.PageSize, total), nil
}

func (s *Service) StartPickTask(ctx context.Context, id, actor string) (dto.PickTaskResponse, error) {
	if !validID(id, 140) || !validUUID(actor) {
		return dto.PickTaskResponse{}, invalid("invalid pick task start request")
	}
	err := s.transaction(ctx, func(local *Service) error {
		task, err := local.repositories.PickTask.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.PickTask.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.TaskStatusCode != "OPEN" && row.TaskStatusCode != "ASSIGNED" {
			return state("only an OPEN or ASSIGNED pick task can be started")
		}
		replacement, err := replacementWork(local, ctx, task.ReservationID)
		if err != nil {
			return err
		}
		target, err := local.taskTransition(ctx, task.TaskStatusID, "IN_PROGRESS")
		if err != nil {
			return err
		}
		if err := local.repositories.PickTask.Start(ctx, id, target.ID, actor); err != nil {
			return err
		}
		wave, err := local.repositories.Wave.Lock(ctx, task.WaveID)
		if err != nil {
			return err
		}
		waveRow, err := local.repositories.Wave.Get(ctx, wave.ID)
		if err != nil {
			return err
		}
		if waveRow.StatusCode == "RELEASED" {
			progress, err := local.transition(ctx, wave.DocumentTypeID, wave.StatusID, "IN_PROGRESS")
			if err != nil {
				return err
			}
			if err := local.repositories.Wave.SetStatusLocked(ctx, wave.ID, progress.ID, actor, nil); err != nil {
				return err
			}
		} else if waveRow.StatusCode != "IN_PROGRESS" {
			return state("pick task wave is not released")
		}
		order, err := local.repositories.Order.Lock(ctx, row.OutboundID)
		if err != nil {
			return err
		}
		orderRow, err := local.repositories.Order.Get(ctx, order.ID)
		if err != nil {
			return err
		}
		if orderRow.StatusCode == "WAVED" {
			progress, err := local.transition(ctx, order.DocumentTypeID, order.StatusID, "PICKING")
			if err != nil {
				return err
			}
			return local.repositories.Order.SetStatusLocked(ctx, order.ID, progress.ID, actor, nil)
		}
		if replacement != nil && orderRow.StatusCode == "CHECK_FAILED" {
			return nil
		}
		if orderRow.StatusCode != "PICKING" {
			return state("outbound order is not ready for picking")
		}
		return nil
	})
	if err != nil {
		return dto.PickTaskResponse{}, err
	}
	return s.GetPickTask(ctx, id)
}

func fingerprint(values ...string) string {
	h := sha256.New()
	for _, v := range values {
		h.Write([]byte(v))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (s *Service) ConfirmPick(ctx context.Context, id string, request dto.ConfirmPickRequest, actor string) (dto.PickConfirmationResponse, error) {
	if !validID(id, 140) || !validUUID(actor) || request.ExpectedBalanceVersion < 1 {
		return dto.PickConfirmationResponse{}, invalid("invalid pick confirmation request")
	}
	qty, qtyRat, err := quantity(request.PickedQty, "picked_qty", false)
	if err != nil {
		return dto.PickConfirmationResponse{}, err
	}
	businessDate, err := date(request.BusinessDate, "business_date")
	if err != nil {
		return dto.PickConfirmationResponse{}, err
	}
	operationKey, err := clean(request.OperationKey, 160, "operation_key")
	if err != nil {
		return dto.PickConfirmationResponse{}, err
	}
	fp := fingerprint(id, qty, request.BusinessDate)
	if existing, lookupErr := s.repositories.Inventory.Movement.GetByOperationKey(ctx, operationKey); lookupErr == nil {
		if existing.OperationFingerprint == nil || *existing.OperationFingerprint != fp {
			return dto.PickConfirmationResponse{}, repository.ErrConflict
		}
		execution, e := s.repositories.PickExecution.GetByMovement(ctx, existing.ID)
		if e != nil {
			return dto.PickConfirmationResponse{}, e
		}
		task, e := s.GetPickTask(ctx, id)
		return dto.PickConfirmationResponse{ExecutionID: execution.ID, MovementID: existing.ID, StagingBalanceID: execution.StagingBalanceID, Task: task}, e
	} else if !errors.Is(lookupErr, inventoryrepository.ErrNotFound) {
		return dto.PickConfirmationResponse{}, lookupErr
	}
	var executionID, movementID, stagingBalanceID string
	err = s.transaction(ctx, func(local *Service) error {
		task, err := local.repositories.PickTask.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.PickTask.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.TaskStatusCode != "IN_PROGRESS" {
			return state("only an IN_PROGRESS pick task can be confirmed")
		}
		if task.AssignedTo == nil || *task.AssignedTo != actor {
			return state("pick task is assigned to another account")
		}
		_, planned, err := quantity(task.PlannedQty, "planned_qty", false)
		if err != nil {
			return err
		}
		_, picked, err := quantity(task.PickedQty, "picked_qty", true)
		if err != nil {
			return err
		}
		remaining := new(big.Rat).Sub(planned, picked)
		if qtyRat.Cmp(remaining) > 0 {
			return invalid("picked_qty exceeds the task remainder")
		}
		reservation, err := local.repositories.Reservation.Lock(ctx, task.ReservationID)
		if err != nil {
			return err
		}
		replacement, err := replacementWork(local, ctx, reservation.ID)
		if err != nil {
			return err
		}
		active, err := local.status(ctx, reservation.DocumentTypeID, "ACTIVE")
		if err != nil {
			return err
		}
		partial, err := local.status(ctx, reservation.DocumentTypeID, "PARTIALLY_PICKED")
		if err != nil {
			return err
		}
		if reservation.StatusID != active.ID && reservation.StatusID != partial.ID {
			return state("reservation is not active")
		}
		balanceRow, err := local.repositories.Inventory.Balance.Get(ctx, reservation.BalanceID)
		if err != nil {
			return err
		}
		identity := inventoryrepository.BalanceIdentity{OwnerID: balanceRow.OwnerID, WarehouseID: balanceRow.WarehouseID, LocationID: balanceRow.LocationID, ItemID: balanceRow.ItemID, LotID: balanceRow.LotID, HandlingUnitID: balanceRow.HandlingUnitID, InventoryStatusID: balanceRow.InventoryStatusID}
		source, err := local.repositories.Inventory.Balance.FindLocked(ctx, identity)
		if err != nil {
			return err
		}
		if source.VersionNo != request.ExpectedBalanceVersion {
			return repository.ErrConcurrentWrite
		}
		if source.LocationID != task.SourceLocationID {
			return state("reserved balance moved away from the pick source")
		}
		_, onHand, err := quantity(source.OnHandQty, "on_hand_qty", true)
		if err != nil {
			return err
		}
		_, reserved, err := quantity(source.ReservedQty, "reserved_qty", true)
		if err != nil {
			return err
		}
		if onHand.Cmp(qtyRat) < 0 || reserved.Cmp(qtyRat) < 0 {
			return state("source stock no longer covers the pick")
		}
		if source.HandlingUnitID != nil && (onHand.Cmp(qtyRat) != 0 || reserved.Cmp(qtyRat) != 0) {
			return state("handling-unit stock must be picked in full")
		}
		if task.TargetLocationID == nil {
			return state("pick task has no staging target")
		}
		newSourceOnHand := decimal(new(big.Rat).Sub(onHand, qtyRat))
		newSourceReserved := decimal(new(big.Rat).Sub(reserved, qtyRat))
		if err := local.repositories.Inventory.Balance.SetQuantities(ctx, source.ID, newSourceOnHand, newSourceReserved); err != nil {
			return err
		}
		targetIdentity := inventoryrepository.BalanceIdentity{OwnerID: source.OwnerID, WarehouseID: source.WarehouseID, LocationID: *task.TargetLocationID, ItemID: source.ItemID, LotID: source.LotID, HandlingUnitID: source.HandlingUnitID, InventoryStatusID: source.InventoryStatusID}
		target, err := local.repositories.Inventory.Balance.FindLocked(ctx, targetIdentity)
		if err != nil {
			if !errors.Is(err, inventoryrepository.ErrNotFound) {
				return err
			}
			stagingBalanceID, err = local.generateID(ctx, "INVENTORY_BALANCE", businessDate, nil, &source.WarehouseID)
			if err != nil {
				return err
			}
			target = inventorymodel.InventoryBalance{ID: stagingBalanceID, OwnerID: source.OwnerID, WarehouseID: source.WarehouseID, LocationID: *task.TargetLocationID, ItemID: source.ItemID, LotID: source.LotID, HandlingUnitID: source.HandlingUnitID, InventoryStatusID: source.InventoryStatusID, OnHandQty: qty, ReservedQty: "0", UOMID: source.UOMID, VersionNo: 1}
			if err := local.repositories.Inventory.Balance.Create(ctx, &target); err != nil {
				return err
			}
		} else {
			stagingBalanceID = target.ID
			_, targetQty, err := quantity(target.OnHandQty, "staging on_hand_qty", true)
			if err != nil {
				return err
			}
			if err := local.repositories.Inventory.Balance.SetQuantities(ctx, target.ID, decimal(new(big.Rat).Add(targetQty, qtyRat)), target.ReservedQty); err != nil {
				return err
			}
		}
		serialID, err := local.moveSerialIdentity(ctx, source.ID, &stagingBalanceID, qtyRat)
		if err != nil {
			return err
		}
		movementType, err := local.repositories.Inventory.MovementType.ByCodeShared(ctx, "PICK")
		if err != nil || !movementType.IsActive {
			return state("PICK movement type is not configured")
		}
		movementID, err = local.generateID(ctx, "MOVEMENT", businessDate, nil, &source.WarehouseID)
		if err != nil {
			return err
		}
		sourceLocation := source.LocationID
		targetLocation := *task.TargetLocationID
		sourceStatus := source.InventoryStatusID
		targetStatus := source.InventoryStatusID
		lineID := task.OutboundLineID
		movement := inventorymodel.InventoryMovement{ID: movementID, MovementTypeID: movementType.ID, OwnerID: source.OwnerID, WarehouseID: source.WarehouseID, BusinessDate: businessDate, OccurredAt: time.Now(), ItemID: source.ItemID, LotID: source.LotID, SerialID: serialID, HandlingUnitID: source.HandlingUnitID, FromLocationID: &sourceLocation, ToLocationID: &targetLocation, FromStatusID: &sourceStatus, ToStatusID: &targetStatus, Quantity: qty, UOMID: source.UOMID, SourceDocumentID: id, SourceLineID: &lineID, OperationKey: &operationKey, OperationFingerprint: &fp, CreatedBy: actor}
		if err := local.repositories.Inventory.Movement.Create(ctx, &movement); err != nil {
			return err
		}
		executionID, err = randomID(id + "-EX")
		if err != nil {
			return err
		}
		execution := model.PickExecution{ID: executionID, PickTaskID: id, SourceBalanceID: source.ID, StagingBalanceID: stagingBalanceID, PickedQty: qty, UOMID: source.UOMID, MovementID: movementID, PickedBy: actor}
		if err := local.repositories.PickExecution.Create(ctx, &execution); err != nil {
			return err
		}
		if replacement != nil {
			stagingLineID, err := randomID(replacement.StagingID + "-RL")
			if err != nil {
				return err
			}
			if err := local.repositories.StagingLine.Create(ctx, &model.OutboundStagingLine{ID: stagingLineID, StagingID: replacement.StagingID, PickExecutionID: executionID, StagingBalanceID: stagingBalanceID, StagedQty: qty, RemovedQty: "0", UOMID: source.UOMID, CreatedBy: actor}); err != nil {
				return err
			}
		}
		complete := qtyRat.Cmp(remaining) == 0
		nextTaskCode := "IN_PROGRESS"
		if complete {
			nextTaskCode = "COMPLETED"
		}
		nextTaskStatus, err := local.taskStatus(ctx, nextTaskCode)
		if err != nil {
			return err
		}
		if complete {
			nextTaskStatus, err = local.taskTransition(ctx, task.TaskStatusID, "COMPLETED")
			if err != nil {
				return err
			}
		}
		if err := local.repositories.PickTask.AddPicked(ctx, id, nextTaskStatus.ID, qty, complete); err != nil {
			return err
		}
		_, reservationPicked, err := quantity(reservation.PickedQty, "reservation picked_qty", true)
		if err != nil {
			return err
		}
		reservationComplete := new(big.Rat).Add(reservationPicked, qtyRat).Cmp(mustRat(reservation.ReservedQty)) == 0
		reservationCode := "PARTIALLY_PICKED"
		if reservationComplete {
			reservationCode = "CONSUMED"
		}
		var reservationStatusID string
		if reservationCode == "PARTIALLY_PICKED" && reservation.StatusID == partial.ID {
			reservationStatusID = partial.ID
		} else {
			reservationStatus, err := local.transition(ctx, reservation.DocumentTypeID, reservation.StatusID, reservationCode)
			if err != nil {
				return err
			}
			reservationStatusID = reservationStatus.ID
		}
		if err := local.repositories.Reservation.UpdatePicked(ctx, reservation.ID, reservationStatusID, qty); err != nil {
			return err
		}
		if replacement == nil {
			if err := local.repositories.Line.AddPicked(ctx, task.OutboundLineID, qty); err != nil {
				return err
			}
		} else if reservationComplete {
			totals, err := local.repositories.CheckResolution.Totals(ctx, replacement.ExceptionID)
			if err != nil {
				return err
			}
			fulfilled := new(big.Rat).Add(mustRat(totals.AcceptedQty), mustRat(totals.ReplacementQty))
			exceptionRow, err := local.repositories.CheckException.Get(ctx, replacement.ExceptionID)
			if err != nil {
				return err
			}
			if fulfilled.Cmp(mustRat(exceptionRow.ExceptionQty)) == 0 && mustRat(totals.CorrectedQty).Cmp(fulfilled) == 0 {
				resolved, err := local.repositories.ExceptionStatus.ByCode(ctx, "RESOLVED")
				if err != nil {
					return err
				}
				if err := local.repositories.CheckException.Resolve(ctx, replacement.ExceptionID, resolved.ID, actor); err != nil {
					return err
				}
			}
		}
		if source.HandlingUnitID != nil {
			if err := local.repositories.Inventory.HandlingUnit.Relocate(ctx, *source.HandlingUnitID, source.LocationID, *task.TargetLocationID); err != nil {
				return err
			}
		}
		remainingTasks, err := local.repositories.PickTask.RemainingByWave(ctx, task.WaveID)
		if err != nil {
			return err
		}
		if remainingTasks == 0 {
			wave, err := local.repositories.Wave.Lock(ctx, task.WaveID)
			if err != nil {
				return err
			}
			completed, err := local.transition(ctx, wave.DocumentTypeID, wave.StatusID, "COMPLETED")
			if err != nil {
				return err
			}
			return local.repositories.Wave.SetStatusLocked(ctx, wave.ID, completed.ID, actor, map[string]interface{}{"completed_at": time.Now()})
		}
		return nil
	})
	if err != nil {
		return dto.PickConfirmationResponse{}, err
	}
	task, err := s.GetPickTask(ctx, id)
	if err != nil {
		return dto.PickConfirmationResponse{}, err
	}
	return dto.PickConfirmationResponse{ExecutionID: executionID, MovementID: movementID, StagingBalanceID: stagingBalanceID, Task: task}, nil
}

func mustRat(v string) *big.Rat { x, _ := new(big.Rat).SetString(v); return x }

func (s *Service) CloseShortPick(ctx context.Context, id string, request dto.CloseShortPickRequest, actor string) (dto.PickTaskResponse, error) {
	if !validID(id, 140) || !validUUID(actor) {
		return dto.PickTaskResponse{}, invalid("invalid short-pick request")
	}
	reasonCode, err := clean(strings.ToUpper(request.ReasonCode), 40, "reason_code")
	if err != nil {
		return dto.PickTaskResponse{}, err
	}
	notes, err := optional(request.Notes, 4000, "notes")
	if err != nil {
		return dto.PickTaskResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		task, err := local.repositories.PickTask.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.PickTask.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.TaskStatusCode != "IN_PROGRESS" {
			return state("only an IN_PROGRESS pick task can be short-closed")
		}
		if task.AssignedTo == nil || *task.AssignedTo != actor {
			return state("pick task is assigned to another account")
		}
		reason, err := local.repositories.Master.ReasonCode.ByModuleCode(ctx, "OUTBOUND", reasonCode)
		if err != nil {
			return invalid("reason_code is not an active OUTBOUND reason")
		}
		_, planned, err := quantity(task.PlannedQty, "planned_qty", false)
		if err != nil {
			return err
		}
		_, picked, err := quantity(task.PickedQty, "picked_qty", true)
		if err != nil {
			return err
		}
		short := new(big.Rat).Sub(planned, picked)
		if short.Sign() <= 0 {
			return state("pick task has no short quantity")
		}
		reservation, err := local.repositories.Reservation.Lock(ctx, task.ReservationID)
		if err != nil {
			return err
		}
		balanceRow, err := local.repositories.Inventory.Balance.Get(ctx, reservation.BalanceID)
		if err != nil {
			return err
		}
		identity := inventoryrepository.BalanceIdentity{OwnerID: balanceRow.OwnerID, WarehouseID: balanceRow.WarehouseID, LocationID: balanceRow.LocationID, ItemID: balanceRow.ItemID, LotID: balanceRow.LotID, HandlingUnitID: balanceRow.HandlingUnitID, InventoryStatusID: balanceRow.InventoryStatusID}
		balance, err := local.repositories.Inventory.Balance.FindLocked(ctx, identity)
		if err != nil {
			return err
		}
		_, balanceReserved, err := quantity(balance.ReservedQty, "reserved_qty", true)
		if err != nil || balanceReserved.Cmp(short) < 0 {
			return state("reserved inventory balance is inconsistent")
		}
		if err := local.repositories.Inventory.Balance.SetQuantities(ctx, balance.ID, balance.OnHandQty, decimal(new(big.Rat).Sub(balanceReserved, short))); err != nil {
			return err
		}
		completed, err := local.taskTransition(ctx, task.TaskStatusID, "COMPLETED")
		if err != nil {
			return err
		}
		if err := local.repositories.PickTask.CloseShort(ctx, id, completed.ID, reason.ID, decimal(short), notes); err != nil {
			return err
		}
		released, err := local.transition(ctx, reservation.DocumentTypeID, reservation.StatusID, "RELEASED")
		if err != nil {
			return err
		}
		if err := local.repositories.Reservation.Release(ctx, reservation.ID, released.ID); err != nil {
			return err
		}
		if err := local.repositories.Line.SubtractAllocated(ctx, task.OutboundLineID, decimal(short)); err != nil {
			return err
		}
		remaining, err := local.repositories.PickTask.RemainingByWave(ctx, task.WaveID)
		if err != nil {
			return err
		}
		if remaining == 0 {
			wave, err := local.repositories.Wave.Lock(ctx, task.WaveID)
			if err != nil {
				return err
			}
			done, err := local.transition(ctx, wave.DocumentTypeID, wave.StatusID, "COMPLETED")
			if err != nil {
				return err
			}
			return local.repositories.Wave.SetStatusLocked(ctx, wave.ID, done.ID, actor, map[string]interface{}{"completed_at": time.Now()})
		}
		return nil
	})
	if err != nil {
		return dto.PickTaskResponse{}, err
	}
	return s.GetPickTask(ctx, id)
}
