package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDocumentationRoutesAndPathParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/inventory/lots/:id", func(c *gin.Context) {})
	registerDocumentationRoutes(router)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/openapi.json", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("openapi status=%d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, expected := range []string{`/api/v1/inventory/lots/{id}`, `"name":"id"`, `"bearerAuth"`} {
		if !strings.Contains(body, expected) {
			t.Fatalf("OpenAPI output missing %q: %s", expected, body)
		}
	}

	docs := httptest.NewRecorder()
	router.ServeHTTP(docs, httptest.NewRequest(http.MethodGet, "/docs", nil))
	if docs.Code != http.StatusOK || !strings.Contains(docs.Body.String(), "SwaggerUIBundle") {
		t.Fatalf("docs response invalid: status=%d", docs.Code)
	}
}
