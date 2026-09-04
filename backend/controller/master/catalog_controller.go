package master

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"io"
	"net/http"
	"strings"
	dto "wms-api/dto/master"
	repository "wms-api/repository/master"
	service "wms-api/services/master"
	"wms-api/utils"
)

type CatalogController struct{ service *service.CatalogService }

func NewCatalogController(service *service.CatalogService) *CatalogController {
	return &CatalogController{service: service}
}

// Reject unknown/immutable fields and trailing JSON; all HTTP validation stays here.
func bindCatalog(c *gin.Context, request interface{}) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(request); err != nil {
		utils.Failure(c, http.StatusBadRequest, "invalid request JSON or unknown field", nil)
		return false
	}
	var extra interface{}
	if decoder.Decode(&extra) != io.EOF || binding.Validator.ValidateStruct(request) != nil {
		utils.Failure(c, http.StatusBadRequest, "invalid request fields", nil)
		return false
	}
	return true
}
func catalogFilter(c *gin.Context) (repository.CatalogFilter, bool) {
	page, size, err := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	if err != nil {
		utils.Failure(c, http.StatusBadRequest, err.Error(), nil)
		return repository.CatalogFilter{}, false
	}
	active, err := optionalBool(c.Query("active"))
	if err != nil {
		utils.Failure(c, http.StatusBadRequest, "active must be true or false", nil)
		return repository.CatalogFilter{}, false
	}
	return repository.CatalogFilter{Page: page, PageSize: size, Active: active, Search: strings.TrimSpace(c.Query("search")),
		OwnerID: c.Query("owner_id"), CategoryID: c.Query("category_id"), PartnerTypeCode: c.Query("partner_type_code")}, true
}

func (controller *CatalogController) CreatePartnerType(c *gin.Context) {
	var request dto.CreatePartnerTypeRequest
	if !bindCatalog(c, &request) {
		return
	}

	response, err := controller.service.CreatePartnerType(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "PartnerType create successful", response)
}

func (controller *CatalogController) GetPartnerType(c *gin.Context) {

	response, err := controller.service.GetPartnerType(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "PartnerType get successful", response)
}

func (controller *CatalogController) ListPartnerType(c *gin.Context) {

	filter, ok := catalogFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListPartnerType(c.Request.Context(), filter)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "PartnerType list successful", response)
}

func (controller *CatalogController) UpdatePartnerType(c *gin.Context) {
	var request dto.UpdatePartnerTypeRequest
	if !bindCatalog(c, &request) {
		return
	}

	response, err := controller.service.UpdatePartnerType(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "PartnerType update successful", response)
}

func (controller *CatalogController) DeactivatePartnerType(c *gin.Context) {

	response, err := controller.service.DeactivatePartnerType(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "PartnerType deactivate successful", response)
}

func (controller *CatalogController) CreateBusinessPartner(c *gin.Context) {
	var request dto.CreateBusinessPartnerRequest
	if !bindCatalog(c, &request) {
		return
	}

	response, err := controller.service.CreateBusinessPartner(c.Request.Context(), request, actorID(c))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "BusinessPartner create successful", response)
}

func (controller *CatalogController) GetBusinessPartner(c *gin.Context) {

	response, err := controller.service.GetBusinessPartner(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "BusinessPartner get successful", response)
}

func (controller *CatalogController) ListBusinessPartner(c *gin.Context) {

	filter, ok := catalogFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListBusinessPartner(c.Request.Context(), filter)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "BusinessPartner list successful", response)
}

func (controller *CatalogController) UpdateBusinessPartner(c *gin.Context) {
	var request dto.UpdateBusinessPartnerRequest
	if !bindCatalog(c, &request) {
		return
	}

	response, err := controller.service.UpdateBusinessPartner(c.Request.Context(), c.Param("id"), request, actorID(c))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "BusinessPartner update successful", response)
}

func (controller *CatalogController) DeactivateBusinessPartner(c *gin.Context) {
	var request dto.DeactivateRequest
	if !bindCatalog(c, &request) {
		return
	}

	response, err := controller.service.DeactivateBusinessPartner(c.Request.Context(), c.Param("id"), request, actorID(c))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "BusinessPartner deactivate successful", response)
}

func (controller *CatalogController) CreateUOM(c *gin.Context) {
	var request dto.CreateUOMRequest
	if !bindCatalog(c, &request) {
		return
	}

	response, err := controller.service.CreateUOM(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "UOM create successful", response)
}

func (controller *CatalogController) GetUOM(c *gin.Context) {

	response, err := controller.service.GetUOM(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "UOM get successful", response)
}

