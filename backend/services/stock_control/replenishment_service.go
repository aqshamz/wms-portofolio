package stockcontrol

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strings"
	"time"

	inventorydto "wms-api/dto/inventory"
	dto "wms-api/dto/stock_control"
	inventorymodel "wms-api/models/inventory"
	mastermodel "wms-api/models/master"
	model "wms-api/models/stock_control"
	inventoryrepo "wms-api/repository/inventory"
	repository "wms-api/repository/stock_control"
	inventory "wms-api/services/inventory"
)

func replenishmentID() (string, error) {
	data := make([]byte, 12)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return "RPL-" + strings.ToUpper(hex.EncodeToString(data)), nil
}

func replenishmentPage[T any](items []T, page, size int, total int64) dto.PageResponse[T] {
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(size) - 1) / int64(size))
	}
	return dto.PageResponse[T]{Items: items, Page: page, PageSize: size, TotalItems: total, TotalPages: totalPages}
}

func mapReplenishment(row repository.ReplenishmentTaskRow) dto.ReplenishmentTaskResponse {
	return dto.ReplenishmentTaskResponse{
		ID: row.ID, SourceBalanceID: row.SourceBalanceID, SourceBalanceVersionNo: row.SourceBalanceVersionNo,
		OwnerID: row.OwnerID, WarehouseID: row.WarehouseID, ItemID: row.ItemID, ItemCode: row.ItemCode, ItemName: row.ItemName,
		LotID: row.LotID, LotNumber: row.LotNumber, SerialID: row.SerialID, SerialNumber: row.SerialNumber,
		HandlingUnitID: row.HandlingUnitID, HandlingUnitBarcode: row.HandlingUnitBarcode,
		InventoryStatusID: row.InventoryStatusID, InventoryStatusCode: row.InventoryStatusCode,
		SourceLocationID: row.SourceLocationID, SourceLocationCode: row.SourceLocationCode,
		TargetLocationID: row.TargetLocationID, TargetLocationCode: row.TargetLocationCode,
		PlannedQty: row.PlannedQty, CompletedQty: row.CompletedQty, UOMID: row.UOMID, UOMCode: row.UOMCode,
		TaskStatusCode: row.TaskStatusCode, TaskPriorityCode: row.TaskPriorityCode,
		AssignedTo: row.AssignedTo, AssignedUsername: row.AssignedUsername, AssignedDisplayName: row.AssignedDisplayName,
		Notes: row.Notes, StartedAt: row.StartedAt, CompletedAt: row.CompletedAt, CancelledAt: row.CancelledAt,
		CancelledBy: row.CancelledBy, CancellationReason: row.CancellationReason,
		InventoryMovementID: row.InventoryMovementID, ResultingBalanceID: row.ResultingBalanceID,
		VersionNo: row.VersionNo, CreatedAt: row.CreatedAt,
	}
}

func (s *Service) GetReplenishment(ctx context.Context, id, actor string) (dto.ReplenishmentTaskResponse, error) {
	if len(strings.TrimSpace(id)) == 0 || len(id) > 120 {
		return dto.ReplenishmentTaskResponse{}, invalid("invalid replenishment_task_id")
	}
	row, err := s.repositories.Replenishments.Get(ctx, id)
	if err != nil {
		return dto.ReplenishmentTaskResponse{}, err
	}
	if err := authorize(ctx, s.repositories, actor, scopeBalance(row.OwnerID, row.WarehouseID)); err != nil {
		return dto.ReplenishmentTaskResponse{}, err
	}
	return mapReplenishment(row), nil
}

func scopeBalance(ownerID, warehouseID string) inventoryrepo.BalanceRow {
	return inventoryrepo.BalanceRow{InventoryBalance: inventorymodel.InventoryBalance{OwnerID: ownerID, WarehouseID: warehouseID}}
}

