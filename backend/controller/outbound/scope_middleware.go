package outbound

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	service "wms-api/services/outbound"
	"wms-api/utils"
)

func (cn *Controller) RequireScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		marker := "/outbound/"
		at := strings.Index(path, marker)
		if at < 0 {
			c.Next()
			return
		}
		relative := path[at+len(marker):]
		kind := strings.Split(relative, "/")[0]
		if kind == "carriers" {
			c.Next()
			return
		}

		ownerID, warehouseID := "", ""
		resourceKind, resourceID := "", ""
		if c.Param("id") != "" {
			resourceKind, resourceID = kind, c.Param("id")
		} else if c.Request.Method == http.MethodGet {
			ownerID, warehouseID = strings.TrimSpace(c.Query("owner_id")), strings.TrimSpace(c.Query("warehouse_id"))
		} else {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				utils.Failure(c, http.StatusBadRequest, "unable to read request body", nil)
				c.Abort()
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			var value map[string]interface{}
			if json.Unmarshal(body, &value) != nil {
				c.Next()
				return
			}
			text := func(key string) string { result, _ := value[key].(string); return strings.TrimSpace(result) }
			switch kind {
			case "orders", "waves", "shipments", "return-policy":
				ownerID, warehouseID = text("owner_id"), text("warehouse_id")
			case "stagings":
				resourceKind, resourceID = "orders", text("outbound_id")
			case "checks":
				resourceKind, resourceID = "stagings", text("staging_id")
			case "packings":
				resourceKind, resourceID = "checks", text("outbound_check_id")
			case "deliveries":
				resourceKind, resourceID = "orders", text("outbound_id")
			default:
				c.Next()
				return
			}
		}
		if resourceID != "" {
			var err error
			ownerID, warehouseID, err = cn.service.ResourceScope(c.Request.Context(), resourceKind, resourceID)
			if err != nil {
				fail(c, err)
				c.Abort()
				return
			}
		}
		if ownerID == "" || warehouseID == "" {
			utils.Failure(c, http.StatusBadRequest, "owner_id and warehouse_id are required for outbound access", nil)
			c.Abort()
			return
		}
		allowed, err := cn.service.CanAccess(c.Request.Context(), actor(c), ownerID, warehouseID)
		if err != nil {
			fail(c, err)
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
