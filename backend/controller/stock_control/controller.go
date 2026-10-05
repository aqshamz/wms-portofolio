package stockcontrol

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"io"
	"net/http"
	"strconv"
	authdto "wms-api/dto/authentication"
	dto "wms-api/dto/stock_control"
	"wms-api/middleware"
	inventoryrepo "wms-api/repository/inventory"
	repository "wms-api/repository/stock_control"
	service "wms-api/services/stock_control"
	"wms-api/utils"
)

type Controller struct{ service *service.Service }

func NewController(s *service.Service) *Controller { return &Controller{service: s} }
func bind(c *gin.Context, request interface{}) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(request); err != nil {
		utils.Failure(c, 400, "invalid JSON or unknown field", nil)
		return false
	}
	var extra interface{}
	if decoder.Decode(&extra) != io.EOF || binding.Validator.ValidateStruct(request) != nil {
		utils.Failure(c, 400, "invalid request fields", nil)
		return false
	}
	return true
}

func replenishmentListFilter(c *gin.Context) (repository.ReplenishmentFilter, bool) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		utils.Failure(c, 400, "invalid page", nil)
		return repository.ReplenishmentFilter{}, false
	}
	size, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil {
		utils.Failure(c, 400, "invalid page_size", nil)
		return repository.ReplenishmentFilter{}, false
	}
	allowed := map[string]bool{"owner_id": true, "warehouse_id": true, "status_code": true, "search": true, "page": true, "page_size": true}
	for key, values := range c.Request.URL.Query() {
		if !allowed[key] || len(values) != 1 {
			utils.Failure(c, 400, "unsupported or repeated query parameter", nil)
			return repository.ReplenishmentFilter{}, false
		}
	}
	return repository.ReplenishmentFilter{OwnerID: c.Query("owner_id"), WarehouseID: c.Query("warehouse_id"), StatusCode: c.Query("status_code"), Search: c.Query("search"), Page: page, PageSize: size}, true
}

func replenishmentLookup(c *gin.Context) (string, int, int, bool) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		utils.Failure(c, 400, "invalid page", nil)
		return "", 0, 0, false
	}
	size, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil {
		utils.Failure(c, 400, "invalid page_size", nil)
		return "", 0, 0, false
	}
	for key, values := range c.Request.URL.Query() {
		if (key != "search" && key != "page" && key != "page_size") || len(values) != 1 {
			utils.Failure(c, 400, "unsupported or repeated query parameter", nil)
			return "", 0, 0, false
		}
	}
	return c.Query("search"), page, size, true
}

func adjustmentListFilter(c *gin.Context) (repository.AdjustmentFilter, bool) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		utils.Failure(c, 400, "invalid page", nil)
		return repository.AdjustmentFilter{}, false
	}
	size, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil {
		utils.Failure(c, 400, "invalid page_size", nil)
		return repository.AdjustmentFilter{}, false
	}
	allowed := map[string]bool{"owner_id": true, "warehouse_id": true, "status_code": true, "search": true, "page": true, "page_size": true}
	for key, values := range c.Request.URL.Query() {
		if !allowed[key] || len(values) != 1 {
			utils.Failure(c, 400, "unsupported or repeated query parameter", nil)
			return repository.AdjustmentFilter{}, false
		}
	}
	return repository.AdjustmentFilter{OwnerID: c.Query("owner_id"), WarehouseID: c.Query("warehouse_id"), StatusCode: c.Query("status_code"), Search: c.Query("search"), Page: page, PageSize: size}, true
}

func cycleCountListFilter(c *gin.Context) (repository.CycleCountFilter, bool) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		utils.Failure(c, 400, "invalid page", nil)
		return repository.CycleCountFilter{}, false
	}
	size, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil {
		utils.Failure(c, 400, "invalid page_size", nil)
		return repository.CycleCountFilter{}, false
	}
	allowed := map[string]bool{"owner_id": true, "warehouse_id": true, "count_type_code": true, "status_code": true, "search": true, "page": true, "page_size": true}
	for key, values := range c.Request.URL.Query() {
		if !allowed[key] || len(values) != 1 {
			utils.Failure(c, 400, "unsupported or repeated query parameter", nil)
			return repository.CycleCountFilter{}, false
		}
	}
	return repository.CycleCountFilter{OwnerID: c.Query("owner_id"), WarehouseID: c.Query("warehouse_id"), CountTypeCode: c.Query("count_type_code"), StatusCode: c.Query("status_code"), Search: c.Query("search"), Page: page, PageSize: size}, true
}

