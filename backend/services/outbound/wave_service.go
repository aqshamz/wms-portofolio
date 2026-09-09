package outbound

import (
	"context"
	"math/big"
	"strings"
	"time"

	dto "wms-api/dto/outbound"
	model "wms-api/models/outbound"
	repository "wms-api/repository/outbound"
)

func (s *Service) CreateWave(ctx context.Context, request dto.CreateWaveRequest, actor string) (dto.WaveResponse, error) {
	if !validUUID(request.OwnerID) || !validUUID(request.WarehouseID) || !validUUID(actor) {
		return dto.WaveResponse{}, invalid("owner_id, warehouse_id and actor must be UUIDs")
	}
	request.OwnerID = strings.ToLower(request.OwnerID)
	request.WarehouseID = strings.ToLower(request.WarehouseID)
	businessDate, err := date(request.BusinessDate, "business_date")
	if err != nil {
		return dto.WaveResponse{}, err
	}
	waveCode, err := clean(strings.ToUpper(request.WaveTypeCode), 40, "wave_type_code")
	if err != nil {
		return dto.WaveResponse{}, err
	}
	planned, err := optionalTimestamp(request.PlannedReleaseAt, "planned_release_at")
	if err != nil {
		return dto.WaveResponse{}, err
	}
	notes, err := optional(request.Notes, 4000, "notes")
	if err != nil {
		return dto.WaveResponse{}, err
	}
	if len(request.OutboundIDs) == 0 || len(request.OutboundIDs) > 1000 {
		return dto.WaveResponse{}, invalid("outbound_ids must contain 1..1000 entries")
	}
	var id string
	err = s.transaction(ctx, func(local *Service) error {
		owner, err := local.repositories.Master.Organization.FindByID(ctx, request.OwnerID)
		if err != nil || !owner.IsActive {
			return invalid("owner does not exist or is inactive")
		}
		warehouse, err := local.repositories.Master.Warehouse.FindByID(ctx, request.WarehouseID)
		if err != nil || !warehouse.IsActive {
			return invalid("warehouse does not exist or is inactive")
		}
		assigned, err := local.repositories.Master.WarehouseOwner.IsActive(ctx, request.WarehouseID, request.OwnerID)
		if err != nil || !assigned {
			return invalid("owner is not active for the warehouse")
		}
		waveType, err := local.repositories.WaveType.ByCode(ctx, waveCode)
		if err != nil {
			return invalid("wave_type_code is not configured")
		}
		strategy, err := local.pickingStrategy(ctx, request.OwnerID, request.WarehouseID, request.PickingStrategyID)
		if err != nil {
			return err
		}
		kind, err := local.documentType(ctx, "OUTBOUND_WAVE")
		if err != nil {
			return err
		}
		initial, err := local.initialStatus(ctx, kind.ID)
		if err != nil {
			return err
		}
		id, err = local.generateID(ctx, "OUTBOUND_WAVE", businessDate, nil, &warehouse.ID)
		if err != nil {
			return err
		}
		wave := model.OutboundWave{ID: id, DocumentTypeID: kind.ID, StatusID: initial.ID, WaveTypeID: waveType.ID, OwnerID: request.OwnerID, WarehouseID: request.WarehouseID, PickingStrategyID: &strategy.ID, BusinessDate: businessDate, PlannedReleaseAt: planned, Notes: notes, VersionNo: 1, CreatedBy: actor, UpdatedBy: actor}
		if err := local.repositories.Wave.Create(ctx, &wave); err != nil {
			return err
		}
		links := make([]model.OutboundWaveOrder, 0, len(request.OutboundIDs))
		seen := map[string]bool{}
		for _, outboundID := range request.OutboundIDs {
			if !validID(outboundID, 120) || seen[outboundID] {
				return invalid("outbound_ids contains an invalid or duplicate ID")
			}
			seen[outboundID] = true
			order, err := local.repositories.Order.Lock(ctx, outboundID)
			if err != nil {
				return err
			}
			row, err := local.repositories.Order.Get(ctx, outboundID)
			if err != nil {
				return err
			}
			if order.OwnerID != request.OwnerID || order.WarehouseID != request.WarehouseID {
				return invalid("all outbound orders must match the wave owner and warehouse")
			}
			if row.StatusCode != "ALLOCATED" {
				return state("only fully ALLOCATED outbound orders can enter a wave")
			}
			activeWave, err := local.repositories.WaveOrder.HasActiveWave(ctx, outboundID)
			if err != nil {
				return err
			}
			if activeWave {
				return state("outbound order already belongs to an active wave")
			}
			links = append(links, model.OutboundWaveOrder{WaveID: id, OutboundID: outboundID, AddedBy: actor})
		}
		return local.repositories.WaveOrder.CreateBatch(ctx, links)
	})
	if err != nil {
		return dto.WaveResponse{}, err
	}
	return s.GetWave(ctx, id)
}