func (s *Service) ListReplenishments(ctx context.Context, filter repository.ReplenishmentFilter, actor string) (dto.PageResponse[dto.ReplenishmentTaskResponse], error) {
	filter.OwnerID = strings.ToLower(strings.TrimSpace(filter.OwnerID))
	filter.WarehouseID = strings.ToLower(strings.TrimSpace(filter.WarehouseID))
	filter.StatusCode = strings.ToUpper(strings.TrimSpace(filter.StatusCode))
	filter.Search = strings.TrimSpace(filter.Search)
	if filter.OwnerID == "" || filter.WarehouseID == "" || filter.Page < 1 || filter.PageSize < 1 || filter.PageSize > 100 || len(filter.Search) > 160 {
		return dto.PageResponse[dto.ReplenishmentTaskResponse]{}, invalid("invalid replenishment filters")
	}
	allowed, err := s.repositories.Scope.Allowed(ctx, actor, filter.OwnerID, filter.WarehouseID)
	if err != nil {
		return dto.PageResponse[dto.ReplenishmentTaskResponse]{}, err
	}
	if !allowed {
		return dto.PageResponse[dto.ReplenishmentTaskResponse]{}, ErrForbidden
	}
	rows, total, err := s.repositories.Replenishments.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.ReplenishmentTaskResponse]{}, err
	}
	items := make([]dto.ReplenishmentTaskResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapReplenishment(row))
	}
	return replenishmentPage(items, filter.Page, filter.PageSize, total), nil
}

