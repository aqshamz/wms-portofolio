package billing

import (
	"context"
	"strings"
	"time"
	dto "wms-api/dto/billing"
	model "wms-api/models/billing"
	repository "wms-api/repository/billing"
)

func (s *Service) validatedRateLine(ctx context.Context, cardID string, q dto.CreateRateCardLineRequest, actor string) (model.RateCardLine, error) {
	code := strings.ToUpper(strings.TrimSpace(q.ServiceCode))
	source := strings.ToUpper(strings.TrimSpace(q.SourceKind))
	basis := strings.ToUpper(strings.TrimSpace(q.BillingBasis))
	if !codePattern.MatchString(code) {
		return model.RateCardLine{}, invalid("invalid service_code")
	}
	if source != "MOVEMENT" && source != "STORAGE" && source != "MANUAL" {
		return model.RateCardLine{}, invalid("source_kind must be MOVEMENT, STORAGE, or MANUAL")
	}
	if basis != "QUANTITY" && basis != "EVENT" {
		return model.RateCardLine{}, invalid("billing_basis must be QUANTITY or EVENT")
	}
	if source == "MOVEMENT" && q.MovementTypeID == nil {
		return model.RateCardLine{}, invalid("movement_type_id is required for MOVEMENT")
	}
	if source != "MOVEMENT" && q.MovementTypeID != nil {
		return model.RateCardLine{}, invalid("movement_type_id must be omitted for STORAGE and MANUAL")
	}
	if q.MovementTypeID != nil {
		v, e := s.repositories.Inventory.MovementType.Get(ctx, *q.MovementTypeID)
		if e != nil || !v.IsActive {
			return model.RateCardLine{}, invalid("movement_type_id is unavailable")
		}
	}
	rate, _, e := amount(q.UnitRate, "unit_rate", true)
	if e != nil {
		return model.RateCardLine{}, e
	}
	minimum, _, e := amount(q.MinimumCharge, "minimum_charge", true)
	if strings.TrimSpace(q.MinimumCharge) == "" {
		minimum = "0.000000"
		e = nil
	}
	if e != nil {
		return model.RateCardLine{}, e
	}
	tax, _, e := percentage(q.TaxPercent, "tax_percent")
	if e != nil {
		return model.RateCardLine{}, e
	}
	description := strings.TrimSpace(q.Description)
	if description == "" {
		return model.RateCardLine{}, invalid("description is required")
	}
	return model.RateCardLine{RateCardID: cardID, ServiceCode: code, Description: description, SourceKind: source, MovementTypeID: q.MovementTypeID, BillingBasis: basis, UnitRate: rate, MinimumCharge: minimum, TaxPercent: tax, IsActive: true, CreatedBy: actor}, nil
}
func (s *Service) CreateRateCard(ctx context.Context, q dto.CreateRateCardRequest, actor string) (out dto.RateCardResponse, err error) {
	business, e := date(q.BusinessDate, "business_date")
	if e != nil {
		return out, e
	}
	from, e := date(q.EffectiveFrom, "effective_from")
	if e != nil {
		return out, e
	}
	until, e := optionalDate(q.EffectiveUntil, "effective_until")
	if e != nil {
		return out, e
	}
	if until != nil && until.Before(from) {
		return out, invalid("effective_until cannot precede effective_from")
	}
	name := strings.TrimSpace(q.Name)
	if name == "" {
		return out, invalid("name is required")
	}
	err = s.transaction(ctx, func(local *Service) error {
		contract, e := local.repositories.Contract.Lock(ctx, q.BillingContractID)
		if e != nil {
			return e
		}
		if from.Before(contract.EffectiveFrom) || (contract.EffectiveUntil != nil && (until == nil || until.After(*contract.EffectiveUntil))) {
			return invalid("rate card effective period must be inside the contract period")
		}
		kind, e := local.documentType(ctx, "RATE_CARD")
		if e != nil {
			return e
		}
		status, e := local.initialStatus(ctx, kind.ID)
		if e != nil {
			return e
		}
		id, e := local.number(ctx, "RATE_CARD", business)
		if e != nil {
			return e
		}
		v := model.RateCard{ID: id, BillingContractID: contract.ID, DocumentTypeID: kind.ID, StatusID: status.ID, Name: name, EffectiveFrom: from, EffectiveUntil: until, VersionNo: 1, CreatedBy: actor}
		if e = local.repositories.RateCard.Create(ctx, &v); e != nil {
			return e
		}
		for _, line := range q.Lines {
			x, e := local.validatedRateLine(ctx, id, line, actor)
			if e != nil {
				return e
			}
			if e = local.repositories.RateCardLine.Create(ctx, &x); e != nil {
				return e
			}
		}
		var e2 error
		out, e2 = local.GetRateCard(ctx, id)
		return e2
	})
	return out, err
}
func (s *Service) GetRateCard(ctx context.Context, id string) (dto.RateCardResponse, error) {
	v, e := s.repositories.RateCard.Get(ctx, id)
	if e != nil {
		return dto.RateCardResponse{}, e
	}
	out := mapRateCard(v)
	lines, e := s.repositories.RateCardLine.List(ctx, id, false)
	if e != nil {
		return out, e
	}
	out.Lines = make([]dto.RateCardLineResponse, 0, len(lines))
	for _, x := range lines {
		out.Lines = append(out.Lines, mapRateLine(x))
	}
	return out, nil
}
func (s *Service) ListRateCards(ctx context.Context, f repository.ListFilter) (dto.PageResponse[dto.RateCardResponse], error) {
	v, n, e := s.repositories.RateCard.List(ctx, f)
	if e != nil {
		return dto.PageResponse[dto.RateCardResponse]{}, e
	}
	x := make([]dto.RateCardResponse, 0, len(v))
	for _, r := range v {
		x = append(x, mapRateCard(r))
	}
	return page(x, f.Page, f.PageSize, n), nil
}
func (s *Service) UpdateRateCard(ctx context.Context, id string, q dto.UpdateRateCardRequest, actor string) (dto.RateCardResponse, error) {
	from, e := date(q.EffectiveFrom, "effective_from")
	if e != nil {
		return dto.RateCardResponse{}, e
	}
	until, e := optionalDate(q.EffectiveUntil, "effective_until")
	if e != nil {
		return dto.RateCardResponse{}, e
	}
	if until != nil && until.Before(from) {
		return dto.RateCardResponse{}, invalid("effective_until cannot precede effective_from")
	}
	if e = s.repositories.RateCard.UpdateDraft(ctx, id, actor, q.ExpectedVersion, map[string]interface{}{"name": strings.TrimSpace(q.Name), "effective_from": from, "effective_until": until}); e != nil {
		return dto.RateCardResponse{}, e
	}
	return s.GetRateCard(ctx, id)
}
func (s *Service) AddRateCardLine(ctx context.Context, id string, q dto.AddRateCardLineRequest, actor string) (dto.RateCardResponse, error) {
	err := s.transaction(ctx, func(local *Service) error {
		v, e := local.repositories.RateCard.Lock(ctx, id)
		if e != nil {
			return e
		}
		status, e := local.status(ctx, v.DocumentTypeID, "DRAFT")
		if e != nil || v.StatusID != status.ID {
			return state("rate card lines can only change in DRAFT")
		}
		if v.VersionNo != q.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		line, e := local.validatedRateLine(ctx, id, q.CreateRateCardLineRequest, actor)
		if e != nil {
			return e
		}
		if e = local.repositories.RateCardLine.Create(ctx, &line); e != nil {
			return e
		}
		return local.repositories.RateCard.UpdateDraft(ctx, id, actor, q.ExpectedVersion, map[string]interface{}{})
	})
	if err != nil {
		return dto.RateCardResponse{}, err
	}
	return s.GetRateCard(ctx, id)
}
func (s *Service) UpdateRateCardLine(ctx context.Context, id, lineID string, q dto.UpdateRateCardLineRequest, actor string) (dto.RateCardResponse, error) {
	err := s.transaction(ctx, func(local *Service) error {
		v, e := local.repositories.RateCard.Lock(ctx, id)
		if e != nil {
			return e
		}
		draft, e := local.status(ctx, v.DocumentTypeID, "DRAFT")
		if e != nil || v.StatusID != draft.ID {
			return state("rate card lines can only change in DRAFT")
		}
		if v.VersionNo != q.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		old, e := local.repositories.RateCardLine.Get(ctx, lineID)
		if e != nil || old.RateCardID != id {
			return repository.ErrNotFound
		}
		line, e := local.validatedRateLine(ctx, id, q.CreateRateCardLineRequest, actor)
		if e != nil {
			return e
		}
		e = local.repositories.RateCardLine.Update(ctx, lineID, actor, map[string]interface{}{"service_code": line.ServiceCode, "description": line.Description, "source_kind": line.SourceKind, "movement_type_id": line.MovementTypeID, "billing_basis": line.BillingBasis, "unit_rate": line.UnitRate, "minimum_charge": line.MinimumCharge, "tax_percent": line.TaxPercent, "is_active": q.IsActive})
		if e != nil {
			return e
		}
		return local.repositories.RateCard.UpdateDraft(ctx, id, actor, q.ExpectedVersion, map[string]interface{}{})
	})
	if err != nil {
		return dto.RateCardResponse{}, err
	}
	return s.GetRateCard(ctx, id)
}
func (s *Service) DeleteRateCardLine(ctx context.Context, id, lineID string, version int64, actor string) (dto.RateCardResponse, error) {
	err := s.transaction(ctx, func(local *Service) error {
		v, e := local.repositories.RateCard.Lock(ctx, id)
		if e != nil {
			return e
		}
		draft, e := local.status(ctx, v.DocumentTypeID, "DRAFT")
		if e != nil || v.StatusID != draft.ID {
			return state("rate card lines can only change in DRAFT")
		}
		if v.VersionNo != version {
			return repository.ErrConcurrentWrite
		}
		lines, e := local.repositories.RateCardLine.List(ctx, id, false)
		if e != nil {
			return e
		}
		if len(lines) <= 1 {
			return state("rate card requires at least one line")
		}
		line, e := local.repositories.RateCardLine.Get(ctx, lineID)
		if e != nil || line.RateCardID != id {
			return repository.ErrNotFound
		}
		if e = local.repositories.RateCardLine.Delete(ctx, lineID); e != nil {
			return e
		}
		return local.repositories.RateCard.UpdateDraft(ctx, id, actor, version, map[string]interface{}{})
	})
	if err != nil {
		return dto.RateCardResponse{}, err
	}
	return s.GetRateCard(ctx, id)
}
func (s *Service) TransitionRateCard(ctx context.Context, id, to string, version int64, actor string) (out dto.RateCardResponse, err error) {
	err = s.transaction(ctx, func(local *Service) error {
		v, e := local.repositories.RateCard.Lock(ctx, id)
		if e != nil {
			return e
		}
		if v.VersionNo != version {
			return repository.ErrConcurrentWrite
		}
		target, e := local.transition(ctx, v.DocumentTypeID, v.StatusID, to)
		if e != nil {
			return e
		}
		if to == "APPROVED" {
			lines, e := local.repositories.RateCardLine.List(ctx, id, true)
			if e != nil {
				return e
			}
			if len(lines) == 0 {
				return state("rate card requires at least one active line")
			}
		}
		if to == "ACTIVE" {
			contract, e := local.repositories.Contract.Get(ctx, v.BillingContractID)
			if e != nil {
				return e
			}
			if contract.StatusCode != "ACTIVE" {
				return state("billing contract must be ACTIVE before its rate card")
			}
			if v.EffectiveFrom.Before(contract.EffectiveFrom) || (contract.EffectiveUntil != nil && (v.EffectiveUntil == nil || v.EffectiveUntil.After(*contract.EffectiveUntil))) {
				return state("rate card effective period is outside the contract period")
			}
			until := time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
			if v.EffectiveUntil != nil {
				until = *v.EffectiveUntil
			}
			cards, e := local.repositories.RateCard.ActiveForPeriod(ctx, v.BillingContractID, v.EffectiveFrom, until)
			if e != nil {
				return e
			}
			for _, c := range cards {
				if c.ID != id {
					return state("an active rate card overlaps this effective period")
				}
			}
		}
		if e = local.repositories.RateCard.SetStatus(ctx, id, target.ID, actor, version); e != nil {
			return e
		}
		out, e = local.GetRateCard(ctx, id)
		return e
	})
	return out, err
}
