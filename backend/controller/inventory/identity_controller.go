package inventory

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"io"
	"net/http"
	"strconv"
	authdto "wms-api/dto/authentication"
	dto "wms-api/dto/inventory"
	"wms-api/middleware"
	repository "wms-api/repository/inventory"
	service "wms-api/services/inventory"
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
	case errors.Is(err, repository.ErrConstraint):
		utils.Failure(c, 400, "identity references invalid or incompatible data", nil)
	case errors.Is(err, repository.ErrNotFound):
		utils.Failure(c, 404, err.Error(), nil)
	case errors.Is(err, repository.ErrConflict):
		utils.Failure(c, 409, "identity already exists; use GET with its exact number/barcode", nil)
	default:
		_ = c.Error(err)
		utils.Failure(c, 500, "unable to process inventory identity", nil)
	}
}
func filter(c *gin.Context, numberKey string, hu bool) (repository.Filter, bool) {
	allowed := map[string]bool{"page": true, "page_size": true, "search": true, "owner_id": true, numberKey: true}
	if hu {
		for _, key := range []string{"warehouse_id", "current_location_id", "parent_handling_unit_id", "closed"} {
			allowed[key] = true
		}
	} else {
		allowed["item_id"] = true
	}
	for key, values := range c.Request.URL.Query() {
		if !allowed[key] || len(values) != 1 {
			utils.Failure(c, 400, "unknown or repeated query parameter: "+key, nil)
			return repository.Filter{}, false
		}
	}
	page, size, err := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	if err != nil {
		utils.Failure(c, 400, err.Error(), nil)
		return repository.Filter{}, false
	}
	f := repository.Filter{Page: page, PageSize: size, OwnerID: c.Query("owner_id"), ItemID: c.Query("item_id"), WarehouseID: c.Query("warehouse_id"), LocationID: c.Query("current_location_id"), ParentID: c.Query("parent_handling_unit_id"), Search: c.Query("search"), Number: c.Query(numberKey)}
	if values, exists := c.Request.URL.Query()["closed"]; exists {
		v := values[0]
		if v != "true" && v != "false" {
			utils.Failure(c, 400, "closed must be true or false", nil)
			return f, false
		}
		b, _ := strconv.ParseBool(v)
		f.Closed = &b
	}
	return f, true
}
func (controller *Controller) CreateLot(c *gin.Context) {
	var request dto.CreateLotRequest
	if !bind(c, &request) {
		return
	}
	response, err := controller.service.CreateLot(c.Request.Context(), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "inventory identity created", response)
}
func (controller *Controller) GetLot(c *gin.Context) {
	response, err := controller.service.GetLot(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "inventory identity found", response)
}
func (controller *Controller) ListLot(c *gin.Context) {
	f, ok := filter(c, "lot_number", false)
	if !ok {
		return
	}
	response, err := controller.service.ListLot(c.Request.Context(), f)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "inventory identities retrieved", response)
}
func (controller *Controller) CreateSerial(c *gin.Context) {
	var request dto.CreateSerialRequest
	if !bind(c, &request) {
		return
	}
	response, err := controller.service.CreateSerial(c.Request.Context(), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "inventory identity created", response)
}
func (controller *Controller) GetSerial(c *gin.Context) {
	response, err := controller.service.GetSerial(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "inventory identity found", response)
}
func (controller *Controller) ListSerial(c *gin.Context) {
	f, ok := filter(c, "serial_no", false)
	if !ok {
		return
	}
	response, err := controller.service.ListSerial(c.Request.Context(), f)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "inventory identities retrieved", response)
}
func (controller *Controller) CreateHandlingUnit(c *gin.Context) {
	var request dto.CreateHandlingUnitRequest
	if !bind(c, &request) {
		return
	}
	response, err := controller.service.CreateHandlingUnit(c.Request.Context(), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "inventory identity created", response)
}
func (controller *Controller) GetHandlingUnit(c *gin.Context) {
	response, err := controller.service.GetHandlingUnit(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "inventory identity found", response)
}
func (controller *Controller) ListHandlingUnit(c *gin.Context) {
	f, ok := filter(c, "barcode", true)
	if !ok {
		return
	}
	response, err := controller.service.ListHandlingUnit(c.Request.Context(), f)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "inventory identities retrieved", response)
}
