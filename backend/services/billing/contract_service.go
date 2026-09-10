package billing

import (
	"context"
	"strings"
	"time"
	dto "wms-api/dto/billing"
	model "wms-api/models/billing"
	repository "wms-api/repository/billing"
)

func (s *Service) CreateContract(ctx context.Context, q dto.CreateContractRequest, actor string) (out dto.ContractResponse, err error) {
	if !uuidPattern.MatchString(q.OwnerID) || !uuidPattern.MatchString(q.WarehouseID) || !uuidPattern.MatchString(actor) {
		return out, invalid("owner_id, warehouse_id and actor must be UUIDs")
	}
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
	currency := strings.ToUpper(strings.TrimSpace(q.CurrencyCode))
	cycle := strings.ToUpper(strings.TrimSpace(q.BillingCycle))
	if len(currency) != 3 || (cycle != "WEEKLY" && cycle != "MONTHLY") {
		return out, invalid("currency_code must be three letters and billing_cycle must be WEEKLY or MONTHLY")
	}
	err = s.transaction(ctx, func(local *Service) error {
		kind, e := local.documentType(ctx, "BILLING_CONTRACT")
		if e != nil {
			return e
		}
		status, e := local.initialStatus(ctx, kind.ID)
		if e != nil {
			return e
		}
		id, e := local.number(ctx, "BILLING_CONTRACT", business)
		if e != nil {
			return e
		}
		v := model.BillingContract{ID: id, DocumentTypeID: kind.ID, StatusID: status.ID, OwnerID: q.OwnerID, WarehouseID: q.WarehouseID, CurrencyCode: currency, BillingCycle: cycle, PaymentTermDays: q.PaymentTermDays, EffectiveFrom: from, EffectiveUntil: until, Notes: q.Notes, VersionNo: 1, CreatedBy: actor}
		if e = local.repositories.Contract.Create(ctx, &v); e != nil {
			return e
		}
		row, e := local.repositories.Contract.Get(ctx, id)
		if e == nil {
			out = mapContract(row)
		}
		return e
	})
	return out, err
}
func (s *Service) GetContract(ctx context.Context, id string) (dto.ContractResponse, error) {
	v, e := s.repositories.Contract.Get(ctx, id)
	return mapContract(v), e
}
func (s *Service) ListContracts(ctx context.Context, f repository.ListFilter) (dto.PageResponse[dto.ContractResponse], error) {
	v, n, e := s.repositories.Contract.List(ctx, f)
	if e != nil {
		return dto.PageResponse[dto.ContractResponse]{}, e
	}
	x := make([]dto.ContractResponse, 0, len(v))
	for _, r := range v {
		x = append(x, mapContract(r))
	}
	return page(x, f.Page, f.PageSize, n), nil
}
func (s *Service) UpdateContract(ctx context.Context, id string, q dto.UpdateContractRequest, actor string) (dto.ContractResponse, error) {
	if !validID(id, 120) {
		return dto.ContractResponse{}, invalid("invalid billing contract id")
	}
	from, e := date(q.EffectiveFrom, "effective_from")
	if e != nil {
		return dto.ContractResponse{}, e
	}
	until, e := optionalDate(q.EffectiveUntil, "effective_until")
	if e != nil {
		return dto.ContractResponse{}, e
	}
	if until != nil && until.Before(from) {
		return dto.ContractResponse{}, invalid("effective_until cannot precede effective_from")
	}
	currency := strings.ToUpper(strings.TrimSpace(q.CurrencyCode))
	cycle := strings.ToUpper(strings.TrimSpace(q.BillingCycle))
	if len(currency) != 3 || (cycle != "WEEKLY" && cycle != "MONTHLY") {
		return dto.ContractResponse{}, invalid("invalid currency_code or billing_cycle")
	}
	if e = s.repositories.Contract.UpdateDraft(ctx, id, actor, q.ExpectedVersion, map[string]interface{}{"currency_code": currency, "billing_cycle": cycle, "payment_term_days": q.PaymentTermDays, "effective_from": from, "effective_until": until, "notes": q.Notes}); e != nil {
		return dto.ContractResponse{}, e
	}
	return s.GetContract(ctx, id)
}
func (s *Service) TransitionContract(ctx context.Context, id, to string, version int64, actor string) (out dto.ContractResponse, err error) {
	err = s.transaction(ctx, func(local *Service) error {
		v, e := local.repositories.Contract.Lock(ctx, id)
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
		if to == "ACTIVE" {
			until := time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
			if v.EffectiveUntil != nil {
				until = *v.EffectiveUntil
			}
			rows, e := local.repositories.Contract.ActiveForPeriod(ctx, v.OwnerID, v.WarehouseID, v.EffectiveFrom, until)
			if e != nil {
				return e
			}
			for _, row := range rows {
				if row.ID != id {
					return state("an active billing contract overlaps this effective period")
				}
			}
		}
		if e = local.repositories.Contract.SetStatus(ctx, id, target.ID, actor, version); e != nil {
			return e
		}
		row, e := local.repositories.Contract.Get(ctx, id)
		if e == nil {
			out = mapContract(row)
		}
		return e
	})
	return out, err
}
