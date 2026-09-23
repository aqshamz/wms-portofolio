package inbound

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"wms-api/utils"
)

func (controller *Controller) ListQuarantineTargets(c *gin.Context) {
	search, page, size, ok := putawayLookupQuery(c)
	if !ok {
		return
	}
	response, err := controller.service.ListQuarantineTargets(c.Request.Context(), c.Param("id"), search, page, size)
	if err != nil {
		fail(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "eligible quarantine storage targets retrieved", response)
}
