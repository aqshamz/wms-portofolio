package billing

import (
	"context"
	"fmt"
	"strings"
	"time"
	dto "wms-api/dto/billing"
	model "wms-api/models/billing"
	repository "wms-api/repository/billing"
)

func (s *Service) CollectEvents(ctx context.Context, q dto.CollectEventsRequest, actor string) (out dto.CollectEventsResponse, err error) {
	from, e := date(q.DateFrom, "date_from")
	if e != nil {
		return out, e
	}
	until, e := date(q.DateUntil, "date_until")
	if e != nil {
		return out, e
	}
	if until.Before(from) {
		return out, invalid("date_until cannot precede date_from")
	}
	err = s.transaction(ctx, func(local *Service) error {
		kind, e := local.documentType(ctx, "BILLABLE_EVENT")
		if e != nil {
			return e
		}
		pending, e := local.status(ctx, kind.ID, "PENDING")
		if e != nil {
			return e
		}
		cards, e := local.repositories.RateCard.ActiveByScope(ctx, q.OwnerID, q.WarehouseID, from, until)
		if e != nil {
			return e
		}
		if len(cards) == 0 {
			return state("no active contract and rate card covers the requested period")
		}
		for _, card := range cards {
			cardFrom, cardUntil := from, until
			if card.EffectiveFrom.After(cardFrom) {
				cardFrom = card.EffectiveFrom
			}
			if card.EffectiveUntil != nil && card.EffectiveUntil.Before(cardUntil) {
				cardUntil = *card.EffectiveUntil
			}
			lines, e := local.repositories.RateCardLine.List(ctx, card.ID, true)
			if e != nil {
				return e
			}
			for _, line := range lines {
				if line.SourceKind != "MOVEMENT" || line.MovementTypeID == nil {
					continue
				}
				moves, e := local.repositories.Source.Movements(ctx, q.OwnerID, q.WarehouseID, *line.MovementTypeID, cardFrom, cardUntil)
				if e != nil {
					return e
				}
				for _, m := range moves {
					id, e := randomID("BEV")
					if e != nil {
						return e
					}
					key := fmt.Sprintf("movement:%s:rate:%s", m.ID, line.ID)
					v := model.BillableEvent{ID: id, DocumentTypeID: kind.ID, StatusID: pending.ID, OwnerID: q.OwnerID, WarehouseID: q.WarehouseID, RateCardLineID: line.ID, InventoryMovementID: &m.ID, EventKey: key, BusinessDate: m.BusinessDate, SourceDocumentID: m.SourceDocumentID, SourceLineID: m.SourceLineID, Quantity: m.Quantity, UOMID: &m.UOMID, CreatedBy: actor}
					created, e := local.repositories.Event.CreateIgnore(ctx, &v)
					if e != nil {
						return e
					}
					if created {
						out.Created++
					} else {
						out.Existing++
					}
				}
			}
		}
		return nil
	})
	return out, err
}

func (s *Service) SnapshotStorage(ctx context.Context, q dto.StorageSnapshotRequest, actor string) (out dto.CollectEventsResponse, err error) {
	businessDate, e := date(q.BusinessDate, "business_date")
	if e != nil {
		return out, e
	}
	location, _ := time.LoadLocation(s.timezone)
	if q.BusinessDate != time.Now().In(location).Format("2006-01-02") {
		return out, invalid("storage snapshot business_date must be today's warehouse business date")
	}
	err = s.transaction(ctx, func(local *Service) error {
		kind, e := local.documentType(ctx, "BILLABLE_EVENT")
		if e != nil {
			return e
		}
		pending, e := local.status(ctx, kind.ID, "PENDING")
		if e != nil {
			return e
		}
		cards, e := local.repositories.RateCard.ActiveByScope(ctx, q.OwnerID, q.WarehouseID, businessDate, businessDate)
		if e != nil {
			return e
		}
		if len(cards) == 0 {
			return state("no active contract and rate card covers the snapshot date")
		}
		balances, e := local.repositories.Source.StorageBalances(ctx, q.OwnerID, q.WarehouseID)
		if e != nil {
			return e
		}
		for _, card := range cards {
			lines, e := local.repositories.RateCardLine.List(ctx, card.ID, true)
			if e != nil {
				return e
			}
			for _, line := range lines {
				if line.SourceKind != "STORAGE" {
					continue
				}
				for _, balance := range balances {
					id, e := randomID("BEV")
					if e != nil {
						return e
					}
					key := fmt.Sprintf("storage:%s:balance:%s:rate:%s", q.BusinessDate, balance.ID, line.ID)
					notes := "end-of-day storage snapshot"
					v := model.BillableEvent{ID: id, DocumentTypeID: kind.ID, StatusID: pending.ID, OwnerID: q.OwnerID, WarehouseID: q.WarehouseID, RateCardLineID: line.ID, EventKey: key, BusinessDate: businessDate, SourceDocumentID: balance.ID, Quantity: balance.Quantity, UOMID: &balance.UOMID, Notes: &notes, CreatedBy: actor}
					created, e := local.repositories.Event.CreateIgnore(ctx, &v)
					if e != nil {
						return e
					}
					if created {
						out.Created++
					} else {
						out.Existing++
					}
				}
			}
		}
		return nil
	})
	return out, err
}

