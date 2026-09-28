package outbound

import (
	"net/http"

	"github.com/gin-gonic/gin"
	dto "wms-api/dto/outbound"
	"wms-api/utils"
)

func (cn *Controller) ListDisposals(c *gin.Context) {
	filter, _, ok := listFilter(c, false)
	if !ok {
		return
	}
	value, err := cn.service.ListDisposals(c.Request.Context(), filter)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "disposal transactions retrieved", value)
}

func (cn *Controller) GetDisposal(c *gin.Context) {
	value, err := cn.service.GetDisposal(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "disposal transaction retrieved", value)
}

func (cn *Controller) CompleteDisposal(c *gin.Context) {
	var request dto.CompleteDisposalRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	value, err := cn.service.CompleteDisposal(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "disposal completed and inventory removed", value)
}

func (cn *Controller) CancelDisposal(c *gin.Context) {
	var request dto.CancelDisposalRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	value, err := cn.service.CancelDisposal(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "disposal cancelled", value)
}
