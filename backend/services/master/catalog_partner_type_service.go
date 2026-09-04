package master

import (
	"context"
	dto "wms-api/dto/master"
	repository "wms-api/repository/master"
)

func (s *CatalogService) ListBusinessPartnerType(ctx context.Context, partnerID string) ([]dto.PartnerTypeResponse, error) {
	if validateID(partnerID) != nil {
		return nil, ErrInvalidInput
	}
	if _, err := s.repositories.BusinessPartner.Get(ctx, partnerID); err != nil {
		return nil, catalogError(err)
	}
	rows, err := s.repositories.BusinessPartnerType.List(ctx, partnerID)
	result := make([]dto.PartnerTypeResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapPartnerType(row))
	}
	return result, catalogError(err)
}
func (s *CatalogService) AssignBusinessPartnerType(ctx context.Context, partnerID string, request dto.AssignPartnerTypeRequest) ([]dto.PartnerTypeResponse, error) {
	if validateID(partnerID) != nil || validateID(request.PartnerTypeID) != nil {
		return nil, ErrInvalidInput
	}
	err := s.repositories.Transaction(ctx, func(repos *repository.CatalogRepositories) error {
		partner, err := repos.BusinessPartner.Get(ctx, partnerID)
		if err != nil {
			return err
		}
		if !partner.IsActive {
			return invalidCatalog("partner is inactive")
		}
		kind, err := repos.PartnerType.Get(ctx, request.PartnerTypeID)
		if err != nil {
			return err
		}
		if !kind.IsActive {
			return invalidCatalog("partner type is inactive")
		}
		return repos.BusinessPartnerType.Assign(ctx, partnerID, request.PartnerTypeID)
	})
	if err != nil {
		return nil, catalogError(err)
	}
	return s.ListBusinessPartnerType(ctx, partnerID)
}
func (s *CatalogService) RemoveBusinessPartnerType(ctx context.Context, partnerID, typeID string) error {
	if validateID(partnerID) != nil || validateID(typeID) != nil {
		return ErrInvalidInput
	}
	return catalogError(s.repositories.BusinessPartnerType.Remove(ctx, partnerID, typeID))
}
