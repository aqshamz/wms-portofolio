package billing

import (
	"context"
	"math/big"
	dto "wms-api/dto/billing"
	model "wms-api/models/billing"
	repository "wms-api/repository/billing"
)

func (s *Service) CreateRun(ctx context.Context, q dto.CreateBillingRunRequest, actor string) (out dto.BillingRunResponse, err error) {
	business, e := date(q.BusinessDate, "business_date")
	if e != nil {
		return out, e
	}
	from, e := date(q.PeriodFrom, "period_from")
	if e != nil {
		return out, e
	}
	until, e := date(q.PeriodUntil, "period_until")
	if e != nil {
		return out, e
	}
	if until.Before(from) {
		return out, invalid("period_until cannot precede period_from")
	}
	err = s.transaction(ctx, func(local *Service) error {
		c, e := local.repositories.Contract.Lock(ctx, q.BillingContractID)
		if e != nil {
			return e
		}
		cs, e := local.status(ctx, c.DocumentTypeID, "ACTIVE")
		if e != nil {
			return e
		}
		expired, e := local.status(ctx, c.DocumentTypeID, "EXPIRED")
		if e != nil || (c.StatusID != cs.ID && c.StatusID != expired.ID) {
			return state("billing contract must be ACTIVE or EXPIRED")
		}
		if from.Before(c.EffectiveFrom) || (c.EffectiveUntil != nil && until.After(*c.EffectiveUntil)) {
			return invalid("billing period is outside contract effective dates")
		}
		kind, e := local.documentType(ctx, "BILLING_RUN")
		if e != nil {
			return e
		}
		status, e := local.initialStatus(ctx, kind.ID)
		if e != nil {
			return e
		}
		id, e := local.number(ctx, "BILLING_RUN", business)
		if e != nil {
			return e
		}
		v := model.BillingRun{ID: id, BillingContractID: c.ID, DocumentTypeID: kind.ID, StatusID: status.ID, OwnerID: c.OwnerID, WarehouseID: c.WarehouseID, PeriodFrom: from, PeriodUntil: until, BusinessDate: business, CurrencyCode: c.CurrencyCode, SubtotalAmount: "0.000000", TaxAmount: "0.000000", TotalAmount: "0.000000", Notes: q.Notes, VersionNo: 1, CreatedBy: actor}
		if e = local.repositories.Run.Create(ctx, &v); e != nil {
			return e
		}
		row, e := local.repositories.Run.Get(ctx, id)
		if e == nil {
			out = mapRun(row)
		}
		return e
	})
	return out, err
}
func (s *Service) GetRun(ctx context.Context, id string) (dto.BillingRunResponse, error) {
	v, e := s.repositories.Run.Get(ctx, id)
	if e != nil {
		return dto.BillingRunResponse{}, e
	}
	out := mapRun(v)
	charges, e := s.repositories.Charge.List(ctx, id)
	if e != nil {
		return out, e
	}
	out.Charges = make([]dto.ChargeResponse, 0, len(charges))
	for _, x := range charges {
		out.Charges = append(out.Charges, mapCharge(x))
	}
	return out, nil
}
func (s *Service) ListRuns(ctx context.Context, f repository.ListFilter) (dto.PageResponse[dto.BillingRunResponse], error) {
	v, n, e := s.repositories.Run.List(ctx, f)
	if e != nil {
		return dto.PageResponse[dto.BillingRunResponse]{}, e
	}
	x := make([]dto.BillingRunResponse, 0, len(v))
	for _, r := range v {
		x = append(x, mapRun(r))
	}
	return page(x, f.Page, f.PageSize, n), nil
}
func maxRat(a, b *big.Rat) *big.Rat {
	if a.Cmp(b) >= 0 {
		return new(big.Rat).Set(a)
	}
	return new(big.Rat).Set(b)
}
func (s *Service) CalculateRun(ctx context.Context, id string, version int64, actor string) (out dto.BillingRunResponse, err error) {
	err = s.transaction(ctx, func(local *Service) error {
		run, e := local.repositories.Run.Lock(ctx, id)
		if e != nil {
			return e
		}
		if run.VersionNo != version {
			return repository.ErrConcurrentWrite
		}
		draft, e := local.status(ctx, run.DocumentTypeID, "DRAFT")
		if e != nil || run.StatusID != draft.ID {
			return state("only a DRAFT billing run can be calculated")
		}
		events, e := local.repositories.Event.PendingForRun(ctx, run.BillingContractID, run.PeriodFrom, run.PeriodUntil)
		if e != nil {
			return e
		}
		if len(events) == 0 {
			return state("no PENDING billable events exist for this run")
		}
		ratedType, e := local.documentType(ctx, "BILLABLE_EVENT")
		if e != nil {
			return e
		}
		rated, e := local.status(ctx, ratedType.ID, "RATED")
		if e != nil {
			return e
		}
		subtotalTotal, taxTotal := new(big.Rat), new(big.Rat)
		for _, event := range events {
			line, e := local.repositories.RateCardLine.Get(ctx, event.RateCardLineID)
			if e != nil {
				return e
			}
			qtyText := event.Quantity
			qty, _ := new(big.Rat).SetString(qtyText)
			if line.BillingBasis == "EVENT" {
				qty = big.NewRat(1, 1)
				qtyText = "1.000000"
			}
			rate, _ := new(big.Rat).SetString(line.UnitRate)
			minimum, _ := new(big.Rat).SetString(line.MinimumCharge)
			sub := maxRat(new(big.Rat).Mul(qty, rate), minimum)
			taxPct, _ := new(big.Rat).SetString(line.TaxPercent)
			tax := new(big.Rat).Quo(new(big.Rat).Mul(sub, taxPct), big.NewRat(100, 1))
			subText, taxText := sub.FloatString(6), tax.FloatString(6)
			subRounded, _ := new(big.Rat).SetString(subText)
			taxRounded, _ := new(big.Rat).SetString(taxText)
			total := new(big.Rat).Add(subRounded, taxRounded)
			chargeID, e := randomID("CHG")
			if e != nil {
				return e
			}
			charge := model.BillingCharge{ID: chargeID, BillingRunID: id, BillableEventID: event.ID, ServiceCode: line.ServiceCode, Description: line.Description, SourceDocumentID: event.SourceDocumentID, Quantity: qtyText, UnitRate: line.UnitRate, SubtotalAmount: subText, TaxPercent: line.TaxPercent, TaxAmount: taxText, TotalAmount: total.FloatString(6), CreatedBy: actor}
			if e = local.repositories.Charge.Create(ctx, &charge); e != nil {
				return e
			}
			if e = local.repositories.Event.SetStatus(ctx, event.ID, rated.ID, event.Notes); e != nil {
				return e
			}
			subtotalTotal.Add(subtotalTotal, subRounded)
			taxTotal.Add(taxTotal, taxRounded)
		}
		target, e := local.transition(ctx, run.DocumentTypeID, run.StatusID, "CALCULATED")
		if e != nil {
			return e
		}
		if e = local.repositories.Run.SetStatusAndTotals(ctx, id, target.ID, actor, version, subtotalTotal.FloatString(6), taxTotal.FloatString(6), new(big.Rat).Add(subtotalTotal, taxTotal).FloatString(6)); e != nil {
			return e
		}
		out, e = local.GetRun(ctx, id)
		return e
	})
	return out, err
}
func (s *Service) TransitionRun(ctx context.Context, id, to string, version int64, actor string) (out dto.BillingRunResponse, err error) {
	err = s.transaction(ctx, func(local *Service) error {
		run, e := local.repositories.Run.Lock(ctx, id)
		if e != nil {
			return e
		}
		if run.VersionNo != version {
			return repository.ErrConcurrentWrite
		}
		target, e := local.transition(ctx, run.DocumentTypeID, run.StatusID, to)
		if e != nil {
			return e
		}
		if to == "DRAFT" || to == "CANCELLED" {
			ratedType, e := local.documentType(ctx, "BILLABLE_EVENT")
			if e != nil {
				return e
			}
			pending, e := local.status(ctx, ratedType.ID, "PENDING")
			if e != nil {
				return e
			}
			if e = local.repositories.Event.ResetRunEvents(ctx, id, pending.ID); e != nil {
				return e
			}
			if e = local.repositories.Charge.DeleteByRun(ctx, id); e != nil {
				return e
			}
			if e = local.repositories.Run.SetStatusAndTotals(ctx, id, target.ID, actor, version, "0.000000", "0.000000", "0.000000"); e != nil {
				return e
			}
		} else {
			if e = local.repositories.Run.SetStatus(ctx, id, target.ID, actor, version); e != nil {
				return e
			}
		}
		out, e = local.GetRun(ctx, id)
		return e
	})
	return out, err
}
