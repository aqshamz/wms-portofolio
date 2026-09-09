package outbound

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

	masterdto "wms-api/dto/master"
	dto "wms-api/dto/outbound"
	mastermodel "wms-api/models/master"
	masterrepository "wms-api/repository/master"
	repository "wms-api/repository/outbound"
	masterservice "wms-api/services/master"
)

var (
	ErrInvalidInput = errors.New("invalid outbound request")
	ErrInvalidState = errors.New("invalid outbound workflow state")
	ErrForbidden    = errors.New("outbound owner or warehouse access denied")
)
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]*$`)
var quantityPattern = regexp.MustCompile(`^(0|[1-9][0-9]{0,13})(\.[0-9]{1,6})?$`)

type Service struct {
	repositories *repository.Repositories
	timezone     string
}

func (s *Service) CanAccess(ctx context.Context, accountID, ownerID, warehouseID string) (bool, error) {
	if !validUUID(accountID) || !validUUID(ownerID) || !validUUID(warehouseID) {
		return false, invalid("owner_id and warehouse_id are required UUIDs")
	}
	return s.repositories.Scope.Allowed(ctx, accountID, ownerID, warehouseID)
}

func (s *Service) ResourceScope(ctx context.Context, kind, id string) (string, string, error) {
	if !validID(id, 190) {
		return "", "", invalid("invalid outbound resource id")
	}
	return s.repositories.Scope.ResourceScope(ctx, kind, id)
}

func NewService(repositories *repository.Repositories, timezone string) (*Service, error) {
	if _, err := time.LoadLocation(timezone); err != nil {
		return nil, err
	}
	return &Service{repositories: repositories, timezone: timezone}, nil
}
func invalid(m string) error       { return fmt.Errorf("%w: %s", ErrInvalidInput, m) }
func state(m string) error         { return fmt.Errorf("%w: %s", ErrInvalidState, m) }
func validUUID(v string) bool      { return uuidPattern.MatchString(v) }
func validID(v string, n int) bool { return len(v) <= n && idPattern.MatchString(v) }
func clean(v string, n int, label string) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" || !utf8.ValidString(v) || utf8.RuneCountInString(v) > n {
		return "", invalid(label + " is required or too long")
	}
	for _, r := range v {
		if unicode.IsControl(r) {
			return "", invalid(label + " contains control characters")
		}
	}
	return v, nil
}
func optional(v *string, n int, label string) (*string, error) {
	if v == nil {
		return nil, nil
	}
	x := strings.TrimSpace(*v)
	if x == "" {
		return nil, nil
	}
	y, err := clean(x, n, label)
	return &y, err
}
func date(v, label string) (time.Time, error) {
	x, err := time.Parse("2006-01-02", v)
	if err != nil {
		return time.Time{}, invalid(label + " must be YYYY-MM-DD")
	}
	return x, nil
}
func timestamp(v, label string) (time.Time, error) {
	x, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return time.Time{}, invalid(label + " must be RFC3339 with timezone")
	}
	return x, nil
}
func optionalTimestamp(v *string, label string) (*time.Time, error) {
	if v == nil {
		return nil, nil
	}
	x, err := timestamp(*v, label)
	if err != nil {
		return nil, err
	}
	return &x, nil
}
func quantity(v, label string, zero bool) (string, *big.Rat, error) {
	v = strings.TrimSpace(v)
	if !quantityPattern.MatchString(v) {
		return "", nil, invalid(label + " must be numeric(20,6)")
	}
	x, ok := new(big.Rat).SetString(v)
	if !ok || (!zero && x.Sign() <= 0) {
		return "", nil, invalid(label + " must be greater than zero")
	}
	return decimal(x), x, nil
}
func decimal(v *big.Rat) string { return v.FloatString(6) }
func randomID(prefix string) (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return prefix + "-" + hex.EncodeToString(b[:]), nil
}
func page[T any](items []T, p, size int, total int64) dto.PageResponse[T] {
	pages := int64(0)
	if total > 0 {
		pages = (total + int64(size) - 1) / int64(size)
	}
	return dto.PageResponse[T]{Items: items, Page: p, PageSize: size, TotalItems: total, TotalPages: pages}
}
func (s *Service) transaction(ctx context.Context, work func(*Service) error) error {
	return s.repositories.Transaction(ctx, func(r *repository.Repositories) error { return work(&Service{repositories: r, timezone: s.timezone}) })
}

// moveSerialIdentity keeps serialized stock's current-state pointer aligned with
// the balance movement. A serialized balance must move in full.
func (s *Service) moveSerialIdentity(ctx context.Context, sourceBalanceID string, targetBalanceID *string, qty *big.Rat) (*string, error) {
	ids, err := s.repositories.Inventory.SerialState.IDsByBalance(ctx, sourceBalanceID)
	if err != nil || len(ids) == 0 {
		return nil, err
	}
	if qty.Denom().Cmp(big.NewInt(1)) != 0 || qty.Num().Int64() != int64(len(ids)) {
		return nil, state("serialized stock must move as the complete balance identity")
	}
	if len(ids) != 1 {
		return nil, state("an inventory movement supports exactly one serial identity")
	}
	if targetBalanceID == nil {
		if err := s.repositories.Inventory.SerialState.DeleteBalance(ctx, sourceBalanceID); err != nil {
			return nil, err
		}
	} else if err := s.repositories.Inventory.SerialState.MoveBalance(ctx, sourceBalanceID, *targetBalanceID); err != nil {
		return nil, err
	}
	return &ids[0], nil
}

func (s *Service) generateID(ctx context.Context, code string, businessDate time.Time, partnerID, warehouseID *string) (string, error) {
	kind, err := s.documentType(ctx, code)
	if err != nil {
		return "", err
	}
	numbering, err := masterservice.NewOperationalService(s.repositories.Master, s.timezone)
	if err != nil {
		return "", err
	}
	result, err := numbering.GenerateDocumentID(ctx, kind.ID, masterdto.GenerateDocumentIDRequest{BusinessDate: businessDate.Format("2006-01-02"), PartnerID: partnerID, WarehouseID: warehouseID})
	if err != nil {
		return "", err
	}
	return result.DocumentID, nil
}

func (s *Service) documentType(ctx context.Context, code string) (mastermodel.DocumentType, error) {
	v, err := s.repositories.Master.DocumentType.ByCode(ctx, code)
	if err != nil || !v.IsActive {
		return v, state("document type " + code + " is not configured")
	}
	return v, nil
}

func (s *Service) document(ctx context.Context, code string) (mastermodel.DocumentType, mastermodel.DocumentStatus, error) {
	kind, err := s.documentType(ctx, code)
	if err != nil {
		return kind, mastermodel.DocumentStatus{}, err
	}
	initial, err := s.initialStatus(ctx, kind.ID)
	return kind, initial, err
}

func (s *Service) status(ctx context.Context, typeID, code string) (mastermodel.DocumentStatus, error) {
	active := true
	rows, _, err := s.repositories.Master.DocumentStatus.List(ctx, masterrepository.OperationalFilter{ParentID: typeID, Active: &active})
	if err != nil {
		return mastermodel.DocumentStatus{}, repository.Error(err)
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
	rows, _, err := s.repositories.Master.DocumentStatus.List(ctx, masterrepository.OperationalFilter{ParentID: typeID, Active: &active})
	if err != nil {
		return mastermodel.DocumentStatus{}, repository.Error(err)
	}
	for _, v := range rows {
		if v.IsInitial {
			return v, nil
		}
	}
	return mastermodel.DocumentStatus{}, state("initial status is not configured")
}
func (s *Service) transition(ctx context.Context, typeID, fromID, toCode string) (mastermodel.DocumentStatus, error) {
	target, err := s.status(ctx, typeID, toCode)
	if err != nil {
		return target, err
	}
	active := true
	rows, _, err := s.repositories.Master.DocumentStatusTransition.List(ctx, masterrepository.OperationalFilter{ParentID: typeID, Active: &active})
	if err != nil {
		return target, repository.Error(err)
	}
	for _, v := range rows {
		if v.FromStatusID == fromID && v.ToStatusID == target.ID {
			return target, nil
		}
	}
	return target, state("workflow transition to " + toCode + " is not configured")
}
func (s *Service) taskStatus(ctx context.Context, code string) (mastermodel.TaskStatus, error) {
	v, err := s.repositories.Master.TaskStatus.ByCode(ctx, code)
	if err != nil || !v.IsActive {
		return v, state("task status " + code + " is not configured")
	}
	return v, nil
}
func (s *Service) taskTransition(ctx context.Context, fromID, toCode string) (mastermodel.TaskStatus, error) {
	target, err := s.taskStatus(ctx, toCode)
	if err != nil {
		return target, err
	}
	active := true
	rows, _, err := s.repositories.Master.TaskStatusTransition.List(ctx, masterrepository.OperationalFilter{Active: &active})
	if err != nil {
		return target, repository.Error(err)
	}
	for _, v := range rows {
		if v.FromStatusID == fromID && v.ToStatusID == target.ID {
			return target, nil
		}
	}
	return target, state("task workflow transition to " + toCode + " is not configured")
}
func (s *Service) pickingStrategy(ctx context.Context, ownerID, warehouseID string, requested *string) (*mastermodel.PickingStrategy, error) {
	active := true
	rows, _, err := s.repositories.Master.PickingStrategy.List(ctx, masterrepository.OperationalFilter{Active: &active})
	if err != nil {
		return nil, repository.Error(err)
	}
	var best *mastermodel.PickingStrategy
	score := -1
	for i := range rows {
		v := rows[i]
		if requested != nil && v.ID != strings.ToLower(*requested) {
			continue
		}
		if v.OwnerID != nil && *v.OwnerID != ownerID || v.WarehouseID != nil && *v.WarehouseID != warehouseID {
			continue
		}
		n := 0
		if v.OwnerID != nil {
			n += 2
		}
		if v.WarehouseID != nil {
			n++
		}
		if n > score {
			copy := v
			best = &copy
			score = n
		}
	}
	if best == nil {
		return nil, state("no active picking strategy matches owner and warehouse")
	}
	return best, nil
}
