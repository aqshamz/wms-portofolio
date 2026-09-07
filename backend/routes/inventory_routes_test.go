package routes

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
	controller "wms-api/controller/inventory"
	"wms-api/middleware"
)

func TestAllInventoryRoutesRequireSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerInventoryRoutes(router.Group("/api/v1"), controller.NewController(nil), middleware.NewAuthentication(nil).RequireSession())
	if len(router.Routes()) != 16 {
		t.Fatalf("expected 16 routes, got %d", len(router.Routes()))
	}
	for _, route := range router.Routes() {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(route.Method, strings.ReplaceAll(route.Path, ":id", "LOT-TEST"), nil))
		if w.Code != 401 {
			t.Fatalf("%s %s = %d", route.Method, route.Path, w.Code)
		}
	}
}
