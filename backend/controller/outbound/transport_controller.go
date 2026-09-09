package outbound

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	dto "wms-api/dto/outbound"
	"wms-api/utils"
)

func transportFilters(c *gin.Context, searchAllowed bool) (*bool, string, bool) {
	for key, values := range c.Request.URL.Query() {
		if (key != "active" && (!searchAllowed || key != "search")) || len(values) != 1 {
			utils.Failure(c, http.StatusBadRequest, "unknown or repeated query parameter: "+key, nil)
			return nil, "", false
		}
	}
	var active *bool
	if raw := strings.TrimSpace(c.Query("active")); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			utils.Failure(c, http.StatusBadRequest, "active must be true or false", nil)
			return nil, "", false
		}
		active = &value
	}
	return active, strings.TrimSpace(c.Query("search")), true
}

func (cn *Controller) CreateCarrier(c *gin.Context) {
	var q dto.CreateCarrierRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.CreateCarrier(c.Request.Context(), q)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "carrier created", v)
}
func (cn *Controller) ListCarriers(c *gin.Context) {
	active, search, ok := transportFilters(c, true)
	if !ok {
		return
	}
	v, err := cn.service.ListCarriers(c.Request.Context(), active, search)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "carriers retrieved", v)
}
func (cn *Controller) GetCarrier(c *gin.Context) {
	v, err := cn.service.GetCarrier(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "carrier retrieved", v)
}
func (cn *Controller) UpdateCarrier(c *gin.Context) {
	var q dto.UpdateCarrierRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.UpdateCarrier(c.Request.Context(), c.Param("id"), q)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "carrier updated", v)
}
func (cn *Controller) CreateCarrierService(c *gin.Context) {
	var q dto.CreateCarrierServiceRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.CreateCarrierService(c.Request.Context(), c.Param("id"), q)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "carrier service created", v)
}
func (cn *Controller) ListCarrierServices(c *gin.Context) {
	active, _, ok := transportFilters(c, false)
	if !ok {
		return
	}
	v, err := cn.service.ListCarrierServices(c.Request.Context(), c.Param("id"), active)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "carrier services retrieved", v)
}
func (cn *Controller) GetCarrierService(c *gin.Context) {
	v, err := cn.service.GetCarrierService(c.Request.Context(), c.Param("id"), c.Param("service_id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "carrier service retrieved", v)
}
func (cn *Controller) UpdateCarrierService(c *gin.Context) {
	var q dto.UpdateCarrierServiceRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.UpdateCarrierService(c.Request.Context(), c.Param("id"), c.Param("service_id"), q)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "carrier service updated", v)
}
func (cn *Controller) CreateCarrierDriver(c *gin.Context) {
	var q dto.CreateCarrierDriverRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.CreateCarrierDriver(c.Request.Context(), c.Param("id"), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "carrier driver created", v)
}
func (cn *Controller) ListCarrierDrivers(c *gin.Context) {
	active, _, ok := transportFilters(c, false)
	if !ok {
		return
	}
	v, err := cn.service.ListCarrierDrivers(c.Request.Context(), c.Param("id"), active)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "carrier drivers retrieved", v)
}
func (cn *Controller) GetCarrierDriver(c *gin.Context) {
	v, err := cn.service.GetCarrierDriver(c.Request.Context(), c.Param("id"), c.Param("driver_id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "carrier driver retrieved", v)
}
func (cn *Controller) UpdateCarrierDriver(c *gin.Context) {
	var q dto.UpdateCarrierDriverRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.UpdateCarrierDriver(c.Request.Context(), c.Param("id"), c.Param("driver_id"), q)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "carrier driver updated", v)
}
func (cn *Controller) AssignShipmentDriver(c *gin.Context) {
	var q dto.AssignShipmentDriverRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.AssignShipmentDriver(c.Request.Context(), c.Param("id"), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "shipment driver assigned", v)
}
func (cn *Controller) ListShipmentDrivers(c *gin.Context) {
	if len(c.Request.URL.Query()) != 0 {
		utils.Failure(c, http.StatusBadRequest, "query parameters are not supported", nil)
		return
	}
	v, err := cn.service.ListShipmentDrivers(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "shipment drivers retrieved", v)
}
func (cn *Controller) RemoveShipmentDriver(c *gin.Context) {
	v, err := cn.service.RemoveShipmentDriver(c.Request.Context(), c.Param("id"), c.Param("driver_id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "shipment driver removed", v)
}
