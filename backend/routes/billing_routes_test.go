package routes

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	controller "wms-api/controller/billing"
	"wms-api/middleware"
)

func TestAllBillingRoutesRequireSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerBillingRoutes(router.Group("/api/v1"), controller.NewController(nil), middleware.NewAuthentication(nil).RequireSession())
	if len(router.Routes()) != 45 {
		t.Fatalf("expected 45 routes got %d", len(router.Routes()))
	}
	for _, route := range router.Routes() {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(route.Method, route.Path, nil))
		if response.Code != 401 {
			t.Fatalf("%s %s=%d", route.Method, route.Path, response.Code)
		}
	}
}
