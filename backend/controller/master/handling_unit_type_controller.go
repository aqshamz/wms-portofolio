package master

import (
	"github.com/gin-gonic/gin"
	"net/http"
	dto "wms-api/dto/master"
	"wms-api/utils"
)

func (controller *CatalogController) CreateHandlingUnitType(c *gin.Context) {
	var request dto.CreateHandlingUnitTypeRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.CreateHandlingUnitType(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "handling unit type create successful", response)
}
func (controller *CatalogController) GetHandlingUnitType(c *gin.Context) {
	response, err := controller.service.GetHandlingUnitType(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "handling unit type get successful", response)
}
func (controller *CatalogController) ListHandlingUnitType(c *gin.Context) {
	filter, ok := catalogFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListHandlingUnitType(c.Request.Context(), filter)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "handling unit type list successful", response)
}
func (controller *CatalogController) UpdateHandlingUnitType(c *gin.Context) {
	var request dto.UpdateHandlingUnitTypeRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.UpdateHandlingUnitType(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "handling unit type update successful", response)
}
func (controller *CatalogController) DeactivateHandlingUnitType(c *gin.Context) {
	response, err := controller.service.DeactivateHandlingUnitType(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "handling unit type deactivate successful", response)
}
