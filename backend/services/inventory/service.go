package inventory

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
	dto "wms-api/dto/inventory"
	model "wms-api/models/inventory"
	master "wms-api/models/master"
	repository "wms-api/repository/inventory"
)

var ErrInvalidInput = errors.New("invalid inventory identity")

type Service struct{ repositories *repository.Repositories }

func NewService(r *repository.Repositories) *Service { return &Service{repositories: r} }
func invalid(message string) error                   { return fmt.Errorf("%w: %s", ErrInvalidInput, message) }

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
var identityPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
var businessCodePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]*$`)

func uuid(value string) bool { return uuidPattern.MatchString(value) }
func identityID(value string, max int) bool {
	return len(value) <= max && identityPattern.MatchString(value)
}
func normalizeNumber(value string, max int) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || !utf8.ValidString(value) || utf8.RuneCountInString(value) > max {
		return "", invalid("number/barcode is empty or too long")
	}
	for _, c := range value {
		if unicode.IsControl(c) {
			return "", invalid("number/barcode contains control characters")
		}
	}
	return value, nil
}
func newID(prefix string) (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return prefix + "-" + hex.EncodeToString(b[:]), nil
}
func date(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", *value)
	if err != nil || parsed.Year() < 1 {
		return nil, invalid("dates must be valid YYYY-MM-DD values")
	}
	return &parsed, nil
}
func reference(err error, label string) error {
	if errors.Is(repository.Error(err), repository.ErrNotFound) {
		return invalid(label + " does not exist")
	}
	return repository.Error(err)
}
func requireItem(ctx context.Context, r *repository.Repositories, ownerID, itemID string) (master.Item, error) {
	owner, err := r.Owner.GetShared(ctx, ownerID)
	if err != nil {
		return master.Item{}, reference(err, "owner")
	}
	if !owner.IsActive {
		return master.Item{}, invalid("owner is inactive")
	}
	item, err := r.Catalog.Item.GetShared(ctx, itemID)
	if err != nil {
		return item, reference(err, "item")
	}
	if !item.IsActive {
		return item, invalid("item is inactive")
	}
	if item.OwnerID != ownerID {
		return item, invalid("item does not belong to owner")
	}
	return item, nil
}
func validateFilter(f *repository.Filter, hu bool) error {
	if f.Page < 1 || f.Page > 1000000 || f.PageSize < 1 || f.PageSize > 100 {
		return invalid("page must be 1..1000000 and page_size 1..100")
	}
	for _, field := range []*string{&f.OwnerID, &f.ItemID, &f.WarehouseID, &f.LocationID} {
		if *field != "" && !uuid(*field) {
			return invalid("filters must use UUIDs")
		}
		*field = strings.ToLower(*field)
	}
	if f.ParentID != "" && !identityID(f.ParentID, 120) {
		return invalid("invalid parent_handling_unit_id")
	}
	if hu && f.ItemID != "" {
		return invalid("handling units have no item_id filter")
	}
	if !hu && (f.WarehouseID != "" || f.LocationID != "" || f.ParentID != "" || f.Closed != nil) {
		return invalid("lots and serial identities have no warehouse/location/parent/closed filters")
	}
	if utf8.RuneCountInString(f.Search) > 160 || utf8.RuneCountInString(f.Number) > 120 {
		return invalid("search or exact number is too long")
	}
	f.Search = strings.TrimSpace(f.Search)
	f.Number = strings.TrimSpace(f.Number)
	return nil
}
func page[T any](items []T, f repository.Filter, total int64) dto.PageResponse[T] {
	return dto.PageResponse[T]{Items: items, Page: f.Page, PageSize: f.PageSize, TotalItems: total, TotalPages: (total + int64(f.PageSize) - 1) / int64(f.PageSize)}
}
func dateText(v *time.Time) *string {
	if v == nil {
		return nil
	}
	text := v.Format("2006-01-02")
	return &text
}
func mapLot(v model.InventoryLot) dto.LotResponse {
	return dto.LotResponse{ID: v.ID, OwnerID: v.OwnerID, ItemID: v.ItemID, LotNumber: v.LotNumber, ManufactureDate: dateText(v.ManufactureDate), ExpiryDate: dateText(v.ExpiryDate), QualityStatusID: v.QualityStatusID, CreatedAt: v.CreatedAt, CreatedBy: v.CreatedBy}
}
func mapSerial(v model.SerialNumber) dto.SerialResponse {
	return dto.SerialResponse{ID: v.ID, OwnerID: v.OwnerID, ItemID: v.ItemID, SerialNo: v.SerialNo, CreatedAt: v.CreatedAt, CreatedBy: v.CreatedBy}
}
func mapHandlingUnit(v model.HandlingUnit) dto.HandlingUnitResponse {
	return dto.HandlingUnitResponse{ID: v.ID, WarehouseID: v.WarehouseID, OwnerID: v.OwnerID, HandlingUnitTypeID: v.HandlingUnitTypeID, ParentHandlingUnitID: v.ParentHandlingUnitID, CurrentLocationID: v.CurrentLocationID, Barcode: v.Barcode, IsClosed: v.IsClosed, CreatedAt: v.CreatedAt, CreatedBy: v.CreatedBy}
}
