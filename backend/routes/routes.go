package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	authcontroller "wms-api/controller/authentication"
	inboundcontroller "wms-api/controller/inbound"
	inventorycontroller "wms-api/controller/inventory"
	mastercontroller "wms-api/controller/master"
	outboundcontroller "wms-api/controller/outbound"
	stockcontrolcontroller "wms-api/controller/stock_control"
	"wms-api/middleware"
	"wms-api/utils"
)

type Dependencies struct {
	AuthenticationController *authcontroller.Controller
	AuthenticationMiddleware *middleware.Authentication
	MasterController         *mastercontroller.Controller
	CatalogController        *mastercontroller.CatalogController
	OperationalController    *mastercontroller.OperationalController
	InventoryController      *inventorycontroller.Controller
	InboundController        *inboundcontroller.Controller
	OutboundController       *outboundcontroller.Controller
	StockControlController   *stockcontrolcontroller.Controller
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
	registerMasterRoutes(
		api,
		dependencies.MasterController,
		dependencies.AuthenticationMiddleware.RequireSession(),
	)
	registerCatalogRoutes(api, dependencies.CatalogController, dependencies.AuthenticationMiddleware.RequireSession())
	registerOperationalRoutes(api, dependencies.OperationalController, dependencies.AuthenticationMiddleware.RequireSession())
	registerInventoryRoutes(api, dependencies.InventoryController, dependencies.AuthenticationMiddleware.RequireSession())
	registerInboundRoutes(api, dependencies.InboundController, dependencies.AuthenticationMiddleware.RequireSession())
	registerStockControlRoutes(api, dependencies.StockControlController, dependencies.AuthenticationMiddleware.RequireSession())
	registerOutboundRoutes(api, dependencies.OutboundController, dependencies.AuthenticationMiddleware.RequireSession())

	return router
}
