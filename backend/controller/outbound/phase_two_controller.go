package outbound

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	dto "wms-api/dto/outbound"
	"wms-api/utils"
)

func (cn *Controller) CreateCheck(c *gin.Context) {
	var q dto.CreateCheckRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.CreateCheck(c.Request.Context(), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "outbound check created", v)
}
func (cn *Controller) ListChecks(c *gin.Context) {
	f, _, ok := listFilter(c, false)
	if !ok {
		return
	}
	v, err := cn.service.ListChecks(c.Request.Context(), f)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "outbound checks retrieved", v)
}
func (cn *Controller) GetCheck(c *gin.Context) {
	v, err := cn.service.GetCheck(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "outbound check retrieved", v)
}
func (cn *Controller) RecordCheckLine(c *gin.Context) {
	var q dto.RecordCheckLineRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.RecordCheckLine(c.Request.Context(), c.Param("id"), c.Param("line_id"), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "check line recorded", v)
}
func (cn *Controller) CompleteCheck(c *gin.Context) {
	v, err := cn.service.CompleteCheck(c.Request.Context(), c.Param("id"), actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "outbound check completed", v)
}
func (cn *Controller) ListCheckExceptions(c *gin.Context) {
	v, err := cn.service.ListCheckExceptions(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "check exceptions retrieved", v)
}
func (cn *Controller) ResolveCheckException(c *gin.Context) {
	var q dto.ResolveCheckExceptionRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.ResolveCheckException(c.Request.Context(), c.Param("id"), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "check exception resolution recorded", v)
}
func (cn *Controller) CreateReplacementPick(c *gin.Context) {
	var q dto.CreateReplacementPickRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.CreateReplacementPick(c.Request.Context(), c.Param("id"), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "replacement pick created", v)
}
func (cn *Controller) CreatePacking(c *gin.Context) {
	var q dto.CreatePackingRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.CreatePacking(c.Request.Context(), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "packing document created", v)
}
func (cn *Controller) ListPackings(c *gin.Context) {
	f, _, ok := listFilter(c, false)
	if !ok {
		return
	}
	v, err := cn.service.ListPackings(c.Request.Context(), f)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "packing documents retrieved", v)
}
func (cn *Controller) GetPacking(c *gin.Context) {
	v, err := cn.service.GetPacking(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "packing document retrieved", v)
}
func (cn *Controller) PackLine(c *gin.Context) {
	var q dto.PackLineRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.PackLine(c.Request.Context(), c.Param("id"), c.Param("line_id"), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "checked stock packed", v)
}
func (cn *Controller) CompletePacking(c *gin.Context) {
	v, err := cn.service.CompletePacking(c.Request.Context(), c.Param("id"), actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "packing completed", v)
}
func (cn *Controller) CancelPacking(c *gin.Context) {
	v, err := cn.service.CancelPacking(c.Request.Context(), c.Param("id"), actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "empty packing cancelled", v)
}
func (cn *Controller) CreateShipment(c *gin.Context) {
	var q dto.CreateShipmentRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.CreateShipment(c.Request.Context(), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "shipment manifest created", v)
}
func (cn *Controller) ListShipments(c *gin.Context) {
	f, _, ok := listFilter(c, false)
	if !ok {
		return
	}
	v, err := cn.service.ListShipments(c.Request.Context(), f)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "shipments retrieved", v)
}
func (cn *Controller) GetShipment(c *gin.Context) {
	v, err := cn.service.GetShipment(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "shipment retrieved", v)
}
func (cn *Controller) DispatchShipmentLine(c *gin.Context) {
	var q dto.DispatchShipmentLineRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.DispatchShipmentLine(c.Request.Context(), c.Param("id"), c.Param("line_id"), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "packing line dispatched", v)
}
func (cn *Controller) CompleteShipment(c *gin.Context) {
	v, err := cn.service.CompleteShipment(c.Request.Context(), c.Param("id"), actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "shipment dispatched", v)
}
func (cn *Controller) CancelShipment(c *gin.Context) {
	v, err := cn.service.CancelShipment(c.Request.Context(), c.Param("id"), actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "empty shipment cancelled", v)
}
func (cn *Controller) CreateDelivery(c *gin.Context) {
	var q dto.CreateDeliveryRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.CreateDelivery(c.Request.Context(), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "delivery created", v)
}
func (cn *Controller) ListDeliveries(c *gin.Context) {
	f, _, ok := listFilter(c, false)
	if !ok {
		return
	}
	v, err := cn.service.ListDeliveries(c.Request.Context(), f)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "deliveries retrieved", v)
}
func (cn *Controller) GetDelivery(c *gin.Context) {
	v, err := cn.service.GetDelivery(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "delivery retrieved", v)
}
func (cn *Controller) DepartDelivery(c *gin.Context) {
	var q dto.DeliveryEventRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.DepartDelivery(c.Request.Context(), c.Param("id"), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "delivery departed", v)
}
func (cn *Controller) ArriveDelivery(c *gin.Context) {
	var q dto.DeliveryEventRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.ArriveDelivery(c.Request.Context(), c.Param("id"), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "delivery arrived", v)
}
func (cn *Controller) DeliverLine(c *gin.Context) {
	var q dto.DeliverLineRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.DeliverLine(c.Request.Context(), c.Param("id"), c.Param("line_id"), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "proof of delivery recorded", v)
}
func (cn *Controller) CompleteDelivery(c *gin.Context) {
	var q dto.CompleteDeliveryRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.CompleteDelivery(c.Request.Context(), c.Param("id"), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "delivery completed", v)
}
func (cn *Controller) FailDelivery(c *gin.Context) {
	var q dto.FailDeliveryRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.FailDelivery(c.Request.Context(), c.Param("id"), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "delivery failure recorded", v)
}
func (cn *Controller) CancelDelivery(c *gin.Context) {
	v, err := cn.service.CancelDelivery(c.Request.Context(), c.Param("id"), actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "planned delivery cancelled", v)
}
func (cn *Controller) ReturnDeliveryLine(c *gin.Context) {
	var q dto.ReturnDeliveryLineRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.ReturnDeliveryLine(c.Request.Context(), c.Param("id"), c.Param("line_id"), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "delivery stock returned to depot", v)
}
func (cn *Controller) CloseReturnedDelivery(c *gin.Context) {
	var q dto.DeliveryEventRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.CloseReturnedDelivery(c.Request.Context(), c.Param("id"), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "returned delivery closed", v)
}
func (cn *Controller) UpsertReturnPolicy(c *gin.Context) {
	var q dto.UpsertReturnPolicyRequest
	if !utils.BindStrictJSON(c, &q) {
		return
	}
	v, err := cn.service.UpsertReturnPolicy(c.Request.Context(), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "outbound return policy saved", v)
}
func (cn *Controller) GetReturnPolicy(c *gin.Context) {
	for key, values := range c.Request.URL.Query() {
		if (key != "owner_id" && key != "warehouse_id") || len(values) != 1 {
			utils.Failure(c, http.StatusBadRequest, "unknown or repeated query parameter: "+key, nil)
			return
		}
	}
	v, err := cn.service.GetReturnPolicy(c.Request.Context(), strings.TrimSpace(c.Query("owner_id")), strings.TrimSpace(c.Query("warehouse_id")))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "outbound return policy retrieved", v)
}
