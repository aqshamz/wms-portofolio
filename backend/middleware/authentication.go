package middleware

import (
	"net/http"
	"strings"

	authdto "wms-api/dto/authentication"
	service "wms-api/services/authentication"
	"wms-api/utils"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserKey  = "authenticated_user"
	ContextTokenKey = "session_token"
)

type Authentication struct {
	service               *service.Service
	authorizationEnforced bool
}

func NewAuthentication(service *service.Service) *Authentication {
	return &Authentication{service: service}
}

func NewAuthenticationWithAuthorization(service *service.Service, enforced bool) *Authentication {
	return &Authentication{service: service, authorizationEnforced: enforced}
}

func (m *Authentication) RequireSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.authenticate(c) {
			return
		}
		c.Next()
	}
}

func (m *Authentication) RequireModule(module string) gin.HandlerFunc {
	module = strings.ToUpper(strings.TrimSpace(module))
	return func(c *gin.Context) {
		if !m.authenticate(c) {
			return
		}
		if m.authorizationEnforced {
			value, _ := c.Get(ContextUserKey)
			user, _ := value.(authdto.UserResponse)
			required, allowed := hasModulePermission(user.Permissions, module, c.Request.Method)
			if !allowed {
				utils.Failure(c, http.StatusForbidden, "permission required: "+required, nil)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func hasModulePermission(permissions []string, module, method string) (string, bool) {
	action := "WRITE"
	if method == http.MethodGet || method == http.MethodHead {
		action = "READ"
	}
	required := module + "." + action
	for _, code := range permissions {
		if code == required || code == "*" {
			return required, true
		}
	}
	return required, false
}

func (m *Authentication) authenticate(c *gin.Context) bool {
	token, ok := bearerToken(c.GetHeader("Authorization"))
	if !ok {
		utils.Failure(c, http.StatusUnauthorized, "authentication required", nil)
		c.Abort()
		return false
	}

	user, err := m.service.Authenticate(c.Request.Context(), token)
	if err != nil {
		utils.Failure(c, http.StatusUnauthorized, "invalid or expired session", nil)
		c.Abort()
		return false
	}

	c.Set(ContextUserKey, user)
	c.Set(ContextTokenKey, token)
	return true
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}
