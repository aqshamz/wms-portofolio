package master

import (
	"net/http"

	dto "wms-api/dto/master"
	"wms-api/utils"

	"github.com/gin-gonic/gin"
)

func (controller *Controller) AssignWarehouseOwner(c *gin.Context) {
	var request dto.AssignWarehouseOwnerRequest
	if c.ShouldBindJSON(&request) != nil {
		utils.Failure(c, http.StatusBadRequest, "owner_id is required", nil)
		return
	}
	if err := controller.structure.AssignOwner(c, c.Param("id"), request, actorID(c)); err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "warehouse owner assigned", nil)
}

func (controller *Controller) ListWarehouseOwners(c *gin.Context) {
	response, err := controller.structure.ListOwners(c, c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "warehouse owners retrieved", response)
}

func (controller *Controller) DeactivateWarehouseOwner(c *gin.Context) {
	if err := controller.structure.DeactivateOwner(c, c.Param("id"), c.Param("owner_id")); err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "warehouse owner deactivated", nil)
}

func (controller *Controller) CreateLocationType(c *gin.Context) {
	var request dto.CreateLocationTypeRequest
	if c.ShouldBindJSON(&request) != nil {
		utils.Failure(c, http.StatusBadRequest, "invalid location type request", nil)
		return
	}
	response, err := controller.structure.CreateLocationType(c, request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "location type created", response)
}

func (controller *Controller) ListLocationTypes(c *gin.Context) {
	active, err := optionalBool(c.Query("active"))
	if err != nil {
		utils.Failure(c, http.StatusBadRequest, "active must be true or false", nil)
		return
	}
	response, err := controller.structure.ListLocationTypes(c, active)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "location types retrieved", response)
}

func (controller *Controller) UpdateLocationType(c *gin.Context) {
	var request dto.UpdateLocationTypeRequest
	if c.ShouldBindJSON(&request) != nil {
		utils.Failure(c, http.StatusBadRequest, "invalid location type request", nil)
		return
	}
	response, err := controller.structure.UpdateLocationType(c, c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "location type updated", response)
}

func (controller *Controller) CreateWarehouseZone(c *gin.Context) {
	var request dto.CreateWarehouseZoneRequest
	if c.ShouldBindJSON(&request) != nil {
		utils.Failure(c, http.StatusBadRequest, "invalid zone request", nil)
		return
	}
	response, err := controller.structure.CreateZone(c, c.Param("id"), request, actorID(c))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "warehouse zone created", response)
}

func (controller *Controller) ListWarehouseZones(c *gin.Context) {
	active, err := optionalBool(c.Query("active"))
	if err != nil {
		utils.Failure(c, http.StatusBadRequest, "active must be true or false", nil)
		return
	}
	response, err := controller.structure.ListZones(
		c, c.Param("id"), optionalQuery(c.Query("search")), active,
	)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "warehouse zones retrieved", response)
}

func (controller *Controller) UpdateWarehouseZone(c *gin.Context) {
	var request dto.UpdateWarehouseZoneRequest
	if c.ShouldBindJSON(&request) != nil {
		utils.Failure(c, http.StatusBadRequest, "invalid zone request", nil)
		return
	}
	response, err := controller.structure.UpdateZone(c, c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "warehouse zone updated", response)
}

func (controller *Controller) DeactivateWarehouseZone(c *gin.Context) {
	if err := controller.structure.DeactivateZone(c, c.Param("id")); err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "warehouse zone deactivated", nil)
}

func (controller *Controller) CreateWarehouseLocation(c *gin.Context) {
	var request dto.CreateWarehouseLocationRequest
	if c.ShouldBindJSON(&request) != nil {
		utils.Failure(c, http.StatusBadRequest, "invalid location request", nil)
		return
	}
	response, err := controller.structure.CreateLocation(c, c.Param("id"), request, actorID(c))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "warehouse location created", response)
}

func (controller *Controller) GetWarehouseLocation(c *gin.Context) {
	response, err := controller.structure.GetLocation(c, c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "warehouse location found", response)
}

func (controller *Controller) ListWarehouseLocations(c *gin.Context) {
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
	response, err := controller.structure.ListLocations(c, c.Param("id"),
		optionalQuery(c.Query("zone_id")), optionalQuery(c.Query("location_type_id")),
		optionalQuery(c.Query("search")), active, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "warehouse locations retrieved", response)
}

func (controller *Controller) UpdateWarehouseLocation(c *gin.Context) {
	var request dto.UpdateWarehouseLocationRequest
	if c.ShouldBindJSON(&request) != nil {
		utils.Failure(c, http.StatusBadRequest, "invalid location request", nil)
		return
	}
	response, err := controller.structure.UpdateLocation(c, c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "warehouse location updated", response)
}

func (controller *Controller) DeactivateWarehouseLocation(c *gin.Context) {
	if err := controller.structure.DeactivateLocation(c, c.Param("id")); err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "warehouse location deactivated", nil)
}
