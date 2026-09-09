package inbound

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	dto "wms-api/dto/inbound"
	masterdto "wms-api/dto/master"
	mastermodel "wms-api/models/master"
	repository "wms-api/repository/inbound"
	masterrepository "wms-api/repository/master"
	masterservice "wms-api/services/master"
)

var (
	ErrInvalidInput = errors.New("invalid inbound request")
	ErrInvalidState = errors.New("invalid inbound document state")
	ErrForbidden    = errors.New("inbound owner or warehouse access denied")
)

var inboundUUIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
var inboundIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]*$`)
var inboundQuantityPattern = regexp.MustCompile(`^(0|[1-9][0-9]{0,13})(\.[0-9]{1,6})?$`)

type Service struct {
	repositories *repository.Repositories
	timezone     string
}

func (s *Service) CanAccess(ctx context.Context, accountID, ownerID, warehouseID string) (bool, error) {
	if !inboundUUID(accountID) || !inboundUUID(ownerID) || !inboundUUID(warehouseID) {
		return false, invalid("owner_id and warehouse_id are required UUIDs")
	}
	return s.repositories.Scope.Allowed(ctx, accountID, ownerID, warehouseID)
}

func (s *Service) ResourceScope(ctx context.Context, kind, id string) (string, string, error) {
	if !inboundID(id, 190) {
		return "", "", invalid("invalid inbound resource id")
	}
	return s.repositories.Scope.ResourceScope(ctx, kind, id)
}

func NewService(repositories *repository.Repositories, timezone string) (*Service, error) {
	if _, err := time.LoadLocation(timezone); err != nil {
		return nil, err
	}
	return &Service{repositories: repositories, timezone: timezone}, nil
}

func invalid(message string) error { return fmt.Errorf("%w: %s", ErrInvalidInput, message) }
func state(message string) error   { return fmt.Errorf("%w: %s", ErrInvalidState, message) }

func newInboundID(prefix string) (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return prefix + "-" + hex.EncodeToString(bytes[:]), nil
}

func (s *Service) transaction(ctx context.Context, work func(*Service) error) error {
	return repository.Error(s.repositories.Transaction(ctx, func(repositories *repository.Repositories) error {
		return work(&Service{repositories: repositories, timezone: s.timezone})
	}))
}

func inboundUUID(value string) bool { return inboundUUIDPattern.MatchString(value) }
func inboundID(value string, max int) bool {
	return len(value) <= max && inboundIDPattern.MatchString(value)
}
func clean(value string, max int, label string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || !utf8.ValidString(value) || utf8.RuneCountInString(value) > max {
		return "", invalid(label + " is required or too long")
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return "", invalid(label + " contains control characters")
		}
	}
	return value, nil
}
func optional(value *string, max int, label string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	cleaned := strings.TrimSpace(*value)
	if cleaned == "" {
		return nil, nil
	}
	result, err := clean(cleaned, max, label)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
func inboundDate(value, label string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, invalid(label + " must be YYYY-MM-DD")
	}
	return parsed, nil
}
func inboundOptionalDate(value *string, label string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := inboundDate(*value, label)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
func inboundTimestamp(value, label string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, invalid(label + " must be RFC3339 with timezone")
	}
	return parsed, nil
}
func inboundOptionalTimestamp(value *string, label string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := inboundTimestamp(*value, label)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
func inboundQuantity(value, label string, allowZero bool) (string, *big.Rat, error) {
	value = strings.TrimSpace(value)
	if !inboundQuantityPattern.MatchString(value) {
		return "", nil, invalid(label + " must be numeric(20,6)")
	}
	number, ok := new(big.Rat).SetString(value)
	if !ok || number.Sign() < 0 || (!allowZero && number.Sign() == 0) {
		return "", nil, invalid(label + " must be greater than zero")
	}
	return number.FloatString(6), number, nil
}

func inboundPercentage(value *string, label string) (string, *big.Rat, error) {
	if value == nil {
		return "0.0000", new(big.Rat), nil
	}
	_, number, err := inboundQuantity(*value, label, true)
	if err != nil {
		return "", nil, err
	}
	if number.Cmp(big.NewRat(100, 1)) > 0 {
		return "", nil, invalid(label + " cannot exceed 100")
	}
	return number.FloatString(4), number, nil
}
func inboundMultiply(quantity, conversion *big.Rat) (string, *big.Rat, error) {
	result := new(big.Rat).Mul(quantity, conversion)
	formatted := result.FloatString(6)
	parsed, exact, err := inboundQuantity(formatted, "base quantity", false)
	if err != nil || exact.Cmp(result) != 0 {
		return "", nil, invalid("converted base quantity exceeds numeric(20,6) precision")
	}
	return parsed, exact, nil
}

func (s *Service) documentType(ctx context.Context, code string) (mastermodel.DocumentType, error) {
	value, err := s.repositories.Master.DocumentType.ByCode(ctx, code)
	if err != nil {
		return value, repository.Error(err)
	}
	if !value.IsActive {
		return value, state(code + " document type is inactive")
	}
	return value, nil
}
func (s *Service) documentStatus(ctx context.Context, typeID, code string) (mastermodel.DocumentStatus, error) {
	active := true
	rows, _, err := s.repositories.Master.DocumentStatus.List(ctx, masterrepository.OperationalFilter{ParentID: typeID, Active: &active})
	if err != nil {
		return mastermodel.DocumentStatus{}, repository.Error(err)
	}
	for _, row := range rows {
		if row.Code == code {
			return row, nil
		}
	}
	return mastermodel.DocumentStatus{}, state("status " + code + " is not configured")
}
func (s *Service) initialStatus(ctx context.Context, typeID string) (mastermodel.DocumentStatus, error) {
	active := true
	rows, _, err := s.repositories.Master.DocumentStatus.List(ctx, masterrepository.OperationalFilter{ParentID: typeID, Active: &active})
	if err != nil {
		return mastermodel.DocumentStatus{}, repository.Error(err)
	}
	var result *mastermodel.DocumentStatus
	for index := range rows {
		if rows[index].IsInitial {
			if result != nil {
				return mastermodel.DocumentStatus{}, state("multiple initial statuses are configured")
			}
			result = &rows[index]
		}
	}
	if result == nil {
		return mastermodel.DocumentStatus{}, state("initial status is not configured")
	}
	return *result, nil
}
func (s *Service) transitionTarget(ctx context.Context, typeID, fromID, code string) (mastermodel.DocumentStatus, error) {
	target, err := s.documentStatus(ctx, typeID, code)
	if err != nil {
		return target, err
	}
	active := true
	rows, _, err := s.repositories.Master.DocumentStatusTransition.List(ctx, masterrepository.OperationalFilter{ParentID: typeID, Active: &active})
	if err != nil {
		return target, repository.Error(err)
	}
	for _, row := range rows {
		if row.FromStatusID == fromID && row.ToStatusID == target.ID {
			return target, nil
		}
	}
	return target, state("workflow transition to " + code + " is not configured")
}
func (s *Service) generateID(ctx context.Context, typeID, businessDate, partnerID, warehouseID string) (string, error) {
	service, err := masterservice.NewOperationalService(s.repositories.Master, s.timezone)
	if err != nil {
		return "", err
	}
	generated, err := service.GenerateDocumentID(ctx, typeID, masterdto.GenerateDocumentIDRequest{BusinessDate: businessDate, PartnerID: &partnerID, WarehouseID: &warehouseID})
	if err != nil {
		return "", invalid("cannot allocate document ID: " + err.Error())
	}
	return generated.DocumentID, nil
}

func page[T any](items []T, page, size int, total int64) dto.PageResponse[T] {
	return dto.PageResponse[T]{Items: items, Page: page, PageSize: size, TotalItems: total, TotalPages: (total + int64(size) - 1) / int64(size)}
}
