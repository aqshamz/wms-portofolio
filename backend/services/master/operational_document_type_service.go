package master

import (
	"context"
	"strings"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
)

func (s *OperationalService) GetDocumentType(ctx context.Context, id string) (response dto.DocumentTypeResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}

	value, err := s.repositories.DocumentType.Get(ctx, id)
	if err != nil {
		return response, catalogError(err)
	}

	return mapDocumentType(value), nil
}
func (s *OperationalService) ListDocumentType(ctx context.Context, request dto.OperationalListRequest) (dto.PageResponse[dto.DocumentTypeResponse], error) {
	filter, err := operationalFilter(request, "")
	if err != nil {
		return dto.PageResponse[dto.DocumentTypeResponse]{}, err
	}

	rows, total, err := s.repositories.DocumentType.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.DocumentTypeResponse]{}, catalogError(err)
	}
	items := make([]dto.DocumentTypeResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapDocumentType(row))
	}
	return pageResponse(items, request.Page, request.PageSize, total), nil
}

func (s *OperationalService) CreateDocumentType(ctx context.Context, request dto.CreateDocumentTypeRequest) (response dto.DocumentTypeResponse, err error) {

	code, name, err := catalogIdentity(request.Code, request.Name)
	if err != nil {
		return response, err
	}
	err = s.transaction(ctx, func(local *OperationalService) error {

		value := model.DocumentType{
			Code:        code,
			Name:        name,
			ModuleCode:  request.ModuleCode,
			Description: request.Description,
			IsActive:    true}
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		if err := local.repositories.DocumentType.Create(ctx, &value); err != nil {
			return err
		}
		response = mapDocumentType(value)
		return nil
	})
	return response, err
}
func (s *OperationalService) UpdateDocumentType(ctx context.Context, id string, request dto.UpdateDocumentTypeRequest) (response dto.DocumentTypeResponse, err error) {
	if validateID(id) != nil || request.IsActive == nil {
		return response, ErrInvalidInput
	}

	if strings.TrimSpace(request.Name) == "" {
		return response, ErrInvalidInput
	}
	err = s.transaction(ctx, func(local *OperationalService) error {

		value, err := local.repositories.DocumentType.Lock(ctx, id)
		if err != nil {
			return err
		}

		value.Name = strings.TrimSpace(request.Name)
		value.ModuleCode = request.ModuleCode
		value.Description = request.Description
		value.IsActive = *request.IsActive
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		updated, err := local.repositories.DocumentType.Update(ctx, id, map[string]interface{}{"name": value.Name,
			"module_code": value.ModuleCode,
			"description": value.Description, "is_active": value.IsActive}, "", nil)
		response = mapDocumentType(updated)
		return err
	})
	return response, err
}
func (s *OperationalService) DeactivateDocumentType(ctx context.Context, id string) (response dto.DocumentTypeResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}

	err = s.transaction(ctx, func(local *OperationalService) error {

		value, err := local.repositories.DocumentType.Lock(ctx, id)
		if err != nil {
			return err
		}
		_ = value
		updated, err := local.repositories.DocumentType.Update(ctx, id, map[string]interface{}{"is_active": false}, "", nil)
		response = mapDocumentType(updated)
		return err
	})
	return response, err
}