func (s *Service) CreateReplenishment(ctx context.Context, request dto.CreateReplenishmentRequest, actor string) (dto.ReplenishmentTaskResponse, error) {
	qty, qtyNumber, err := number(request.Quantity, true)
	if err != nil {
		return dto.ReplenishmentTaskResponse{}, err
	}
	request.PriorityCode = strings.ToUpper(strings.TrimSpace(request.PriorityCode))
	if !reasonPattern.MatchString(request.PriorityCode) {
		return dto.ReplenishmentTaskResponse{}, invalid("invalid priority_code")
	}
	if request.Notes != nil {
		notes := strings.TrimSpace(*request.Notes)
		request.Notes = &notes
	}
	id, err := replenishmentID()
	if err != nil {
		return dto.ReplenishmentTaskResponse{}, err
	}
	err = s.repositories.Transaction(ctx, func(r *repository.Repositories, ir *inventoryrepo.Repositories) error {
		source, err := r.Balances.Get(ctx, request.SourceBalanceID)
		if err != nil {
			return err
		}
		if err := authorize(ctx, r, actor, source); err != nil {
			return err
		}
		eligible, err := r.Replenishments.SourceEligible(ctx, source.ID)
		if err != nil || !eligible {
			return invalid("source must be active, unlocked, pickable reserve storage outside a pick face")
		}
		eligible, err = r.Replenishments.TargetEligible(ctx, source.WarehouseID, request.TargetLocationID)
		if err != nil || !eligible || source.LocationID == request.TargetLocationID {
			return invalid("target must be an active, unlocked storage-capable pick face in the same warehouse")
		}
		if err := ir.Balance.LockStockKey(ctx, source.OwnerID, source.WarehouseID, source.ItemID); err != nil {
			return err
		}
		locked, err := ir.Balance.FindLocked(ctx, inventoryrepo.BalanceIdentity{OwnerID: source.OwnerID, WarehouseID: source.WarehouseID, LocationID: source.LocationID, ItemID: source.ItemID, LotID: source.LotID, HandlingUnitID: source.HandlingUnitID, InventoryStatusID: source.InventoryStatusID})
		if err != nil || locked.ID != source.ID || locked.VersionNo != request.ExpectedVersion {
			return invalid("source balance changed; reload before creating the task")
		}
		onHand, _ := new(big.Rat).SetString(locked.OnHandQty)
		reserved, _ := new(big.Rat).SetString(locked.ReservedQty)
		available := new(big.Rat).Sub(onHand, reserved)
		if available.Cmp(qtyNumber) < 0 {
			return invalid("replenishment quantity exceeds available stock")
		}
		if source.SerialControlled {
			if request.SerialID == nil || qtyNumber.Cmp(big.NewRat(1, 1)) != 0 {
				return invalid("serialized replenishment requires one serial and quantity 1")
			}
			state, stateErr := ir.SerialState.Get(ctx, *request.SerialID)
			if stateErr != nil || state.BalanceID != source.ID {
				return invalid("serial is not in the source balance")
			}
		} else if request.SerialID != nil {
			return invalid("serial_id is only allowed for serialized stock")
		}
		if source.HandlingUnitID != nil {
			positive, countErr := ir.Balance.CountPositiveForHandlingUnit(ctx, *source.HandlingUnitID)
			if countErr != nil {
				return countErr
			}
			children, childErr := ir.HandlingUnit.CountChildren(ctx, *source.HandlingUnitID)
			if childErr != nil {
				return childErr
			}
			if positive != 1 || children != 0 || reserved.Sign() != 0 || onHand.Cmp(qtyNumber) != 0 {
				return invalid("handling-unit replenishment requires its one full unreserved balance and no child handling units")
			}
		}
		var taskType mastermodel.TaskType
		if err := r.Replenishments.Reference(ctx, "task_type", "REPLENISHMENT", &taskType); err != nil {
			return err
		}
		var status mastermodel.TaskStatus
		if err := r.Replenishments.Reference(ctx, "task_status", "OPEN", &status); err != nil {
			return err
		}
		var priority mastermodel.TaskPriority
		if err := r.Replenishments.Reference(ctx, "task_priority", request.PriorityCode, &priority); err != nil {
			return invalid("priority_code does not exist or is inactive")
		}
		reserved.Add(reserved, qtyNumber)
		if err := ir.Balance.SetQuantities(ctx, locked.ID, locked.OnHandQty, reserved.FloatString(6)); err != nil {
			return err
		}
		return r.Replenishments.Create(ctx, &model.ReplenishmentTask{ID: id, TaskTypeID: taskType.ID, TaskStatusID: status.ID, TaskPriorityID: priority.ID, SourceBalanceID: source.ID, OwnerID: source.OwnerID, WarehouseID: source.WarehouseID, ItemID: source.ItemID, LotID: source.LotID, SerialID: request.SerialID, HandlingUnitID: source.HandlingUnitID, InventoryStatusID: source.InventoryStatusID, SourceLocationID: source.LocationID, TargetLocationID: request.TargetLocationID, PlannedQty: qty, CompletedQty: "0.000000", UOMID: source.UOMID, Notes: request.Notes, CreatedBy: actor, UpdatedBy: &actor, VersionNo: 1})
	})
	if err != nil {
		return dto.ReplenishmentTaskResponse{}, err
	}
	return s.GetReplenishment(ctx, id, actor)
}

func (s *Service) StartReplenishment(ctx context.Context, id string, request dto.ReplenishmentTransitionRequest, actor string) (dto.ReplenishmentTaskResponse, error) {
	err := s.repositories.Transaction(ctx, func(r *repository.Repositories, _ *inventoryrepo.Repositories) error {
		task, err := r.Replenishments.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := r.Replenishments.Get(ctx, id)
		if err != nil {
			return err
		}
		if err := authorize(ctx, r, actor, scopeBalance(task.OwnerID, task.WarehouseID)); err != nil {
			return err
		}
		if row.TaskStatusCode == "IN_PROGRESS" && task.AssignedTo != nil && *task.AssignedTo == actor {
			return nil
		}
		if row.TaskStatusCode != "OPEN" && row.TaskStatusCode != "ASSIGNED" {
			return invalid("only an OPEN or ASSIGNED replenishment can be started")
		}
		if row.TaskStatusCode == "ASSIGNED" && (task.AssignedTo == nil || *task.AssignedTo != actor) {
			return ErrForbidden
		}
		var status mastermodel.TaskStatus
		if err := r.Replenishments.Reference(ctx, "task_status", "IN_PROGRESS", &status); err != nil {
			return err
		}
		return r.Replenishments.Start(ctx, id, status.ID, actor, request.ExpectedVersion)
	})
	if err != nil {
		return dto.ReplenishmentTaskResponse{}, err
	}
	return s.GetReplenishment(ctx, id, actor)
}