func (s *Service) GetWave(ctx context.Context, id string) (dto.WaveResponse, error) {
	if !validID(id, 120) {
		return dto.WaveResponse{}, invalid("invalid wave_id")
	}
	row, err := s.repositories.Wave.Get(ctx, id)
	if err != nil {
		return dto.WaveResponse{}, err
	}
	orders, err := s.repositories.WaveOrder.List(ctx, id)
	if err != nil {
		return dto.WaveResponse{}, err
	}
	return mapWave(row, orders), nil
}
func (s *Service) ListWaves(ctx context.Context, f repository.ListFilter) (dto.PageResponse[dto.WaveResponse], error) {
	if !validUUID(f.OwnerID) {
		return dto.PageResponse[dto.WaveResponse]{}, invalid("owner_id is required and must be a UUID")
	}
	rows, total, err := s.repositories.Wave.List(ctx, f)
	if err != nil {
		return dto.PageResponse[dto.WaveResponse]{}, err
	}
	items := make([]dto.WaveResponse, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapWave(v, nil))
	}
	return page(items, f.Page, f.PageSize, total), nil
}

func (s *Service) ReleaseWave(ctx context.Context, id string, request dto.ReleaseWaveRequest, actor string) (dto.WaveResponse, error) {
	if !validID(id, 120) || !validUUID(request.StagingLocationID) || !validUUID(actor) || request.ExpectedVersion < 1 {
		return dto.WaveResponse{}, invalid("invalid wave release request")
	}
	priorityCode, err := clean(strings.ToUpper(request.PriorityCode), 40, "priority_code")
	if err != nil {
		return dto.WaveResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		wave, err := local.repositories.Wave.Lock(ctx, id)
		if err != nil {
			return err
		}
		if wave.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		row, err := local.repositories.Wave.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "DRAFT" {
			return state("only a DRAFT wave can be released")
		}
		location, err := local.repositories.Inventory.Location.FindByID(ctx, strings.ToLower(request.StagingLocationID))
		if err != nil || !location.IsActive || location.IsLocked || location.WarehouseID != wave.WarehouseID || location.LocationTypeCode != "STAGING" {
			return invalid("staging_location_id must be an active unlocked STAGING location in the wave warehouse")
		}
		taskType, err := local.repositories.Master.TaskType.ByCode(ctx, "PICK")
		if err != nil || !taskType.IsActive {
			return state("PICK task type is not configured")
		}
		open, err := local.repositories.Master.TaskStatus.ByCode(ctx, "OPEN")
		if err != nil || !open.IsActive || !open.IsInitial {
			return state("OPEN initial task status is not configured")
		}
		priority, err := local.repositories.Master.TaskPriority.ByCode(ctx, priorityCode)
		if err != nil || !priority.IsActive {
			return invalid("priority_code is not configured")
		}
		orders, err := local.repositories.WaveOrder.List(ctx, id)
		if err != nil {
			return err
		}
		warehouse, err := local.repositories.Master.Warehouse.FindByID(ctx, wave.WarehouseID)
		if err != nil {
			return err
		}
		tasks := make([]model.PickTask, 0)
		for _, link := range orders {
			order, err := local.repositories.Order.Lock(ctx, link.OutboundID)
			if err != nil {
				return err
			}
			orderRow, err := local.repositories.Order.Get(ctx, order.ID)
			if err != nil {
				return err
			}
			if orderRow.StatusCode != "ALLOCATED" {
				return state("every wave order must remain ALLOCATED")
			}
			reservations, err := local.repositories.Reservation.ListActiveByOrder(ctx, order.ID)
			if err != nil {
				return err
			}
			if len(reservations) == 0 {
				return state("wave order has no active reservations")
			}
			for _, reservation := range reservations {
				if wave.PickingStrategyID != nil && (reservation.PickingStrategyID == nil || *reservation.PickingStrategyID != *wave.PickingStrategyID) {
					return state("reservation picking strategy does not match the wave")
				}
				balance, err := local.repositories.Inventory.Balance.Get(ctx, reservation.BalanceID)
				if err != nil {
					return err
				}
				if balance.LocationID == location.ID {
					return state("reservation source is already the staging location")
				}
				_, reserved, err := quantity(reservation.ReservedQty, "reserved_qty", false)
				if err != nil {
					return err
				}
				_, picked, err := quantity(reservation.PickedQty, "picked_qty", true)
				if err != nil {
					return err
				}
				remaining := new(big.Rat).Sub(reserved, picked)
				taskID, err := local.generateID(ctx, "PICK_TASK", wave.BusinessDate, nil, &warehouse.ID)
				if err != nil {
					return err
				}
				tasks = append(tasks, model.PickTask{ID: taskID, TaskTypeID: taskType.ID, TaskStatusID: open.ID, TaskPriorityID: priority.ID, WaveID: id, ReservationID: reservation.ID, OutboundLineID: reservation.OutboundLineID, SourceLocationID: balance.LocationID, TargetLocationID: &location.ID, PlannedQty: decimal(remaining), PickedQty: "0", UOMID: reservation.UOMID, ShortQty: "0", CreatedBy: actor})
			}
		}
		if len(tasks) == 0 {
			return state("wave has no pick tasks to create")
		}
		if err := local.repositories.PickTask.CreateBatch(ctx, tasks); err != nil {
			return err
		}
		released, err := local.transition(ctx, wave.DocumentTypeID, wave.StatusID, "RELEASED")
		if err != nil {
			return err
		}
		if err := local.repositories.Wave.SetStatus(ctx, id, released.ID, actor, wave.VersionNo, map[string]interface{}{"released_at": time.Now()}); err != nil {
			return err
		}
		for _, link := range orders {
			order, err := local.repositories.Order.Lock(ctx, link.OutboundID)
			if err != nil {
				return err
			}
			target, err := local.transition(ctx, order.DocumentTypeID, order.StatusID, "WAVED")
			if err != nil {
				return err
			}
			if err := local.repositories.Order.SetStatusLocked(ctx, order.ID, target.ID, actor, nil); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return dto.WaveResponse{}, err
	}
	return s.GetWave(ctx, id)
}

func (s *Service) CancelWave(ctx context.Context, id string, request dto.TransitionRequest, actor string) (dto.WaveResponse, error) {
	err := s.transaction(ctx, func(local *Service) error {
		wave, err := local.repositories.Wave.Lock(ctx, id)
		if err != nil {
			return err
		}
		if wave.VersionNo != request.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		row, err := local.repositories.Wave.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "DRAFT" || row.PickTaskCount > 0 {
			return state("only an empty DRAFT wave can be cancelled")
		}
		target, err := local.transition(ctx, wave.DocumentTypeID, wave.StatusID, "CANCELLED")
		if err != nil {
			return err
		}
		return local.repositories.Wave.SetStatus(ctx, id, target.ID, actor, wave.VersionNo, nil)
	})
	if err != nil {
		return dto.WaveResponse{}, err
	}
	return s.GetWave(ctx, id)
}
