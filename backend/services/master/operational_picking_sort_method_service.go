package master

import (
	"context"
	"strings"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
)

func (s *OperationalService) GetPickingSortMethod(ctx context.Context, id string) (response dto.PickingSortMethodResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}

	value, err := s.repositories.PickingSortMethod.Get(ctx, id)
	if err != nil {
		return response, catalogError(err)
	}

	return mapPickingSortMethod(value), nil
}
func (s *OperationalService) ListPickingSortMethod(ctx context.Context, request dto.OperationalListRequest) (dto.PageResponse[dto.PickingSortMethodResponse], error) {
	filter, err := operationalFilter(request, "")
	if err != nil {
		return dto.PageResponse[dto.PickingSortMethodResponse]{}, err
	}

	rows, total, err := s.repositories.PickingSortMethod.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.PickingSortMethodResponse]{}, catalogError(err)
	}
	items := make([]dto.PickingSortMethodResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapPickingSortMethod(row))
	}
	return pageResponse(items, request.Page, request.PageSize, total), nil
}

func (s *OperationalService) CreatePickingSortMethod(ctx context.Context, request dto.CreatePickingSortMethodRequest) (response dto.PickingSortMethodResponse, err error) {

	code, name, err := catalogIdentity(request.Code, request.Name)
	if err != nil {
		return response, err
	}
	err = s.transaction(ctx, func(local *OperationalService) error {

		value := model.PickingSortMethod{
			Code:        code,
			Name:        name,
			Description: request.Description,
			IsActive:    true}
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		if err := local.repositories.PickingSortMethod.Create(ctx, &value); err != nil {
			return err
		}
		response = mapPickingSortMethod(value)
		return nil
	})
	return response, err
}
func (s *OperationalService) UpdatePickingSortMethod(ctx context.Context, id string, request dto.UpdatePickingSortMethodRequest) (response dto.PickingSortMethodResponse, err error) {
	if validateID(id) != nil || request.IsActive == nil {
		return response, ErrInvalidInput
	}

	if strings.TrimSpace(request.Name) == "" {
		return response, ErrInvalidInput
	}
	err = s.transaction(ctx, func(local *OperationalService) error {

		value, err := local.repositories.PickingSortMethod.Lock(ctx, id)
		if err != nil {
			return err
		}

		value.Name = strings.TrimSpace(request.Name)
		value.Description = request.Description
		value.IsActive = *request.IsActive
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		updated, err := local.repositories.PickingSortMethod.Update(ctx, id, map[string]interface{}{"name": value.Name,
			"description": value.Description, "is_active": value.IsActive}, "", nil)
		response = mapPickingSortMethod(updated)
		return err
	})
	return response, err
}
func (s *OperationalService) DeactivatePickingSortMethod(ctx context.Context, id string) (response dto.PickingSortMethodResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}

	err = s.transaction(ctx, func(local *OperationalService) error {

		value, err := local.repositories.PickingSortMethod.Lock(ctx, id)
		if err != nil {
			return err
		}
		_ = value
		updated, err := local.repositories.PickingSortMethod.Update(ctx, id, map[string]interface{}{"is_active": false}, "", nil)
		response = mapPickingSortMethod(updated)
		return err
	})
	return response, err
}
