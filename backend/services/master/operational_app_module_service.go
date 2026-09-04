package master

import (
	"context"
	"strings"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
)

func (s *OperationalService) GetAppModule(ctx context.Context, id string) (response dto.AppModuleResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}

	value, err := s.repositories.AppModule.Get(ctx, id)
	if err != nil {
		return response, catalogError(err)
	}

	return mapAppModule(value), nil
}
func (s *OperationalService) ListAppModule(ctx context.Context, request dto.OperationalListRequest) (dto.PageResponse[dto.AppModuleResponse], error) {
	filter, err := operationalFilter(request, "")
	if err != nil {
		return dto.PageResponse[dto.AppModuleResponse]{}, err
	}

	rows, total, err := s.repositories.AppModule.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.AppModuleResponse]{}, catalogError(err)
	}
	items := make([]dto.AppModuleResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapAppModule(row))
	}
	return pageResponse(items, request.Page, request.PageSize, total), nil
}

func (s *OperationalService) CreateAppModule(ctx context.Context, request dto.CreateAppModuleRequest) (response dto.AppModuleResponse, err error) {

	code, name, err := catalogIdentity(request.Code, request.Name)
	if err != nil {
		return response, err
	}
	err = s.transaction(ctx, func(local *OperationalService) error {

		value := model.AppModule{
			Code:         code,
			Name:         name,
			DisplayOrder: request.DisplayOrder,
			IsActive:     true}
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		if err := local.repositories.AppModule.Create(ctx, &value); err != nil {
			return err
		}
		response = mapAppModule(value)
		return nil
	})
	return response, err
}
func (s *OperationalService) UpdateAppModule(ctx context.Context, id string, request dto.UpdateAppModuleRequest) (response dto.AppModuleResponse, err error) {
	if validateID(id) != nil || request.IsActive == nil {
		return response, ErrInvalidInput
	}

	if strings.TrimSpace(request.Name) == "" {
		return response, ErrInvalidInput
	}
	err = s.transaction(ctx, func(local *OperationalService) error {

		value, err := local.repositories.AppModule.Lock(ctx, id)
		if err != nil {
			return err
		}

		value.Name = strings.TrimSpace(request.Name)
		value.DisplayOrder = request.DisplayOrder
		value.IsActive = *request.IsActive
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		updated, err := local.repositories.AppModule.Update(ctx, id, map[string]interface{}{"name": value.Name,
			"display_order": value.DisplayOrder, "is_active": value.IsActive}, "", nil)
		response = mapAppModule(updated)
		return err
	})
	return response, err
}
func (s *OperationalService) DeactivateAppModule(ctx context.Context, id string) (response dto.AppModuleResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}

	err = s.transaction(ctx, func(local *OperationalService) error {

		value, err := local.repositories.AppModule.Lock(ctx, id)
		if err != nil {
			return err
		}
		_ = value
		updated, err := local.repositories.AppModule.Update(ctx, id, map[string]interface{}{"is_active": false}, "", nil)
		response = mapAppModule(updated)
		return err
	})
	return response, err
}
