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
	inventory.GET("/balances", read, controller.ListBalances)
	inventory.GET("/balances/:id", read, controller.GetBalance)
	inventory.GET("/movements", read, controller.ListMovements)
	inventory.GET("/movements/:id", read, controller.GetMovement)
	inventory.GET("/serial-states", read, controller.ListSerialStates)
	inventory.GET("/serial-states/:id", read, controller.GetSerialState)
	inventory.GET("/movement-types", read, controller.ListMovementTypes)
	inventory.POST("/lots", identity, controller.CreateLot)
	inventory.GET("/lots", read, controller.ListLot)
	inventory.GET("/lots/:id", read, controller.GetLot)
	inventory.POST("/serials", identity, controller.CreateSerial)
	inventory.GET("/serials", read, controller.ListSerial)
	inventory.GET("/serials/:id", read, controller.GetSerial)
	inventory.POST("/handling-units", identity, controller.CreateHandlingUnit)
	inventory.GET("/handling-units", read, controller.ListHandlingUnit)
	inventory.GET("/handling-units/:id", read, controller.GetHandlingUnit)
}
