package routes

import (
	"github.com/gin-gonic/gin"
	controller "wms-api/controller/inventory"
	"wms-api/middleware"
)

func registerInventoryRoutes(api *gin.RouterGroup, controller *controller.Controller, authorization *middleware.Authentication) {
	inventory := api.Group("/inventory")
	inventory.Use(authorization.RequireSession())
	read := authorization.RequirePermission(middleware.PermissionInventoryRead)
	identity := authorization.RequirePermission(middleware.PermissionInventoryIdentity)
	balanceScope := controller.RequireBalanceScope()
	inventory.GET("/balances", read, balanceScope, controller.ListBalances)
	inventory.GET("/balances/:id", read, balanceScope, controller.GetBalance)
	movementScope := controller.RequireMovementScope()
	inventory.GET("/movements", read, movementScope, controller.ListMovements)
	inventory.GET("/movements/:id", read, movementScope, controller.GetMovement)
	serialStateScope := controller.RequireSerialStateScope()
	inventory.GET("/serial-states", read, serialStateScope, controller.ListSerialStates)
	inventory.GET("/serial-states/:id", read, serialStateScope, controller.GetSerialState)
	inventory.GET("/movement-types", read, controller.ListMovementTypes)
	lotScope := controller.RequireLotScope()
	inventory.POST("/lots", identity, lotScope, controller.CreateLot)
	inventory.GET("/lots", read, lotScope, controller.ListLot)
	inventory.GET("/lots/:id", read, lotScope, controller.GetLot)
	serialScope := controller.RequireSerialScope()
	inventory.POST("/serials", identity, serialScope, controller.CreateSerial)
	inventory.GET("/serials", read, serialScope, controller.ListSerial)
	inventory.GET("/serials/:id", read, serialScope, controller.GetSerial)
	handlingUnitScope := controller.RequireHandlingUnitScope()
	inventory.POST("/handling-units", identity, handlingUnitScope, controller.CreateHandlingUnit)
	inventory.GET("/handling-units", read, handlingUnitScope, controller.ListHandlingUnit)
	inventory.GET("/handling-units/:id", read, handlingUnitScope, controller.GetHandlingUnit)
}
