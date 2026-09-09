package outbound

import (
	"context"
	"fmt"

	dto "wms-api/dto/outbound"
	model "wms-api/models/outbound"
	repository "wms-api/repository/outbound"
)

func (s *Service) CreateStaging(ctx context.Context, request dto.CreateStagingRequest, actor string) (dto.StagingResponse, error) {
	if !validID(request.OutboundID, 120) || !validID(request.WaveID, 120) || !validUUID(actor) {
		return dto.StagingResponse{}, invalid("invalid staging request")
	}
	notes, err := optional(request.Notes, 4000, "notes")
	if err != nil {
		return dto.StagingResponse{}, err
	}
	var id string
	err = s.transaction(ctx, func(local *Service) error {
		exists, err := local.repositories.Staging.ExistsForOrderWave(ctx, request.OutboundID, request.WaveID)
		if err != nil {
			return err
		}
		if exists {
			return repository.ErrConflict
		}
		order, err := local.repositories.Order.Lock(ctx, request.OutboundID)
		if err != nil {
			return err
		}
		orderRow, err := local.repositories.Order.Get(ctx, order.ID)
		if err != nil {
			return err
		}
		if orderRow.StatusCode != "PICKING" {
			return state("outbound order must be PICKING before staging is created")
		}
		wave, err := local.repositories.Wave.Lock(ctx, request.WaveID)
		if err != nil {
			return err
		}
		waveRow, err := local.repositories.Wave.Get(ctx, wave.ID)
		if err != nil {
			return err
		}
		if waveRow.StatusCode != "COMPLETED" {
			return state("wave must be COMPLETED before staging is created")
		}
		links, err := local.repositories.WaveOrder.List(ctx, wave.ID)
		if err != nil {
			return err
		}
		found := false
		for _, v := range links {
			if v.OutboundID == order.ID {
				found = true
			}
		}
		if !found {
			return invalid("outbound order does not belong to the wave")
		}
		remaining, err := local.repositories.PickTask.RemainingByOrderWave(ctx, order.ID, wave.ID)
		if err != nil {
			return err
		}
		if remaining != 0 {
			return state("all order pick tasks must be final")
		}
		executions, err := local.repositories.PickExecution.ListForOrderWave(ctx, order.ID, wave.ID)
		if err != nil {
			return err
		}
		if len(executions) == 0 {
			return state("order has no picked stock to stage")
		}
		var locationID string
		for _, execution := range executions {
			task, err := local.repositories.PickTask.Get(ctx, execution.PickTaskID)
			if err != nil {
				return err
			}
			if task.TargetLocationID == nil {
				return state("pick execution has no staging target")
			}
			if locationID == "" {
				locationID = *task.TargetLocationID
			} else if locationID != *task.TargetLocationID {
				return state("one staging document cannot span multiple staging locations")
			}
		}
		kind, err := local.documentType(ctx, "OUTBOUND_STAGING")
		if err != nil {
			return err
		}
		open, err := local.initialStatus(ctx, kind.ID)
		if err != nil {
			return err
		}
		id, err = local.generateID(ctx, "OUTBOUND_STAGING", order.BusinessDate, nil, &order.WarehouseID)
		if err != nil {
			return err
		}
		header := model.OutboundStaging{ID: id, DocumentTypeID: kind.ID, StatusID: open.ID, OutboundID: order.ID, WaveID: wave.ID, WarehouseID: order.WarehouseID, StagingLocationID: locationID, Notes: notes, CreatedBy: actor}
		if err := local.repositories.Staging.Create(ctx, &header); err != nil {
			return err
		}
		lines := make([]model.OutboundStagingLine, 0, len(executions))
		for i, x := range executions {
			lines = append(lines, model.OutboundStagingLine{ID: fmt.Sprintf("%s-L%06d", id, i+1), StagingID: id, PickExecutionID: x.ID, StagingBalanceID: x.StagingBalanceID, StagedQty: x.PickedQty, RemovedQty: "0", UOMID: x.UOMID, CreatedBy: actor})
		}
		return local.repositories.StagingLine.CreateBatch(ctx, lines)
	})
	if err != nil {
		return dto.StagingResponse{}, err
	}
	return s.GetStaging(ctx, id)
}

func (s *Service) GetStaging(ctx context.Context, id string) (dto.StagingResponse, error) {
	if !validID(id, 120) {
		return dto.StagingResponse{}, invalid("invalid staging_id")
	}
	row, err := s.repositories.Staging.Get(ctx, id)
	if err != nil {
		return dto.StagingResponse{}, err
	}
	lines, err := s.repositories.StagingLine.List(ctx, id)
	if err != nil {
		return dto.StagingResponse{}, err
	}
	return mapStaging(row, lines), nil
}
func (s *Service) ListStagings(ctx context.Context, f repository.ListFilter) (dto.PageResponse[dto.StagingResponse], error) {
	if !validUUID(f.OwnerID) {
		return dto.PageResponse[dto.StagingResponse]{}, invalid("owner_id is required and must be a UUID")
	}
	rows, total, err := s.repositories.Staging.List(ctx, f)
	if err != nil {
		return dto.PageResponse[dto.StagingResponse]{}, err
	}
	items := make([]dto.StagingResponse, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapStaging(v, nil))
	}
	return page(items, f.Page, f.PageSize, total), nil
}

func (s *Service) CompleteStaging(ctx context.Context, id, actor string) (dto.StagingResponse, error) {
	if !validID(id, 120) || !validUUID(actor) {
		return dto.StagingResponse{}, invalid("invalid staging completion request")
	}
	err := s.transaction(ctx, func(local *Service) error {
		staging, err := local.repositories.Staging.Get(ctx, id)
		if err != nil {
			return err
		}
		if staging.StatusCode != "OPEN" {
			return state("only an OPEN staging document can be completed")
		}
		lines, err := local.repositories.StagingLine.List(ctx, id)
		if err != nil {
			return err
		}
		if len(lines) == 0 {
			return state("staging document has no lines")
		}
		order, err := local.repositories.Order.Lock(ctx, staging.OutboundID)
		if err != nil {
			return err
		}
		orderRow, err := local.repositories.Order.Get(ctx, order.ID)
		if err != nil {
			return err
		}
		if orderRow.StatusCode != "PICKING" {
			return state("outbound order is not in PICKING")
		}
		completed, err := local.transition(ctx, staging.DocumentTypeID, staging.StatusID, "COMPLETED")
		if err != nil {
			return err
		}
		if err := local.repositories.Staging.Complete(ctx, id, completed.ID, actor); err != nil {
			return err
		}
		target, err := local.transition(ctx, order.DocumentTypeID, order.StatusID, "STAGED")
		if err != nil {
			return err
		}
		return local.repositories.Order.SetStatusLocked(ctx, order.ID, target.ID, actor, nil)
	})
	if err != nil {
		return dto.StagingResponse{}, err
	}
	return s.GetStaging(ctx, id)
}
