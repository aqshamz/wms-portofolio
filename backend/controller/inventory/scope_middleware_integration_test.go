package inventory

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"wms-api/config"
	authdto "wms-api/dto/authentication"
	"wms-api/middleware"
	repository "wms-api/repository/inventory"
	"wms-api/requestscope"
	service "wms-api/services/inventory"
)

func TestInventoryHTTPScopePostgreSQL(t *testing.T) {
	if os.Getenv("WMS_INTEGRATION_TEST") != "1" {
		t.Skip("set WMS_INTEGRATION_TEST=1")
	}
	check := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	_ = godotenv.Load("../../.env")
	cfg, err := config.Load()
	check(err)
	cfg.Database.AutoCreate = false
	db, err := config.OpenDatabase(cfg.Database)
	check(err)
	sqlDB, err := db.DB()
	check(err)
	defer sqlDB.Close()
	tx := db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Begin()
	check(tx.Error)
	defer tx.Rollback()
	schema := fmt.Sprintf("inventory_balance_scope_test_%d", time.Now().UnixNano())
	check(tx.Exec("CREATE SCHEMA " + schema).Error)
	check(tx.Exec("SET LOCAL search_path TO " + schema).Error)
	for _, statement := range []string{
		"CREATE TABLE inventory_balance (balance_id text PRIMARY KEY, owner_id uuid NOT NULL, warehouse_id uuid NOT NULL)",
		"CREATE TABLE inventory_movement (movement_id text PRIMARY KEY, owner_id uuid NOT NULL, warehouse_id uuid NOT NULL)",
		"CREATE TABLE serial_inventory (serial_id text PRIMARY KEY, owner_id uuid NOT NULL, balance_id text NOT NULL)",
		"CREATE TABLE account_owner_access (account_id uuid, owner_id uuid)",
		"CREATE TABLE account_warehouse_access (account_id uuid, warehouse_id uuid)",
		"CREATE TABLE warehouse_owner (owner_id uuid, warehouse_id uuid, is_active boolean)",
	} {
		check(tx.Exec(statement).Error)
	}
	worker := "11111111-1111-4111-8111-111111111111"
	denied := "22222222-2222-4222-8222-222222222222"
	owner := "33333333-3333-4333-8333-333333333333"
	warehouse := "44444444-4444-4444-8444-444444444444"
	check(tx.Exec("INSERT INTO inventory_balance VALUES ('BAL-1',?,?)", owner, warehouse).Error)
	check(tx.Exec("INSERT INTO inventory_movement VALUES ('MOV-1',?,?)", owner, warehouse).Error)
	check(tx.Exec("INSERT INTO serial_inventory VALUES ('SER-1',?,'BAL-1')", owner).Error)
	check(tx.Exec("INSERT INTO account_owner_access VALUES (?,?)", worker, owner).Error)
	check(tx.Exec("INSERT INTO account_warehouse_access VALUES (?,?)", worker, warehouse).Error)
	check(tx.Exec("INSERT INTO warehouse_owner VALUES (?,?,true)", owner, warehouse).Error)
	controller := NewController(service.NewService(repository.NewRepositories(tx)))
	gin.SetMode(gin.TestMode)

	for _, test := range []struct {
		name, actor, path string
		unrestricted      bool
		want              int
	}{
		{"granted list", worker, "/api/v1/inventory/balances?owner_id=" + owner + "&warehouse_id=" + warehouse, false, http.StatusNoContent},
		{"denied list", denied, "/api/v1/inventory/balances?owner_id=" + owner + "&warehouse_id=" + warehouse, false, http.StatusForbidden},
		{"missing list scope", worker, "/api/v1/inventory/balances", false, http.StatusBadRequest},
		{"granted detail", worker, "/api/v1/inventory/balances/BAL-1", false, http.StatusNoContent},
		{"denied detail", denied, "/api/v1/inventory/balances/BAL-1", false, http.StatusForbidden},
		{"superadmin detail", denied, "/api/v1/inventory/balances/BAL-1", true, http.StatusNoContent},
		{"missing detail", worker, "/api/v1/inventory/balances/BAL-MISSING", false, http.StatusNotFound},
		{"granted movement list", worker, "/api/v1/inventory/movements?owner_id=" + owner + "&warehouse_id=" + warehouse, false, http.StatusNoContent},
		{"denied movement list", denied, "/api/v1/inventory/movements?owner_id=" + owner + "&warehouse_id=" + warehouse, false, http.StatusForbidden},
		{"granted movement detail", worker, "/api/v1/inventory/movements/MOV-1", false, http.StatusNoContent},
		{"denied movement detail", denied, "/api/v1/inventory/movements/MOV-1", false, http.StatusForbidden},
		{"granted serial-state list", worker, "/api/v1/inventory/serial-states?owner_id=" + owner + "&warehouse_id=" + warehouse, false, http.StatusNoContent},
		{"denied serial-state list", denied, "/api/v1/inventory/serial-states?owner_id=" + owner + "&warehouse_id=" + warehouse, false, http.StatusForbidden},
		{"granted serial-state detail", worker, "/api/v1/inventory/serial-states/SER-1", false, http.StatusNoContent},
		{"denied serial-state detail", denied, "/api/v1/inventory/serial-states/SER-1", false, http.StatusForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set(middleware.ContextUserKey, authdto.UserResponse{AccountID: test.actor})
				c.Request = c.Request.WithContext(requestscope.WithPrincipal(c.Request.Context(), requestscope.Principal{AccountID: test.actor, Unrestricted: test.unrestricted}))
				c.Next()
			})
			reached := false
			router.GET("/api/v1/inventory/balances", controller.RequireBalanceScope(), func(c *gin.Context) {
				reached = true
				c.Status(http.StatusNoContent)
			})
			router.GET("/api/v1/inventory/balances/:id", controller.RequireBalanceScope(), func(c *gin.Context) {
				reached = true
				c.Status(http.StatusNoContent)
			})
			router.GET("/api/v1/inventory/movements", controller.RequireMovementScope(), func(c *gin.Context) {
				reached = true
				c.Status(http.StatusNoContent)
			})
			router.GET("/api/v1/inventory/movements/:id", controller.RequireMovementScope(), func(c *gin.Context) {
				reached = true
				c.Status(http.StatusNoContent)
			})
			router.GET("/api/v1/inventory/serial-states", controller.RequireSerialStateScope(), func(c *gin.Context) {
				reached = true
				c.Status(http.StatusNoContent)
			})
			router.GET("/api/v1/inventory/serial-states/:id", controller.RequireSerialStateScope(), func(c *gin.Context) {
				reached = true
				c.Status(http.StatusNoContent)
			})
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
			if response.Code != test.want || reached != (test.want == http.StatusNoContent) {
				t.Fatalf("status=%d reached=%t want=%d body=%s", response.Code, reached, test.want, response.Body.String())
			}
		})
	}
}
