package stockcontrol

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
	dto "wms-api/dto/stock_control"
)

func TestStrictCommandJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	valid := `{"operation_key":"move:1","business_date":"2026-09-07","source_document_id":"DOC-1","source_balance_id":"BAL-1","target_location_id":"11111111-1111-4111-8111-111111111111","quantity":"1","expected_version":1}`
	for _, body := range []string{"{}", `{"unexpected":true}`, valid + " {}", strings.Repeat(" ", (1<<20)+1)} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", strings.NewReader(body))
		var q dto.InternalMoveRequest
		if bind(c, &q) || w.Code != 400 {
			t.Fatalf("accepted invalid JSON status=%d", w.Code)
		}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", strings.NewReader(valid))
	var q dto.InternalMoveRequest
	if !bind(c, &q) {
		t.Fatalf("rejected valid JSON: %s", w.Body.String())
	}
}
