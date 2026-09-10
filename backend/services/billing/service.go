package billing

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

	dto "wms-api/dto/billing"
	masterdto "wms-api/dto/master"
	mastermodel "wms-api/models/master"
	repository "wms-api/repository/billing"
	masterrepository "wms-api/repository/master"
	masterservice "wms-api/services/master"
)

var (
	ErrInvalidInput = errors.New("invalid billing request")
	ErrInvalidState = errors.New("invalid billing workflow state")
	ErrForbidden    = errors.New("billing owner or warehouse access denied")
)
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]*$`)
var amountPattern = regexp.MustCompile(`^(0|[1-9][0-9]{0,13})(\.[0-9]{1,6})?$`)
var codePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,39}$`)

type Service struct {
	repositories *repository.Repositories
	timezone     string
}

func NewService(r *repository.Repositories, timezone string) (*Service, error) {
	if _, e := time.LoadLocation(timezone); e != nil {
		return nil, e
	}
	return &Service{repositories: r, timezone: timezone}, nil
}
func (s *Service) transaction(ctx context.Context, work func(*Service) error) error {
	return repository.Error(s.repositories.Transaction(ctx, func(r *repository.Repositories) error { return work(&Service{repositories: r, timezone: s.timezone}) }))
}
func (s *Service) CanAccess(ctx context.Context, account, owner, warehouse string) (bool, error) {
	if !uuidPattern.MatchString(account) || !uuidPattern.MatchString(owner) || !uuidPattern.MatchString(warehouse) {
		return false, invalid("owner_id and warehouse_id are required UUIDs")
	}
	return s.repositories.Scope.Allowed(ctx, account, owner, warehouse)
}
func (s *Service) ResourceScope(ctx context.Context, kind, id string) (string, string, error) {
	if !validID(id, 190) {
		return "", "", invalid("invalid billing resource id")
	}
	return s.repositories.Scope.ResourceScope(ctx, kind, id)
}
func invalid(v string) error         { return fmt.Errorf("%w: %s", ErrInvalidInput, v) }
func state(v string) error           { return fmt.Errorf("%w: %s", ErrInvalidState, v) }
func validID(v string, max int) bool { return len(v) <= max && idPattern.MatchString(v) }
func date(v, label string) (time.Time, error) {
	x, e := time.Parse("2006-01-02", strings.TrimSpace(v))
	if e != nil {
		return time.Time{}, invalid(label + " must be YYYY-MM-DD")
	}
	return x, nil
}
func optionalDate(v *string, label string) (*time.Time, error) {
	if v == nil {
		return nil, nil
	}
	x, e := date(*v, label)
	if e != nil {
		return nil, e
	}
	return &x, nil
}
func amount(v, label string, zero bool) (string, *big.Rat, error) {
	v = strings.TrimSpace(v)
	if !amountPattern.MatchString(v) {
		return "", nil, invalid(label + " must be numeric(20,6)")
	}
	n, ok := new(big.Rat).SetString(v)
	if !ok || n.Sign() < 0 || (!zero && n.Sign() == 0) {
		return "", nil, invalid(label + " must be greater than zero")
	}
	return n.FloatString(6), n, nil
}
func percentage(v, label string) (string, *big.Rat, error) {
	if strings.TrimSpace(v) == "" {
		v = "0"
	}
	x, n, e := amount(v, label, true)
	if e != nil {
		return "", nil, e
	}
	if n.Cmp(big.NewRat(100, 1)) > 0 {
		return "", nil, invalid(label + " cannot exceed 100")
	}
	return x, n, nil
}
func randomID(prefix string) (string, error) {
	var b [16]byte
	if _, e := rand.Read(b[:]); e != nil {
		return "", e
	}
	return prefix + "-" + hex.EncodeToString(b[:]), nil
}
func (s *Service) documentType(ctx context.Context, code string) (mastermodel.DocumentType, error) {
	v, e := s.repositories.Master.DocumentType.ByCode(ctx, code)
	if e != nil {
		return v, repository.Error(e)
	}
	if !v.IsActive {
		return v, state(code + " document type is inactive")
	}
	return v, nil
}
func (s *Service) status(ctx context.Context, typeID, code string) (mastermodel.DocumentStatus, error) {
	active := true
	rows, _, e := s.repositories.Master.DocumentStatus.List(ctx, masterrepository.OperationalFilter{ParentID: typeID, Active: &active})
	if e != nil {
		return mastermodel.DocumentStatus{}, repository.Error(e)
	}
	for _, v := range rows {
		if v.Code == code {
			return v, nil
		}
	}
	return mastermodel.DocumentStatus{}, state("status " + code + " is not configured")
}
func (s *Service) initialStatus(ctx context.Context, typeID string) (mastermodel.DocumentStatus, error) {
	active := true
	rows, _, e := s.repositories.Master.DocumentStatus.List(ctx, masterrepository.OperationalFilter{ParentID: typeID, Active: &active})
	if e != nil {
		return mastermodel.DocumentStatus{}, repository.Error(e)
	}
	for _, v := range rows {
		if v.IsInitial {
			return v, nil
		}
	}
	return mastermodel.DocumentStatus{}, state("initial status is not configured")
}
func (s *Service) transition(ctx context.Context, typeID, fromID, toCode string) (mastermodel.DocumentStatus, error) {
	to, e := s.status(ctx, typeID, toCode)
	if e != nil {
		return to, e
	}
	active := true
	rows, _, e := s.repositories.Master.DocumentStatusTransition.List(ctx, masterrepository.OperationalFilter{ParentID: typeID, Active: &active})
	if e != nil {
		return to, repository.Error(e)
	}
	for _, v := range rows {
		if v.FromStatusID == fromID && v.ToStatusID == to.ID {
			return to, nil
		}
	}
	return to, state("workflow transition to " + toCode + " is not configured")
}
func (s *Service) number(ctx context.Context, code string, businessDate time.Time) (string, error) {
	kind, e := s.documentType(ctx, code)
	if e != nil {
		return "", e
	}
	svc, e := masterservice.NewOperationalService(s.repositories.Master, s.timezone)
	if e != nil {
		return "", e
	}
	v, e := svc.GenerateDocumentID(ctx, kind.ID, masterdto.GenerateDocumentIDRequest{BusinessDate: businessDate.Format("2006-01-02")})
	if e != nil {
		return "", e
	}
	return v.DocumentID, nil
}
func page[T any](items []T, p, size int, total int64) dto.PageResponse[T] {
	pages := int64(0)
	if size > 0 {
		pages = (total + int64(size) - 1) / int64(size)
	}
	return dto.PageResponse[T]{Items: items, Page: p, PageSize: size, TotalItems: total, TotalPages: pages}
}
