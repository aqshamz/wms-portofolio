package master

import (
	"net/http"

	dto "wms-api/dto/master"
	"wms-api/utils"

	"github.com/gin-gonic/gin"
)

func (controller *Controller) GrantOwnerAccess(c *gin.Context) {
	var request dto.GrantAccountAccessRequest
	if c.ShouldBindJSON(&request) != nil {
		utils.Failure(c, http.StatusBadRequest, "account_id is required", nil)
		return
	}
	if err := controller.accessScopes.GrantOwner(
		c.Request.Context(), c.Param("owner_id"), request.AccountID, actorID(c),
	); err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "owner access granted", nil)
}

func (controller *Controller) ListAccountOwnerAccess(c *gin.Context) {
	response, err := controller.accessScopes.ListOwners(c.Request.Context(), c.Param("account_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "account owner access retrieved", response)
}

func (controller *Controller) RevokeOwnerAccess(c *gin.Context) {
	err := controller.accessScopes.RevokeOwner(
		c.Request.Context(), c.Param("account_id"), c.Param("owner_id"),
	)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "owner access revoked", nil)
}

func (controller *Controller) GrantWarehouseAccess(c *gin.Context) {
	var request dto.GrantAccountAccessRequest
	if c.ShouldBindJSON(&request) != nil {
		utils.Failure(c, http.StatusBadRequest, "account_id is required", nil)
		return
	}
	if err := controller.accessScopes.GrantWarehouse(
		c.Request.Context(), c.Param("warehouse_id"), request.AccountID, actorID(c),
	); err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "warehouse access granted", nil)
}

func (controller *Controller) ListAccountWarehouseAccess(c *gin.Context) {
	response, err := controller.accessScopes.ListWarehouses(c.Request.Context(), c.Param("account_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "account warehouse access retrieved", response)
}

func (controller *Controller) RevokeWarehouseAccess(c *gin.Context) {
	err := controller.accessScopes.RevokeWarehouse(
		c.Request.Context(), c.Param("account_id"), c.Param("warehouse_id"),
	)
	if err != nil {
		handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "warehouse access revoked", nil)
}
