package master

import (
	"context"
	"strings"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
	repository "wms-api/repository/master"
)

func (s *CatalogService) CreateBusinessPartner(ctx context.Context, request dto.CreateBusinessPartnerRequest, actor string) (response dto.BusinessPartnerResponse, err error) {
	code, name, err := catalogIdentity(request.Code, request.Name)
	if err != nil || validateID(actor) != nil {
		return response, ErrInvalidInput
	}

	err = s.repositories.Transaction(ctx, func(repos *repository.CatalogRepositories) error {
		local := NewCatalogService(repos)
		if err := local.requireOwner(ctx, request.OwnerID, true); err != nil {
			return err
		}

		value := model.BusinessPartner{OwnerID: request.OwnerID,
			Code:         code,
			Name:         name,
			LegalName:    request.LegalName,
			TaxNumber:    request.TaxNumber,
			Email:        request.Email,
			Phone:        request.Phone,
			AddressLine1: request.AddressLine1,
			AddressLine2: request.AddressLine2,
			City:         request.City,
			Province:     request.Province,
			PostalCode:   request.PostalCode,
			CountryCode:  request.CountryCode,
			IsActive:     true, CreatedBy: &actor, UpdatedBy: &actor,
		}
		if err := repos.BusinessPartner.Create(ctx, &value); err != nil {
			return err
		}

		response = mapBusinessPartner(value)
		return nil
	})
	return response, catalogError(err)
}
func (s *CatalogService) GetBusinessPartner(ctx context.Context, id string) (response dto.PartnerDetailResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}
	value, err := s.repositories.BusinessPartner.Get(ctx, id)
	if err != nil {
		return response, catalogError(err)
	}
	response.BusinessPartnerResponse = mapBusinessPartner(value)
	response.PartnerTypes, err = s.ListBusinessPartnerType(ctx, id)
	return response, err
}
func (s *CatalogService) ListBusinessPartner(ctx context.Context, filter repository.CatalogFilter) (dto.PageResponse[dto.BusinessPartnerResponse], error) {
	if err := catalogPage(filter); err != nil {
		return dto.PageResponse[dto.BusinessPartnerResponse]{}, err
	}
	if filter.OwnerID != "" && validateID(filter.OwnerID) != nil {
		return dto.PageResponse[dto.BusinessPartnerResponse]{}, ErrInvalidInput
	}
	filter.CategoryID = ""
	filter.PartnerTypeCode = strings.ToUpper(strings.TrimSpace(filter.PartnerTypeCode))
	filter.ItemID = ""
	rows, total, err := s.repositories.BusinessPartner.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.BusinessPartnerResponse]{}, catalogError(err)
	}
	items := make([]dto.BusinessPartnerResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapBusinessPartner(row))
	}
	return pageResponse(items, filter.Page, filter.PageSize, total), nil
}
func (s *CatalogService) UpdateBusinessPartner(ctx context.Context, id string, request dto.UpdateBusinessPartnerRequest, actor string) (response dto.BusinessPartnerResponse, err error) {
	name := strings.TrimSpace(request.Name)
	if validateID(id) != nil || validateID(actor) != nil || name == "" || request.IsActive == nil || request.ExpectedUpdatedAt.IsZero() {
		return response, ErrInvalidInput
	}

	err = s.repositories.Transaction(ctx, func(repos *repository.CatalogRepositories) error {
		current, err := repos.BusinessPartner.Get(ctx, id)
		if err != nil {
			return err
		}
		local := NewCatalogService(repos)
		if err := local.requireOwner(ctx, current.OwnerID, true); err != nil {
			return err
		}

		value, err := repos.BusinessPartner.Update(ctx, id, map[string]interface{}{
			"name":           name,
			"legal_name":     request.LegalName,
			"tax_number":     request.TaxNumber,
			"email":          request.Email,
			"phone":          request.Phone,
			"address_line_1": request.AddressLine1,
			"address_line_2": request.AddressLine2,
			"city":           request.City,
			"province":       request.Province,
			"postal_code":    request.PostalCode,
			"country_code":   request.CountryCode,
			"is_active":      *request.IsActive,
		}, actor, &request.ExpectedUpdatedAt)
		response = mapBusinessPartner(value)
		return err
	})
	return response, catalogError(err)
}
func (s *CatalogService) DeactivateBusinessPartner(ctx context.Context, id string, request dto.DeactivateRequest, actor string) (dto.BusinessPartnerResponse, error) {
	if validateID(id) != nil || validateID(actor) != nil || request.ExpectedUpdatedAt.IsZero() {
		return dto.BusinessPartnerResponse{}, ErrInvalidInput
	}
	if _, err := s.repositories.BusinessPartner.Get(ctx, id); err != nil {
		return dto.BusinessPartnerResponse{}, catalogError(err)
	}
	value, err := s.repositories.BusinessPartner.Update(ctx, id, map[string]interface{}{"is_active": false}, actor, &request.ExpectedUpdatedAt)
	return mapBusinessPartner(value), catalogError(err)
}
