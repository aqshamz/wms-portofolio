package inbound

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestQuarantineTargetLookupRejectsInvalidParameters(t *testing.T) {
	service := &Service{}
	for _, test := range []struct {
		name, id, search string
		page, size       int
	}{
		{"missing case", "", "", 1, 20},
		{"long case", strings.Repeat("a", 141), "", 1, 20},
		{"long search", "QCASE-1", strings.Repeat("a", 161), 1, 20},
		{"zero page", "QCASE-1", "", 0, 20},
		{"large page", "QCASE-1", "", 1000001, 20},
		{"zero size", "QCASE-1", "", 1, 0},
		{"large size", "QCASE-1", "", 1, 101},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.ListQuarantineTargets(context.Background(), test.id, test.search, test.page, test.size)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("invalid lookup was accepted: %v", err)
			}
		})
	}
}
