package routes

import (
	"github.com/gin-gonic/gin"
	controller "wms-api/controller/outbound"
)

func registerOutboundRoutes(api *gin.RouterGroup, controller *controller.Controller, requireSession gin.HandlerFunc) {
	group := api.Group("/outbound")
	group.Use(requireSession)
	group.POST("/orders", controller.CreateOrder)
	group.GET("/orders", controller.ListOrders)
	group.GET("/orders/:id", controller.GetOrder)
	group.POST("/orders/:id/validate", controller.ValidateOrder)
	group.POST("/orders/:id/release", controller.ReleaseOrder)
	group.POST("/orders/:id/cancel", controller.CancelOrder)
	group.POST("/orders/:id/allocate", controller.AllocateOrder)
	group.GET("/validation-runs/:id", controller.GetValidationRun)
	group.GET("/reservations", controller.ListReservations)
	group.POST("/reservations/:id/release", controller.ReleaseReservation)
	group.POST("/waves", controller.CreateWave)
	group.GET("/waves", controller.ListWaves)
	group.GET("/waves/:id", controller.GetWave)
	group.POST("/waves/:id/release", controller.ReleaseWave)
	group.POST("/waves/:id/cancel", controller.CancelWave)
	group.GET("/pick-tasks", controller.ListPickTasks)
	group.GET("/pick-tasks/:id", controller.GetPickTask)
	group.POST("/pick-tasks/:id/start", controller.StartPickTask)
	group.POST("/pick-tasks/:id/confirm", controller.ConfirmPick)
	group.POST("/pick-tasks/:id/short-close", controller.CloseShortPick)
	group.POST("/stagings", controller.CreateStaging)
	group.GET("/stagings", controller.ListStagings)
	group.GET("/stagings/:id", controller.GetStaging)
	group.POST("/stagings/:id/complete", controller.CompleteStaging)
}
