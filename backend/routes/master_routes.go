package routes

import (
	controller "wms-api/controller/master"

	"github.com/gin-gonic/gin"
)

func registerMasterRoutes(
	api *gin.RouterGroup,
	controller *controller.Controller,
	requireSession gin.HandlerFunc,
) {
	master := api.Group("/master")
	master.Use(requireSession)

	organizations := master.Group("/organizations")
	organizations.POST("", controller.CreateOrganization)
	organizations.GET("", controller.ListOrganizations)
	organizations.GET("/:id", controller.GetOrganization)
	organizations.PUT("/:id", controller.UpdateOrganization)
	organizations.PATCH("/:id/deactivate", controller.DeactivateOrganization)

	warehouses := master.Group("/warehouses")
	warehouses.POST("", controller.CreateWarehouse)
	warehouses.GET("", controller.ListWarehouses)
	warehouses.GET("/:id", controller.GetWarehouse)
	warehouses.PUT("/:id", controller.UpdateWarehouse)
	warehouses.PATCH("/:id/deactivate", controller.DeactivateWarehouse)
	warehouses.POST("/:id/owners", controller.AssignWarehouseOwner)
	warehouses.GET("/:id/owners", controller.ListWarehouseOwners)
	warehouses.PATCH("/:id/owners/:owner_id/deactivate", controller.DeactivateWarehouseOwner)
	warehouses.POST("/:id/zones", controller.CreateWarehouseZone)
	warehouses.GET("/:id/zones", controller.ListWarehouseZones)
	warehouses.POST("/:id/locations", controller.CreateWarehouseLocation)
	warehouses.GET("/:id/locations", controller.ListWarehouseLocations)

	locationTypes := master.Group("/location-types")
	locationTypes.POST("", controller.CreateLocationType)
	locationTypes.GET("", controller.ListLocationTypes)
	locationTypes.PUT("/:id", controller.UpdateLocationType)

	zones := master.Group("/zones")
	zones.PUT("/:id", controller.UpdateWarehouseZone)
	zones.PATCH("/:id/deactivate", controller.DeactivateWarehouseZone)

	locations := master.Group("/locations")
	locations.GET("/:id", controller.GetWarehouseLocation)
	locations.PUT("/:id", controller.UpdateWarehouseLocation)
	locations.PATCH("/:id/deactivate", controller.DeactivateWarehouseLocation)

	owners := master.Group("/owners")
	owners.POST("/:owner_id/account-access", controller.GrantOwnerAccess)

	warehouseAccess := master.Group("/warehouse-access")
	warehouseAccess.POST("/:warehouse_id", controller.GrantWarehouseAccess)

	accounts := master.Group("/accounts")
	accounts.GET("/:account_id/owner-access", controller.ListAccountOwnerAccess)
	accounts.DELETE("/:account_id/owner-access/:owner_id", controller.RevokeOwnerAccess)
	accounts.GET("/:account_id/warehouse-access", controller.ListAccountWarehouseAccess)
	accounts.DELETE("/:account_id/warehouse-access/:warehouse_id", controller.RevokeWarehouseAccess)
}
