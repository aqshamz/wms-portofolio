package master

import (
	"errors"
	"net/http"
	"strconv"

	authdto "wms-api/dto/authentication"
	dto "wms-api/dto/master"
	"wms-api/middleware"
	service "wms-api/services/master"
	"wms-api/utils"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	organizations *service.OrganizationService
	warehouses    *service.WarehouseService
	structure     *service.WarehouseStructureService
	accessScopes  *service.AccessScopeService
}

func NewController(
	organizations *service.OrganizationService,
	warehouses *service.WarehouseService,
	structure *service.WarehouseStructureService,
	accessScopes *service.AccessScopeService,
) *Controller {
	return &Controller{
		organizations: organizations,
		warehouses:    warehouses,
		structure:     structure,
		accessScopes:  accessScopes,
	}
}

func (controller *Controller) CreateOrganization(c *gin.Context) {
	var request dto.CreateOrganizationRequest
	if c.ShouldBindJSON(&request) != nil {
		utils.Failure(c, http.StatusBadRequest, "invalid organization request", nil)
		return
	}
	response, err := controller.organizations.Create(c.Request.Context(), request, actorID(c))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "organization created", response)
}

func (controller *Controller) GetOrganization(c *gin.Context) {
	response, err := controller.organizations.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "organization found", response)
}

func (controller *Controller) ListOrganizations(c *gin.Context) {
	page, pageSize, err := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	if err != nil {
		utils.Failure(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	active, err := optionalBool(c.Query("active"))
	if err != nil {
		utils.Failure(c, http.StatusBadRequest, "active must be true or false", nil)
		return
	}
	response, err := controller.organizations.List(
		c.Request.Context(), optionalQuery(c.Query("search")), active, page, pageSize,
	)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "organizations retrieved", response)
}

func (controller *Controller) UpdateOrganization(c *gin.Context) {
	var request dto.UpdateOrganizationRequest
	if c.ShouldBindJSON(&request) != nil {
		utils.Failure(c, http.StatusBadRequest, "invalid organization request", nil)
		return
	}
	response, err := controller.organizations.Update(
		c.Request.Context(), c.Param("id"), request, actorID(c),
	)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "organization updated", response)
}

func (controller *Controller) DeactivateOrganization(c *gin.Context) {
	var request dto.DeactivateRequest
	if c.ShouldBindJSON(&request) != nil {
		utils.Failure(c, http.StatusBadRequest, "expected_updated_at is required", nil)
		return
	}
	response, err := controller.organizations.Deactivate(
		c.Request.Context(), c.Param("id"), request, actorID(c),
	)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "organization deactivated", response)
}

func (controller *Controller) CreateWarehouse(c *gin.Context) {
	var request dto.CreateWarehouseRequest
	if c.ShouldBindJSON(&request) != nil {
		utils.Failure(c, http.StatusBadRequest, "invalid warehouse request", nil)
		return
	}
	response, err := controller.warehouses.Create(c.Request.Context(), request, actorID(c))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "warehouse created", response)
}

func (controller *Controller) GetWarehouse(c *gin.Context) {
	response, err := controller.warehouses.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "warehouse found", response)
}

func (controller *Controller) ListWarehouses(c *gin.Context) {
	page, pageSize, err := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	if err != nil {
		utils.Failure(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	active, err := optionalBool(c.Query("active"))
	if err != nil {
		utils.Failure(c, http.StatusBadRequest, "active must be true or false", nil)
		return
	}
	response, err := controller.warehouses.List(
		c.Request.Context(),
		optionalQuery(c.Query("operator_id")),
		optionalQuery(c.Query("search")),
		active,
		page,
		pageSize,
	)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "warehouses retrieved", response)
}

func (controller *Controller) UpdateWarehouse(c *gin.Context) {
	var request dto.UpdateWarehouseRequest
	if c.ShouldBindJSON(&request) != nil {
		utils.Failure(c, http.StatusBadRequest, "invalid warehouse request", nil)
		return
	}
	response, err := controller.warehouses.Update(
		c.Request.Context(), c.Param("id"), request, actorID(c),
	)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "warehouse updated", response)
}

func (controller *Controller) DeactivateWarehouse(c *gin.Context) {
	var request dto.DeactivateRequest
	if c.ShouldBindJSON(&request) != nil {
		utils.Failure(c, http.StatusBadRequest, "expected_updated_at is required", nil)
		return
	}
	response, err := controller.warehouses.Deactivate(
		c.Request.Context(), c.Param("id"), request, actorID(c),
	)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "warehouse deactivated", response)
}

func actorID(c *gin.Context) string {
	value, exists := c.Get(middleware.ContextUserKey)
	if !exists {
		return ""
	}
	user, ok := value.(authdto.UserResponse)
	if !ok {
		return ""
	}
	return user.AccountID
}

func optionalQuery(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func optionalBool(value string) (*bool, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		utils.Failure(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, service.ErrNotFound):
		utils.Failure(c, http.StatusNotFound, err.Error(), nil)
	case errors.Is(err, service.ErrConflict), errors.Is(err, service.ErrConcurrentUpdate), errors.Is(err, service.ErrCounterExhausted):
		utils.Failure(c, http.StatusConflict, err.Error(), nil)
	default:
		_ = c.Error(err)
		utils.Failure(c, http.StatusInternalServerError, "unable to process master data", nil)
	}
}
