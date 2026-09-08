package inbound

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	authdto "wms-api/dto/authentication"
	dto "wms-api/dto/inbound"
	"wms-api/middleware"
	repository "wms-api/repository/inbound"
	inventoryrepository "wms-api/repository/inventory"
	service "wms-api/services/inbound"
	inventoryservice "wms-api/services/inventory"
	"wms-api/utils"
)

type Controller struct{ service *service.Service }

func NewController(service *service.Service) *Controller { return &Controller{service: service} }
func actor(c *gin.Context) string {
	value, _ := c.Get(middleware.ContextUserKey)
	user, _ := value.(authdto.UserResponse)
	return user.AccountID
}
func fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput), errors.Is(err, inventoryservice.ErrInvalidInput), errors.Is(err, repository.ErrConstraint), errors.Is(err, inventoryrepository.ErrConstraint):
		utils.Failure(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, repository.ErrNotFound), errors.Is(err, inventoryrepository.ErrNotFound):
		utils.Failure(c, http.StatusNotFound, err.Error(), nil)
	case errors.Is(err, service.ErrInvalidState), errors.Is(err, repository.ErrConcurrentWrite), errors.Is(err, repository.ErrConflict), errors.Is(err, inventoryrepository.ErrConflict):
		utils.Failure(c, http.StatusConflict, err.Error(), nil)
	default:
		_ = c.Error(err)
		utils.Failure(c, http.StatusInternalServerError, "unable to process inbound request", nil)
	}
}
func listFilter(c *gin.Context) (repository.ListFilter, bool) {
	allowed := map[string]bool{"owner_id": true, "warehouse_id": true, "status_code": true, "search": true, "page": true, "page_size": true}
	for key, values := range c.Request.URL.Query() {
		if !allowed[key] || len(values) != 1 {
			utils.Failure(c, http.StatusBadRequest, "unknown or repeated query parameter: "+key, nil)
			return repository.ListFilter{}, false
		}
	}
	page, size, err := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	if err != nil {
		utils.Failure(c, http.StatusBadRequest, err.Error(), nil)
		return repository.ListFilter{}, false
	}
	return repository.ListFilter{OwnerID: strings.TrimSpace(c.Query("owner_id")), WarehouseID: strings.TrimSpace(c.Query("warehouse_id")), StatusCode: strings.TrimSpace(c.Query("status_code")), Search: strings.TrimSpace(c.Query("search")), Page: page, PageSize: size}, true
}

func (controller *Controller) CreatePurchaseOrder(c *gin.Context) {
	var request dto.CreatePurchaseOrderRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.CreatePurchaseOrder(c.Request.Context(), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "purchase order created", response)
}
func (controller *Controller) ListPurchaseOrders(c *gin.Context) {
	filter, ok := listFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListPurchaseOrders(c.Request.Context(), filter)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "purchase orders retrieved", response)
}
func (controller *Controller) GetPurchaseOrder(c *gin.Context) {
	response, err := controller.service.GetPurchaseOrder(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "purchase order retrieved", response)
}
func (controller *Controller) ApprovePurchaseOrder(c *gin.Context) {
	var request dto.TransitionRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.ApprovePurchaseOrder(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "purchase order approved", response)
}
func (controller *Controller) CreateInboundOrder(c *gin.Context) {
	var request dto.CreateInboundOrderRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.CreateInboundOrder(c.Request.Context(), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "inbound order created", response)
}
func (controller *Controller) ListInboundOrders(c *gin.Context) {
	filter, ok := listFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListInboundOrders(c.Request.Context(), filter)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "inbound orders retrieved", response)
}
func (controller *Controller) GetInboundOrder(c *gin.Context) {
	response, err := controller.service.GetInboundOrder(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "inbound order retrieved", response)
}
func (controller *Controller) ReleaseInboundOrder(c *gin.Context) {
	var request dto.TransitionRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.ReleaseInboundOrder(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "inbound order released", response)
}
func (controller *Controller) CreateReceipt(c *gin.Context) {
	var request dto.CreateReceiptRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.CreateReceipt(c.Request.Context(), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "receipt opened", response)
}
func (controller *Controller) ListReceipts(c *gin.Context) {
	filter, ok := listFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListReceipts(c.Request.Context(), filter)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "receipts retrieved", response)
}
func (controller *Controller) GetReceipt(c *gin.Context) {
	response, err := controller.service.GetReceipt(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "receipt retrieved", response)
}
func (controller *Controller) CompleteReceipt(c *gin.Context) {
	var request dto.TransitionRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.CompleteReceipt(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "receipt completed and inventory posted", response)
}

func (controller *Controller) CreateQualityInspection(c *gin.Context) {
	var request dto.CreateQualityInspectionRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.CreateQualityInspection(c.Request.Context(), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "quality inspection opened", response)
}
func (controller *Controller) ListQualityInspections(c *gin.Context) {
	filter, ok := listFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListQualityInspections(c.Request.Context(), filter)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "quality inspections retrieved", response)
}
func (controller *Controller) GetQualityInspection(c *gin.Context) {
	response, err := controller.service.GetQualityInspection(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "quality inspection retrieved", response)
}
func (controller *Controller) CompleteQualityInspection(c *gin.Context) {
	var request dto.CompleteQualityInspectionRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.CompleteQualityInspection(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "quality inspection completed", response)
}
func (controller *Controller) ListPutawayTasks(c *gin.Context) {
	filter, ok := listFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListPutawayTasks(c.Request.Context(), filter)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "putaway tasks retrieved", response)
}
func (controller *Controller) GetPutawayTask(c *gin.Context) {
	response, err := controller.service.GetPutawayTask(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "putaway task retrieved", response)
}
func (controller *Controller) StartPutawayTask(c *gin.Context) {
	var request dto.PutawayTransitionRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.StartPutawayTask(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "putaway task started", response)
}
func (controller *Controller) CompletePutawayTask(c *gin.Context) {
	var request dto.CompletePutawayRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.CompletePutawayTask(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "putaway task completed", response)
}
func (controller *Controller) ListQuarantineDispositionTypes(c *gin.Context) {
	var active *bool
	for key, values := range c.Request.URL.Query() {
		if key != "active" || len(values) != 1 || (values[0] != "true" && values[0] != "false") {
			utils.Failure(c, http.StatusBadRequest, "only active=true or active=false is allowed", nil)
			return
		}
		value := values[0] == "true"
		active = &value
	}
	response, err := controller.service.ListQuarantineDispositionTypes(c.Request.Context(), active)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "quarantine disposition types retrieved", response)
}
func (controller *Controller) ListQuarantineCases(c *gin.Context) {
	filter, ok := listFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListQuarantineCases(c.Request.Context(), filter)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "quarantine cases retrieved", response)
}
func (controller *Controller) GetQuarantineCase(c *gin.Context) {
	response, err := controller.service.GetQuarantineCase(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "quarantine case retrieved", response)
}
func (controller *Controller) CreateQuarantineDisposition(c *gin.Context) {
	var request dto.CreateQuarantineDispositionRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.service.CreateQuarantineDisposition(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "quarantine disposition processed", response)
}
