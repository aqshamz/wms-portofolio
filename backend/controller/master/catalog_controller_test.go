package master

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	dto "wms-api/dto/master"
)

func TestCatalogRequestValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name, body string
		request    func() interface{}
		valid      bool
	}{
		{"valid unit", `{"code":"EA","name":"Each","decimal_scale":0}`, func() interface{} { return &dto.CreateUOMRequest{} }, true},
		{"scale range", `{"code":"EA","name":"Each","decimal_scale":7}`, func() interface{} { return &dto.CreateUOMRequest{} }, false},
		{"immutable code", `{"code":"NEW","name":"Each","decimal_scale":0,"is_active":true}`, func() interface{} { return &dto.UpdateUOMRequest{} }, false},
		{"missing active", `{"name":"Each","decimal_scale":0}`, func() interface{} { return &dto.UpdateUOMRequest{} }, false},
		{"explicit false", `{"name":"Each","decimal_scale":0,"is_active":false}`, func() interface{} { return &dto.UpdateUOMRequest{} }, true},
		{"missing conversion", `{"uom_id":"123e4567-e89b-42d3-a456-426614174000"}`, func() interface{} { return &dto.CreateItemUOMRequest{} }, false},
		{"float conversion", `{"uom_id":"123e4567-e89b-42d3-a456-426614174000","conversion_to_base":1.2}`, func() interface{} { return &dto.CreateItemUOMRequest{} }, false},
		{"invalid UUID", `{"uom_id":"oops","conversion_to_base":"1"}`, func() interface{} { return &dto.CreateItemUOMRequest{} }, false},
		{"trailing JSON", `{"code":"EA","name":"Each"} {}`, func() interface{} { return &dto.CreateUOMRequest{} }, false},
		{"missing optimistic timestamp", `{"name":"Updated","is_active":true}`, func() interface{} { return &dto.UpdateItemRequest{} }, false},
		{"null object", `null`, func() interface{} { return &dto.CreateUOMRequest{} }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			if got := bindCatalog(c, tt.request()); got != tt.valid {
				t.Fatalf("valid=%v got=%v body=%s", tt.valid, got, recorder.Body)
			}
			if !tt.valid && recorder.Code != http.StatusBadRequest {
				t.Fatalf("status=%d", recorder.Code)
			}
		})
	}
}
