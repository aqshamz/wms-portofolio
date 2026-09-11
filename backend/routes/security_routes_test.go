package routes

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	controller "wms-api/controller/authentication"
	"wms-api/middleware"
)

func TestAllSecurityRoutesRequireSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerSecurityRoutes(router.Group("/api/v1"), controller.NewSecurityController(nil, nil, nil), middleware.NewAuthentication(nil).RequireModule("SECURITY"))
	if len(router.Routes()) != 28 {
		t.Fatalf("expected 28 security routes, got %d", len(router.Routes()))
	}
	for _, route := range router.Routes() {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(route.Method, route.Path, nil))
		if response.Code != 401 {
			t.Fatalf("%s %s=%d", route.Method, route.Path, response.Code)
		}
	}
}
