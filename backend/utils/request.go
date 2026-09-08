package utils

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func BindStrictJSON(c *gin.Context, request interface{}) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(request); err != nil {
		Failure(c, http.StatusBadRequest, "invalid JSON or unknown field", nil)
		return false
	}
	var extra interface{}
	if decoder.Decode(&extra) != io.EOF || binding.Validator.ValidateStruct(request) != nil {
		Failure(c, http.StatusBadRequest, "invalid request fields", nil)
		return false
	}
	return true
}
