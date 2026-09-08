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
