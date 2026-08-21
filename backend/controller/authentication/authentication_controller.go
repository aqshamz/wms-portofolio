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

type Controller struct {
	service *service.Service
}

func NewController(service *service.Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) Login(c *gin.Context) {
	var request dto.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Failure(c, http.StatusBadRequest, "identifier and password are required", nil)
		return
	}

	response, err := controller.service.Login(c.Request.Context(), request, service.ClientInfo{
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			utils.Failure(c, http.StatusUnauthorized, err.Error(), nil)
			return
		}
		_ = c.Error(err)
		utils.Failure(c, http.StatusInternalServerError, "unable to complete login", nil)
		return
	}

	utils.Success(c, http.StatusOK, "login successful", response)
}

func (controller *Controller) Me(c *gin.Context) {
	user, ok := c.Get(middleware.ContextUserKey)
	if !ok {
		utils.Failure(c, http.StatusUnauthorized, "authentication required", nil)
		return
	}
	utils.Success(c, http.StatusOK, "authenticated user", user)
}

func (controller *Controller) Logout(c *gin.Context) {
	token, ok := c.Get(middleware.ContextTokenKey)
	if !ok {
		utils.Failure(c, http.StatusUnauthorized, "authentication required", nil)
		return
	}
	if err := controller.service.Logout(c.Request.Context(), token.(string)); err != nil {
		_ = c.Error(err)
		utils.Failure(c, http.StatusInternalServerError, "unable to log out", nil)
		return
	}
	utils.Success(c, http.StatusOK, "logout successful", nil)
}

func (controller *Controller) LogoutAll(c *gin.Context) {
	token, ok := c.Get(middleware.ContextTokenKey)
	if !ok {
		utils.Failure(c, http.StatusUnauthorized, "authentication required", nil)
		return
	}
	if err := controller.service.LogoutAll(c.Request.Context(), token.(string)); err != nil {
		_ = c.Error(err)
		utils.Failure(c, http.StatusInternalServerError, "unable to log out all sessions", nil)
		return
	}
	utils.Success(c, http.StatusOK, "all sessions logged out", nil)
}
