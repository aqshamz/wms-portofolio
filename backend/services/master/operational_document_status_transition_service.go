package master

import (
	"context"
	"strings"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
)

func (s *OperationalService) GetDocumentStatusTransition(ctx context.Context, parentID string, id string) (response dto.DocumentStatusTransitionResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}
	if validateID(parentID) != nil {
		return response, ErrInvalidInput
	}
	value, err := s.repositories.DocumentStatusTransition.Get(ctx, id)
	if err != nil {
		return response, catalogError(err)
	}
	if !strings.EqualFold(value.DocumentTypeID, parentID) {
		return response, ErrNotFound
	}
	return mapDocumentStatusTransition(value), nil
}
func (s *OperationalService) ListDocumentStatusTransition(ctx context.Context, parentID string, request dto.OperationalListRequest) (dto.PageResponse[dto.DocumentStatusTransitionResponse], error) {
	filter, err := operationalFilter(request, parentID)
	if err != nil {
		return dto.PageResponse[dto.DocumentStatusTransitionResponse]{}, err
	}
	if _, err := s.repositories.DocumentType.Get(ctx, parentID); err != nil {
		return dto.PageResponse[dto.DocumentStatusTransitionResponse]{}, catalogError(err)
	}
	rows, total, err := s.repositories.DocumentStatusTransition.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.DocumentStatusTransitionResponse]{}, catalogError(err)
	}
	items := make([]dto.DocumentStatusTransitionResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapDocumentStatusTransition(row))
	}
	return pageResponse(items, request.Page, request.PageSize, total), nil
}

func (s *OperationalService) CreateDocumentStatusTransition(ctx context.Context, parentID string, request dto.CreateDocumentStatusTransitionRequest) (response dto.DocumentStatusTransitionResponse, err error) {
	if validateID(parentID) != nil {
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
		value := model.DocumentStatusTransition{
			FromStatusID:         request.FromStatusID,
			ToStatusID:           request.ToStatusID,
			RequiredPermissionID: request.RequiredPermissionID,
			DocumentTypeID:       parentID, IsActive: true}
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		if err := local.repositories.DocumentStatusTransition.Create(ctx, &value); err != nil {
			return err
		}
		response = mapDocumentStatusTransition(value)
		return nil
	})
	return response, err
}
func (s *OperationalService) UpdateDocumentStatusTransition(ctx context.Context, parentID string, id string, request dto.UpdateDocumentStatusTransitionRequest) (response dto.DocumentStatusTransitionResponse, err error) {
	if validateID(id) != nil || request.IsActive == nil {
		return response, ErrInvalidInput
	}
	if validateID(parentID) != nil {
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
		value, err := local.repositories.DocumentStatusTransition.Lock(ctx, id)
		if err != nil {
			return err
		}
		if !strings.EqualFold(value.DocumentTypeID, parentID) {
			return ErrNotFound
		}
		value.RequiredPermissionID = request.RequiredPermissionID
		value.IsActive = *request.IsActive
		if err := local.prepareWrite(ctx, &value); err != nil {
			return err
		}
		updated, err := local.repositories.DocumentStatusTransition.Update(ctx, id, map[string]interface{}{"required_permission_id": value.RequiredPermissionID, "is_active": value.IsActive}, "", nil)
		response = mapDocumentStatusTransition(updated)
		return err
	})
	return response, err
}
func (s *OperationalService) DeactivateDocumentStatusTransition(ctx context.Context, parentID string, id string) (response dto.DocumentStatusTransitionResponse, err error) {
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
		value, err := local.repositories.DocumentStatusTransition.Lock(ctx, id)
		if err != nil {
			return err
		}
		if !strings.EqualFold(value.DocumentTypeID, parentID) {
			return ErrNotFound
		}
		updated, err := local.repositories.DocumentStatusTransition.Update(ctx, id, map[string]interface{}{"is_active": false}, "", nil)
		response = mapDocumentStatusTransition(updated)
		return err
	})
	return response, err
}
