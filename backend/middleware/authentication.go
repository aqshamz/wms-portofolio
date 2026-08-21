package middleware

import (
	"net/http"
	"strings"

	service "wms-api/services/authentication"
	"wms-api/utils"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserKey  = "authenticated_user"
	ContextTokenKey = "session_token"
)

type Authentication struct {
	service *service.Service
}

func NewAuthentication(service *service.Service) *Authentication {
	return &Authentication{service: service}
}

func (m *Authentication) RequireSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			utils.Failure(c, http.StatusUnauthorized, "authentication required", nil)
			c.Abort()
			return
		}

		user, err := m.service.Authenticate(c.Request.Context(), token)
		if err != nil {
			utils.Failure(c, http.StatusUnauthorized, "invalid or expired session", nil)
			c.Abort()
			return
		}

		c.Set(ContextUserKey, user)
		c.Set(ContextTokenKey, token)
		c.Next()
	}
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}
