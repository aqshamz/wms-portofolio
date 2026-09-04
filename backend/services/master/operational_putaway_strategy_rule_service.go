package master

import (
	"context"
	"strings"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
)

func (s *OperationalService) GetPutawayStrategyRule(ctx context.Context, parentID string, id string) (response dto.PutawayStrategyRuleResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}
	if validateID(parentID) != nil {
		return response, ErrInvalidInput
	}
	value, err := s.repositories.PutawayStrategyRule.Get(ctx, id)
	if err != nil {
		return response, catalogError(err)
	}
	if !strings.EqualFold(value.PutawayStrategyID, parentID) {
		return response, ErrNotFound
	}
	return mapPutawayStrategyRule(value), nil
}
func (s *OperationalService) ListPutawayStrategyRule(ctx context.Context, parentID string, request dto.OperationalListRequest) (dto.PageResponse[dto.PutawayStrategyRuleResponse], error) {
	filter, err := operationalFilter(request, parentID)
	if err != nil {
		return dto.PageResponse[dto.PutawayStrategyRuleResponse]{}, err
	}
	if _, err := s.repositories.PutawayStrategy.Get(ctx, parentID); err != nil {
		return dto.PageResponse[dto.PutawayStrategyRuleResponse]{}, catalogError(err)
	}
	rows, total, err := s.repositories.PutawayStrategyRule.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.PutawayStrategyRuleResponse]{}, catalogError(err)
	}
	items := make([]dto.PutawayStrategyRuleResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapPutawayStrategyRule(row))
	}
	return pageResponse(items, request.Page, request.PageSize, total), nil
}

func (s *OperationalService) CreatePutawayStrategyRule(ctx context.Context, parentID string, request dto.CreatePutawayStrategyRuleRequest) (response dto.PutawayStrategyRuleResponse, err error) {
	if validateID(parentID) != nil {
		return response, ErrInvalidInput
	}

	err = s.transaction(ctx, func(local *OperationalService) error {

		parent, err := local.repositories.PutawayStrategy.Lock(ctx, parentID)
		if err != nil {
			return err
		}
		if !parent.IsActive {
			return invalidCatalog("parent configuration is inactive")
		}
		value := model.PutawayStrategyRule{
			SequenceNo:          request.SequenceNo,
			CategoryID:          request.CategoryID,
			LocationTypeID:      request.LocationTypeID,
			ZoneID:              request.ZoneID,
			MinimumEmptyPercent: request.MinimumEmptyPercent,
			PutawayStrategyID:   parentID, IsActive: true}
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		if err := local.repositories.PutawayStrategyRule.Create(ctx, &value); err != nil {
			return err
		}
		response = mapPutawayStrategyRule(value)
		return nil
	})
	return response, err
}
func (s *OperationalService) UpdatePutawayStrategyRule(ctx context.Context, parentID string, id string, request dto.UpdatePutawayStrategyRuleRequest) (response dto.PutawayStrategyRuleResponse, err error) {
	if validateID(id) != nil || request.IsActive == nil {
		return response, ErrInvalidInput
	}
	if validateID(parentID) != nil {
		return response, ErrInvalidInput
	}

	err = s.transaction(ctx, func(local *OperationalService) error {

		parent, err := local.repositories.PutawayStrategy.Lock(ctx, parentID)
		if err != nil {
			return err
		}
		if !parent.IsActive {
			return invalidCatalog("parent configuration is inactive")
		}
		value, err := local.repositories.PutawayStrategyRule.Lock(ctx, id)
		if err != nil {
			return err
		}
		if !strings.EqualFold(value.PutawayStrategyID, parentID) {
			return ErrNotFound
		}
		value.SequenceNo = request.SequenceNo
		value.CategoryID = request.CategoryID
		value.LocationTypeID = request.LocationTypeID
		value.ZoneID = request.ZoneID
		value.MinimumEmptyPercent = request.MinimumEmptyPercent
		value.IsActive = *request.IsActive
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		updated, err := local.repositories.PutawayStrategyRule.Update(ctx, id, map[string]interface{}{"sequence_no": value.SequenceNo,
			"category_id":           value.CategoryID,
			"location_type_id":      value.LocationTypeID,
			"zone_id":               value.ZoneID,
			"minimum_empty_percent": value.MinimumEmptyPercent, "is_active": value.IsActive}, "", nil)
		response = mapPutawayStrategyRule(updated)
		return err
	})
	return response, err
}
func (s *OperationalService) DeactivatePutawayStrategyRule(ctx context.Context, parentID string, id string) (response dto.PutawayStrategyRuleResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}
	if validateID(parentID) != nil {
		return response, ErrInvalidInput
	}
	err = s.transaction(ctx, func(local *OperationalService) error {

		if _, err := local.repositories.PutawayStrategy.Lock(ctx, parentID); err != nil {
			return err
		}
		value, err := local.repositories.PutawayStrategyRule.Lock(ctx, id)
		if err != nil {
			return err
		}
		if !strings.EqualFold(value.PutawayStrategyID, parentID) {
			return ErrNotFound
		}
		updated, err := local.repositories.PutawayStrategyRule.Update(ctx, id, map[string]interface{}{"is_active": false}, "", nil)
		response = mapPutawayStrategyRule(updated)
		return err
	})
	return response, err
}
