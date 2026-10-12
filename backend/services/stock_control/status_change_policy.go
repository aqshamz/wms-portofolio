package stockcontrol

// Workflow-owned statuses must only change through their owning transactions.
func manualInventoryStatus(code string) bool {
	switch code {
	case "AVAILABLE", "HOLD", "DAMAGED", "EXPIRED":
		return true
	default:
		return false
	}
}

func validateManualStatusChange(source, target string) error {
	if !manualInventoryStatus(source) || !manualInventoryStatus(target) {
		return invalid("manual status changes are limited to AVAILABLE, HOLD, DAMAGED and EXPIRED; use the owning workflow for other statuses")
	}
	if source == target {
		return invalid("target inventory status must differ from the source status")
	}
	return nil
}
