package stockcontrol

import (
	"github.com/gin-gonic/gin"
	"net/http"
	dto "wms-api/dto/stock_control"
	repository "wms-api/repository/stock_control"
	"wms-api/utils"
)

func (controller *Controller) AddFoundStock(c *gin.Context) {
	var request dto.AddFoundStockRequest
	if !bind(c, &request) {
		return
	}
	response, err := controller.service.AddFoundStock(c.Request.Context(), c.Param("id"), request, actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "found stock recorded for count review", response)
}

func (controller *Controller) GetFoundStockOptions(c *gin.Context) {
	var query struct {
		ItemID         string `form:"item_id" binding:"omitempty,uuid"`
		LocationID     string `form:"location_id" binding:"omitempty,uuid"`
		ItemSearch     string `form:"item_search" binding:"max=120"`
		LocationSearch string `form:"location_search" binding:"max=120"`
		LotSearch      string `form:"lot_search" binding:"max=120"`
		HUSearch       string `form:"hu_search" binding:"max=120"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.Failure(c, http.StatusBadRequest, "invalid found-stock lookup filters", nil)
		return
	}
	response, err := controller.service.GetFoundStockOptions(c.Request.Context(), c.Param("id"), repository.FoundStockOptionsFilter(query), actor(c))
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "found-stock options retrieved", response)
}
