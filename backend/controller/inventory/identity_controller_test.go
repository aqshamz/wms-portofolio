package inventory

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
	dto "wms-api/dto/inventory"
)

func TestStrictJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, body := range []string{
		"{}", "null", "{", `{"owner_id":"not-uuid","item_id":"not-uuid","lot_number":"A"}`,
		`{"owner_id":"11111111-1111-4111-8111-111111111111","item_id":"11111111-1111-4111-8111-111111111111","lot_number":"A","lot_id":"injected"}`,
		`{"owner_id":"11111111-1111-4111-8111-111111111111","item_id":"11111111-1111-4111-8111-111111111111","lot_number":"A"} {}`,
		strings.Repeat(" ", (1<<20)+1),
	} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", strings.NewReader(body))
		var request dto.CreateLotRequest
		if bind(c, &request) || w.Code != 400 {
			t.Fatalf("accepted malformed request; status=%d", w.Code)
		}
	}
}
func TestFilterRejectsIgnoredParameters(t *testing.T) {
	for _, query := range []string{"?active=true", "?owner_id=a&owner_id=b", "?warehouse_id=anything", "?page=0", "?page_size=101"} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/"+query, nil)
		if _, ok := filter(c, "lot_number", false); ok {
			t.Fatalf("accepted %s", query)
		}
	}
	for _, query := range []string{"?closed=", "?closed=1", "?closed=wrong"} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/"+query, nil)
		if _, ok := filter(c, "barcode", true); ok {
			t.Fatalf("accepted %s", query)
		}
	}
}
