package stockcontrol

import (
	"testing"
	model "wms-api/models/stock_control"
)

func TestFoundStockCannotBeSelfReviewed(t *testing.T) {
	header := model.CycleCount{CreatedBy: "creator"}
	counter := "counter"
	line := model.CycleCountLine{FoundStock: true, CreatedBy: "recorder", CountedBy: &counter}
	for _, actor := range []string{"creator", "recorder", "counter"} {
		if !countLineSelfReview(header, line, actor) {
			t.Errorf("%s could review their own found stock", actor)
		}
	}
	if countLineSelfReview(header, line, "independent") {
		t.Fatal("independent reviewer was blocked")
	}
	line.FoundStock = false
	if countLineSelfReview(header, line, "counter") {
		t.Fatal("ordinary count review rules changed")
	}
}
