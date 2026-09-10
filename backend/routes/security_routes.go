package routes

import (
	controller "wms-api/controller/authentication"

	"github.com/gin-gonic/gin"
)

func registerSecurityRoutes(
	api *gin.RouterGroup,
	controller *controller.SecurityController,
	requirePermission gin.HandlerFunc,
) {
	security := api.Group("/security")
	security.Use(requirePermission)
	security.GET("/permissions", controller.ListPermissions)
	security.GET("/accounts/:account_id/permissions", controller.ListAccountPermissions)
	security.POST("/accounts/:account_id/permissions", controller.GrantAccountPermission)
	security.DELETE("/accounts/:account_id/permissions/:permission_id", controller.RevokeAccountPermission)
}
