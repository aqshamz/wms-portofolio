package inbound

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"
)

func TestPutawayLookupQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, query := range []string{"?owner_id=other", "?search=a&search=b", "?page=0", "?page_size=101"} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/"+query, nil)
		if _, _, _, ok := putawayLookupQuery(c); ok || w.Code != 400 {
			t.Fatalf("accepted %s", query)
		}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/?search=bulk&page=2&page_size=10", nil)
	search, page, size, ok := putawayLookupQuery(c)
	if !ok || search != "bulk" || page != 2 || size != 10 {
		t.Fatal("valid lookup query failed")
	}
}