func warehouseTransferListFilter(c *gin.Context) (repository.WarehouseTransferFilter, bool) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil { utils.Failure(c, 400, "invalid page", nil); return repository.WarehouseTransferFilter{}, false }
	size, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil { utils.Failure(c, 400, "invalid page_size", nil); return repository.WarehouseTransferFilter{}, false }
	allowed := map[string]bool{"owner_id": true, "warehouse_id": true, "side": true, "status_code": true, "search": true, "page": true, "page_size": true}
	for key, values := range c.Request.URL.Query() {
		if !allowed[key] || len(values) != 1 { utils.Failure(c, 400, "unsupported or repeated query parameter", nil); return repository.WarehouseTransferFilter{}, false }
	}
	return repository.WarehouseTransferFilter{OwnerID: c.Query("owner_id"), WarehouseID: c.Query("warehouse_id"), Side: c.Query("side"), StatusCode: c.Query("status_code"), Search: c.Query("search"), Page: page, PageSize: size}, true
}

func replenishmentTargetLookup(c *gin.Context) (string, int, int, bool) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		utils.Failure(c, 400, "invalid page", nil)
		return "", 0, 0, false
	}
	size, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil {
		utils.Failure(c, 400, "invalid page_size", nil)
		return "", 0, 0, false
	}
	allowed := map[string]bool{"owner_id": true, "warehouse_id": true, "search": true, "page": true, "page_size": true}
	for key, values := range c.Request.URL.Query() {
		if !allowed[key] || len(values) != 1 {
			utils.Failure(c, 400, "unsupported or repeated query parameter", nil)
			return "", 0, 0, false
		}
	}
	return c.Query("search"), page, size, true
}

func (controller *Controller) ListReplenishments(c *gin.Context) {
	filter, ok := replenishmentListFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListReplenishments(c.Request.Context(), filter, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "replenishment tasks retrieved", response)
}

func (controller *Controller) GetReplenishment(c *gin.Context) {
	response, err := controller.service.GetReplenishment(c.Request.Context(), c.Param("id"), actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "replenishment task retrieved", response)
}

func (controller *Controller) CreateReplenishment(c *gin.Context) {
	var request dto.CreateReplenishmentRequest
	if !bind(c, &request) {
		return
	}
	response, err := controller.service.CreateReplenishment(c.Request.Context(), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "replenishment task created", response)
}

func (controller *Controller) StartReplenishment(c *gin.Context) {
	var request dto.ReplenishmentTransitionRequest
	if !bind(c, &request) {
		return
	}
	response, err := controller.service.StartReplenishment(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "replenishment task started", response)
}

func (controller *Controller) AssignReplenishment(c *gin.Context) {
	var request dto.AssignReplenishmentRequest
	if !bind(c, &request) {
		return
	}
	response, err := controller.service.AssignReplenishment(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "replenishment task assigned", response)
}

func (controller *Controller) CompleteReplenishment(c *gin.Context) {
	var request dto.CompleteReplenishmentRequest
	if !bind(c, &request) {
		return
	}
	response, err := controller.service.CompleteReplenishment(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "replenishment task completed", response)
}

func (controller *Controller) CancelReplenishment(c *gin.Context) {
	var request dto.CancelReplenishmentRequest
	if !bind(c, &request) {
		return
	}
	response, err := controller.service.CancelReplenishment(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "replenishment task cancelled", response)
}

func (controller *Controller) ListReplenishmentAssignees(c *gin.Context) {
	search, page, size, ok := replenishmentLookup(c)
	if !ok {
		return
	}
	response, err := controller.service.ListReplenishmentAssignees(c.Request.Context(), c.Param("id"), search, page, size, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "replenishment assignees retrieved", response)
}

