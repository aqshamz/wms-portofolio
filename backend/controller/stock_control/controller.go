package stockcontrol

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"io"
	"net/http"
	authdto "wms-api/dto/authentication"
	dto "wms-api/dto/stock_control"
	"wms-api/middleware"
	inventoryrepo "wms-api/repository/inventory"
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
func actor(c *gin.Context) string {
	value, _ := c.Get(middleware.ContextUserKey)
	user, _ := value.(authdto.UserResponse)
	return user.AccountID
}
func fail(c *gin.Context, err error) {
	switch {
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
