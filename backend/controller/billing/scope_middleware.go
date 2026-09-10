package billing

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
	service "wms-api/services/billing"
	"wms-api/utils"
)

func (cn *Controller) RequireScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		at := strings.Index(path, "/billing/")
		if at < 0 {
			c.Next()
			return
		}
		relative := path[at+len("/billing/"):]
		kind := strings.Split(relative, "/")[0]
		owner, warehouse, resourceKind, resourceID := "", "", "", ""
		if c.Param("id") != "" {
			resourceKind, resourceID = kind, c.Param("id")
		} else if c.Request.Method == http.MethodGet {
			owner, warehouse = strings.TrimSpace(c.Query("owner_id")), strings.TrimSpace(c.Query("warehouse_id"))
		} else {
			body, e := io.ReadAll(c.Request.Body)
			if e != nil {
				utils.Failure(c, http.StatusBadRequest, "unable to read request body", nil)
				c.Abort()
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			var v map[string]interface{}
			if json.Unmarshal(body, &v) != nil {
				c.Next()
				return
			}
			text := func(k string) string { x, _ := v[k].(string); return strings.TrimSpace(x) }
			switch kind {
			case "contracts":
				owner, warehouse = text("owner_id"), text("warehouse_id")
			case "rate-cards", "runs":
				resourceKind, resourceID = "contracts", text("billing_contract_id")
			case "events":
				owner, warehouse = text("owner_id"), text("warehouse_id")
			default:
				c.Next()
				return
			}
		}
		if resourceID != "" {
			var e error
			owner, warehouse, e = cn.service.ResourceScope(c.Request.Context(), resourceKind, resourceID)
			if e != nil {
				fail(c, e)
				c.Abort()
				return
			}
		}
		if owner == "" || warehouse == "" {
			utils.Failure(c, http.StatusBadRequest, "owner_id and warehouse_id are required for billing access", nil)
			c.Abort()
			return
		}
		allowed, e := cn.service.CanAccess(c.Request.Context(), actor(c), owner, warehouse)
		if e != nil {
			fail(c, e)
			c.Abort()
			return
		}
		if !allowed {
			fail(c, service.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}
