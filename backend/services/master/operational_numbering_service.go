package master

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
	repository "wms-api/repository/master"
)

var numberSeparator = regexp.MustCompile(`^[./_-]{0,3}$`)

func parseOperationalDate(value string) (time.Time, error) {
	date, err := time.Parse("2006-01-02", value)
	if err != nil || date.Year() < 1 {
		return time.Time{}, invalidCatalog("date must use YYYY-MM-DD")
	}
	return date, nil
}
func sequenceLimit(length int16) int64 {
	var limit int64 = 1
	for i := int16(0); i < length; i++ {
		limit *= 10
	}
	return limit - 1
}
func (s *OperationalService) ListDocumentNumberRule(ctx context.Context, typeID string, request dto.OperationalListRequest) (dto.PageResponse[dto.DocumentNumberRuleResponse], error) {
	filter, err := operationalFilter(request, typeID)
	if err != nil {
		return dto.PageResponse[dto.DocumentNumberRuleResponse]{}, err
	}
	if _, err := s.repositories.DocumentType.Get(ctx, typeID); err != nil {
		return dto.PageResponse[dto.DocumentNumberRuleResponse]{}, catalogError(err)
	}
	rows, total, err := s.repositories.DocumentNumberRule.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.DocumentNumberRuleResponse]{}, catalogError(err)
	}
	items := make([]dto.DocumentNumberRuleResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapDocumentNumberRule(row))
	}
	return pageResponse(items, request.Page, request.PageSize, total), nil
}
func (s *OperationalService) ReplaceDocumentNumberRule(ctx context.Context, typeID string, request dto.ReplaceDocumentNumberRuleRequest, actor string) (response dto.DocumentNumberRuleResponse, err error) {
	if validateID(typeID) != nil || validateID(actor) != nil || request.SequenceLength < 3 || request.SequenceLength > 18 || request.Separator == nil || request.IncludePartnerCode == nil || request.IncludeWarehouseCode == nil {
		return response, ErrInvalidInput
	}
	prefix, err := normalizeCode(request.Prefix)
	if err != nil || len(prefix) > 20 || !numberSeparator.MatchString(*request.Separator) {
		return response, invalidCatalog("invalid prefix or separator")
	}
	from, err := parseOperationalDate(request.EffectiveFrom)
	if err != nil {
		return response, err
	}
	today, _ := parseOperationalDate(time.Now().In(s.location).Format("2006-01-02"))
	if from.After(today) {
		return response, invalidCatalog("future scheduling is not supported; replacement activates immediately")
	}
	err = s.transaction(ctx, func(local *OperationalService) error {
		kind, err := local.repositories.DocumentType.Lock(ctx, typeID)
		if err != nil {
			return err
		}
		if !kind.IsActive {
			return invalidCatalog("document type is inactive")
		}
		history, _, err := local.repositories.DocumentNumberRule.List(ctx, repository.OperationalFilter{ParentID: typeID})
		if err != nil {
			return err
		}
		if len(history) > 0 && request.EffectiveFrom < operationalDate(history[0].EffectiveFrom) {
			return invalidCatalog("effective_from cannot precede the latest numbering rule")
		}
		if err := local.repositories.DocumentNumberRule.Retire(ctx, typeID, from.AddDate(0, 0, -1)); err != nil {
			return err
		}
		value := model.DocumentNumberRule{DocumentTypeID: typeID, Prefix: prefix, Separator: *request.Separator, SequenceLength: request.SequenceLength,
			IncludePartnerCode: *request.IncludePartnerCode, IncludeWarehouseCode: *request.IncludeWarehouseCode, EffectiveFrom: from, IsActive: true, CreatedBy: &actor}
		if err := local.repositories.DocumentNumberRule.Create(ctx, &value); err != nil {
			return err
		}
		response = mapDocumentNumberRule(value)
		return nil
	})
	return response, err
}
func (s *OperationalService) DeactivateDocumentNumberRule(ctx context.Context, typeID, id string) (response dto.DocumentNumberRuleResponse, err error) {
	if validateID(typeID) != nil || validateID(id) != nil {
		return response, ErrInvalidInput
	}
	err = s.transaction(ctx, func(local *OperationalService) error {
		if _, err := local.repositories.DocumentType.Lock(ctx, typeID); err != nil {
			return err
		}
		value, err := local.repositories.DocumentNumberRule.Get(ctx, id)
		if err != nil {
			return err
		}
		if !strings.EqualFold(value.DocumentTypeID, typeID) {
			return ErrNotFound
		}
		updated, err := local.repositories.DocumentNumberRule.Update(ctx, id, map[string]interface{}{"is_active": false}, "", nil)
		response = mapDocumentNumberRule(updated)
		return err
	})
	return response, err
}

