package master

import (
	"context"
	"strings"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
)

func (s *OperationalService) GetPickingStrategy(ctx context.Context, id string) (response dto.PickingStrategyResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}

	value, err := s.repositories.PickingStrategy.Get(ctx, id)
	if err != nil {
		return response, catalogError(err)
	}

	return mapPickingStrategy(value), nil
}
func (s *OperationalService) ListPickingStrategy(ctx context.Context, request dto.OperationalListRequest) (dto.PageResponse[dto.PickingStrategyResponse], error) {
	filter, err := operationalFilter(request, "")
	if err != nil {
		return dto.PageResponse[dto.PickingStrategyResponse]{}, err
	}

	rows, total, err := s.repositories.PickingStrategy.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.PickingStrategyResponse]{}, catalogError(err)
	}
	items := make([]dto.PickingStrategyResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapPickingStrategy(row))
	}
	return pageResponse(items, request.Page, request.PageSize, total), nil
}

func (s *OperationalService) CreatePickingStrategy(ctx context.Context, request dto.CreatePickingStrategyRequest) (response dto.PickingStrategyResponse, err error) {

	code, name, err := catalogIdentity(request.Code, request.Name)
	if err != nil {
		return response, err
	}
	err = s.transaction(ctx, func(local *OperationalService) error {

		value := model.PickingStrategy{
			OwnerID:     request.OwnerID,
			WarehouseID: request.WarehouseID,
			Code:        code,
			Name:        name,
			Description: request.Description,
			IsActive:    true}
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		if err := local.repositories.PickingStrategy.Create(ctx, &value); err != nil {
			return err
		}
		response = mapPickingStrategy(value)
		return nil
	})
	return response, err
}
func (s *OperationalService) UpdatePickingStrategy(ctx context.Context, id string, request dto.UpdatePickingStrategyRequest) (response dto.PickingStrategyResponse, err error) {
	if validateID(id) != nil || request.IsActive == nil {
		return response, ErrInvalidInput
	}

	if strings.TrimSpace(request.Name) == "" {
		return response, ErrInvalidInput
	}
	err = s.transaction(ctx, func(local *OperationalService) error {

		value, err := local.repositories.PickingStrategy.Lock(ctx, id)
		if err != nil {
			return err
		}

		value.Name = strings.TrimSpace(request.Name)
		value.Description = request.Description
		value.IsActive = *request.IsActive
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		updated, err := local.repositories.PickingStrategy.Update(ctx, id, map[string]interface{}{"name": value.Name,
			"description": value.Description, "is_active": value.IsActive}, "", nil)
		response = mapPickingStrategy(updated)
		return err
	})
	return response, err
}
func (s *OperationalService) DeactivatePickingStrategy(ctx context.Context, id string) (response dto.PickingStrategyResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}

	err = s.transaction(ctx, func(local *OperationalService) error {

		value, err := local.repositories.PickingStrategy.Lock(ctx, id)
		if err != nil {
			return err
		}
		_ = value
		updated, err := local.repositories.PickingStrategy.Update(ctx, id, map[string]interface{}{"is_active": false}, "", nil)
		response = mapPickingStrategy(updated)
		return err
	})
	return response, err
}
