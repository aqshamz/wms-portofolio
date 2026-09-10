package authentication

import (
	"errors"
	"net/http"

	dto "wms-api/dto/authentication"
	"wms-api/middleware"
	service "wms-api/services/authentication"
	"wms-api/utils"

	"github.com/gin-gonic/gin"
)

type SecurityController struct{ permissions *service.PermissionService }

func NewSecurityController(permissions *service.PermissionService) *SecurityController {
	return &SecurityController{permissions: permissions}
}

func (controller *SecurityController) ListPermissions(c *gin.Context) {
	response, err := controller.permissions.Available(c.Request.Context())
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "permissions retrieved", response)
}

func (controller *SecurityController) ListAccountPermissions(c *gin.Context) {
	response, err := controller.permissions.List(c.Request.Context(), c.Param("account_id"))
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "account permissions retrieved", response)
}

func (controller *SecurityController) GrantAccountPermission(c *gin.Context) {
	var request dto.GrantAccountPermissionRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	user, _ := c.Get(middleware.ContextUserKey)
	actor, _ := user.(dto.UserResponse)
	if err := controller.permissions.Grant(
		c.Request.Context(), c.Param("account_id"), request.PermissionID, actor.AccountID,
	); err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "account permission granted", nil)
}

func (controller *SecurityController) RevokeAccountPermission(c *gin.Context) {
	err := controller.permissions.Revoke(
		c.Request.Context(), c.Param("account_id"), c.Param("permission_id"),
	)
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "account permission revoked", nil)
}

func securityError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidPermissionGrant):
		utils.Failure(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, service.ErrPermissionGrantMissing):
		utils.Failure(c, http.StatusNotFound, err.Error(), nil)
	case errors.Is(err, service.ErrLastSecurityAdmin):
		utils.Failure(c, http.StatusConflict, err.Error(), nil)
	default:
		_ = c.Error(err)
		utils.Failure(c, http.StatusInternalServerError, "unable to process security configuration", nil)
	}
}
