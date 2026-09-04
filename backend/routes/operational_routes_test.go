package routes

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	controller "wms-api/controller/master"
	"wms-api/middleware"
)

func TestOperationalRoutesRequireSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerOperationalRoutes(router.Group("/api/v1"), controller.NewOperationalController(nil), middleware.NewAuthentication(nil).RequireSession())
	registered := router.Routes()
	if len(registered) != 72 {
		t.Fatalf("registered %d routes, expected 72", len(registered))
	}
	for _, route := range registered {
		t.Run(route.Method+" "+route.Path, func(t *testing.T) {
			path := strings.NewReplacer(":child_id", "123e4567-e89b-42d3-a456-426614174000", ":id", "123e4567-e89b-42d3-a456-426614174000").Replace(route.Path)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(route.Method, path, nil))
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("got status %d", recorder.Code)
			}
		})
	}
}
