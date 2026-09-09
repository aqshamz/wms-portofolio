package outbound

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	authdto "wms-api/dto/authentication"
	dto "wms-api/dto/outbound"
	"wms-api/middleware"
	inventoryrepository "wms-api/repository/inventory"
	repository "wms-api/repository/outbound"
	service "wms-api/services/outbound"
	"wms-api/utils"
)

type Controller struct{ service *service.Service }

func NewController(service *service.Service) *Controller { return &Controller{service: service} }
func actor(c *gin.Context) string {
	v, _ := c.Get(middleware.ContextUserKey)
	user, _ := v.(authdto.UserResponse)
	return user.AccountID
}
func fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput), errors.Is(err, repository.ErrConstraint), errors.Is(err, inventoryrepository.ErrConstraint):
		utils.Failure(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, repository.ErrNotFound), errors.Is(err, inventoryrepository.ErrNotFound):
		utils.Failure(c, http.StatusNotFound, err.Error(), nil)
	case errors.Is(err, service.ErrInvalidState), errors.Is(err, repository.ErrConcurrentWrite), errors.Is(err, repository.ErrConflict), errors.Is(err, inventoryrepository.ErrConflict):
		utils.Failure(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, service.ErrForbidden):
		utils.Failure(c, http.StatusForbidden, err.Error(), nil)
	default:
		_ = c.Error(err)
		utils.Failure(c, http.StatusInternalServerError, "unable to process outbound request", nil)
	}
}
func listFilter(c *gin.Context, allowAssignee bool) (repository.ListFilter, string, bool) {
	allowed := map[string]bool{"owner_id": true, "warehouse_id": true, "status_code": true, "search": true, "page": true, "page_size": true}
	if allowAssignee {
		allowed["assignee_id"] = true
	}
	for key, values := range c.Request.URL.Query() {
		if !allowed[key] || len(values) != 1 {
			utils.Failure(c, http.StatusBadRequest, "unknown or repeated query parameter: "+key, nil)
			return repository.ListFilter{}, "", false
		}
	}
	p, size, err := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	if err != nil {
		utils.Failure(c, http.StatusBadRequest, err.Error(), nil)
		return repository.ListFilter{}, "", false
	}
	return repository.ListFilter{OwnerID: strings.TrimSpace(c.Query("owner_id")), WarehouseID: strings.TrimSpace(c.Query("warehouse_id")), StatusCode: strings.TrimSpace(c.Query("status_code")), Search: strings.TrimSpace(c.Query("search")), Page: p, PageSize: size}, strings.TrimSpace(c.Query("assignee_id")), true
}