func (controller *Controller) ListReplenishmentTargets(c *gin.Context) {
	search, page, size, ok := replenishmentTargetLookup(c)
	if !ok {
		return
	}
	response, err := controller.service.ListReplenishmentTargets(c.Request.Context(), c.Query("owner_id"), c.Query("warehouse_id"), search, page, size, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "eligible replenishment targets retrieved", response)
}
func actor(c *gin.Context) string {
	value, _ := c.Get(middleware.ContextUserKey)
	user, _ := value.(authdto.UserResponse)
	return user.AccountID
}
func fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrForbidden):
		utils.Failure(c, 403, err.Error(), nil)
	case errors.Is(err, service.ErrInvalidInput):
		utils.Failure(c, 400, err.Error(), nil)
	case errors.Is(err, inventoryrepo.ErrConstraint):
		utils.Failure(c, 400, "stock command references invalid or incompatible data", nil)
	case errors.Is(err, inventoryrepo.ErrNotFound):
		utils.Failure(c, 404, err.Error(), nil)
	case errors.Is(err, inventoryrepo.ErrConflict):
		utils.Failure(c, 409, "stock command conflicts with existing data", nil)
	default:
		_ = c.Error(err)
		utils.Failure(c, 500, "unable to process stock-control command", nil)
	}
}
func (controller *Controller) InternalMove(c *gin.Context) {
	var q dto.InternalMoveRequest
	if !bind(c, &q) {
		return
	}
	r, err := controller.service.InternalMove(c.Request.Context(), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "internal move posted", r)
}
func (controller *Controller) StatusChange(c *gin.Context) {
	var q dto.StatusChangeRequest
	if !bind(c, &q) {
		return
	}
	r, err := controller.service.StatusChange(c.Request.Context(), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "status change posted", r)
}
func (controller *Controller) Adjustment(c *gin.Context) {
	var q dto.AdjustmentRequest
	if !bind(c, &q) {
		return
	}
	r, err := controller.service.Adjustment(c.Request.Context(), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "adjustment posted", r)
}

func (controller *Controller) ListAdjustments(c *gin.Context) {
	filter, ok := adjustmentListFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListAdjustments(c.Request.Context(), filter, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "inventory adjustments retrieved", response)
}

func (controller *Controller) GetAdjustment(c *gin.Context) {
	response, err := controller.service.GetAdjustment(c.Request.Context(), c.Param("id"), actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "inventory adjustment retrieved", response)
}

func (controller *Controller) CreateAdjustment(c *gin.Context) {
	var request dto.CreateAdjustmentRequest
	if !bind(c, &request) {
		return
	}
	response, err := controller.service.CreateAdjustment(c.Request.Context(), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "inventory adjustment requested", response)
}

func (controller *Controller) ApproveAdjustment(c *gin.Context) {
	var request dto.AdjustmentLineSelectionRequest
	if !bind(c, &request) {
		return
	}
	response, err := controller.service.ApproveAdjustment(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "inventory adjustment approved and posted", response)
}

func (controller *Controller) RejectAdjustment(c *gin.Context) {
	var request dto.RejectAdjustmentLinesRequest
	if !bind(c, &request) {
		return
	}
	response, err := controller.service.RejectAdjustment(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "inventory adjustment rejected", response)
}

func (controller *Controller) CancelAdjustment(c *gin.Context) {
	var request dto.CancelAdjustmentRequest
	if !bind(c, &request) {
		return
	}
	response, err := controller.service.CancelAdjustment(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "inventory adjustment cancelled", response)
}

func (controller *Controller) ListCycleCounts(c *gin.Context) {
	filter, ok := cycleCountListFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListCycleCounts(c.Request.Context(), filter, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "cycle counts retrieved", response)
}
func (controller *Controller) GetCycleCount(c *gin.Context) {
	response, err := controller.service.GetCycleCount(c.Request.Context(), c.Param("id"), actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "cycle count retrieved", response)
}
func (controller *Controller) CreateCycleCount(c *gin.Context) {
	var request dto.CreateCycleCountRequest
	if !bind(c, &request) {
		return
	}
	response, err := controller.service.CreateCycleCount(c.Request.Context(), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "cycle count created", response)
}
func (controller *Controller) CreateGrandStockOpname(c *gin.Context) {
	var request dto.CreateGrandStockOpnameRequest
	if !bind(c, &request) {
		return
	}
	response, err := controller.service.CreateGrandStockOpname(c.Request.Context(), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "grand stock opname created", response)
}
func (controller *Controller) RecordCycleCount(c *gin.Context) {
	var request dto.RecordCycleCountRequest
	if !bind(c, &request) {
		return
	}
	response, err := controller.service.RecordCycleCount(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "cycle count entries recorded", response)
}
func (controller *Controller) ApproveCycleCount(c *gin.Context) {
	var request dto.CycleCountDecisionRequest
	if !bind(c, &request) {
		return
	}
	response, err := controller.service.ApproveCycleCount(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "cycle count lines approved", response)
}
func (controller *Controller) RejectCycleCount(c *gin.Context) {
	var request dto.RejectCycleCountRequest
	if !bind(c, &request) {
		return
	}
	response, err := controller.service.RejectCycleCount(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "cycle count lines rejected", response)
}
func (controller *Controller) CancelCycleCount(c *gin.Context) {
	var request dto.CancelCycleCountRequest
	if !bind(c, &request) {
		return
	}
	response, err := controller.service.CancelCycleCount(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "cycle count cancelled", response)
}
func (controller *Controller) ReconcileCount(c *gin.Context) {
	var q dto.StockCountReconcileRequest
	if !bind(c, &q) {
		return
	}
	r, err := controller.service.ReconcileCount(c.Request.Context(), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "stock count reconciled", r)
}
func (controller *Controller) WarehouseTransfer(c *gin.Context) {
	var q dto.WarehouseTransferRequest
	if !bind(c, &q) {
		return
	}
	r, err := controller.service.WarehouseTransfer(c.Request.Context(), q, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "warehouse transfer posted", r)
}

