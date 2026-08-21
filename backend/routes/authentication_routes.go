package routes

import (
	controller "wms-api/controller/authentication"

	"github.com/gin-gonic/gin"
)

func registerAuthenticationRoutes(
	api *gin.RouterGroup,
	controller *controller.Controller,
	requireSession gin.HandlerFunc,
) {
	auth := api.Group("/auth")
	auth.POST("/login", controller.Login)
	auth.GET("/me", requireSession, controller.Me)
	auth.POST("/logout", requireSession, controller.Logout)
	auth.POST("/logout-all", requireSession, controller.LogoutAll)
}
