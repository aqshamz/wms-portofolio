package inventory

import (
	"strings"

	"github.com/gin-gonic/gin"
	service "wms-api/services/inventory"
	"wms-api/utils"
)

// RequireBalanceScope limits balance inquiry to the account's owner and
// warehouse grants. Superadmins remain unrestricted through request scope.
func (controller *Controller) RequireBalanceScope() gin.HandlerFunc {
	return controller.requireInventoryScope("balance")
}

// RequireMovementScope applies the same access grants to the immutable
// movement ledger and derives detail scope from the stored movement.
func (controller *Controller) RequireMovementScope() gin.HandlerFunc {
	return controller.requireInventoryScope("movement")
}

// RequireSerialStateScope protects the current warehouse position of a serial
// while leaving immutable movement history available through its own guard.
func (controller *Controller) RequireSerialStateScope() gin.HandlerFunc {
	return controller.requireInventoryScope("serial-state")
}

func (controller *Controller) requireInventoryScope(resource string) gin.HandlerFunc {
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
		}
		if ownerID == "" || warehouseID == "" {
			utils.Failure(c, 400, "owner_id and warehouse_id are required for inventory access", nil)
			c.Abort()
			return
		}
		allowed, err := controller.service.CanAccess(c.Request.Context(), actor(c), ownerID, warehouseID)
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
