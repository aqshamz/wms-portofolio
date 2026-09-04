package master

import (
	"context"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
)

func (s *OperationalService) GetTaskStatusTransition(ctx context.Context, id string) (response dto.TaskStatusTransitionResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}

	value, err := s.repositories.TaskStatusTransition.Get(ctx, id)
	if err != nil {
		return response, catalogError(err)
	}

	return mapTaskStatusTransition(value), nil
}
func (s *OperationalService) ListTaskStatusTransition(ctx context.Context, request dto.OperationalListRequest) (dto.PageResponse[dto.TaskStatusTransitionResponse], error) {
	filter, err := operationalFilter(request, "")
	if err != nil {
		return dto.PageResponse[dto.TaskStatusTransitionResponse]{}, err
	}

	rows, total, err := s.repositories.TaskStatusTransition.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.TaskStatusTransitionResponse]{}, catalogError(err)
	}
	items := make([]dto.TaskStatusTransitionResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapTaskStatusTransition(row))
	}
	return pageResponse(items, request.Page, request.PageSize, total), nil
}

func (s *OperationalService) CreateTaskStatusTransition(ctx context.Context, request dto.CreateTaskStatusTransitionRequest) (response dto.TaskStatusTransitionResponse, err error) {

	err = s.transaction(ctx, func(local *OperationalService) error {
		if err := local.repositories.TaskStatus.LockConfiguration(ctx); err != nil {
			return err
		}

		value := model.TaskStatusTransition{
			FromStatusID:         request.FromStatusID,
			ToStatusID:           request.ToStatusID,
			RequiredPermissionID: request.RequiredPermissionID,
			IsActive:             true}
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		if err := local.repositories.TaskStatusTransition.Create(ctx, &value); err != nil {
			return err
		}
		response = mapTaskStatusTransition(value)
		return nil
	})
	return response, err
}
func (s *OperationalService) UpdateTaskStatusTransition(ctx context.Context, id string, request dto.UpdateTaskStatusTransitionRequest) (response dto.TaskStatusTransitionResponse, err error) {
	if validateID(id) != nil || request.IsActive == nil {
		return response, ErrInvalidInput
	}

	err = s.transaction(ctx, func(local *OperationalService) error {
		if err := local.repositories.TaskStatus.LockConfiguration(ctx); err != nil {
			return err
		}

		value, err := local.repositories.TaskStatusTransition.Lock(ctx, id)
		if err != nil {
			return err
		}

		value.RequiredPermissionID = request.RequiredPermissionID
		value.IsActive = *request.IsActive
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		updated, err := local.repositories.TaskStatusTransition.Update(ctx, id, map[string]interface{}{"required_permission_id": value.RequiredPermissionID, "is_active": value.IsActive}, "", nil)
		response = mapTaskStatusTransition(updated)
		return err
	})
	return response, err
}
func (s *OperationalService) DeactivateTaskStatusTransition(ctx context.Context, id string) (response dto.TaskStatusTransitionResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}

	err = s.transaction(ctx, func(local *OperationalService) error {
		if err := local.repositories.TaskStatus.LockConfiguration(ctx); err != nil {
			return err
		}

		value, err := local.repositories.TaskStatusTransition.Lock(ctx, id)
		if err != nil {
			return err
		}
		_ = value
		updated, err := local.repositories.TaskStatusTransition.Update(ctx, id, map[string]interface{}{"is_active": false}, "", nil)
		response = mapTaskStatusTransition(updated)
		return err
	})
	return response, err
}
