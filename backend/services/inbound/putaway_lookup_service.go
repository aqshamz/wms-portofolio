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
