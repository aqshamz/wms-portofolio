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

type SecurityController struct {
	permissions *service.PermissionService
	accounts    *service.AccountAdminService
	roles       *service.RoleService
}

func NewSecurityController(
	permissions *service.PermissionService,
	accounts *service.AccountAdminService,
	roles *service.RoleService,
) *SecurityController {
	return &SecurityController{permissions: permissions, accounts: accounts, roles: roles}
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
	if err := controller.permissions.Grant(
		c.Request.Context(), c.Param("account_id"), request.PermissionID, securityActorID(c),
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
	case errors.Is(err, service.ErrInvalidAccount), errors.Is(err, service.ErrInvalidRole):
		utils.Failure(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, service.ErrAccountNotFound), errors.Is(err, service.ErrRoleNotFound),
		errors.Is(err, service.ErrAccessNotFound), errors.Is(err, service.ErrAccountRoleNotFound):
		utils.Failure(c, http.StatusNotFound, err.Error(), nil)
	case errors.Is(err, service.ErrAccountConflict), errors.Is(err, service.ErrRoleConflict),
		errors.Is(err, service.ErrConcurrentAccountUpdate), errors.Is(err, service.ErrConcurrentRoleUpdate):
		utils.Failure(c, http.StatusConflict, err.Error(), nil)
	default:
		_ = c.Error(err)
		utils.Failure(c, http.StatusInternalServerError, "unable to process security configuration", nil)
	}
}

func securityActorID(c *gin.Context) string {
	value, _ := c.Get(middleware.ContextUserKey)
	actor, _ := value.(dto.UserResponse)
	return actor.AccountID
}