func (controller *Controller) ListWarehouseTransfers(c *gin.Context) {
	filter, ok := warehouseTransferListFilter(c); if !ok { return }
	response, err := controller.service.ListWarehouseTransfers(c.Request.Context(), filter, actor(c))
	if err != nil { fail(c, err); return }
	utils.Success(c, http.StatusOK, "warehouse transfers retrieved", response)
}
func (controller *Controller) GetWarehouseTransfer(c *gin.Context) {
	response, err := controller.service.GetWarehouseTransfer(c.Request.Context(), c.Param("id"), actor(c))
	if err != nil { fail(c, err); return }
	utils.Success(c, http.StatusOK, "warehouse transfer retrieved", response)
}
func (controller *Controller) CreateWarehouseTransfer(c *gin.Context) {
	var request dto.CreateWarehouseTransferRequest; if !bind(c, &request) { return }
	response, err := controller.service.CreateWarehouseTransfer(c.Request.Context(), request, actor(c))
	if err != nil { fail(c, err); return }
	utils.Success(c, http.StatusCreated, "warehouse transfer created", response)
}
func (controller *Controller) ApproveWarehouseTransfer(c *gin.Context) {
	var request dto.WarehouseTransferTransitionRequest; if !bind(c, &request) { return }
	response, err := controller.service.ApproveWarehouseTransfer(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil { fail(c, err); return }
	utils.Success(c, http.StatusOK, "warehouse transfer approved", response)
}
func (controller *Controller) DispatchWarehouseTransfer(c *gin.Context) {
	var request dto.WarehouseTransferTransitionRequest; if !bind(c, &request) { return }
	response, err := controller.service.DispatchWarehouseTransfer(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil { fail(c, err); return }
	utils.Success(c, http.StatusOK, "warehouse transfer dispatched", response)
}
func (controller *Controller) CancelWarehouseTransfer(c *gin.Context) {
	var request dto.CancelWarehouseTransferRequest; if !bind(c, &request) { return }
	response, err := controller.service.CancelWarehouseTransfer(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil { fail(c, err); return }
	utils.Success(c, http.StatusOK, "warehouse transfer cancelled", response)
}
func (controller *Controller) ReceiveWarehouseTransfer(c *gin.Context) {
	var request dto.ReceiveWarehouseTransferRequest; if !bind(c, &request) { return }
	response, err := controller.service.ReceiveWarehouseTransfer(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil { fail(c, err); return }
	utils.Success(c, http.StatusOK, "warehouse transfer received", response)
}
func (controller *Controller) PutawayWarehouseTransfer(c *gin.Context) {
	var request dto.PutawayWarehouseTransferRequest; if !bind(c, &request) { return }
	response, err := controller.service.PutawayWarehouseTransfer(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil { fail(c, err); return }
	utils.Success(c, http.StatusOK, "warehouse transfer put away", response)
}

func (controller *Controller) ListReasons(c *gin.Context) {
	var active *bool
	for key, values := range c.Request.URL.Query() {
		if key != "active" || len(values) != 1 || (values[0] != "true" && values[0] != "false") {
			utils.Failure(c, http.StatusBadRequest, "only active=true|false is supported", nil)
			return
		}
		value := values[0] == "true"
		active = &value
	}
	response, err := controller.service.ListReasons(c.Request.Context(), active)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "stock-control reasons retrieved", response)
}
