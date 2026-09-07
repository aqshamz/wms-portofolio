package routes

import (
	"github.com/gin-gonic/gin"
	controller "wms-api/controller/stock_control"
)

func registerStockControlRoutes(api *gin.RouterGroup, controller *controller.Controller, requireSession gin.HandlerFunc) {
	group := api.Group("/stock-control")
	group.Use(requireSession)
	group.GET("/reason-codes", controller.ListReasons)
	group.POST("/internal-moves", controller.InternalMove)
	group.POST("/status-changes", controller.StatusChange)
	group.POST("/adjustments", controller.Adjustment)
	group.POST("/stock-count-reconciliations", controller.ReconcileCount)
	group.POST("/warehouse-transfers", controller.WarehouseTransfer)
}