func (s *Service) AssignReplenishment(ctx context.Context, id string, request dto.AssignReplenishmentRequest, actor string) (dto.ReplenishmentTaskResponse, error) {
	err := s.repositories.Transaction(ctx, func(r *repository.Repositories, _ *inventoryrepo.Repositories) error {
		task, err := r.Replenishments.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := r.Replenishments.Get(ctx, id)
		if err != nil {
			return err
		}
		if err := authorize(ctx, r, actor, scopeBalance(task.OwnerID, task.WarehouseID)); err != nil {
			return err
		}
		if row.TaskStatusCode != "OPEN" && row.TaskStatusCode != "ASSIGNED" {
			return invalid("only an OPEN or ASSIGNED replenishment can be assigned")
		}
		allowed, err := r.Replenishments.AccountCanExecute(ctx, request.AccountID, task.OwnerID, task.WarehouseID)
		if err != nil {
			return err
		}
		if !allowed {
			return invalid("assigned account requires owner and warehouse access and INVENTORY.MOVE permission")
		}
		var status mastermodel.TaskStatus
		if err := r.Replenishments.Reference(ctx, "task_status", "ASSIGNED", &status); err != nil {
			return err
		}
		return r.Replenishments.Assign(ctx, id, status.ID, request.AccountID, actor, request.ExpectedVersion)
	})
	if err != nil {
		return dto.ReplenishmentTaskResponse{}, err
	}
	return s.GetReplenishment(ctx, id, actor)
}

