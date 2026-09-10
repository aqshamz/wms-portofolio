package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"wms-api/config"
	authdto "wms-api/dto/authentication"
	auditmodel "wms-api/models/audit"
	"wms-api/utils"
)

const ContextRequestIDKey = "request_id"

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{8,64}$`)

type AuditWriter func(context.Context, *auditmodel.APIAuditLog) error
type rateWindow struct {
	reset time.Time
	count int
}
type Hardening struct {
	security config.SecurityConfig
	origins  map[string]bool
	audit    AuditWriter
	mu       sync.Mutex
	windows  map[string]rateWindow
}

func NewHardening(security config.SecurityConfig, audit AuditWriter) *Hardening {
	origins := map[string]bool{}
	for _, v := range security.AllowedOrigins {
		origins[v] = true
	}
	return &Hardening{security: security, origins: origins, audit: audit, windows: map[string]rateWindow{}}
}

func (m *Hardening) RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if !requestIDPattern.MatchString(id) {
			var b [16]byte
			if _, e := rand.Read(b[:]); e == nil {
				id = hex.EncodeToString(b[:])
			} else {
				id = time.Now().UTC().Format("20060102150405.000000000")
			}
		}
		c.Set(ContextRequestIDKey, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}
func (m *Hardening) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Header("Cache-Control", "no-store")
		if c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}
func (m *Hardening) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := strings.TrimSuffix(strings.TrimSpace(c.GetHeader("Origin")), "/")
		if origin == "" {
			c.Next()
			return
		}
		if !m.origins[origin] && !m.origins["*"] {
			utils.Failure(c, http.StatusForbidden, "origin is not allowed", nil)
			c.Abort()
			return
		}
		c.Header("Vary", "Origin")
		if m.origins["*"] {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Max-Age", "600")
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}
		c.Next()
	}
}
func (m *Hardening) BodyLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > m.security.MaxRequestBodyBytes {
			utils.Failure(c, http.StatusRequestEntityTooLarge, "request body is too large", nil)
			c.Abort()
			return
		}
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, m.security.MaxRequestBodyBytes)
		}
		c.Next()
	}
}
func (m *Hardening) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		key := c.ClientIP()
		m.mu.Lock()
		w := m.windows[key]
		if now.After(w.reset) {
			w = rateWindow{reset: now.Add(time.Minute)}
		}
		w.count++
		m.windows[key] = w
		limited := w.count > m.security.RateLimitPerMinute
		remaining := m.security.RateLimitPerMinute - w.count
		if remaining < 0 {
			remaining = 0
		}
		if len(m.windows) > 10000 {
			for k, v := range m.windows {
				if now.After(v.reset) {
					delete(m.windows, k)
				}
			}
		}
		m.mu.Unlock()
		c.Header("X-RateLimit-Limit", fmtInt(m.security.RateLimitPerMinute))
		c.Header("X-RateLimit-Remaining", fmtInt(remaining))
		if limited {
			c.Header("Retry-After", fmtInt(int(time.Until(w.reset).Seconds())+1))
			utils.Failure(c, http.StatusTooManyRequests, "rate limit exceeded", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}
func (m *Hardening) Audit() gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		if m.audit == nil || c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			return
		}
		id, _ := c.Get(ContextRequestIDKey)
		requestID, _ := id.(string)
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}
		var accountID *string
		if value, ok := c.Get(ContextUserKey); ok {
			if user, yes := value.(authdto.UserResponse); yes && user.AccountID != "" {
				accountID = &user.AccountID
			}
		}
		ip := optionalAudit(c.ClientIP())
		ua := optionalAudit(c.Request.UserAgent())
		entry := auditmodel.APIAuditLog{RequestID: requestID, AccountID: accountID, Method: c.Request.Method, Route: route, StatusCode: c.Writer.Status(), IPAddress: ip, UserAgent: ua, DurationMS: time.Since(started).Milliseconds()}
		auditContext, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := m.audit(auditContext, &entry); err != nil {
			log.Printf("write API audit request_id=%s: %v", requestID, err)
		}
	}
}
func optionalAudit(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
}
func fmtInt(v int) string {
	if v == 0 {
		return "0"
	}
	negative := v < 0
	if negative {
		v = -v
	}
	var b [20]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	if negative {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
