package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	authcontroller "wms-api/controller/authentication"
	"wms-api/middleware"
	"wms-api/utils"
)

type Dependencies struct {
	AuthenticationController *authcontroller.Controller
	AuthenticationMiddleware *middleware.Authentication
}

func New(db *gorm.DB, dependencies Dependencies) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	if err := router.SetTrustedProxies(nil); err != nil {
		panic("configure trusted proxies: " + err.Error())
	}

	router.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.PingContext(c.Request.Context()) != nil {
			utils.Failure(c, http.StatusServiceUnavailable, "database is unavailable", nil)
			return
		}

		utils.Success(c, http.StatusOK, "WMS API is healthy", gin.H{
			"database": "connected",
		})
	})

	api := router.Group("/api/v1")
	registerAuthenticationRoutes(
		api,
		dependencies.AuthenticationController,
		dependencies.AuthenticationMiddleware.RequireSession(),
	)
	registerMasterRoutes(api)
	registerInventoryRoutes(api)
	registerInboundRoutes(api)
	registerStockControlRoutes(api)
	registerOutboundRoutes(api)

	return router
}

func registerMasterRoutes(_ *gin.RouterGroup)       {}
func registerInventoryRoutes(_ *gin.RouterGroup)    {}
func registerInboundRoutes(_ *gin.RouterGroup)      {}
func registerStockControlRoutes(_ *gin.RouterGroup) {}
func registerOutboundRoutes(_ *gin.RouterGroup)     {}
