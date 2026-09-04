package master

import (
	"context"
	"strings"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
)

func (s *OperationalService) GetTaskPriority(ctx context.Context, id string) (response dto.TaskPriorityResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}

	value, err := s.repositories.TaskPriority.Get(ctx, id)
	if err != nil {
		return response, catalogError(err)
	}

	return mapTaskPriority(value), nil
}
func (s *OperationalService) ListTaskPriority(ctx context.Context, request dto.OperationalListRequest) (dto.PageResponse[dto.TaskPriorityResponse], error) {
	filter, err := operationalFilter(request, "")
	if err != nil {
		return dto.PageResponse[dto.TaskPriorityResponse]{}, err
	}

	rows, total, err := s.repositories.TaskPriority.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.TaskPriorityResponse]{}, catalogError(err)
	}
	items := make([]dto.TaskPriorityResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapTaskPriority(row))
	}
	return pageResponse(items, request.Page, request.PageSize, total), nil
}

func (s *OperationalService) CreateTaskPriority(ctx context.Context, request dto.CreateTaskPriorityRequest) (response dto.TaskPriorityResponse, err error) {

	code, name, err := catalogIdentity(request.Code, request.Name)
	if err != nil {
		return response, err
	}
	err = s.transaction(ctx, func(local *OperationalService) error {

		value := model.TaskPriority{
			Code:          code,
			Name:          name,
			PriorityValue: request.PriorityValue,
			IsActive:      true}
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		if err := local.repositories.TaskPriority.Create(ctx, &value); err != nil {
			return err
		}
		response = mapTaskPriority(value)
		return nil
	})
	return response, err
}
func (s *OperationalService) UpdateTaskPriority(ctx context.Context, id string, request dto.UpdateTaskPriorityRequest) (response dto.TaskPriorityResponse, err error) {
	if validateID(id) != nil || request.IsActive == nil {
		return response, ErrInvalidInput
	}

	if strings.TrimSpace(request.Name) == "" {
		return response, ErrInvalidInput
	}
	err = s.transaction(ctx, func(local *OperationalService) error {

		value, err := local.repositories.TaskPriority.Lock(ctx, id)
		if err != nil {
			return err
		}

		value.Name = strings.TrimSpace(request.Name)
		value.PriorityValue = request.PriorityValue
		value.IsActive = *request.IsActive
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		updated, err := local.repositories.TaskPriority.Update(ctx, id, map[string]interface{}{"name": value.Name,
			"priority_value": value.PriorityValue, "is_active": value.IsActive}, "", nil)
		response = mapTaskPriority(updated)
		return err
	})
	return response, err
}
func (s *OperationalService) DeactivateTaskPriority(ctx context.Context, id string) (response dto.TaskPriorityResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}

	err = s.transaction(ctx, func(local *OperationalService) error {

		value, err := local.repositories.TaskPriority.Lock(ctx, id)
		if err != nil {
			return err
		}
		_ = value
		updated, err := local.repositories.TaskPriority.Update(ctx, id, map[string]interface{}{"is_active": false}, "", nil)
		response = mapTaskPriority(updated)
		return err
	})
	return response, err
}