func (controller *CatalogController) ListUOM(c *gin.Context) {

	filter, ok := catalogFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListUOM(c.Request.Context(), filter)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "UOM list successful", response)
}

func (controller *CatalogController) UpdateUOM(c *gin.Context) {
	var request dto.UpdateUOMRequest
	if !bindCatalog(c, &request) {
		return
	}

	response, err := controller.service.UpdateUOM(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "UOM update successful", response)
}

func (controller *CatalogController) DeactivateUOM(c *gin.Context) {

	response, err := controller.service.DeactivateUOM(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "UOM deactivate successful", response)
}

func (controller *CatalogController) CreateItemCategory(c *gin.Context) {
	var request dto.CreateItemCategoryRequest
	if !bindCatalog(c, &request) {
		return
	}

	response, err := controller.service.CreateItemCategory(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "ItemCategory create successful", response)
}

func (controller *CatalogController) GetItemCategory(c *gin.Context) {

	response, err := controller.service.GetItemCategory(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ItemCategory get successful", response)
}

func (controller *CatalogController) ListItemCategory(c *gin.Context) {

	filter, ok := catalogFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListItemCategory(c.Request.Context(), filter)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ItemCategory list successful", response)
}

func (controller *CatalogController) UpdateItemCategory(c *gin.Context) {
	var request dto.UpdateItemCategoryRequest
	if !bindCatalog(c, &request) {
		return
	}

	response, err := controller.service.UpdateItemCategory(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ItemCategory update successful", response)
}

func (controller *CatalogController) DeactivateItemCategory(c *gin.Context) {

	response, err := controller.service.DeactivateItemCategory(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ItemCategory deactivate successful", response)
}

func (controller *CatalogController) CreateItem(c *gin.Context) {
	var request dto.CreateItemRequest
	if !bindCatalog(c, &request) {
		return
	}

	response, err := controller.service.CreateItem(c.Request.Context(), request, actorID(c))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "Item create successful", response)
}

func (controller *CatalogController) GetItem(c *gin.Context) {

	response, err := controller.service.GetItem(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "Item get successful", response)
}

func (controller *CatalogController) ListItem(c *gin.Context) {

	filter, ok := catalogFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListItem(c.Request.Context(), filter)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "Item list successful", response)
}

func (controller *CatalogController) UpdateItem(c *gin.Context) {
	var request dto.UpdateItemRequest
	if !bindCatalog(c, &request) {
		return
	}

	response, err := controller.service.UpdateItem(c.Request.Context(), c.Param("id"), request, actorID(c))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "Item update successful", response)
}

func (controller *CatalogController) DeactivateItem(c *gin.Context) {
	var request dto.DeactivateRequest
	if !bindCatalog(c, &request) {
		return
	}

	response, err := controller.service.DeactivateItem(c.Request.Context(), c.Param("id"), request, actorID(c))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "Item deactivate successful", response)
}

func (controller *CatalogController) CreateInventoryStatus(c *gin.Context) {
	var request dto.CreateInventoryStatusRequest
	if !bindCatalog(c, &request) {
		return
	}

	response, err := controller.service.CreateInventoryStatus(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "InventoryStatus create successful", response)
}

func (controller *CatalogController) GetInventoryStatus(c *gin.Context) {

	response, err := controller.service.GetInventoryStatus(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "InventoryStatus get successful", response)
}

func (controller *CatalogController) ListInventoryStatus(c *gin.Context) {

	filter, ok := catalogFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListInventoryStatus(c.Request.Context(), filter)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "InventoryStatus list successful", response)
}

func (controller *CatalogController) UpdateInventoryStatus(c *gin.Context) {
	var request dto.UpdateInventoryStatusRequest
	if !bindCatalog(c, &request) {
		return
	}

	response, err := controller.service.UpdateInventoryStatus(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "InventoryStatus update successful", response)
}

func (controller *CatalogController) DeactivateInventoryStatus(c *gin.Context) {

	response, err := controller.service.DeactivateInventoryStatus(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "InventoryStatus deactivate successful", response)
}

func (controller *CatalogController) CreateQualityStatus(c *gin.Context) {
	var request dto.CreateQualityStatusRequest
	if !bindCatalog(c, &request) {
		return
	}

	response, err := controller.service.CreateQualityStatus(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "QualityStatus create successful", response)
}

func (controller *CatalogController) GetQualityStatus(c *gin.Context) {

	response, err := controller.service.GetQualityStatus(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "QualityStatus get successful", response)
}

