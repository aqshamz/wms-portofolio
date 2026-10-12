package stockcontrol

import (
	"errors"
	"testing"
)

func TestManualInventoryStatusChange(t *testing.T) {
	for _, source := range []string{"AVAILABLE", "HOLD", "DAMAGED", "EXPIRED"} {
		for _, target := range []string{"AVAILABLE", "HOLD", "DAMAGED", "EXPIRED"} {
			err := validateManualStatusChange(source, target)
			if (source == target) != errors.Is(err, ErrInvalidInput) {
				t.Fatalf("%s -> %s: %v", source, target, err)
			}
		}
	}
	for _, workflow := range []string{"QC_PENDING", "PUTAWAY_PENDING", "QUARANTINE", "RETURN_PENDING", "PICKED", "PACKED", "SHIPPED", ""} {
		if !errors.Is(validateManualStatusChange(workflow, "AVAILABLE"), ErrInvalidInput) || !errors.Is(validateManualStatusChange("AVAILABLE", workflow), ErrInvalidInput) {
			t.Fatalf("workflow-owned status %s was allowed", workflow)
		}
	}
}
