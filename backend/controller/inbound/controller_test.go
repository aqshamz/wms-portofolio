package inbound

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	dto "wms-api/dto/inbound"
	"wms-api/utils"
)

func TestStrictInboundJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	valid := `{"expected_version":1}`
	for _, body := range []string{"{}", `{"expected_version":1,"unexpected":true}`, valid + " {}", strings.Repeat(" ", (1<<20)+1)} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", strings.NewReader(body))
		var request dto.TransitionRequest
		if utils.BindStrictJSON(c, &request) || w.Code != 400 {
			t.Fatalf("accepted invalid JSON status=%d", w.Code)
		}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", strings.NewReader(valid))
	var request dto.TransitionRequest
	if !utils.BindStrictJSON(c, &request) {
		t.Fatalf("rejected valid JSON: %s", w.Body.String())
	}
}

func TestReceiptInspectionEligibilityFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		query               string
		allow, ok, eligible bool
	}{
		{"", true, true, false},
		{"?inspection_eligible=true", true, true, true},
		{"?inspection_eligible=false", true, true, false},
		{"?inspection_eligible=", true, false, false},
		{"?inspection_eligible=1", true, false, false},
		{"?inspection_eligible=true&inspection_eligible=false", true, false, false},
		{"?inspection_eligible=true&unknown=true", true, false, false},
		{"?inspection_eligible=true", false, false, false},
	} {
		t.Run(test.query+"_allow_"+strings.ToUpper(map[bool]string{true: "yes", false: "no"}[test.allow]), func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/"+test.query, nil)
			filter, ok := listFilter(c, test.allow)
			if ok != test.ok || filter.InspectionEligible != test.eligible || (!ok && w.Code != 400) {
				t.Fatalf("ok=%t eligible=%t status=%d", ok, filter.InspectionEligible, w.Code)
			}
		})
	}
}
