package master

import (
	"github.com/gin-gonic/gin"
	"net/http"
	dto "wms-api/dto/master"
	service "wms-api/services/master"
	"wms-api/utils"
)

type OperationalController struct{ service *service.OperationalService }

func NewOperationalController(service *service.OperationalService) *OperationalController {
	return &OperationalController{service: service}
}
func operationalQuery(c *gin.Context) (dto.OperationalListRequest, bool) {
	page, size, err := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	if err != nil {
		utils.Failure(c, http.StatusBadRequest, err.Error(), nil)
		return dto.OperationalListRequest{}, false
	}
	active, err := optionalBool(c.Query("active"))
	if err != nil {
		utils.Failure(c, http.StatusBadRequest, "active must be true or false", nil)
		return dto.OperationalListRequest{}, false
	}
	return dto.OperationalListRequest{OwnerID: c.Query("owner_id"), WarehouseID: c.Query("warehouse_id"), ModuleCode: c.Query("module_code"), Search: c.Query("search"), Active: active, Page: page, PageSize: size}, true
}
func (controller *OperationalController) ListAppModule(c *gin.Context) {
	request, ok := operationalQuery(c)
	if !ok {
		return
	}
	response, err := controller.service.ListAppModule(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) GetAppModule(c *gin.Context) {

	response, err := controller.service.GetAppModule(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) CreateAppModule(c *gin.Context) {
	var request dto.CreateAppModuleRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.CreateAppModule(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "configuration created", response)
}
func (controller *OperationalController) UpdateAppModule(c *gin.Context) {
	var request dto.UpdateAppModuleRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.UpdateAppModule(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration updated", response)
}
func (controller *OperationalController) DeactivateAppModule(c *gin.Context) {

	response, err := controller.service.DeactivateAppModule(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration deactivated", response)
}
func (controller *OperationalController) ListAppPermission(c *gin.Context) {
	request, ok := operationalQuery(c)
	if !ok {
		return
	}
	response, err := controller.service.ListAppPermission(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) GetAppPermission(c *gin.Context) {

	response, err := controller.service.GetAppPermission(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) ListDocumentType(c *gin.Context) {
	request, ok := operationalQuery(c)
	if !ok {
		return
	}
	response, err := controller.service.ListDocumentType(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) GetDocumentType(c *gin.Context) {

	response, err := controller.service.GetDocumentType(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) CreateDocumentType(c *gin.Context) {
	var request dto.CreateDocumentTypeRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.CreateDocumentType(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "configuration created", response)
}
func (controller *OperationalController) UpdateDocumentType(c *gin.Context) {
	var request dto.UpdateDocumentTypeRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.UpdateDocumentType(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration updated", response)
}
func (controller *OperationalController) DeactivateDocumentType(c *gin.Context) {

	response, err := controller.service.DeactivateDocumentType(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration deactivated", response)
}
func (controller *OperationalController) ListDocumentStatus(c *gin.Context) {
	request, ok := operationalQuery(c)
	if !ok {
		return
	}
	response, err := controller.service.ListDocumentStatus(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) GetDocumentStatus(c *gin.Context) {

	response, err := controller.service.GetDocumentStatus(c.Request.Context(), c.Param("id"), c.Param("child_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) CreateDocumentStatus(c *gin.Context) {
	var request dto.CreateDocumentStatusRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.CreateDocumentStatus(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "configuration created", response)
}
func (controller *OperationalController) UpdateDocumentStatus(c *gin.Context) {
	var request dto.UpdateDocumentStatusRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.UpdateDocumentStatus(c.Request.Context(), c.Param("id"), c.Param("child_id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration updated", response)
}
func (controller *OperationalController) DeactivateDocumentStatus(c *gin.Context) {

	response, err := controller.service.DeactivateDocumentStatus(c.Request.Context(), c.Param("id"), c.Param("child_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration deactivated", response)
}
func (controller *OperationalController) ListDocumentStatusTransition(c *gin.Context) {
	request, ok := operationalQuery(c)
	if !ok {
		return
	}
	response, err := controller.service.ListDocumentStatusTransition(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) GetDocumentStatusTransition(c *gin.Context) {

	response, err := controller.service.GetDocumentStatusTransition(c.Request.Context(), c.Param("id"), c.Param("child_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) CreateDocumentStatusTransition(c *gin.Context) {
	var request dto.CreateDocumentStatusTransitionRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.CreateDocumentStatusTransition(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "configuration created", response)
}
func (controller *OperationalController) UpdateDocumentStatusTransition(c *gin.Context) {
	var request dto.UpdateDocumentStatusTransitionRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.UpdateDocumentStatusTransition(c.Request.Context(), c.Param("id"), c.Param("child_id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration updated", response)
}
func (controller *OperationalController) DeactivateDocumentStatusTransition(c *gin.Context) {

	response, err := controller.service.DeactivateDocumentStatusTransition(c.Request.Context(), c.Param("id"), c.Param("child_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration deactivated", response)
}
func (controller *OperationalController) ListTaskType(c *gin.Context) {
	request, ok := operationalQuery(c)
	if !ok {
		return
	}
	response, err := controller.service.ListTaskType(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) GetTaskType(c *gin.Context) {

	response, err := controller.service.GetTaskType(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) CreateTaskType(c *gin.Context) {
	var request dto.CreateTaskTypeRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.CreateTaskType(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "configuration created", response)
}
func (controller *OperationalController) UpdateTaskType(c *gin.Context) {
	var request dto.UpdateTaskTypeRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.UpdateTaskType(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration updated", response)
}
func (controller *OperationalController) DeactivateTaskType(c *gin.Context) {

	response, err := controller.service.DeactivateTaskType(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration deactivated", response)
}
func (controller *OperationalController) ListTaskStatus(c *gin.Context) {
	request, ok := operationalQuery(c)
	if !ok {
		return
	}
	response, err := controller.service.ListTaskStatus(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) GetTaskStatus(c *gin.Context) {

	response, err := controller.service.GetTaskStatus(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) CreateTaskStatus(c *gin.Context) {
	var request dto.CreateTaskStatusRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.CreateTaskStatus(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "configuration created", response)
}
func (controller *OperationalController) UpdateTaskStatus(c *gin.Context) {
	var request dto.UpdateTaskStatusRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.UpdateTaskStatus(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration updated", response)
}
func (controller *OperationalController) DeactivateTaskStatus(c *gin.Context) {

	response, err := controller.service.DeactivateTaskStatus(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration deactivated", response)
}
func (controller *OperationalController) ListTaskStatusTransition(c *gin.Context) {
	request, ok := operationalQuery(c)
	if !ok {
		return
	}
	response, err := controller.service.ListTaskStatusTransition(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) GetTaskStatusTransition(c *gin.Context) {

	response, err := controller.service.GetTaskStatusTransition(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) CreateTaskStatusTransition(c *gin.Context) {
	var request dto.CreateTaskStatusTransitionRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.CreateTaskStatusTransition(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "configuration created", response)
}
func (controller *OperationalController) UpdateTaskStatusTransition(c *gin.Context) {
	var request dto.UpdateTaskStatusTransitionRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.UpdateTaskStatusTransition(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration updated", response)
}
func (controller *OperationalController) DeactivateTaskStatusTransition(c *gin.Context) {

	response, err := controller.service.DeactivateTaskStatusTransition(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration deactivated", response)
}
func (controller *OperationalController) ListTaskPriority(c *gin.Context) {
	request, ok := operationalQuery(c)
	if !ok {
		return
	}
	response, err := controller.service.ListTaskPriority(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) GetTaskPriority(c *gin.Context) {

	response, err := controller.service.GetTaskPriority(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) CreateTaskPriority(c *gin.Context) {
	var request dto.CreateTaskPriorityRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.CreateTaskPriority(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "configuration created", response)
}
func (controller *OperationalController) UpdateTaskPriority(c *gin.Context) {
	var request dto.UpdateTaskPriorityRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.UpdateTaskPriority(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration updated", response)
}
func (controller *OperationalController) DeactivateTaskPriority(c *gin.Context) {

	response, err := controller.service.DeactivateTaskPriority(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration deactivated", response)
}
func (controller *OperationalController) ListPickingSortMethod(c *gin.Context) {
	request, ok := operationalQuery(c)
	if !ok {
		return
	}
	response, err := controller.service.ListPickingSortMethod(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) GetPickingSortMethod(c *gin.Context) {

	response, err := controller.service.GetPickingSortMethod(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) CreatePickingSortMethod(c *gin.Context) {
	var request dto.CreatePickingSortMethodRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.CreatePickingSortMethod(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "configuration created", response)
}
func (controller *OperationalController) UpdatePickingSortMethod(c *gin.Context) {
	var request dto.UpdatePickingSortMethodRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.UpdatePickingSortMethod(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration updated", response)
}
func (controller *OperationalController) DeactivatePickingSortMethod(c *gin.Context) {

	response, err := controller.service.DeactivatePickingSortMethod(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration deactivated", response)
}
func (controller *OperationalController) ListPickingStrategy(c *gin.Context) {
	request, ok := operationalQuery(c)
	if !ok {
		return
	}
	response, err := controller.service.ListPickingStrategy(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) GetPickingStrategy(c *gin.Context) {

	response, err := controller.service.GetPickingStrategy(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) CreatePickingStrategy(c *gin.Context) {
	var request dto.CreatePickingStrategyRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.CreatePickingStrategy(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "configuration created", response)
}
func (controller *OperationalController) UpdatePickingStrategy(c *gin.Context) {
	var request dto.UpdatePickingStrategyRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.UpdatePickingStrategy(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration updated", response)
}
func (controller *OperationalController) DeactivatePickingStrategy(c *gin.Context) {

	response, err := controller.service.DeactivatePickingStrategy(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration deactivated", response)
}
func (controller *OperationalController) ListPickingStrategyRule(c *gin.Context) {
	request, ok := operationalQuery(c)
	if !ok {
		return
	}
	response, err := controller.service.ListPickingStrategyRule(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) GetPickingStrategyRule(c *gin.Context) {

	response, err := controller.service.GetPickingStrategyRule(c.Request.Context(), c.Param("id"), c.Param("child_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) CreatePickingStrategyRule(c *gin.Context) {
	var request dto.CreatePickingStrategyRuleRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.CreatePickingStrategyRule(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "configuration created", response)
}
func (controller *OperationalController) UpdatePickingStrategyRule(c *gin.Context) {
	var request dto.UpdatePickingStrategyRuleRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.UpdatePickingStrategyRule(c.Request.Context(), c.Param("id"), c.Param("child_id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration updated", response)
}
func (controller *OperationalController) DeactivatePickingStrategyRule(c *gin.Context) {

	response, err := controller.service.DeactivatePickingStrategyRule(c.Request.Context(), c.Param("id"), c.Param("child_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration deactivated", response)
}
func (controller *OperationalController) ListPutawayStrategy(c *gin.Context) {
	request, ok := operationalQuery(c)
	if !ok {
		return
	}
	response, err := controller.service.ListPutawayStrategy(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) GetPutawayStrategy(c *gin.Context) {

	response, err := controller.service.GetPutawayStrategy(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) CreatePutawayStrategy(c *gin.Context) {
	var request dto.CreatePutawayStrategyRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.CreatePutawayStrategy(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "configuration created", response)
}
func (controller *OperationalController) UpdatePutawayStrategy(c *gin.Context) {
	var request dto.UpdatePutawayStrategyRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.UpdatePutawayStrategy(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration updated", response)
}
func (controller *OperationalController) DeactivatePutawayStrategy(c *gin.Context) {

	response, err := controller.service.DeactivatePutawayStrategy(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration deactivated", response)
}
func (controller *OperationalController) ListPutawayStrategyRule(c *gin.Context) {
	request, ok := operationalQuery(c)
	if !ok {
		return
	}
	response, err := controller.service.ListPutawayStrategyRule(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) GetPutawayStrategyRule(c *gin.Context) {

	response, err := controller.service.GetPutawayStrategyRule(c.Request.Context(), c.Param("id"), c.Param("child_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration retrieved", response)
}
func (controller *OperationalController) CreatePutawayStrategyRule(c *gin.Context) {
	var request dto.CreatePutawayStrategyRuleRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.CreatePutawayStrategyRule(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "configuration created", response)
}
func (controller *OperationalController) UpdatePutawayStrategyRule(c *gin.Context) {
	var request dto.UpdatePutawayStrategyRuleRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.UpdatePutawayStrategyRule(c.Request.Context(), c.Param("id"), c.Param("child_id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration updated", response)
}
func (controller *OperationalController) DeactivatePutawayStrategyRule(c *gin.Context) {

	response, err := controller.service.DeactivatePutawayStrategyRule(c.Request.Context(), c.Param("id"), c.Param("child_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "configuration deactivated", response)
}

func (controller *OperationalController) ListDocumentNumberRule(c *gin.Context) {
	request, ok := operationalQuery(c)
	if !ok {
		return
	}
	response, err := controller.service.ListDocumentNumberRule(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "numbering rules retrieved", response)
}
func (controller *OperationalController) ReplaceDocumentNumberRule(c *gin.Context) {
	var request dto.ReplaceDocumentNumberRuleRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.ReplaceDocumentNumberRule(c.Request.Context(), c.Param("id"), request, actorID(c))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "numbering rule replaced", response)
}
func (controller *OperationalController) DeactivateDocumentNumberRule(c *gin.Context) {
	response, err := controller.service.DeactivateDocumentNumberRule(c.Request.Context(), c.Param("id"), c.Param("child_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "numbering rule deactivated", response)
}
func (controller *OperationalController) GenerateDocumentID(c *gin.Context) {
	var request dto.GenerateDocumentIDRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.GenerateDocumentID(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "document ID allocated", response)
}
func (controller *OperationalController) ListDocumentDailyCounter(c *gin.Context) {
	request, ok := operationalQuery(c)
	if !ok {
		return
	}
	response, err := controller.service.ListDocumentDailyCounter(c.Request.Context(), c.Param("id"), c.Query("date_from"), c.Query("date_to"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "daily counters retrieved", response)
}
