package utils

import "github.com/gin-gonic/gin"

type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     interface{} `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

func requestID(c *gin.Context) string {
	value, _ := c.Get("request_id")
	id, _ := value.(string)
	return id
}

func Success(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: requestID(c),
	})
}

func Failure(c *gin.Context, status int, message string, details interface{}) {
	c.JSON(status, APIResponse{
		Success:   false,
		Message:   message,
		Error:     details,
		RequestID: requestID(c),
	})
}
