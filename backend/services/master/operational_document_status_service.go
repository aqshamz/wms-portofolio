package master

import (
	"context"
	"strings"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
)

func (s *OperationalService) GetDocumentStatus(ctx context.Context, parentID string, id string) (response dto.DocumentStatusResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}
	if validateID(parentID) != nil {
		return response, ErrInvalidInput
	}
	value, err := s.repositories.DocumentStatus.Get(ctx, id)
	if err != nil {
		return response, catalogError(err)
	}
	if !strings.EqualFold(value.DocumentTypeID, parentID) {
		return response, ErrNotFound
	}
	return mapDocumentStatus(value), nil
}
func (s *OperationalService) ListDocumentStatus(ctx context.Context, parentID string, request dto.OperationalListRequest) (dto.PageResponse[dto.DocumentStatusResponse], error) {
	filter, err := operationalFilter(request, parentID)
	if err != nil {
		return dto.PageResponse[dto.DocumentStatusResponse]{}, err
	}
	if _, err := s.repositories.DocumentType.Get(ctx, parentID); err != nil {
		return dto.PageResponse[dto.DocumentStatusResponse]{}, catalogError(err)
	}
	rows, total, err := s.repositories.DocumentStatus.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.DocumentStatusResponse]{}, catalogError(err)
	}
	items := make([]dto.DocumentStatusResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapDocumentStatus(row))
	}
	return pageResponse(items, request.Page, request.PageSize, total), nil
}

func (s *OperationalService) CreateDocumentStatus(ctx context.Context, parentID string, request dto.CreateDocumentStatusRequest) (response dto.DocumentStatusResponse, err error) {
	if validateID(parentID) != nil {
		return response, ErrInvalidInput
	}
	code, name, err := catalogIdentity(request.Code, request.Name)
	if err != nil {
		return response, err
	}
	err = s.transaction(ctx, func(local *OperationalService) error {

		parent, err := local.repositories.DocumentType.Lock(ctx, parentID)
		if err != nil {
			return err
		}
		if !parent.IsActive {
			return invalidCatalog("parent configuration is inactive")
		}
		value := model.DocumentStatus{
			Code:           code,
			Name:           name,
			Description:    request.Description,
			IsInitial:      request.IsInitial,
			IsFinal:        request.IsFinal,
			IsCancelled:    request.IsCancelled,
			DisplayOrder:   request.DisplayOrder,
			DocumentTypeID: parentID, IsActive: true}
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		if err := local.repositories.DocumentStatus.Create(ctx, &value); err != nil {
			return err
		}
		response = mapDocumentStatus(value)
		return nil
	})
	return response, err
}
func (s *OperationalService) UpdateDocumentStatus(ctx context.Context, parentID string, id string, request dto.UpdateDocumentStatusRequest) (response dto.DocumentStatusResponse, err error) {
	if validateID(id) != nil || request.IsActive == nil {
		return response, ErrInvalidInput
	}
	if validateID(parentID) != nil {
		return response, ErrInvalidInput
	}
	if strings.TrimSpace(request.Name) == "" {
		return response, ErrInvalidInput
	}
	err = s.transaction(ctx, func(local *OperationalService) error {

		parent, err := local.repositories.DocumentType.Lock(ctx, parentID)
		if err != nil {
			return err
		}
		if !parent.IsActive {
			return invalidCatalog("parent configuration is inactive")
		}
		value, err := local.repositories.DocumentStatus.Lock(ctx, id)
		if err != nil {
			return err
		}
		if !strings.EqualFold(value.DocumentTypeID, parentID) {
			return ErrNotFound
		}
		value.Name = strings.TrimSpace(request.Name)
		value.Description = request.Description
		value.IsInitial = request.IsInitial
		value.IsFinal = request.IsFinal
		value.IsCancelled = request.IsCancelled
		value.DisplayOrder = request.DisplayOrder
		value.IsActive = *request.IsActive
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		updated, err := local.repositories.DocumentStatus.Update(ctx, id, map[string]interface{}{"name": value.Name,
			"description":   value.Description,
			"is_initial":    value.IsInitial,
			"is_final":      value.IsFinal,
			"is_cancelled":  value.IsCancelled,
			"display_order": value.DisplayOrder, "is_active": value.IsActive}, "", nil)
		response = mapDocumentStatus(updated)
		return err
	})
	return response, err
}
func (s *OperationalService) DeactivateDocumentStatus(ctx context.Context, parentID string, id string) (response dto.DocumentStatusResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}
	if validateID(parentID) != nil {
		return response, ErrInvalidInput
	}
	err = s.transaction(ctx, func(local *OperationalService) error {

		if _, err := local.repositories.DocumentType.Lock(ctx, parentID); err != nil {
			return err
		}
		value, err := local.repositories.DocumentStatus.Lock(ctx, id)
		if err != nil {
			return err
		}
		if !strings.EqualFold(value.DocumentTypeID, parentID) {
			return ErrNotFound
		}
		updated, err := local.repositories.DocumentStatus.Update(ctx, id, map[string]interface{}{"is_active": false, "is_initial": false}, "", nil)
		response = mapDocumentStatus(updated)
		return err
	})
	return response, err
}
