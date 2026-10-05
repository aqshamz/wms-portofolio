package outbound

import (
	"net/http"

	"github.com/gin-gonic/gin"
	dto "wms-api/dto/outbound"
	"wms-api/utils"
)

func (cn *Controller) ListVendorReturns(c *gin.Context) {
	filter, _, ok := listFilter(c, false)
	if !ok {
		return
	}
	value, err := cn.service.ListVendorReturns(c.Request.Context(), filter)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "return to vendor transactions retrieved", value)
}

func (cn *Controller) GetVendorReturn(c *gin.Context) {
	value, err := cn.service.GetVendorReturn(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "return to vendor transaction retrieved", value)
}

func (cn *Controller) CompleteVendorReturn(c *gin.Context) {
	var request dto.CompleteVendorReturnRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	value, err := cn.service.CompleteVendorReturn(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "return to vendor dispatched and inventory removed", value)
}

func (cn *Controller) CancelVendorReturn(c *gin.Context) {
	var request dto.CancelVendorReturnRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	value, err := cn.service.CancelVendorReturn(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "return to vendor cancelled", value)
}
