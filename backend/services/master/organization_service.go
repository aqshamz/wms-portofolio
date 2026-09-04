package master

import (
	"context"
	"errors"
	"strings"

	dto "wms-api/dto/master"
	model "wms-api/models/master"
	repository "wms-api/repository/master"

	"gorm.io/gorm"
)

type OrganizationService struct {
	repository *repository.OrganizationRepository
}

func NewOrganizationService(repository *repository.OrganizationRepository) *OrganizationService {
	return &OrganizationService{repository: repository}
}

func (s *OrganizationService) Create(
	ctx context.Context,
	request dto.CreateOrganizationRequest,
	actorID string,
) (dto.OrganizationResponse, error) {
	code, err := normalizeCode(request.Code)
	if err != nil || strings.TrimSpace(request.Name) == "" || validateTimezone(request.TimezoneName) != nil {
		return dto.OrganizationResponse{}, ErrInvalidInput
	}

	organization := model.Organization{
		Code:         code,
		Name:         strings.TrimSpace(request.Name),
		LegalName:    request.LegalName,
		TaxNumber:    request.TaxNumber,
		TimezoneName: request.TimezoneName,
		AddressLine1: request.AddressLine1,
		AddressLine2: request.AddressLine2,
		City:         request.City,
		Province:     request.Province,
		PostalCode:   request.PostalCode,
		CountryCode:  normalizeCountry(request.CountryCode),
		IsActive:     true,
		CreatedBy:    &actorID,
		UpdatedBy:    &actorID,
	}
	if err := s.repository.Create(ctx, &organization); err != nil {
		if repository.IsUniqueViolation(err) {
			return dto.OrganizationResponse{}, ErrConflict
		}
		return dto.OrganizationResponse{}, err
	}
	return organizationResponse(organization), nil
}

func (s *OrganizationService) Get(ctx context.Context, id string) (dto.OrganizationResponse, error) {
	if validateID(id) != nil {
		return dto.OrganizationResponse{}, ErrInvalidInput
	}
	organization, err := s.repository.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.OrganizationResponse{}, ErrNotFound
	}
	if err != nil {
		return dto.OrganizationResponse{}, err
	}
	return organizationResponse(organization), nil
}

func (s *OrganizationService) List(
	ctx context.Context,
	search *string,
	active *bool,
	page, pageSize int,
) (dto.PageResponse[dto.OrganizationResponse], error) {
	if search != nil {
		value := strings.TrimSpace(*search)
		search = &value
	}
	rows, total, err := s.repository.List(ctx, search, active, pageSize, (page-1)*pageSize)
	if err != nil {
		return dto.PageResponse[dto.OrganizationResponse]{}, err
	}
	items := make([]dto.OrganizationResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.OrganizationResponse{
			ID:           row.ID,
			Code:         row.Code,
			Name:         row.Name,
			LegalName:    row.LegalName,
			TimezoneName: row.TimezoneName,
			City:         row.City,
			CountryCode:  row.CountryCode,
			IsActive:     row.IsActive,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		})
	}
	return pageResponse(items, page, pageSize, total), nil
}

func (s *OrganizationService) Update(
	ctx context.Context,
	id string,
	request dto.UpdateOrganizationRequest,
	actorID string,
) (dto.OrganizationResponse, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return dto.OrganizationResponse{}, err
	}
	if strings.TrimSpace(request.Name) == "" || validateTimezone(request.TimezoneName) != nil {
		return dto.OrganizationResponse{}, ErrInvalidInput
	}
	organization, err := s.repository.Update(ctx, id, request.ExpectedUpdatedAt, map[string]interface{}{
		"name":           strings.TrimSpace(request.Name),
		"legal_name":     request.LegalName,
		"tax_number":     request.TaxNumber,
		"timezone_name":  request.TimezoneName,
		"address_line_1": request.AddressLine1,
		"address_line_2": request.AddressLine2,
		"city":           request.City,
		"province":       request.Province,
		"postal_code":    request.PostalCode,
		"country_code":   normalizeCountry(request.CountryCode),
		"is_active":      request.IsActive,
		"updated_at":     gorm.Expr("clock_timestamp()"),
		"updated_by":     actorID,
	})
	if errors.Is(err, repository.ErrConcurrentUpdate) {
		return dto.OrganizationResponse{}, ErrConcurrentUpdate
	}
	if err != nil {
		return dto.OrganizationResponse{}, err
	}
	return organizationResponse(organization), nil
}

func (s *OrganizationService) Deactivate(
	ctx context.Context,
	id string,
	request dto.DeactivateRequest,
	actorID string,
) (dto.OrganizationResponse, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return dto.OrganizationResponse{}, err
	}
	organization, err := s.repository.Deactivate(ctx, id, actorID, request.ExpectedUpdatedAt)
	if errors.Is(err, repository.ErrConcurrentUpdate) {
		return dto.OrganizationResponse{}, ErrConcurrentUpdate
	}
	if err != nil {
		return dto.OrganizationResponse{}, err
	}
	return organizationResponse(organization), nil
}

func organizationResponse(value model.Organization) dto.OrganizationResponse {
	return dto.OrganizationResponse{
		ID:           value.ID,
		Code:         value.Code,
		Name:         value.Name,
		LegalName:    value.LegalName,
		TaxNumber:    value.TaxNumber,
		TimezoneName: value.TimezoneName,
		AddressLine1: value.AddressLine1,
		AddressLine2: value.AddressLine2,
		City:         value.City,
		Province:     value.Province,
		PostalCode:   value.PostalCode,
		CountryCode:  value.CountryCode,
		IsActive:     value.IsActive,
		CreatedAt:    value.CreatedAt,
		UpdatedAt:    value.UpdatedAt,
	}
}
