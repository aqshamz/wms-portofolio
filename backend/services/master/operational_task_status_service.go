package master

import (
	"context"
	"strings"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
)

func (s *OperationalService) GetTaskStatus(ctx context.Context, id string) (response dto.TaskStatusResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}

	value, err := s.repositories.TaskStatus.Get(ctx, id)
	if err != nil {
		return response, catalogError(err)
	}

	return mapTaskStatus(value), nil
}
func (s *OperationalService) ListTaskStatus(ctx context.Context, request dto.OperationalListRequest) (dto.PageResponse[dto.TaskStatusResponse], error) {
	filter, err := operationalFilter(request, "")
	if err != nil {
		return dto.PageResponse[dto.TaskStatusResponse]{}, err
	}

	rows, total, err := s.repositories.TaskStatus.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.TaskStatusResponse]{}, catalogError(err)
	}
	items := make([]dto.TaskStatusResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapTaskStatus(row))
	}
	return pageResponse(items, request.Page, request.PageSize, total), nil
}

func (s *OperationalService) CreateTaskStatus(ctx context.Context, request dto.CreateTaskStatusRequest) (response dto.TaskStatusResponse, err error) {

	code, name, err := catalogIdentity(request.Code, request.Name)
	if err != nil {
		return response, err
	}
	err = s.transaction(ctx, func(local *OperationalService) error {
		if err := local.repositories.TaskStatus.LockConfiguration(ctx); err != nil {
			return err
		}

		value := model.TaskStatus{
			Code:        code,
			Name:        name,
			IsInitial:   request.IsInitial,
			IsFinal:     request.IsFinal,
			IsCancelled: request.IsCancelled,
			IsActive:    true}
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		if err := local.repositories.TaskStatus.Create(ctx, &value); err != nil {
			return err
		}
		response = mapTaskStatus(value)
		return nil
	})
	return response, err
}
func (s *OperationalService) UpdateTaskStatus(ctx context.Context, id string, request dto.UpdateTaskStatusRequest) (response dto.TaskStatusResponse, err error) {
	if validateID(id) != nil || request.IsActive == nil {
		return response, ErrInvalidInput
	}

	if strings.TrimSpace(request.Name) == "" {
		return response, ErrInvalidInput
	}
	err = s.transaction(ctx, func(local *OperationalService) error {
		if err := local.repositories.TaskStatus.LockConfiguration(ctx); err != nil {
			return err
		}

		value, err := local.repositories.TaskStatus.Lock(ctx, id)
		if err != nil {
			return err
		}

		value.Name = strings.TrimSpace(request.Name)
		value.IsInitial = request.IsInitial
		value.IsFinal = request.IsFinal
		value.IsCancelled = request.IsCancelled
		value.IsActive = *request.IsActive
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		updated, err := local.repositories.TaskStatus.Update(ctx, id, map[string]interface{}{"name": value.Name,
			"is_initial":   value.IsInitial,
			"is_final":     value.IsFinal,
			"is_cancelled": value.IsCancelled, "is_active": value.IsActive}, "", nil)
		response = mapTaskStatus(updated)
		return err
	})
	return response, err
}
func (s *OperationalService) DeactivateTaskStatus(ctx context.Context, id string) (response dto.TaskStatusResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}

	err = s.transaction(ctx, func(local *OperationalService) error {
		if err := local.repositories.TaskStatus.LockConfiguration(ctx); err != nil {
			return err
		}

		value, err := local.repositories.TaskStatus.Lock(ctx, id)
		if err != nil {
			return err
		}
		_ = value
		updated, err := local.repositories.TaskStatus.Update(ctx, id, map[string]interface{}{"is_active": false, "is_initial": false}, "", nil)
		response = mapTaskStatus(updated)
		return err
	})
	return response, err
}
