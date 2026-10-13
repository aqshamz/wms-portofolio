package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
	authdto "wms-api/dto/authentication"
)

func TestCloseShortRequiresSeparateAuthority(t *testing.T) {
	for _, test := range []struct {
		permission string
		allowed    bool
	}{
		{PermissionOutboundCloseShort, true}, {"*", true},
		{PermissionOutboundPlan, false}, {PermissionOutboundCancel, false}, {"OUTBOUND.WRITE", false},
	} {
		t.Run(test.permission, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			auth := NewAuthenticationWithAuthorization(nil, true)
			router.Use(func(c *gin.Context) {
				c.Set(ContextUserKey, authdto.UserResponse{Permissions: []string{test.permission}})
				c.Next()
			})
			router.POST("/close-short", auth.RequirePermission(PermissionOutboundCloseShort), func(c *gin.Context) { c.Status(http.StatusNoContent) })
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/close-short", nil))
			want := http.StatusForbidden
			if test.allowed {
				want = http.StatusNoContent
			}
			if response.Code != want {
				t.Fatalf("status=%d want=%d", response.Code, want)
			}
		})
	}
}
