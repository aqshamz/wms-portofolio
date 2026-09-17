package inbound

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"wms-api/utils"
)

func putawayLookupQuery(c *gin.Context) (string, int, int, bool) {
	for key, values := range c.Request.URL.Query() {
		if (key != "search" && key != "page" && key != "page_size") || len(values) != 1 {
			utils.Failure(c, http.StatusBadRequest, "unknown or repeated query parameter: "+key, nil)
			return "", 0, 0, false
		}
	}
	page, size, err := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	if err != nil {
		utils.Failure(c, http.StatusBadRequest, err.Error(), nil)
		return "", 0, 0, false
	}
	return c.Query("search"), page, size, true
}

func (controller *Controller) ListPutawayAssignees(c *gin.Context) {
	search, page, size, ok := putawayLookupQuery(c)
	if !ok {
		return
	}
	response, err := controller.service.ListPutawayAssignees(c.Request.Context(), c.Param("id"), search, page, size)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "putaway assignees retrieved", response)
}

func (controller *Controller) ListPutawayTargets(c *gin.Context) {
	search, page, size, ok := putawayLookupQuery(c)
	if !ok {
		return
	}
	response, err := controller.service.ListPutawayTargets(c.Request.Context(), c.Param("id"), search, page, size)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "eligible putaway targets retrieved", response)
}