func (controller *CatalogController) ListQualityStatus(c *gin.Context) {

	filter, ok := catalogFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListQualityStatus(c.Request.Context(), filter)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "QualityStatus list successful", response)
}

func (controller *CatalogController) UpdateQualityStatus(c *gin.Context) {
	var request dto.UpdateQualityStatusRequest
	if !bindCatalog(c, &request) {
		return
	}

	response, err := controller.service.UpdateQualityStatus(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "QualityStatus update successful", response)
}

func (controller *CatalogController) DeactivateQualityStatus(c *gin.Context) {

	response, err := controller.service.DeactivateQualityStatus(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "QualityStatus deactivate successful", response)
}

func (controller *CatalogController) CreateInspectionResult(c *gin.Context) {
	var request dto.CreateInspectionResultRequest
	if !bindCatalog(c, &request) {
		return
	}

	response, err := controller.service.CreateInspectionResult(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "InspectionResult create successful", response)
}

func (controller *CatalogController) GetInspectionResult(c *gin.Context) {

	response, err := controller.service.GetInspectionResult(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "InspectionResult get successful", response)
}

func (controller *CatalogController) ListInspectionResult(c *gin.Context) {

	filter, ok := catalogFilter(c)
	if !ok {
		return
	}
	response, err := controller.service.ListInspectionResult(c.Request.Context(), filter)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "InspectionResult list successful", response)
}

func (controller *CatalogController) UpdateInspectionResult(c *gin.Context) {
	var request dto.UpdateInspectionResultRequest
	if !bindCatalog(c, &request) {
		return
	}

	response, err := controller.service.UpdateInspectionResult(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "InspectionResult update successful", response)
}

func (controller *CatalogController) DeactivateInspectionResult(c *gin.Context) {

	response, err := controller.service.DeactivateInspectionResult(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "InspectionResult deactivate successful", response)
}

func (controller *CatalogController) CreateItemUOM(c *gin.Context) {
	var request dto.CreateItemUOMRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.CreateItemUOM(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "ItemUOM create successful", response)
}

func (controller *CatalogController) ListItemUOM(c *gin.Context) {

	response, err := controller.service.ListItemUOM(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ItemUOM list successful", response)
}

func (controller *CatalogController) UpdateItemUOM(c *gin.Context) {
	var request dto.UpdateItemUOMRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.UpdateItemUOM(c.Request.Context(), c.Param("id"), c.Param("item_uom_id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ItemUOM update successful", response)
}

func (controller *CatalogController) DeactivateItemUOM(c *gin.Context) {

	response, err := controller.service.DeactivateItemUOM(c.Request.Context(), c.Param("id"), c.Param("item_uom_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ItemUOM deactivate successful", response)
}

func (controller *CatalogController) CreateItemBarcode(c *gin.Context) {
	var request dto.CreateItemBarcodeRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.CreateItemBarcode(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "ItemBarcode create successful", response)
}

func (controller *CatalogController) ListItemBarcode(c *gin.Context) {

	response, err := controller.service.ListItemBarcode(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ItemBarcode list successful", response)
}

func (controller *CatalogController) UpdateItemBarcode(c *gin.Context) {
	var request dto.UpdateItemBarcodeRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.UpdateItemBarcode(c.Request.Context(), c.Param("id"), c.Param("barcode_id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ItemBarcode update successful", response)
}

func (controller *CatalogController) DeactivateItemBarcode(c *gin.Context) {

	response, err := controller.service.DeactivateItemBarcode(c.Request.Context(), c.Param("id"), c.Param("barcode_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "ItemBarcode deactivate successful", response)
}

func (controller *CatalogController) SetPrimaryItemBarcode(c *gin.Context) {
	response, err := controller.service.SetPrimaryItemBarcode(c.Request.Context(), c.Param("id"), c.Param("barcode_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "primary barcode set", response)
}
func (controller *CatalogController) ListBusinessPartnerType(c *gin.Context) {
	response, err := controller.service.ListBusinessPartnerType(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "partner types retrieved", response)
}
func (controller *CatalogController) AssignBusinessPartnerType(c *gin.Context) {
	var request dto.AssignPartnerTypeRequest
	if !bindCatalog(c, &request) {
		return
	}
	response, err := controller.service.AssignBusinessPartnerType(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "partner type assigned", response)
}
func (controller *CatalogController) RemoveBusinessPartnerType(c *gin.Context) {
	if err := controller.service.RemoveBusinessPartnerType(c.Request.Context(), c.Param("id"), c.Param("partner_type_id")); err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "partner type removed", nil)
}
