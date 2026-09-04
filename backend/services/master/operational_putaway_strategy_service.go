package master

import (
	"context"
	"strings"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
)

func (s *OperationalService) GetPutawayStrategy(ctx context.Context, id string) (response dto.PutawayStrategyResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}

	value, err := s.repositories.PutawayStrategy.Get(ctx, id)
	if err != nil {
		return response, catalogError(err)
	}

	return mapPutawayStrategy(value), nil
}
func (s *OperationalService) ListPutawayStrategy(ctx context.Context, request dto.OperationalListRequest) (dto.PageResponse[dto.PutawayStrategyResponse], error) {
	filter, err := operationalFilter(request, "")
	if err != nil {
		return dto.PageResponse[dto.PutawayStrategyResponse]{}, err
	}

	rows, total, err := s.repositories.PutawayStrategy.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.PutawayStrategyResponse]{}, catalogError(err)
	}
	items := make([]dto.PutawayStrategyResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapPutawayStrategy(row))
	}
	return pageResponse(items, request.Page, request.PageSize, total), nil
}

func (s *OperationalService) CreatePutawayStrategy(ctx context.Context, request dto.CreatePutawayStrategyRequest) (response dto.PutawayStrategyResponse, err error) {

	code, name, err := catalogIdentity(request.Code, request.Name)
	if err != nil {
		return response, err
	}
	err = s.transaction(ctx, func(local *OperationalService) error {

		value := model.PutawayStrategy{
			OwnerID:     request.OwnerID,
			WarehouseID: request.WarehouseID,
			Code:        code,
			Name:        name,
			Description: request.Description,
			IsActive:    true}
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		if err := local.repositories.PutawayStrategy.Create(ctx, &value); err != nil {
			return err
		}
		response = mapPutawayStrategy(value)
		return nil
	})
	return response, err
}
func (s *OperationalService) UpdatePutawayStrategy(ctx context.Context, id string, request dto.UpdatePutawayStrategyRequest) (response dto.PutawayStrategyResponse, err error) {
	if validateID(id) != nil || request.IsActive == nil {
		return response, ErrInvalidInput
	}

	if strings.TrimSpace(request.Name) == "" {
		return response, ErrInvalidInput
	}
	err = s.transaction(ctx, func(local *OperationalService) error {

		value, err := local.repositories.PutawayStrategy.Lock(ctx, id)
		if err != nil {
			return err
		}

		value.Name = strings.TrimSpace(request.Name)
		value.Description = request.Description
		value.IsActive = *request.IsActive
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		updated, err := local.repositories.PutawayStrategy.Update(ctx, id, map[string]interface{}{"name": value.Name,
			"description": value.Description, "is_active": value.IsActive}, "", nil)
		response = mapPutawayStrategy(updated)
		return err
	})
	return response, err
}
func (s *OperationalService) DeactivatePutawayStrategy(ctx context.Context, id string) (response dto.PutawayStrategyResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}

	err = s.transaction(ctx, func(local *OperationalService) error {

		value, err := local.repositories.PutawayStrategy.Lock(ctx, id)
		if err != nil {
			return err
		}
		_ = value
		updated, err := local.repositories.PutawayStrategy.Update(ctx, id, map[string]interface{}{"is_active": false}, "", nil)
		response = mapPutawayStrategy(updated)
		return err
	})
	return response, err
}
