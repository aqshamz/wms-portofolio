package master

import (
	"context"
	"strings"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
)

func (s *OperationalService) GetPickingStrategyRule(ctx context.Context, parentID string, id string) (response dto.PickingStrategyRuleResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}
	if validateID(parentID) != nil {
		return response, ErrInvalidInput
	}
	value, err := s.repositories.PickingStrategyRule.Get(ctx, id)
	if err != nil {
		return response, catalogError(err)
	}
	if !strings.EqualFold(value.PickingStrategyID, parentID) {
		return response, ErrNotFound
	}
	return mapPickingStrategyRule(value), nil
}
func (s *OperationalService) ListPickingStrategyRule(ctx context.Context, parentID string, request dto.OperationalListRequest) (dto.PageResponse[dto.PickingStrategyRuleResponse], error) {
	filter, err := operationalFilter(request, parentID)
	if err != nil {
		return dto.PageResponse[dto.PickingStrategyRuleResponse]{}, err
	}
	if _, err := s.repositories.PickingStrategy.Get(ctx, parentID); err != nil {
		return dto.PageResponse[dto.PickingStrategyRuleResponse]{}, catalogError(err)
	}
	rows, total, err := s.repositories.PickingStrategyRule.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.PickingStrategyRuleResponse]{}, catalogError(err)
	}
	items := make([]dto.PickingStrategyRuleResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapPickingStrategyRule(row))
	}
	return pageResponse(items, request.Page, request.PageSize, total), nil
}

func (s *OperationalService) CreatePickingStrategyRule(ctx context.Context, parentID string, request dto.CreatePickingStrategyRuleRequest) (response dto.PickingStrategyRuleResponse, err error) {
	if validateID(parentID) != nil {
		return response, ErrInvalidInput
	}

	err = s.transaction(ctx, func(local *OperationalService) error {

		parent, err := local.repositories.PickingStrategy.Lock(ctx, parentID)
		if err != nil {
			return err
		}
		if !parent.IsActive {
			return invalidCatalog("parent configuration is inactive")
		}
		value := model.PickingStrategyRule{
			SequenceNo:          request.SequenceNo,
			InventoryStatusID:   request.InventoryStatusID,
			ZoneID:              request.ZoneID,
			PickingSortMethodID: request.PickingSortMethodID,
			PickingStrategyID:   parentID, IsActive: true}
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		if err := local.repositories.PickingStrategyRule.Create(ctx, &value); err != nil {
			return err
		}
		response = mapPickingStrategyRule(value)
		return nil
	})
	return response, err
}
func (s *OperationalService) UpdatePickingStrategyRule(ctx context.Context, parentID string, id string, request dto.UpdatePickingStrategyRuleRequest) (response dto.PickingStrategyRuleResponse, err error) {
	if validateID(id) != nil || request.IsActive == nil {
		return response, ErrInvalidInput
	}
	if validateID(parentID) != nil {
		return response, ErrInvalidInput
	}

	err = s.transaction(ctx, func(local *OperationalService) error {

		parent, err := local.repositories.PickingStrategy.Lock(ctx, parentID)
		if err != nil {
			return err
		}
		if !parent.IsActive {
			return invalidCatalog("parent configuration is inactive")
		}
		value, err := local.repositories.PickingStrategyRule.Lock(ctx, id)
		if err != nil {
			return err
		}
		if !strings.EqualFold(value.PickingStrategyID, parentID) {
			return ErrNotFound
		}
		value.SequenceNo = request.SequenceNo
		value.InventoryStatusID = request.InventoryStatusID
		value.ZoneID = request.ZoneID
		value.PickingSortMethodID = request.PickingSortMethodID
		value.IsActive = *request.IsActive
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		updated, err := local.repositories.PickingStrategyRule.Update(ctx, id, map[string]interface{}{"sequence_no": value.SequenceNo,
			"inventory_status_id":    value.InventoryStatusID,
			"zone_id":                value.ZoneID,
			"picking_sort_method_id": value.PickingSortMethodID, "is_active": value.IsActive}, "", nil)
		response = mapPickingStrategyRule(updated)
		return err
	})
	return response, err
}
func (s *OperationalService) DeactivatePickingStrategyRule(ctx context.Context, parentID string, id string) (response dto.PickingStrategyRuleResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}
	if validateID(parentID) != nil {
		return response, ErrInvalidInput
	}
	err = s.transaction(ctx, func(local *OperationalService) error {

		if _, err := local.repositories.PickingStrategy.Lock(ctx, parentID); err != nil {
			return err
		}
		value, err := local.repositories.PickingStrategyRule.Lock(ctx, id)
		if err != nil {
			return err
		}
		if !strings.EqualFold(value.PickingStrategyID, parentID) {
			return ErrNotFound
		}
		updated, err := local.repositories.PickingStrategyRule.Update(ctx, id, map[string]interface{}{"is_active": false}, "", nil)
		response = mapPickingStrategyRule(updated)
		return err
	})
	return response, err
}
