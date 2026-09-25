package inventory

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	service "wms-api/services/inventory"
	"wms-api/utils"
)

// RequireBalanceScope limits balance inquiry to the account's owner and
// warehouse grants. Superadmins remain unrestricted through request scope.
func (controller *Controller) RequireBalanceScope() gin.HandlerFunc {
	return controller.requireInventoryScope("balance", false)
}

// RequireMovementScope applies the same access grants to the immutable
// movement ledger and derives detail scope from the stored movement.
func (controller *Controller) RequireMovementScope() gin.HandlerFunc {
	return controller.requireInventoryScope("movement", false)
}

// RequireSerialStateScope protects the current warehouse position of a serial
// while leaving immutable movement history available through its own guard.
func (controller *Controller) RequireSerialStateScope() gin.HandlerFunc {
	return controller.requireInventoryScope("serial-state", false)
}

// RequireLotScope protects owner-wide lot identities. Lots are not warehouse
// records, so only an owner grant is required.
func (controller *Controller) RequireLotScope() gin.HandlerFunc {
	return controller.requireInventoryScope("lot", true)
}

// RequireSerialScope protects owner-wide serial identities. A serial identity
// can exist before receipt and after shipment, so it is not warehouse-scoped.
func (controller *Controller) RequireSerialScope() gin.HandlerFunc {
	return controller.requireInventoryScope("serial", true)
}

// RequireHandlingUnitScope protects handling-unit identities using both the
// stored owner and warehouse, including for detail requests.
func (controller *Controller) RequireHandlingUnitScope() gin.HandlerFunc {
	return controller.requireInventoryScope("handling-unit", false)
}

func (controller *Controller) requireInventoryScope(resource string, ownerOnly bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ownerID := strings.TrimSpace(c.Query("owner_id"))
		warehouseID := strings.TrimSpace(c.Query("warehouse_id"))
		if c.Param("id") != "" {
			var err error
			ownerID, warehouseID, err = controller.service.ResourceScope(c.Request.Context(), resource, c.Param("id"))
			if err != nil {
				fail(c, err)
				c.Abort()
				return
			}
		} else if c.Request.Method != http.MethodGet {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				utils.Failure(c, 400, "unable to read request body", nil)
				c.Abort()
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			var value map[string]interface{}
			if json.Unmarshal(body, &value) != nil {
				c.Next()
				return
			}
			if text, ok := value["owner_id"].(string); ok {
				ownerID = strings.TrimSpace(text)
			}
			if text, ok := value["warehouse_id"].(string); ok {
				warehouseID = strings.TrimSpace(text)
			}
		}
		if ownerID == "" || (!ownerOnly && warehouseID == "") {
			message := "owner_id and warehouse_id are required for inventory access"
			if ownerOnly {
				message = "owner_id is required for inventory access"
			}
			utils.Failure(c, 400, message, nil)
			c.Abort()
			return
		}
		var allowed bool
		var err error
		if ownerOnly {
			allowed, err = controller.service.CanAccessOwner(c.Request.Context(), actor(c), ownerID)
		} else {
			allowed, err = controller.service.CanAccess(c.Request.Context(), actor(c), ownerID, warehouseID)
		}
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
