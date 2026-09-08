package inbound

import (
	"context"
	"errors"
	"strings"

	dto "wms-api/dto/inbound"
	repository "wms-api/repository/inbound"
)

func validateList(filter *repository.ListFilter) error {
	if !inboundUUID(filter.OwnerID) {
		return invalid("owner_id is required and must be a UUID")
	}
	if filter.WarehouseID != "" && !inboundUUID(filter.WarehouseID) {
		return invalid("warehouse_id must be a UUID")
	}
	if filter.Page < 1 || filter.Page > 1000000 || filter.PageSize < 1 || filter.PageSize > 100 {
		return invalid("page must be 1..1000000 and page_size 1..100")
	}
	filter.OwnerID = strings.ToLower(filter.OwnerID)
	filter.WarehouseID = strings.ToLower(filter.WarehouseID)
	filter.StatusCode = strings.ToUpper(strings.TrimSpace(filter.StatusCode))
	filter.Search = strings.TrimSpace(filter.Search)
	if len(filter.Search) > 160 || len(filter.StatusCode) > 40 {
		return invalid("search or status_code is too long")
	}
	return nil
}

func (s *Service) GetQualityInspection(ctx context.Context, id string) (dto.QualityInspectionResponse, error) {
	if !inboundID(id, 120) {
		return dto.QualityInspectionResponse{}, invalid("invalid inspection_id")
	}
	row, err := s.repositories.QualityInspection.Get(ctx, id)
	if err != nil {
		return dto.QualityInspectionResponse{}, err
	}
	result := mapQualityInspection(row)
	if child, childErr := s.repositories.QualityInspection.GetChild(ctx, id); childErr == nil {
		result.ReplacementInspectionID = &child.ID
	} else if !errors.Is(childErr, repository.ErrNotFound) {
		return result, childErr
	}
	if task, taskErr := s.repositories.PutawayTask.GetByInspection(ctx, id); taskErr == nil {
		mapped := mapPutawayTask(task)
		result.PutawayTask = &mapped
	} else if !errors.Is(taskErr, repository.ErrNotFound) {
		return result, taskErr
	}
	if quarantine, caseErr := s.repositories.QuarantineCase.GetByInspection(ctx, id); caseErr == nil {
		mapped := mapQuarantineCase(quarantine)
		result.QuarantineCase = &mapped
	} else if !errors.Is(caseErr, repository.ErrNotFound) {
		return result, caseErr
	}
	return result, nil
}

func (s *Service) ListQualityInspections(ctx context.Context, filter repository.ListFilter) (dto.PageResponse[dto.QualityInspectionResponse], error) {
	if err := validateList(&filter); err != nil {
		return dto.PageResponse[dto.QualityInspectionResponse]{}, err
	}
	rows, total, err := s.repositories.QualityInspection.List(ctx, filter)
	items := make([]dto.QualityInspectionResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapQualityInspection(row))
	}
	return page(items, filter.Page, filter.PageSize, total), err
}

func (s *Service) GetPutawayTask(ctx context.Context, id string) (dto.PutawayTaskResponse, error) {
	if !inboundID(id, 120) {
		return dto.PutawayTaskResponse{}, invalid("invalid putaway_task_id")
	}
	row, err := s.repositories.PutawayTask.Get(ctx, id)
	return mapPutawayTask(row), err
}

func (s *Service) ListPutawayTasks(ctx context.Context, filter repository.ListFilter) (dto.PageResponse[dto.PutawayTaskResponse], error) {
	if err := validateList(&filter); err != nil {
		return dto.PageResponse[dto.PutawayTaskResponse]{}, err
	}
	rows, total, err := s.repositories.PutawayTask.List(ctx, filter)
	items := make([]dto.PutawayTaskResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapPutawayTask(row))
	}
	return page(items, filter.Page, filter.PageSize, total), err
}

func (s *Service) GetQuarantineCase(ctx context.Context, id string) (dto.QuarantineCaseResponse, error) {
	if !inboundID(id, 140) {
		return dto.QuarantineCaseResponse{}, invalid("invalid quarantine_case_id")
	}
	row, err := s.repositories.QuarantineCase.Get(ctx, id)
	if err != nil {
		return dto.QuarantineCaseResponse{}, err
	}
	result := mapQuarantineCase(row)
	dispositions, err := s.repositories.Disposition.ListByCase(ctx, id)
	if err != nil {
		return result, err
	}
	for _, disposition := range dispositions {
		mapped := mapDisposition(disposition)
		if task, taskErr := s.repositories.ReworkTask.GetByDisposition(ctx, disposition.ID); taskErr == nil {
			mappedTask := mapReworkTask(task)
			mapped.ReworkTask = &mappedTask
		} else if !errors.Is(taskErr, repository.ErrNotFound) {
			return result, taskErr
		}
		result.Dispositions = append(result.Dispositions, mapped)
	}
	return result, nil
}

func (s *Service) GetReworkTask(ctx context.Context, id string) (dto.ReworkTaskResponse, error) {
	if !inboundID(id, 140) {
		return dto.ReworkTaskResponse{}, invalid("invalid rework_task_id")
	}
	row, err := s.repositories.ReworkTask.Get(ctx, id)
	return mapReworkTask(row), err
}

