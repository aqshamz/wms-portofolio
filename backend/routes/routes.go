package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	authcontroller "wms-api/controller/authentication"
	billingcontroller "wms-api/controller/billing"
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
	SecurityController       *authcontroller.SecurityController
	AuthenticationMiddleware *middleware.Authentication
	MasterController         *mastercontroller.Controller
	CatalogController        *mastercontroller.CatalogController
	OperationalController    *mastercontroller.OperationalController
	InventoryController      *inventorycontroller.Controller
	InboundController        *inboundcontroller.Controller
	OutboundController       *outboundcontroller.Controller
	StockControlController   *stockcontrolcontroller.Controller
	BillingController        *billingcontroller.Controller
	Hardening                *middleware.Hardening
	TrustedProxies           []string
}

func New(db *gorm.DB, dependencies Dependencies) *gin.Engine {
	router := gin.New()
	if dependencies.Hardening != nil {
		router.Use(dependencies.Hardening.RequestID(), dependencies.Hardening.Audit(), dependencies.Hardening.SecurityHeaders(), dependencies.Hardening.CORS(), dependencies.Hardening.BodyLimit(), dependencies.Hardening.RateLimit())
	}
	router.Use(gin.Logger(), gin.CustomRecovery(func(c *gin.Context, _ any) {
		if !c.Writer.Written() {
			utils.Failure(c, http.StatusInternalServerError, "internal server error", nil)
		}
		c.Abort()
	}))
	router.HandleMethodNotAllowed = true
	router.NoRoute(func(c *gin.Context) {
		utils.Failure(c, http.StatusNotFound, "route not found", nil)
	})
	router.NoMethod(func(c *gin.Context) {
		utils.Failure(c, http.StatusMethodNotAllowed, "method not allowed", nil)
	})
	if err := router.SetTrustedProxies(dependencies.TrustedProxies); err != nil {
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
	registerSecurityRoutes(api, dependencies.SecurityController, dependencies.AuthenticationMiddleware.RequireModule("SECURITY"))
	registerMasterRoutes(
		api,
		dependencies.MasterController,
		dependencies.AuthenticationMiddleware.RequireModule("MASTER"),
	)
	registerCatalogRoutes(api, dependencies.CatalogController, dependencies.AuthenticationMiddleware.RequireModule("MASTER"))
	registerOperationalRoutes(api, dependencies.OperationalController, dependencies.AuthenticationMiddleware.RequireModule("MASTER"))
	registerInventoryRoutes(api, dependencies.InventoryController, dependencies.AuthenticationMiddleware.RequireModule("INVENTORY"))
	registerInboundRoutes(api, dependencies.InboundController, dependencies.AuthenticationMiddleware.RequireModule("INBOUND"))
	registerStockControlRoutes(api, dependencies.StockControlController, dependencies.AuthenticationMiddleware.RequireModule("INVENTORY"))
	registerOutboundRoutes(api, dependencies.OutboundController, dependencies.AuthenticationMiddleware.RequireModule("OUTBOUND"))
	registerBillingRoutes(api, dependencies.BillingController, dependencies.AuthenticationMiddleware.RequireModule("BILLING"))
	registerDocumentationRoutes(router)

	return router
}