func (s *Service) CompleteReplenishment(ctx context.Context, id string, request dto.CompleteReplenishmentRequest, actor string) (dto.ReplenishmentTaskResponse, error) {
	if _, err := time.Parse("2006-01-02", request.BusinessDate); err != nil {
		return dto.ReplenishmentTaskResponse{}, invalid("business_date must be YYYY-MM-DD")
	}
	err := s.repositories.Transaction(ctx, func(r *repository.Repositories, ir *inventoryrepo.Repositories) error {
		task, err := r.Replenishments.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := r.Replenishments.Get(ctx, id)
		if err != nil {
			return err
		}
		if err := authorize(ctx, r, actor, scopeBalance(task.OwnerID, task.WarehouseID)); err != nil {
			return err
		}
		if row.TaskStatusCode == "COMPLETED" {
			return nil
		}
		if row.TaskStatusCode != "IN_PROGRESS" {
			return invalid("only an IN_PROGRESS replenishment can be completed")
		}
		if task.AssignedTo == nil || *task.AssignedTo != actor {
			return ErrForbidden
		}
		eligible, err := r.Replenishments.TargetEligible(ctx, task.WarehouseID, task.TargetLocationID)
		if err != nil || !eligible {
			return invalid("replenishment target is no longer an eligible pick face")
		}
		if err := ir.Balance.LockStockKey(ctx, task.OwnerID, task.WarehouseID, task.ItemID); err != nil {
			return err
		}
		balance, err := ir.Balance.FindLocked(ctx, inventoryrepo.BalanceIdentity{OwnerID: task.OwnerID, WarehouseID: task.WarehouseID, LocationID: task.SourceLocationID, ItemID: task.ItemID, LotID: task.LotID, HandlingUnitID: task.HandlingUnitID, InventoryStatusID: task.InventoryStatusID})
		if err != nil || balance.ID != task.SourceBalanceID || balance.VersionNo != request.ExpectedBalanceVersion {
			return invalid("source balance changed; refresh before completing")
		}
		reserved, _ := new(big.Rat).SetString(balance.ReservedQty)
		planned, _ := new(big.Rat).SetString(task.PlannedQty)
		if reserved.Cmp(planned) < 0 {
			return invalid("task reservation is no longer available")
		}
		reserved.Sub(reserved, planned)
		if err := ir.Balance.SetQuantities(ctx, balance.ID, balance.OnHandQty, reserved.FloatString(6)); err != nil {
			return err
		}
		reason, err := r.Reasons.ByCode(ctx, "INVENTORY", "REPLENISHMENT")
		if err != nil {
			return err
		}
		newVersion := balance.VersionNo + 1
		serials := []string{}
		if task.SerialID != nil {
			serials = append(serials, *task.SerialID)
		}
		posted, err := inventory.NewService(ir).PostMovement(ctx, inventorydto.PostingRequest{OperationKey: "stock.replenishment." + task.ID, MovementTypeCode: "REPLENISHMENT", OwnerID: task.OwnerID, WarehouseID: task.WarehouseID, BusinessDate: request.BusinessDate, ItemID: task.ItemID, LotID: task.LotID, HandlingUnitID: task.HandlingUnitID, From: &inventorydto.BalanceDimension{LocationID: task.SourceLocationID, InventoryStatusID: task.InventoryStatusID}, To: &inventorydto.BalanceDimension{LocationID: task.TargetLocationID, InventoryStatusID: task.InventoryStatusID}, Quantity: task.PlannedQty, SerialIDs: serials, ExpectedSourceVersion: &newVersion, SourceDocumentID: task.ID, ReasonCodeID: &reason.ID, Notes: task.Notes, RelocateHandlingUnit: task.HandlingUnitID != nil}, actor)
		if err != nil {
			return err
		}
		var status mastermodel.TaskStatus
		if err := r.Replenishments.Reference(ctx, "task_status", "COMPLETED", &status); err != nil {
			return err
		}
		return r.Replenishments.Complete(ctx, id, status.ID, posted.Movement.ID, posted.ToBalance.ID, task.PlannedQty, actor, task.VersionNo)
	})
	if err != nil {
		return dto.ReplenishmentTaskResponse{}, err
	}
	return s.GetReplenishment(ctx, id, actor)
}

func (s *Service) CancelReplenishment(ctx context.Context, id string, request dto.CancelReplenishmentRequest, actor string) (dto.ReplenishmentTaskResponse, error) {
	reason := strings.TrimSpace(request.Reason)
	if reason == "" {
		return dto.ReplenishmentTaskResponse{}, invalid("reason is required")
	}
	err := s.repositories.Transaction(ctx, func(r *repository.Repositories, ir *inventoryrepo.Repositories) error {
		task, err := r.Replenishments.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := r.Replenishments.Get(ctx, id)
		if err != nil {
			return err
		}
		if err := authorize(ctx, r, actor, scopeBalance(task.OwnerID, task.WarehouseID)); err != nil {
			return err
		}
		if row.TaskStatusCode == "CANCELLED" {
			return nil
		}
		if row.TaskStatusCode != "OPEN" && row.TaskStatusCode != "ASSIGNED" && row.TaskStatusCode != "IN_PROGRESS" {
			return invalid("completed replenishment cannot be cancelled")
		}
		if row.TaskStatusCode == "IN_PROGRESS" && (task.AssignedTo == nil || *task.AssignedTo != actor) {
			return ErrForbidden
		}
		if err := ir.Balance.LockStockKey(ctx, task.OwnerID, task.WarehouseID, task.ItemID); err != nil {
			return err
		}
		balance, err := ir.Balance.FindLocked(ctx, inventoryrepo.BalanceIdentity{OwnerID: task.OwnerID, WarehouseID: task.WarehouseID, LocationID: task.SourceLocationID, ItemID: task.ItemID, LotID: task.LotID, HandlingUnitID: task.HandlingUnitID, InventoryStatusID: task.InventoryStatusID})
		if err != nil || balance.ID != task.SourceBalanceID || balance.VersionNo != request.ExpectedBalanceVersion {
			return invalid("source balance changed; refresh before cancelling")
		}
		reserved, _ := new(big.Rat).SetString(balance.ReservedQty)
		planned, _ := new(big.Rat).SetString(task.PlannedQty)
		if reserved.Cmp(planned) < 0 {
			return invalid("task reservation is no longer available")
		}
		reserved.Sub(reserved, planned)
		if err := ir.Balance.SetQuantities(ctx, balance.ID, balance.OnHandQty, reserved.FloatString(6)); err != nil {
			return err
		}
		var status mastermodel.TaskStatus
		if err := r.Replenishments.Reference(ctx, "task_status", "CANCELLED", &status); err != nil {
			return err
		}
		return r.Replenishments.Cancel(ctx, id, status.ID, actor, reason, task.VersionNo)
	})
	if err != nil {
		return dto.ReplenishmentTaskResponse{}, err
	}
	return s.GetReplenishment(ctx, id, actor)
}

