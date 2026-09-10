package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"wms-api/config"
	authdto "wms-api/dto/authentication"
	auditmodel "wms-api/models/audit"
	"wms-api/utils"
)

func TestHardeningRequestIDAndAllowedCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hardening := NewHardening(config.SecurityConfig{
		AllowedOrigins: []string{"http://localhost:5173"}, RateLimitPerMinute: 10,
		MaxRequestBodyBytes: 64,
	}, nil)
	router := gin.New()
	router.Use(hardening.RequestID(), hardening.CORS(), hardening.RateLimit())
	router.GET("/test", func(c *gin.Context) { utils.Success(c, http.StatusOK, "ok", nil) })

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("X-Request-ID", "frontend-request-123")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK || response.Header().Get("X-Request-ID") != "frontend-request-123" {
		t.Fatalf("unexpected response status=%d request_id=%q", response.Code, response.Header().Get("X-Request-ID"))
	}
	if response.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("missing allowed CORS origin: %q", response.Header().Get("Access-Control-Allow-Origin"))
	}
	if !strings.Contains(response.Body.String(), `"request_id":"frontend-request-123"`) {
		t.Fatalf("request ID missing from response body: %s", response.Body.String())
	}
}

func TestHardeningRejectsOriginLargeBodyAndRateOverflow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hardening := NewHardening(config.SecurityConfig{
		AllowedOrigins: []string{"https://wms.example"}, RateLimitPerMinute: 1,
		MaxRequestBodyBytes: 4,
	}, nil)
	router := gin.New()
	router.Use(hardening.RequestID(), hardening.CORS(), hardening.BodyLimit(), hardening.RateLimit())
	router.POST("/test", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	forbidden := httptest.NewRequest(http.MethodPost, "/test", nil)
	forbidden.Header.Set("Origin", "https://evil.example")
	assertStatus(t, router, forbidden, http.StatusForbidden)

	large := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("12345"))
	assertStatus(t, router, large, http.StatusRequestEntityTooLarge)

	first := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("1"))
	first.RemoteAddr = "192.0.2.1:1234"
	assertStatus(t, router, first, http.StatusNoContent)
	second := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("1"))
	second.RemoteAddr = "192.0.2.1:1235"
	assertStatus(t, router, second, http.StatusTooManyRequests)
}

func TestModulePermissionDecision(t *testing.T) {
	if required, ok := hasModulePermission([]string{"INBOUND.READ"}, "INBOUND", http.MethodGet); !ok || required != "INBOUND.READ" {
		t.Fatalf("expected inbound read permission, got required=%q allowed=%v", required, ok)
	}
	if required, ok := hasModulePermission([]string{"INBOUND.READ"}, "INBOUND", http.MethodPost); ok || required != "INBOUND.WRITE" {
		t.Fatalf("write should be denied, got required=%q allowed=%v", required, ok)
	}
	if _, ok := hasModulePermission([]string{"*"}, "BILLING", http.MethodDelete); !ok {
		t.Fatal("wildcard permission should allow the operation")
	}
}

func TestHardeningAuditsMutationsWithoutRequestBodies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	entries := make([]auditmodel.APIAuditLog, 0, 1)
	hardening := NewHardening(config.SecurityConfig{}, func(_ context.Context, entry *auditmodel.APIAuditLog) error {
		entries = append(entries, *entry)
		return nil
	})
	router := gin.New()
	router.Use(hardening.RequestID(), hardening.Audit(), hardening.SecurityHeaders())
	router.POST("/objects/:id", func(c *gin.Context) {
		c.Set(ContextUserKey, authdto.UserResponse{AccountID: "d9428888-122b-11e1-b85c-61cd3cbb3210"})
		c.Status(http.StatusCreated)
	})

	request := httptest.NewRequest(http.MethodPost, "/objects/secret", strings.NewReader(`{"password":"must-not-be-audited"}`))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated || len(entries) != 1 {
		t.Fatalf("status=%d audit entries=%d", response.Code, len(entries))
	}
	entry := entries[0]
	if entry.Route != "/objects/:id" || entry.AccountID == nil || *entry.AccountID == "" || entry.RequestID == "" {
		t.Fatalf("unexpected audit entry: %#v", entry)
	}
	if response.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("security headers were not applied")
	}
}

func assertStatus(t *testing.T, router http.Handler, request *http.Request, want int) {
	t.Helper()
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != want {
		t.Fatalf("status=%d want=%d body=%s", response.Code, want, response.Body.String())
	}
}
