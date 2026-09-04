package master

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	dto "wms-api/dto/master"
)

func TestOperationalBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name, body string
		request    interface{}
		valid      bool
	}{
		{"number rule", `{"prefix":"PO","separator":"","sequence_length":6,"include_partner_code":false,"include_warehouse_code":false,"effective_from":"2026-09-03"}`, &dto.ReplaceDocumentNumberRuleRequest{}, true},
		{"missing bool", `{"prefix":"PO","separator":"-","sequence_length":6,"include_warehouse_code":false,"effective_from":"2026-09-03"}`, &dto.ReplaceDocumentNumberRuleRequest{}, false},
		{"invalid length", `{"prefix":"PO","separator":"-","sequence_length":19,"include_partner_code":false,"include_warehouse_code":false,"effective_from":"2026-09-03"}`, &dto.ReplaceDocumentNumberRuleRequest{}, false},
		{"counter reset", `{"business_date":"2026-09-03","last_number":0}`, &dto.GenerateDocumentIDRequest{}, false},
		{"generation", `{"business_date":"2026-09-03"}`, &dto.GenerateDocumentIDRequest{}, true},
		{"immutable scope", `{"name":"Changed","owner_id":"123e4567-e89b-42d3-a456-426614174000","is_active":true}`, &dto.UpdatePickingStrategyRequest{}, false},
		{"zero sequence", `{"sequence_no":0}`, &dto.CreatePutawayStrategyRuleRequest{}, false},
		{"float percent", `{"sequence_no":1,"minimum_empty_percent":25.1234}`, &dto.CreatePutawayStrategyRuleRequest{}, false},
		{"string percent", `{"sequence_no":1,"minimum_empty_percent":"25.1234"}`, &dto.CreatePutawayStrategyRuleRequest{}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body))
			if got := bindCatalog(c, tc.request); got != tc.valid {
				t.Fatalf("valid=%v got=%v response=%s", tc.valid, got, recorder.Body)
			}
		})
	}
}