func (s *Service) ListReplenishmentAssignees(ctx context.Context, id, search string, page, size int, actor string) (dto.PageResponse[dto.ReplenishmentAssigneeResponse], error) {
	search = strings.TrimSpace(search)
	if page < 1 || size < 1 || size > 100 || len(search) > 160 {
		return dto.PageResponse[dto.ReplenishmentAssigneeResponse]{}, invalid("invalid assignee filters")
	}
	task, err := s.GetReplenishment(ctx, id, actor)
	if err != nil {
		return dto.PageResponse[dto.ReplenishmentAssigneeResponse]{}, err
	}
	rows, total, err := s.repositories.Replenishments.ListAssignees(ctx, task.OwnerID, task.WarehouseID, search, page, size)
	if err != nil {
		return dto.PageResponse[dto.ReplenishmentAssigneeResponse]{}, err
	}
	items := make([]dto.ReplenishmentAssigneeResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.ReplenishmentAssigneeResponse{AccountID: row.AccountID, Username: row.Username, DisplayName: row.DisplayName})
	}
	return replenishmentPage(items, page, size, total), nil
}

func (s *Service) ListReplenishmentTargets(ctx context.Context, ownerID, warehouseID, search string, page, size int, actor string) (dto.PageResponse[dto.ReplenishmentTargetResponse], error) {
	ownerID = strings.ToLower(strings.TrimSpace(ownerID))
	warehouseID = strings.ToLower(strings.TrimSpace(warehouseID))
	search = strings.TrimSpace(search)
	if ownerID == "" || warehouseID == "" || page < 1 || size < 1 || size > 100 || len(search) > 160 {
		return dto.PageResponse[dto.ReplenishmentTargetResponse]{}, invalid("invalid target filters")
	}
	allowed, err := s.repositories.Scope.Allowed(ctx, actor, ownerID, warehouseID)
	if err != nil {
		return dto.PageResponse[dto.ReplenishmentTargetResponse]{}, err
	}
	if !allowed {
		return dto.PageResponse[dto.ReplenishmentTargetResponse]{}, ErrForbidden
	}
	rows, total, err := s.repositories.Replenishments.ListTargets(ctx, warehouseID, search, page, size)
	if err != nil {
		return dto.PageResponse[dto.ReplenishmentTargetResponse]{}, err
	}
	items := make([]dto.ReplenishmentTargetResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.ReplenishmentTargetResponse{LocationID: row.LocationID, Code: row.Code, ZoneCode: row.ZoneCode, LocationTypeCode: row.LocationTypeCode})
	}
	return replenishmentPage(items, page, size, total), nil
}
