package inventory

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
	repository "wms-api/repository/inventory"
	"wms-api/utils"
)

func coreFilter(c *gin.Context, kind string) (int, int, bool) {
	allowed := map[string]bool{"page": true, "page_size": true, "owner_id": true, "warehouse_id": true, "item_id": true, "search": true}
	switch kind {
	case "balance":
		for _, k := range []string{"location_id", "lot_id", "handling_unit_id", "inventory_status_id", "include_zero"} {
			allowed[k] = true
		}
	case "movement":
		for _, k := range []string{"movement_type_id", "source_document_id", "operation_key", "occurred_from", "occurred_until"} {
			allowed[k] = true
		}
	case "serial-state":
		for _, k := range []string{"location_id", "lot_id", "handling_unit_id", "inventory_status_id"} {
			allowed[k] = true
		}
	}
	for key, values := range c.Request.URL.Query() {
		if !allowed[key] || len(values) != 1 {
			utils.Failure(c, 400, "unknown or repeated query parameter: "+key, nil)
			return 0, 0, false
		}
	}
	page, size, err := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	if err != nil {
		utils.Failure(c, 400, err.Error(), nil)
		return 0, 0, false
	}
	return page, size, true
}
func parseTime(c *gin.Context, key string) (*time.Time, bool) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return nil, true
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		utils.Failure(c, 400, key+" must be RFC3339", nil)
		return nil, false
	}
	return &parsed, true
}
func parseBool(c *gin.Context, key string) (bool, bool) {
	value, exists := c.Request.URL.Query()[key]
	if !exists {
		return false, true
	}
	if value[0] == "true" {
		return true, true
	}
	if value[0] == "false" {
		return false, true
	}
	utils.Failure(c, 400, key+" must be true or false", nil)
	return false, false
}
func (controller *Controller) ListBalances(c *gin.Context) {
	page, size, ok := coreFilter(c, "balance")
	if !ok {
		return
	}
	include, ok := parseBool(c, "include_zero")
	if !ok {
		return
	}
	response, err := controller.service.ListBalances(c.Request.Context(), repository.BalanceFilter{OwnerID: c.Query("owner_id"), WarehouseID: c.Query("warehouse_id"), LocationID: c.Query("location_id"), ItemID: c.Query("item_id"), LotID: c.Query("lot_id"), HandlingUnitID: c.Query("handling_unit_id"), InventoryStatusID: c.Query("inventory_status_id"), Search: strings.TrimSpace(c.Query("search")), IncludeZero: include, Page: page, PageSize: size})
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, 200, "inventory balances retrieved", response)
}
func (controller *Controller) GetBalance(c *gin.Context) {
	response, err := controller.service.GetBalance(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, 200, "inventory balance found", response)
}
func (controller *Controller) ListMovements(c *gin.Context) {
	page, size, ok := coreFilter(c, "movement")
	if !ok {
		return
	}
	from, ok := parseTime(c, "occurred_from")
	if !ok {
		return
	}
	until, ok := parseTime(c, "occurred_until")
	if !ok {
		return
	}
	if from != nil && until != nil && !from.Before(*until) {
		utils.Failure(c, 400, "occurred_from must precede occurred_until", nil)
		return
	}
	response, err := controller.service.ListMovements(c.Request.Context(), repository.MovementFilter{OwnerID: c.Query("owner_id"), WarehouseID: c.Query("warehouse_id"), ItemID: c.Query("item_id"), MovementTypeID: c.Query("movement_type_id"), SourceDocumentID: c.Query("source_document_id"), OperationKey: c.Query("operation_key"), Search: strings.TrimSpace(c.Query("search")), OccurredFrom: from, OccurredUntil: until, Page: page, PageSize: size})
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, 200, "inventory movements retrieved", response)
}
func (controller *Controller) GetMovement(c *gin.Context) {
	response, err := controller.service.GetMovement(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, 200, "inventory movement found", response)
}
func (controller *Controller) ListSerialStates(c *gin.Context) {
	page, size, ok := coreFilter(c, "serial-state")
	if !ok {
		return
	}
	response, err := controller.service.ListSerialStates(c.Request.Context(), repository.SerialStateFilter{OwnerID: c.Query("owner_id"), WarehouseID: c.Query("warehouse_id"), LocationID: c.Query("location_id"), ItemID: c.Query("item_id"), LotID: c.Query("lot_id"), HandlingUnitID: c.Query("handling_unit_id"), InventoryStatusID: c.Query("inventory_status_id"), Search: strings.TrimSpace(c.Query("search")), Page: page, PageSize: size})
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, 200, "serial inventory state retrieved", response)
}
func (controller *Controller) GetSerialState(c *gin.Context) {
	response, err := controller.service.GetSerialState(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, 200, "serial inventory state found", response)
}
func (controller *Controller) ListMovementTypes(c *gin.Context) {
	var active *bool
	if values, exists := c.Request.URL.Query()["active"]; exists {
		if len(values) != 1 || (values[0] != "true" && values[0] != "false") {
			utils.Failure(c, 400, "active must be true or false", nil)
			return
		}
		value := values[0] == "true"
		active = &value
	}
	for key := range c.Request.URL.Query() {
		if key != "active" {
			utils.Failure(c, 400, "unknown query parameter: "+key, nil)
			return
		}
	}
	response, err := controller.service.ListMovementTypes(c.Request.Context(), active)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "movement types retrieved", response)
}