// GenerateDocumentID allocates a number, not a document. Transaction modules can
// construct this service from transaction-bound repositories to commit both together.
func (s *OperationalService) GenerateDocumentID(ctx context.Context, typeID string, request dto.GenerateDocumentIDRequest) (response dto.GeneratedDocumentIDResponse, err error) {
	if validateID(typeID) != nil {
		return response, ErrInvalidInput
	}
	date, err := parseOperationalDate(request.BusinessDate)
	if err != nil {
		return response, err
	}
	for _, id := range []*string{request.PartnerID, request.WarehouseID} {
		if id != nil && validateID(*id) != nil {
			return response, ErrInvalidInput
		}
	}
	err = s.transaction(ctx, func(local *OperationalService) error {
		kind, err := local.repositories.DocumentType.Lock(ctx, typeID)
		if err != nil {
			return err
		}
		if !kind.IsActive {
			return invalidCatalog("document type is inactive")
		}
		active := true
		rules, _, err := local.repositories.DocumentNumberRule.List(ctx, repository.OperationalFilter{ParentID: typeID, Active: &active})
		if err != nil {
			return err
		}
		if len(rules) != 1 {
			return invalidCatalog("no active numbering rule")
		}
		rule := rules[0]
		if request.BusinessDate < operationalDate(rule.EffectiveFrom) || (rule.EffectiveUntil != nil && request.BusinessDate > operationalDate(*rule.EffectiveUntil)) {
			return invalidCatalog("business_date is outside the active numbering rule period")
		}
		if rule.IncludePartnerCode && request.PartnerID == nil {
			return invalidCatalog("partner_id is required by numbering rule")
		}
		if rule.IncludeWarehouseCode && request.WarehouseID == nil {
			return invalidCatalog("warehouse_id is required by numbering rule")
		}
		parts := []string{rule.Prefix}
		var ownerID *string
		if request.PartnerID != nil {
			partner, err := local.repositories.Catalog.BusinessPartner.Get(ctx, *request.PartnerID)
			if err != nil {
				return err
			}
			if !partner.IsActive {
				return invalidCatalog("partner is inactive")
			}
			ownerID = &partner.OwnerID
			if rule.IncludePartnerCode {
				parts = append(parts, strings.ToUpper(strings.TrimSpace(partner.Code)))
			}
		}
		if err := local.scope(ctx, ownerID, request.WarehouseID); err != nil {
			return err
		}
		if request.WarehouseID != nil {
			warehouse, err := local.repositories.Warehouse.FindByID(ctx, *request.WarehouseID)
			if err != nil {
				return err
			}
			if rule.IncludeWarehouseCode {
				parts = append(parts, strings.ToUpper(strings.TrimSpace(warehouse.Code)))
			}
		}
		parts = append(parts, date.Format("20060102"))
		if len(strings.Join(append(append([]string{}, parts...), strings.Repeat("0", int(rule.SequenceLength))), rule.Separator)) > 120 {
			return invalidCatalog("configured document ID exceeds 120 characters")
		}
		next, err := local.repositories.DocumentDailyCounter.Next(ctx, typeID, date, sequenceLimit(rule.SequenceLength))
		if err != nil {
			return err
		}
		number := strings.Join(append(parts, fmt.Sprintf("%0*d", int(rule.SequenceLength), next)), rule.Separator)
		response = dto.GeneratedDocumentIDResponse{DocumentID: number, DocumentTypeID: kind.ID, DocumentNumberRuleID: rule.ID, BusinessDate: request.BusinessDate, SequenceNumber: strconv.FormatInt(next, 10)}
		return nil
	})
	return response, operationalError(err)
}
func (s *OperationalService) ListDocumentDailyCounter(ctx context.Context, typeID, fromText, toText string, request dto.OperationalListRequest) (dto.PageResponse[dto.DocumentDailyCounterResponse], error) {
	if _, err := operationalFilter(request, typeID); err != nil {
		return dto.PageResponse[dto.DocumentDailyCounterResponse]{}, err
	}
	from, err := parseOperationalDate(fromText)
	if err != nil {
		return dto.PageResponse[dto.DocumentDailyCounterResponse]{}, err
	}
	to, err := parseOperationalDate(toText)
	if err != nil {
		return dto.PageResponse[dto.DocumentDailyCounterResponse]{}, err
	}
	if to.Before(from) {
		return dto.PageResponse[dto.DocumentDailyCounterResponse]{}, invalidCatalog("date_to must be on or after date_from")
	}
	if _, err := s.repositories.DocumentType.Get(ctx, typeID); err != nil {
		return dto.PageResponse[dto.DocumentDailyCounterResponse]{}, catalogError(err)
	}
	rows, total, err := s.repositories.DocumentDailyCounter.List(ctx, typeID, from, to, request.Page, request.PageSize)
	if err != nil {
		return dto.PageResponse[dto.DocumentDailyCounterResponse]{}, catalogError(err)
	}
	items := make([]dto.DocumentDailyCounterResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapDocumentDailyCounter(row))
	}
	return pageResponse(items, request.Page, request.PageSize, total), nil
}
