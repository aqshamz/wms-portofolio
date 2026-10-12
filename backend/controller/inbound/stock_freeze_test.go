package inbound

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	inventoryrepo "wms-api/repository/inventory"
	"wms-api/utils"
)

func TestStockFreezeConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	err := fmt.Errorf("%w: finish or cancel grand stock opname GSO-TEST before moving this customer's stock", inventoryrepo.ErrStockFrozen)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	fail(c, err)

	var response utils.APIResponse
	if decodeErr := json.Unmarshal(w.Body.Bytes(), &response); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	if w.Code != http.StatusConflict || response.Success || response.Message != err.Error() {
		t.Fatalf("expected actionable stock-freeze conflict, got status=%d body=%s", w.Code, w.Body.String())
	}
	if len(c.Errors) != 0 {
		t.Fatalf("stock freeze must not be logged as an internal error: %v", c.Errors)
	}
}