func (s *Service) ListReworkTasks(ctx context.Context, filter repository.ListFilter) (dto.PageResponse[dto.ReworkTaskResponse], error) {
	if err := validateList(&filter); err != nil {
		return dto.PageResponse[dto.ReworkTaskResponse]{}, err
	}
	rows, total, err := s.repositories.ReworkTask.List(ctx, filter)
	items := make([]dto.ReworkTaskResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapReworkTask(row))
	}
	return page(items, filter.Page, filter.PageSize, total), err
}

func (s *Service) ListQuarantineCases(ctx context.Context, filter repository.ListFilter) (dto.PageResponse[dto.QuarantineCaseResponse], error) {
	if err := validateList(&filter); err != nil {
		return dto.PageResponse[dto.QuarantineCaseResponse]{}, err
	}
	rows, total, err := s.repositories.QuarantineCase.List(ctx, filter)
	items := make([]dto.QuarantineCaseResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapQuarantineCase(row))
	}
	return page(items, filter.Page, filter.PageSize, total), err
}
func (s *Service) GetPurchaseOrder(ctx context.Context, id string) (dto.PurchaseOrderResponse, error) {
	if !inboundID(id, 120) {
		return dto.PurchaseOrderResponse{}, invalid("invalid purchase_order_id")
	}
	row, err := s.repositories.PurchaseOrder.Get(ctx, id)
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	result := mapPurchaseOrder(row)
	lines, err := s.repositories.PurchaseOrderLine.List(ctx, id)
	if err != nil {
		return result, err
	}
	result.Lines = make([]dto.PurchaseOrderLineResponse, 0, len(lines))
	for _, line := range lines {
		result.Lines = append(result.Lines, mapPurchaseOrderLine(line))
	}
	return result, nil
}
func (s *Service) ListPurchaseOrders(ctx context.Context, filter repository.ListFilter) (dto.PageResponse[dto.PurchaseOrderResponse], error) {
	if err := validateList(&filter); err != nil {
		return dto.PageResponse[dto.PurchaseOrderResponse]{}, err
	}
	rows, total, err := s.repositories.PurchaseOrder.List(ctx, filter)
	items := make([]dto.PurchaseOrderResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapPurchaseOrder(row))
	}
	return page(items, filter.Page, filter.PageSize, total), err
}
func (s *Service) GetInboundOrder(ctx context.Context, id string) (dto.InboundOrderResponse, error) {
	if !inboundID(id, 120) {
		return dto.InboundOrderResponse{}, invalid("invalid inbound_id")
	}
	row, err := s.repositories.InboundOrder.Get(ctx, id)
	if err != nil {
		return dto.InboundOrderResponse{}, err
	}
	result := mapInboundOrder(row)
	lines, err := s.repositories.InboundOrderLine.List(ctx, id)
	if err != nil {
		return result, err
	}
	result.Lines = make([]dto.InboundOrderLineResponse, 0, len(lines))
	for _, line := range lines {
		result.Lines = append(result.Lines, mapInboundOrderLine(line))
	}
	return result, nil
}
func (s *Service) ListInboundOrders(ctx context.Context, filter repository.ListFilter) (dto.PageResponse[dto.InboundOrderResponse], error) {
	if err := validateList(&filter); err != nil {
		return dto.PageResponse[dto.InboundOrderResponse]{}, err
	}
	rows, total, err := s.repositories.InboundOrder.List(ctx, filter)
	items := make([]dto.InboundOrderResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapInboundOrder(row))
	}
	return page(items, filter.Page, filter.PageSize, total), err
}
func (s *Service) GetReceipt(ctx context.Context, id string) (dto.ReceiptResponse, error) {
	if !inboundID(id, 120) {
		return dto.ReceiptResponse{}, invalid("invalid receipt_id")
	}
	row, err := s.repositories.Receipt.Get(ctx, id)
	if err != nil {
		return dto.ReceiptResponse{}, err
	}
	result := mapReceipt(row)
	lines, err := s.repositories.ReceiptLine.List(ctx, id)
	if err != nil {
		return result, err
	}
	result.Lines = make([]dto.ReceiptLineResponse, 0, len(lines))
	for _, line := range lines {
		mapped := mapReceiptLine(line)
		batches, batchErr := s.repositories.ReceiptInventory.ListByLine(ctx, line.ID)
		if batchErr != nil {
			return result, batchErr
		}
		for _, batch := range batches {
			mapped.Batches = append(mapped.Batches, mapReceiptBatch(batch))
		}
		result.Lines = append(result.Lines, mapped)
	}
	return result, nil
}
func (s *Service) ListReceipts(ctx context.Context, filter repository.ListFilter) (dto.PageResponse[dto.ReceiptResponse], error) {
	if err := validateList(&filter); err != nil {
		return dto.PageResponse[dto.ReceiptResponse]{}, err
	}
	rows, total, err := s.repositories.Receipt.List(ctx, filter)
	items := make([]dto.ReceiptResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapReceipt(row))
	}
	return page(items, filter.Page, filter.PageSize, total), err
}
