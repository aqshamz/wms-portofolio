package routes

import (
	"github.com/gin-gonic/gin"
	controller "wms-api/controller/master"
)

func registerCatalogRoutes(api *gin.RouterGroup, controller *controller.CatalogController, requireSession gin.HandlerFunc) {
	master := api.Group("/master")
	master.Use(requireSession)
	master.POST("/handling-unit-types", controller.CreateHandlingUnitType)
	master.GET("/handling-unit-types", controller.ListHandlingUnitType)
	master.GET("/handling-unit-types/:id", controller.GetHandlingUnitType)
	master.PUT("/handling-unit-types/:id", controller.UpdateHandlingUnitType)
	master.PATCH("/handling-unit-types/:id/deactivate", controller.DeactivateHandlingUnitType)

	master.POST("/partner-types", controller.CreatePartnerType)
	master.GET("/partner-types", controller.ListPartnerType)
	master.GET("/partner-types/:id", controller.GetPartnerType)
	master.PUT("/partner-types/:id", controller.UpdatePartnerType)
	master.PATCH("/partner-types/:id/deactivate", controller.DeactivatePartnerType)

	master.POST("/business-partners", controller.CreateBusinessPartner)
	master.GET("/business-partners", controller.ListBusinessPartner)
	master.GET("/business-partners/:id", controller.GetBusinessPartner)
	master.PUT("/business-partners/:id", controller.UpdateBusinessPartner)
	master.PATCH("/business-partners/:id/deactivate", controller.DeactivateBusinessPartner)

	master.POST("/uoms", controller.CreateUOM)
	master.GET("/uoms", controller.ListUOM)
	master.GET("/uoms/:id", controller.GetUOM)
	master.PUT("/uoms/:id", controller.UpdateUOM)
	master.PATCH("/uoms/:id/deactivate", controller.DeactivateUOM)

	master.POST("/item-categories", controller.CreateItemCategory)
	master.GET("/item-categories", controller.ListItemCategory)
	master.GET("/item-categories/:id", controller.GetItemCategory)
	master.PUT("/item-categories/:id", controller.UpdateItemCategory)
	master.PATCH("/item-categories/:id/deactivate", controller.DeactivateItemCategory)

	master.POST("/items", controller.CreateItem)
	master.GET("/items", controller.ListItem)
	master.GET("/items/:id", controller.GetItem)
	master.PUT("/items/:id", controller.UpdateItem)
	master.PATCH("/items/:id/deactivate", controller.DeactivateItem)

	master.POST("/inventory-statuses", controller.CreateInventoryStatus)
	master.GET("/inventory-statuses", controller.ListInventoryStatus)
	master.GET("/inventory-statuses/:id", controller.GetInventoryStatus)
	master.PUT("/inventory-statuses/:id", controller.UpdateInventoryStatus)
	master.PATCH("/inventory-statuses/:id/deactivate", controller.DeactivateInventoryStatus)

	master.POST("/quality-statuses", controller.CreateQualityStatus)
	master.GET("/quality-statuses", controller.ListQualityStatus)
	master.GET("/quality-statuses/:id", controller.GetQualityStatus)
	master.PUT("/quality-statuses/:id", controller.UpdateQualityStatus)
	master.PATCH("/quality-statuses/:id/deactivate", controller.DeactivateQualityStatus)

	master.POST("/inspection-results", controller.CreateInspectionResult)
	master.GET("/inspection-results", controller.ListInspectionResult)
	master.GET("/inspection-results/:id", controller.GetInspectionResult)
	master.PUT("/inspection-results/:id", controller.UpdateInspectionResult)
	master.PATCH("/inspection-results/:id/deactivate", controller.DeactivateInspectionResult)

	master.GET("/business-partners/:id/types", controller.ListBusinessPartnerType)
	master.POST("/business-partners/:id/types", controller.AssignBusinessPartnerType)
	master.DELETE("/business-partners/:id/types/:partner_type_id", controller.RemoveBusinessPartnerType)

	master.GET("/items/:id/uoms", controller.ListItemUOM)
	master.POST("/items/:id/uoms", controller.CreateItemUOM)
	master.PUT("/items/:id/uoms/:item_uom_id", controller.UpdateItemUOM)
	master.PATCH("/items/:id/uoms/:item_uom_id/deactivate", controller.DeactivateItemUOM)

	master.GET("/items/:id/barcodes", controller.ListItemBarcode)
	master.POST("/items/:id/barcodes", controller.CreateItemBarcode)
	master.PUT("/items/:id/barcodes/:barcode_id", controller.UpdateItemBarcode)
	master.PATCH("/items/:id/barcodes/:barcode_id/deactivate", controller.DeactivateItemBarcode)

	master.PATCH("/items/:id/barcodes/:barcode_id/primary", controller.SetPrimaryItemBarcode)
}