func (cn *Controller) CreateOrder(c *gin.Context) {
	var r dto.CreateOutboundOrderRequest
	if !utils.BindStrictJSON(c, &r) {
		return
	}
	v, err := cn.service.CreateOutboundOrder(c.Request.Context(), r, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "outbound order created", v)
}
func (cn *Controller) UpdateOrder(c *gin.Context) {
	var r dto.UpdateOutboundOrderRequest
	if !utils.BindStrictJSON(c, &r) {
		return
	}
	v, err := cn.service.UpdateOutboundOrder(c.Request.Context(), c.Param("id"), r, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "draft outbound order updated", v)
}
func (cn *Controller) AddOrderLine(c *gin.Context) {
	var r dto.AddOutboundOrderLineRequest
	if !utils.BindStrictJSON(c, &r) {
		return
	}
	v, err := cn.service.AddOutboundOrderLine(c.Request.Context(), c.Param("id"), r, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "draft outbound line added", v)
}
func (cn *Controller) UpdateOrderLine(c *gin.Context) {
	var r dto.UpdateOutboundOrderLineRequest
	if !utils.BindStrictJSON(c, &r) {
		return
	}
	v, err := cn.service.UpdateOutboundOrderLine(c.Request.Context(), c.Param("id"), c.Param("line_id"), r, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "draft outbound line updated", v)
}
func (cn *Controller) DeleteOrderLine(c *gin.Context) {
	var r dto.DeleteOutboundOrderLineRequest
	if !utils.BindStrictJSON(c, &r) {
		return
	}
	v, err := cn.service.DeleteOutboundOrderLine(c.Request.Context(), c.Param("id"), c.Param("line_id"), r.ExpectedVersion, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "draft outbound line deleted", v)
}
func (cn *Controller) ListOrders(c *gin.Context) {
	f, _, ok := listFilter(c, false)
	if !ok {
		return
	}
	v, err := cn.service.ListOutboundOrders(c.Request.Context(), f)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "outbound orders retrieved", v)
}
func (cn *Controller) GetOrder(c *gin.Context) {
	v, err := cn.service.GetOutboundOrder(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "outbound order retrieved", v)
}
func (cn *Controller) ValidateOrder(c *gin.Context) {
	var r dto.ValidateOutboundRequest
	if !utils.BindStrictJSON(c, &r) {
		return
	}
	v, err := cn.service.ValidateOutboundOrder(c.Request.Context(), c.Param("id"), r, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "outbound validation completed", v)
}
func (cn *Controller) GetValidationRun(c *gin.Context) {
	v, err := cn.service.GetValidationRun(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "outbound validation retrieved", v)
}
func (cn *Controller) ReleaseOrder(c *gin.Context) {
	var r dto.TransitionRequest
	if !utils.BindStrictJSON(c, &r) {
		return
	}
	v, err := cn.service.ReleaseOutboundOrder(c.Request.Context(), c.Param("id"), r, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "outbound order released", v)
}
func (cn *Controller) CancelOrder(c *gin.Context) {
	var request dto.CancelOrderRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := cn.service.CancelOutboundOrder(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "outbound order cancelled", response)
}
func (cn *Controller) AllocateOrder(c *gin.Context) {
	var r dto.AllocateOutboundRequest
	if !utils.BindStrictJSON(c, &r) {
		return
	}
	v, err := cn.service.AllocateOutboundOrder(c.Request.Context(), c.Param("id"), r, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "outbound allocation completed", v)
}
func (cn *Controller) ListReservations(c *gin.Context) {
	f, _, ok := listFilter(c, false)
	if !ok {
		return
	}
	v, err := cn.service.ListReservations(c.Request.Context(), f)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "reservations retrieved", v)
}
func (cn *Controller) ReleaseReservation(c *gin.Context) {
	var r dto.ReleaseReservationRequest
	if !utils.BindStrictJSON(c, &r) {
		return
	}
	v, err := cn.service.ReleaseReservation(c.Request.Context(), c.Param("id"), r, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "reservation released", v)
}
func (cn *Controller) CreateWave(c *gin.Context) {
	var r dto.CreateWaveRequest
	if !utils.BindStrictJSON(c, &r) {
		return
	}
	v, err := cn.service.CreateWave(c.Request.Context(), r, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "outbound wave created", v)
}
func (cn *Controller) ListWaves(c *gin.Context) {
	f, _, ok := listFilter(c, false)
	if !ok {
		return
	}
	v, err := cn.service.ListWaves(c.Request.Context(), f)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "outbound waves retrieved", v)
}
func (cn *Controller) GetWave(c *gin.Context) {
	v, err := cn.service.GetWave(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "outbound wave retrieved", v)
}
func (cn *Controller) ReleaseWave(c *gin.Context) {
	var r dto.ReleaseWaveRequest
	if !utils.BindStrictJSON(c, &r) {
		return
	}
	v, err := cn.service.ReleaseWave(c.Request.Context(), c.Param("id"), r, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "outbound wave released", v)
}
func (cn *Controller) CancelWave(c *gin.Context) {
	var r dto.TransitionRequest
	if !utils.BindStrictJSON(c, &r) {
		return
	}
	v, err := cn.service.CancelWave(c.Request.Context(), c.Param("id"), r, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "outbound wave cancelled", v)
}
func (cn *Controller) ListPickTasks(c *gin.Context) {
	f, a, ok := listFilter(c, true)
	if !ok {
		return
	}
	v, err := cn.service.ListPickTasks(c.Request.Context(), f, a)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "pick tasks retrieved", v)
}
func (cn *Controller) GetPickTask(c *gin.Context) {
	v, err := cn.service.GetPickTask(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "pick task retrieved", v)
}
func (cn *Controller) StartPickTask(c *gin.Context) {
	v, err := cn.service.StartPickTask(c.Request.Context(), c.Param("id"), actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "pick task started", v)
}
func (cn *Controller) ConfirmPick(c *gin.Context) {
	var r dto.ConfirmPickRequest
	if !utils.BindStrictJSON(c, &r) {
		return
	}
	v, err := cn.service.ConfirmPick(c.Request.Context(), c.Param("id"), r, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "pick confirmed into staging", v)
}
func (cn *Controller) CloseShortPick(c *gin.Context) {
	var r dto.CloseShortPickRequest
	if !utils.BindStrictJSON(c, &r) {
		return
	}
	v, err := cn.service.CloseShortPick(c.Request.Context(), c.Param("id"), r, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "short pick closed", v)
}
func (cn *Controller) CreateStaging(c *gin.Context) {
	var r dto.CreateStagingRequest
	if !utils.BindStrictJSON(c, &r) {
		return
	}
	v, err := cn.service.CreateStaging(c.Request.Context(), r, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "staging document created", v)
}
func (cn *Controller) ListStagings(c *gin.Context) {
	f, _, ok := listFilter(c, false)
	if !ok {
		return
	}
	v, err := cn.service.ListStagings(c.Request.Context(), f)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "staging documents retrieved", v)
}
func (cn *Controller) GetStaging(c *gin.Context) {
	v, err := cn.service.GetStaging(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "staging document retrieved", v)
}
func (cn *Controller) CompleteStaging(c *gin.Context) {
	v, err := cn.service.CompleteStaging(c.Request.Context(), c.Param("id"), actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "outbound stock confirmed on staging", v)
}
