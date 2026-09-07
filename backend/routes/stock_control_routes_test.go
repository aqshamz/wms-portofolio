package routes

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"
	controller "wms-api/controller/stock_control"
	"wms-api/middleware"
)

func TestAllStockControlRoutesRequireSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerStockControlRoutes(router.Group("/api/v1"), controller.NewController(nil), middleware.NewAuthentication(nil).RequireSession())
	if len(router.Routes()) != 6 {
		t.Fatalf("expected 6 routes got %d", len(router.Routes()))
	}
	for _, route := range router.Routes() {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(route.Method, route.Path, nil))
		if w.Code != 401 {
			t.Fatalf("%s %s=%d", route.Method, route.Path, w.Code)
		}
	}
}
