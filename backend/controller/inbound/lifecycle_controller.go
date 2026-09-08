package inbound

import (
	"net/http"

	"github.com/gin-gonic/gin"
	dto "wms-api/dto/inbound"
	"wms-api/utils"
)

func (controller *Controller) ClosePurchaseOrder(c *gin.Context) {
	var request dto.ExceptionTransitionRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.ClosePurchaseOrder(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "purchase order closed", response)
}

func (controller *Controller) CancelPurchaseOrder(c *gin.Context) {
	var request dto.ExceptionTransitionRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.CancelPurchaseOrder(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "purchase order cancelled", response)
}

func (controller *Controller) CloseInboundOrder(c *gin.Context) {
	var request dto.ExceptionTransitionRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.CloseInboundOrder(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "inbound order closed", response)
}

func (controller *Controller) CancelInboundOrder(c *gin.Context) {
	var request dto.ExceptionTransitionRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.CancelInboundOrder(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "inbound order cancelled", response)
}

func (controller *Controller) CancelReceipt(c *gin.Context) {
	var request dto.ExceptionTransitionRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.CancelReceipt(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "receipt cancelled", response)
}

func (controller *Controller) ReverseReceipt(c *gin.Context) {
	var request dto.ReverseReceiptRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.ReverseReceipt(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "receipt reversed", response)
}

func (controller *Controller) CancelQualityInspection(c *gin.Context) {
	var request dto.ExceptionTransitionRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.CancelQualityInspection(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "quality inspection cancelled and replaced", response)
}

func (controller *Controller) AssignPutawayTask(c *gin.Context) {
	var request dto.AssignPutawayRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.AssignPutawayTask(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "putaway task assigned", response)
}

func (controller *Controller) RetargetPutawayTask(c *gin.Context) {
	var request dto.RetargetPutawayRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.RetargetPutawayTask(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "putaway task retargeted", response)
}

func (controller *Controller) CancelPutawayTask(c *gin.Context) {
	var request dto.CancelPutawayRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.CancelPutawayTask(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "putaway task cancelled and stock returned to QC", response)
}

func (controller *Controller) ReversePutawayTask(c *gin.Context) {
	var request dto.CancelPutawayRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.ReversePutawayTask(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "putaway task reversed and stock returned to QC", response)
}

func (controller *Controller) GetInboundException(c *gin.Context) {
	response, err := controller.service.GetInboundException(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "inbound exception retrieved", response)
}

func (controller *Controller) ListInboundExceptions(c *gin.Context) {
	filter, ok := listFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListInboundExceptions(c.Request.Context(), filter)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "inbound exceptions retrieved", response)
}

func (controller *Controller) GetReworkTask(c *gin.Context) {
	response, err := controller.service.GetReworkTask(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "rework task retrieved", response)
}

func (controller *Controller) ListReworkTasks(c *gin.Context) {
	filter, ok := listFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListReworkTasks(c.Request.Context(), filter)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "rework tasks retrieved", response)
}

func (controller *Controller) StartReworkTask(c *gin.Context) {
	var request dto.ReworkTransitionRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.StartReworkTask(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "rework task started", response)
}

func (controller *Controller) CompleteReworkTask(c *gin.Context) {
	var request dto.CompleteReworkRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.CompleteReworkTask(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "rework task completed and reinspection opened", response)
}