func (s *Service) CreateManualEvent(ctx context.Context, contractID string, q dto.CreateManualEventRequest, actor string) (out dto.EventResponse, err error) {
	d, e := date(q.BusinessDate, "business_date")
	if e != nil {
		return out, e
	}
	qty, _, e := amount(q.Quantity, "quantity", false)
	if e != nil {
		return out, e
	}
	err = s.transaction(ctx, func(local *Service) error {
		contract, e := local.repositories.Contract.Lock(ctx, contractID)
		if e != nil {
			return e
		}
		active, e := local.status(ctx, contract.DocumentTypeID, "ACTIVE")
		if e != nil {
			return e
		}
		expired, e := local.status(ctx, contract.DocumentTypeID, "EXPIRED")
		if e != nil || (contract.StatusID != active.ID && contract.StatusID != expired.ID) {
			return state("billing contract must be ACTIVE or EXPIRED")
		}
		line, e := local.repositories.RateCardLine.Get(ctx, q.RateCardLineID)
		if e != nil {
			return e
		}
		card, e := local.repositories.RateCard.Get(ctx, line.RateCardID)
		if e != nil || card.BillingContractID != contract.ID || line.SourceKind != "MANUAL" || !line.IsActive {
			return invalid("rate_card_line_id is not an active MANUAL rate for this contract")
		}
		if (card.StatusCode != "ACTIVE" && card.StatusCode != "EXPIRED") || d.Before(card.EffectiveFrom) || (card.EffectiveUntil != nil && d.After(*card.EffectiveUntil)) {
			return state("manual event date is outside the active rate card period")
		}
		kind, e := local.documentType(ctx, "BILLABLE_EVENT")
		if e != nil {
			return e
		}
		pending, e := local.status(ctx, kind.ID, "PENDING")
		if e != nil {
			return e
		}
		id, e := randomID("BEV")
		if e != nil {
			return e
		}
		key := "manual:" + id
		source := strings.TrimSpace(q.SourceDocumentID)
		v := model.BillableEvent{ID: id, DocumentTypeID: kind.ID, StatusID: pending.ID, OwnerID: contract.OwnerID, WarehouseID: contract.WarehouseID, RateCardLineID: line.ID, EventKey: key, BusinessDate: d, SourceDocumentID: source, Quantity: qty, UOMID: q.UOMID, Notes: q.Notes, CreatedBy: actor}
		_, e = local.repositories.Event.CreateIgnore(ctx, &v)
		if e != nil {
			return e
		}
		row, e := local.repositories.Event.Get(ctx, id)
		if e == nil {
			out = mapEvent(row)
		}
		return e
	})
	return out, err
}
func (s *Service) GetEvent(ctx context.Context, id string) (dto.EventResponse, error) {
	v, e := s.repositories.Event.Get(ctx, id)
	return mapEvent(v), e
}
func (s *Service) ListEvents(ctx context.Context, f repository.ListFilter) (dto.PageResponse[dto.EventResponse], error) {
	v, n, e := s.repositories.Event.List(ctx, f)
	if e != nil {
		return dto.PageResponse[dto.EventResponse]{}, e
	}
	x := make([]dto.EventResponse, 0, len(v))
	for _, r := range v {
		x = append(x, mapEvent(r))
	}
	return page(x, f.Page, f.PageSize, n), nil
}
func (s *Service) ExcludeEvent(ctx context.Context, id, reason, actor string) (dto.EventResponse, error) {
	v, e := s.repositories.Event.Get(ctx, id)
	if e != nil {
		return dto.EventResponse{}, e
	}
	if v.StatusCode != "PENDING" {
		return dto.EventResponse{}, state("only PENDING events can be excluded")
	}
	target, e := s.transition(ctx, v.DocumentTypeID, v.StatusID, "EXCLUDED")
	if e != nil {
		return dto.EventResponse{}, e
	}
	notes := strings.TrimSpace(reason)
	if e = s.repositories.Event.SetStatus(ctx, id, target.ID, &notes); e != nil {
		return dto.EventResponse{}, e
	}
	return s.GetEvent(ctx, id)
}
