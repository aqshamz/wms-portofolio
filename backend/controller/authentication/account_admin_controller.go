package authentication

import (
	"net/http"
	"strconv"

	dto "wms-api/dto/authentication"
	"wms-api/utils"

	"github.com/gin-gonic/gin"
)

func (controller *SecurityController) CreateAccount(c *gin.Context) {
	var request dto.CreateAccountRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.accounts.Create(c.Request.Context(), request, securityActorID(c))
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "account created", response)
}

func (controller *SecurityController) ListAccounts(c *gin.Context) {
	page, pageSize, err := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	if err != nil {
		utils.Failure(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response, err := controller.accounts.List(c.Request.Context(), c.Query("search"), c.Query("status_id"), page, pageSize)
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "accounts retrieved", response)
}

func (controller *SecurityController) GetAccount(c *gin.Context) {
	response, err := controller.accounts.Get(c.Request.Context(), c.Param("account_id"))
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "account retrieved", response)
}

func (controller *SecurityController) UpdateAccount(c *gin.Context) {
	var request dto.UpdateAccountRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.accounts.Update(c.Request.Context(), c.Param("account_id"), request, securityActorID(c))
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "account updated", response)
}

func (controller *SecurityController) ChangeAccountStatus(c *gin.Context) {
	var request dto.ChangeAccountStatusRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.accounts.ChangeStatus(c.Request.Context(), c.Param("account_id"), request, securityActorID(c))
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "account status changed", response)
}

func (controller *SecurityController) DeactivateAccount(c *gin.Context) {
	expected, err := strconv.Atoi(c.Query("expected_version"))
	if err != nil || expected < 1 {
		utils.Failure(c, http.StatusBadRequest, "expected_version must be a positive integer", nil)
		return
	}
	response, err := controller.accounts.Deactivate(c.Request.Context(), c.Param("account_id"), expected, securityActorID(c))
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "account deactivated; historical relationships retained", response)
}

func (controller *SecurityController) ResetAccountPassword(c *gin.Context) {
	var request dto.ResetAccountPasswordRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	if err := controller.accounts.ResetPassword(c.Request.Context(), c.Param("account_id"), request, securityActorID(c)); err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "password reset and active sessions revoked", nil)
}

func (controller *SecurityController) UnlockAccount(c *gin.Context) {
	var request dto.AccountVersionRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	response, err := controller.accounts.Unlock(c.Request.Context(), c.Param("account_id"), request.ExpectedVersion, securityActorID(c))
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "account unlocked", response)
}

func (controller *SecurityController) RevokeAccountSessions(c *gin.Context) {
	if err := controller.accounts.RevokeSessions(c.Request.Context(), c.Param("account_id")); err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "account sessions revoked", nil)
}

func (controller *SecurityController) ListAccountStatuses(c *gin.Context) {
	response, err := controller.accounts.Statuses(c.Request.Context())
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "account statuses retrieved", response)
}

func (controller *SecurityController) ListAuthenticationPolicies(c *gin.Context) {
	response, err := controller.accounts.Policies(c.Request.Context())
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "authentication policies retrieved", response)
}

func (controller *SecurityController) GrantAccountOwner(c *gin.Context) {
	var request dto.AccountOwnerRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	err := controller.accounts.GrantOwner(c.Request.Context(), c.Param("account_id"), request.OwnerID, securityActorID(c))
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "account owner access granted", nil)
}

func (controller *SecurityController) RevokeAccountOwner(c *gin.Context) {
	err := controller.accounts.RevokeOwner(c.Request.Context(), c.Param("account_id"), c.Param("owner_id"))
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "account owner access revoked", nil)
}

func (controller *SecurityController) GrantAccountWarehouse(c *gin.Context) {
	var request dto.AccountWarehouseRequest
	if !utils.BindStrictJSON(c, &request) {
		return
	}
	err := controller.accounts.GrantWarehouse(c.Request.Context(), c.Param("account_id"), request.WarehouseID, securityActorID(c))
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "account warehouse access granted", nil)
}

func (controller *SecurityController) RevokeAccountWarehouse(c *gin.Context) {
	err := controller.accounts.RevokeWarehouse(c.Request.Context(), c.Param("account_id"), c.Param("warehouse_id"))
	if err != nil {
		securityError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "account warehouse access revoked", nil)
}
