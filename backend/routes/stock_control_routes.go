package routes

import (
	"github.com/gin-gonic/gin"
	controller "wms-api/controller/stock_control"
	"wms-api/middleware"
)

func registerStockControlRoutes(api *gin.RouterGroup, controller *controller.Controller, authorization *middleware.Authentication) {
	group := api.Group("/stock-control")
	group.Use(authorization.RequireSession())
	group.GET("/reason-codes", authorization.RequirePermission(middleware.PermissionInventoryRead), controller.ListReasons)
	group.POST("/internal-moves", authorization.RequirePermission(middleware.PermissionInventoryMove), controller.InternalMove)
	group.POST("/status-changes", authorization.RequirePermission(middleware.PermissionInventoryStatusChange), controller.StatusChange)
	group.POST("/adjustments", authorization.RequirePermission(middleware.PermissionInventoryAdjust), controller.Adjustment)
	group.POST("/stock-count-reconciliations", authorization.RequirePermission(middleware.PermissionInventoryCount), controller.ReconcileCount)
	group.POST("/warehouse-transfers", authorization.RequirePermission(middleware.PermissionInventoryTransfer), controller.WarehouseTransfer)
}
