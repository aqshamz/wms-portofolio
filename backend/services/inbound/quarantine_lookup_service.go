package inbound

import (
	"context"
	"strings"
	dto "wms-api/dto/inbound"
)

// ListQuarantineTargets shares the acceptance validator's strategy selection
// and the storage picker query, filtering before count and pagination.
func (s *Service) ListQuarantineTargets(ctx context.Context, id, search string, number, size int) (dto.PageResponse[dto.PutawayTargetResponse], error) {
	search = strings.TrimSpace(search)
	if !inboundID(id, 140) || len(search) > 160 || number < 1 || number > 1000000 || size < 1 || size > 100 {
		return dto.PageResponse[dto.PutawayTargetResponse]{}, invalid("invalid quarantine target lookup or pagination")
	}
	quarantine, err := s.repositories.QuarantineCase.Get(ctx, id)
	if err != nil {
		return dto.PageResponse[dto.PutawayTargetResponse]{}, err
	}
	item, err := s.repositories.Master.Catalog.Item.Get(ctx, quarantine.ItemID)
	if err != nil || !item.IsActive {
		return dto.PageResponse[dto.PutawayTargetResponse]{}, invalid("quarantine item is unavailable")
	}
	rules, err := s.applicablePutawayRules(ctx, quarantine.OwnerID, quarantine.WarehouseID)
	if err != nil {
		return dto.PageResponse[dto.PutawayTargetResponse]{}, err
	}
	rows, total, err := s.repositories.PutawayTask.ListTargets(ctx, quarantine.WarehouseID, item, rules, search, number, size)
	items := make([]dto.PutawayTargetResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.PutawayTargetResponse{LocationID: row.LocationID, Code: row.Code, ZoneCode: row.ZoneCode, LocationTypeCode: row.LocationTypeCode})
	}
	return page(items, number, size, total), err
}
