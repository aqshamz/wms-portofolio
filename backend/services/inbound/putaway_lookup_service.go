package inbound

import (
	"context"
	"strings"
	dto "wms-api/dto/inbound"
	repository "wms-api/repository/inbound"
)

func validatePutawayLookup(id string, search *string, page, size int) error {
	if !inboundID(id, 120) || page < 1 || page > 1000000 || size < 1 || size > 100 {
		return invalid("invalid putaway lookup or pagination")
	}
	*search = strings.TrimSpace(*search)
	if len(*search) > 160 {
		return invalid("search is too long")
	}
	return nil
}

func (s *Service) ListPutawayAssignees(ctx context.Context, id, search string, number, size int) (dto.PageResponse[dto.PutawayAssigneeResponse], error) {
	if err := validatePutawayLookup(id, &search, number, size); err != nil {
		return dto.PageResponse[dto.PutawayAssigneeResponse]{}, err
	}
	task, err := s.repositories.PutawayTask.Get(ctx, id)
	if err != nil {
		return dto.PageResponse[dto.PutawayAssigneeResponse]{}, err
	}
	rows, total, err := s.repositories.PutawayTask.ListAssignees(ctx, task.OwnerID, task.WarehouseID, search, number, size)
	items := make([]dto.PutawayAssigneeResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.PutawayAssigneeResponse{AccountID: row.AccountID, Username: row.Username, DisplayName: row.DisplayName})
	}
	return page(items, number, size, total), err
}

func (s *Service) ListPutawayTargets(ctx context.Context, id, search string, number, size int) (dto.PageResponse[dto.PutawayTargetResponse], error) {
	if err := validatePutawayLookup(id, &search, number, size); err != nil {
		return dto.PageResponse[dto.PutawayTargetResponse]{}, err
	}
	task, err := s.repositories.PutawayTask.Get(ctx, id)
	if err != nil {
		return dto.PageResponse[dto.PutawayTargetResponse]{}, err
	}
	item, err := s.repositories.Master.Catalog.Item.Get(ctx, task.ItemID)
	if err != nil || !item.IsActive {
		return dto.PageResponse[dto.PutawayTargetResponse]{}, invalid("putaway item is unavailable")
	}
	rules, err := s.applicablePutawayRules(ctx, task.OwnerID, task.WarehouseID)
	if err != nil {
		return dto.PageResponse[dto.PutawayTargetResponse]{}, err
	}
	rows, total, err := s.repositories.PutawayTask.ListTargets(ctx, task.WarehouseID, item, rules, search, number, size)
	items := make([]dto.PutawayTargetResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.PutawayTargetResponse{LocationID: row.LocationID, Code: row.Code, ZoneCode: row.ZoneCode, LocationTypeCode: row.LocationTypeCode})
	}
	return page(items, number, size, total), repository.Error(err)
}

// ListQualityInspectionTargets uses the same strategy selection and target
// query as completion validation, so the QC picker cannot offer a location
// that CompleteQualityInspection will reject.
func (s *Service) ListQualityInspectionTargets(ctx context.Context, id, search string, number, size int) (dto.PageResponse[dto.PutawayTargetResponse], error) {
	if err := validatePutawayLookup(id, &search, number, size); err != nil {
		return dto.PageResponse[dto.PutawayTargetResponse]{}, err
	}
	inspection, err := s.repositories.QualityInspection.Get(ctx, id)
	if err != nil {
		return dto.PageResponse[dto.PutawayTargetResponse]{}, err
	}
	if inspection.InspectedAt != nil || inspection.CancelledAt != nil {
		return dto.PageResponse[dto.PutawayTargetResponse]{}, state("only a pending quality inspection has putaway targets")
	}
	item, err := s.repositories.Master.Catalog.Item.Get(ctx, inspection.ItemID)
	if err != nil || !item.IsActive {
		return dto.PageResponse[dto.PutawayTargetResponse]{}, invalid("inspection item is unavailable")
	}
	rules, err := s.applicablePutawayRules(ctx, inspection.OwnerID, inspection.WarehouseID)
	if err != nil {
		return dto.PageResponse[dto.PutawayTargetResponse]{}, err
	}
	rows, total, err := s.repositories.PutawayTask.ListTargets(ctx, inspection.WarehouseID, item, rules, search, number, size)
	items := make([]dto.PutawayTargetResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.PutawayTargetResponse{LocationID: row.LocationID, Code: row.Code, ZoneCode: row.ZoneCode, LocationTypeCode: row.LocationTypeCode})
	}
	return page(items, number, size, total), repository.Error(err)
}
