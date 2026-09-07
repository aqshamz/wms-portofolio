package routes

import (
	"github.com/gin-gonic/gin"
	controller "wms-api/controller/inventory"
)

func registerInventoryRoutes(api *gin.RouterGroup, controller *controller.Controller, requireSession gin.HandlerFunc) {
	inventory := api.Group("/inventory")
	inventory.Use(requireSession)
	inventory.GET("/balances", controller.ListBalances)
	inventory.GET("/balances/:id", controller.GetBalance)
	inventory.GET("/movements", controller.ListMovements)
	inventory.GET("/movements/:id", controller.GetMovement)
	inventory.GET("/serial-states", controller.ListSerialStates)
	inventory.GET("/serial-states/:id", controller.GetSerialState)
	inventory.GET("/movement-types", controller.ListMovementTypes)
	inventory.POST("/lots", controller.CreateLot)
	inventory.GET("/lots", controller.ListLot)
	inventory.GET("/lots/:id", controller.GetLot)
	inventory.POST("/serials", controller.CreateSerial)
	inventory.GET("/serials", controller.ListSerial)
	inventory.GET("/serials/:id", controller.GetSerial)
	inventory.POST("/handling-units", controller.CreateHandlingUnit)
	inventory.GET("/handling-units", controller.ListHandlingUnit)
	inventory.GET("/handling-units/:id", controller.GetHandlingUnit)
}
