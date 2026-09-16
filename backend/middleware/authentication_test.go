package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	authdto "wms-api/dto/authentication"

	"github.com/gin-gonic/gin"
)

func TestRequirePermissionUsesExactActionPermission(t *testing.T) {
	tests := []struct {
		name        string
		permissions []string
		want        int
	}{
		{name: "exact action", permissions: []string{PermissionInboundQuarantineDispose}, want: http.StatusNoContent},
		{name: "wildcard administrator", permissions: []string{"*"}, want: http.StatusNoContent},
		{name: "module write is not an approval bypass", permissions: []string{"INBOUND.WRITE"}, want: http.StatusForbidden},
		{name: "different action", permissions: []string{PermissionInboundReceive}, want: http.StatusForbidden},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			auth := NewAuthenticationWithAuthorization(nil, true)
			router.Use(func(c *gin.Context) {
				c.Set(ContextUserKey, authdto.UserResponse{Permissions: test.permissions})
				c.Next()
			})
			router.POST("/quarantine", auth.RequirePermission(PermissionInboundQuarantineDispose), func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/quarantine", nil))
			if response.Code != test.want {
				t.Fatalf("status=%d want=%d body=%s", response.Code, test.want, response.Body.String())
			}
		})
	}
}

func TestRequireAnyPermissionAcceptsOneListedPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	auth := NewAuthenticationWithAuthorization(nil, true)
	router.Use(func(c *gin.Context) {
		c.Set(ContextUserKey, authdto.UserResponse{Permissions: []string{PermissionBillingIssue}})
		c.Next()
	})
	router.POST("/invoice", auth.RequireAnyPermission(PermissionBillingApprove, PermissionBillingIssue), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/invoice", nil))
	if response.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
