package routes

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	controller "wms-api/controller/inbound"
	"wms-api/middleware"
)

func TestAllInboundRoutesRequireSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerInboundRoutes(router.Group("/api/v1"), controller.NewController(nil), middleware.NewAuthentication(nil).RequireSession())
	if len(router.Routes()) != 41 {
		t.Fatalf("expected 41 routes got %d", len(router.Routes()))
	}
	for _, route := range router.Routes() {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(route.Method, route.Path, nil))
		if w.Code != 401 {
			t.Fatalf("%s %s=%d", route.Method, route.Path, w.Code)
		}
	}
}
