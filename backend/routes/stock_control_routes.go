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
	group.GET("/replenishment-tasks", authorization.RequirePermission(middleware.PermissionInventoryMove), controller.ListReplenishments)
	group.GET("/replenishment-targets", authorization.RequirePermission(middleware.PermissionInventoryMove), controller.ListReplenishmentTargets)
	group.POST("/replenishment-tasks", authorization.RequirePermission(middleware.PermissionInventoryMove), controller.CreateReplenishment)
	group.GET("/replenishment-tasks/:id", authorization.RequirePermission(middleware.PermissionInventoryMove), controller.GetReplenishment)
	group.GET("/replenishment-tasks/:id/assignees", authorization.RequirePermission(middleware.PermissionInventoryMove), controller.ListReplenishmentAssignees)
	group.POST("/replenishment-tasks/:id/assign", authorization.RequirePermission(middleware.PermissionInventoryMove), controller.AssignReplenishment)
	group.POST("/replenishment-tasks/:id/start", authorization.RequirePermission(middleware.PermissionInventoryMove), controller.StartReplenishment)
	group.POST("/replenishment-tasks/:id/complete", authorization.RequirePermission(middleware.PermissionInventoryMove), controller.CompleteReplenishment)
	group.POST("/replenishment-tasks/:id/cancel", authorization.RequirePermission(middleware.PermissionInventoryMove), controller.CancelReplenishment)
	group.POST("/status-changes", authorization.RequirePermission(middleware.PermissionInventoryStatusChange), controller.StatusChange)
	group.GET("/adjustments", authorization.RequireAnyPermission(middleware.PermissionInventoryAdjust, middleware.PermissionInventoryAdjustApprove), controller.ListAdjustments)
	group.POST("/adjustments", authorization.RequirePermission(middleware.PermissionInventoryAdjust), controller.CreateAdjustment)
	group.GET("/adjustments/:id", authorization.RequireAnyPermission(middleware.PermissionInventoryAdjust, middleware.PermissionInventoryAdjustApprove), controller.GetAdjustment)
	group.POST("/adjustments/:id/approve", authorization.RequirePermission(middleware.PermissionInventoryAdjustApprove), controller.ApproveAdjustment)
	group.POST("/adjustments/:id/reject", authorization.RequirePermission(middleware.PermissionInventoryAdjustApprove), controller.RejectAdjustment)
	group.POST("/adjustments/:id/cancel", authorization.RequirePermission(middleware.PermissionInventoryAdjust), controller.CancelAdjustment)
	group.GET("/cycle-counts", authorization.RequireAnyPermission(middleware.PermissionInventoryCount, middleware.PermissionInventoryCountApprove), controller.ListCycleCounts)
	group.POST("/cycle-counts", authorization.RequirePermission(middleware.PermissionInventoryCount), controller.CreateCycleCount)
	group.POST("/cycle-counts/grand", authorization.RequirePermission(middleware.PermissionInventoryCount), controller.CreateGrandStockOpname)
	group.GET("/cycle-counts/:id", authorization.RequireAnyPermission(middleware.PermissionInventoryCount, middleware.PermissionInventoryCountApprove), controller.GetCycleCount)
	group.POST("/cycle-counts/:id/count", authorization.RequirePermission(middleware.PermissionInventoryCount), controller.RecordCycleCount)
	group.POST("/cycle-counts/:id/approve", authorization.RequirePermission(middleware.PermissionInventoryCountApprove), controller.ApproveCycleCount)
	group.POST("/cycle-counts/:id/reject", authorization.RequirePermission(middleware.PermissionInventoryCountApprove), controller.RejectCycleCount)
	group.POST("/cycle-counts/:id/cancel", authorization.RequirePermission(middleware.PermissionInventoryCount), controller.CancelCycleCount)
	group.GET("/warehouse-transfer-documents", authorization.RequireAnyPermission(middleware.PermissionInventoryTransfer, middleware.PermissionInboundReceive, middleware.PermissionInboundPutaway), controller.ListWarehouseTransfers)
	group.POST("/warehouse-transfer-documents", authorization.RequirePermission(middleware.PermissionInventoryTransfer), controller.CreateWarehouseTransfer)
	group.GET("/warehouse-transfer-documents/:id", authorization.RequireAnyPermission(middleware.PermissionInventoryTransfer, middleware.PermissionInboundReceive, middleware.PermissionInboundPutaway), controller.GetWarehouseTransfer)
	group.POST("/warehouse-transfer-documents/:id/approve", authorization.RequirePermission(middleware.PermissionInventoryTransfer), controller.ApproveWarehouseTransfer)
	group.POST("/warehouse-transfer-documents/:id/dispatch", authorization.RequirePermission(middleware.PermissionInventoryTransfer), controller.DispatchWarehouseTransfer)
	group.POST("/warehouse-transfer-documents/:id/cancel", authorization.RequirePermission(middleware.PermissionInventoryTransfer), controller.CancelWarehouseTransfer)
	group.POST("/warehouse-transfer-documents/:id/receive", authorization.RequirePermission(middleware.PermissionInboundReceive), controller.ReceiveWarehouseTransfer)
	group.POST("/warehouse-transfer-documents/:id/putaway", authorization.RequirePermission(middleware.PermissionInboundPutaway), controller.PutawayWarehouseTransfer)
	group.POST("/warehouse-transfers", authorization.RequirePermission(middleware.PermissionInventoryTransferForce), controller.WarehouseTransfer)
}
