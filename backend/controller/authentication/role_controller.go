package authentication

import (
	"net/http"
	"strconv"

	dto "wms-api/dto/authentication"
	"wms-api/utils"

	"github.com/gin-gonic/gin"
)

func (controller *SecurityController) CreateRole(c *gin.Context) {
	var request dto.CreateRoleRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.roles.Create(c.Request.Context(), request, securityActorID(c))
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "role created", response)
}

func (controller *SecurityController) ListRoles(c *gin.Context) {
	page, pageSize, err := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	if err != nil {
		utils.Failure(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	var active *bool
	if value := c.Query("active"); value != "" {
		parsed, parseErr := strconv.ParseBool(value)
		if parseErr != nil {
			utils.Failure(c, http.StatusBadRequest, "active must be true or false", nil)
			return
		}
		active = &parsed
	}
	response, err := controller.roles.List(c.Request.Context(), c.Query("search"), active, page, pageSize)
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "roles retrieved", response)
}

func (controller *SecurityController) GetRole(c *gin.Context) {
	response, err := controller.roles.Get(c.Request.Context(), c.Param("role_id"))
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "role retrieved", response)
}

func (controller *SecurityController) UpdateRole(c *gin.Context) {
	var request dto.UpdateRoleRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.roles.Update(c.Request.Context(), c.Param("role_id"), request, securityActorID(c))
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "role updated", response)
}

func (controller *SecurityController) ReplaceRolePermissions(c *gin.Context) {
	var request dto.ReplaceRolePermissionsRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.roles.ReplacePermissions(c.Request.Context(), c.Param("role_id"), request, securityActorID(c))
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "role permissions replaced", response)
}

func (controller *SecurityController) DeactivateRole(c *gin.Context) {
	expected, err := strconv.Atoi(c.Query("expected_version"))
	if err != nil || expected < 1 {
		utils.Failure(c, http.StatusBadRequest, "expected_version must be a positive integer", nil)
		return
	}
	response, err := controller.roles.Deactivate(c.Request.Context(), c.Param("role_id"), expected, securityActorID(c))
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "role deactivated; assignments retained for history", response)
}

func (controller *SecurityController) ChangeRoleStatus(c *gin.Context) {
	var request dto.ChangeRoleStatusRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.roles.ChangeStatus(c.Request.Context(), c.Param("role_id"), request, securityActorID(c))
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "role status changed", response)
}

func (controller *SecurityController) AssignAccountRole(c *gin.Context) {
	var request dto.AssignAccountRoleRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	err := controller.roles.Assign(c.Request.Context(), c.Param("account_id"), request.RoleID, securityActorID(c))
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "account role assigned", nil)
}

func (controller *SecurityController) RevokeAccountRole(c *gin.Context) {
	err := controller.roles.Revoke(c.Request.Context(), c.Param("account_id"), c.Param("role_id"))
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "account role revoked", nil)
}
