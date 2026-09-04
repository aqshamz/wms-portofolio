package master

import (
	"errors"
	"regexp"
	"strings"
	"time"

	dto "wms-api/dto/master"
)

var (
	ErrNotFound         = errors.New("resource not found")
	ErrConflict         = errors.New("business code already exists")
	ErrConcurrentUpdate = errors.New("resource changed; reload it before updating")
	ErrInvalidInput     = errors.New("invalid master data")
)

var (
	businessCodePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]*$`)
	uuidPattern         = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
)

func normalizeCode(code string) (string, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if !businessCodePattern.MatchString(code) {
		return "", ErrInvalidInput
	}
	return code, nil
}

func validateID(id string) error {
	if !uuidPattern.MatchString(id) {
		return ErrInvalidInput
	}
	return nil
}

func validateTimezone(timezone string) error {
	_, err := time.LoadLocation(timezone)
	if err != nil {
		return ErrInvalidInput
	}
	return nil
}

func normalizeCountry(country *string) *string {
	if country == nil {
		return nil
	}
	value := strings.ToUpper(strings.TrimSpace(*country))
	return &value
}

func pageResponse[T any](items []T, page, pageSize int, total int64) dto.PageResponse[T] {
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return dto.PageResponse[T]{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}
}
