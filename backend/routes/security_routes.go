package routes

import (
	controller "wms-api/controller/authentication"

	"github.com/gin-gonic/gin"
)

func registerSecurityRoutes(
	api *gin.RouterGroup,
	controller *controller.SecurityController,
	requirePermission gin.HandlerFunc,
) {
	security := api.Group("/security")
	security.Use(requirePermission)
	security.GET("/account-statuses", controller.ListAccountStatuses)
	security.GET("/authentication-policies", controller.ListAuthenticationPolicies)
	security.GET("/permissions", controller.ListPermissions)

	security.POST("/accounts", controller.CreateAccount)
	security.GET("/accounts", controller.ListAccounts)
	security.GET("/accounts/:account_id", controller.GetAccount)
	security.PUT("/accounts/:account_id", controller.UpdateAccount)
	security.PATCH("/accounts/:account_id/status", controller.ChangeAccountStatus)
	security.DELETE("/accounts/:account_id", controller.DeactivateAccount)
	security.POST("/accounts/:account_id/password-reset", controller.ResetAccountPassword)
	security.POST("/accounts/:account_id/unlock", controller.UnlockAccount)
	security.POST("/accounts/:account_id/revoke-sessions", controller.RevokeAccountSessions)
	security.GET("/accounts/:account_id/permissions", controller.ListAccountPermissions)
	security.POST("/accounts/:account_id/permissions", controller.GrantAccountPermission)
	security.DELETE("/accounts/:account_id/permissions/:permission_id", controller.RevokeAccountPermission)
	security.POST("/accounts/:account_id/roles", controller.AssignAccountRole)
	security.DELETE("/accounts/:account_id/roles/:role_id", controller.RevokeAccountRole)
	security.POST("/accounts/:account_id/owners", controller.GrantAccountOwner)
	security.DELETE("/accounts/:account_id/owners/:owner_id", controller.RevokeAccountOwner)
	security.POST("/accounts/:account_id/warehouses", controller.GrantAccountWarehouse)
	security.DELETE("/accounts/:account_id/warehouses/:warehouse_id", controller.RevokeAccountWarehouse)

	security.POST("/roles", controller.CreateRole)
	security.GET("/roles", controller.ListRoles)
	security.GET("/roles/:role_id", controller.GetRole)
	security.PUT("/roles/:role_id", controller.UpdateRole)
	security.PATCH("/roles/:role_id/status", controller.ChangeRoleStatus)
	security.PUT("/roles/:role_id/permissions", controller.ReplaceRolePermissions)
	security.DELETE("/roles/:role_id", controller.DeactivateRole)
}
