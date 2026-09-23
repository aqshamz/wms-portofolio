package inbound

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
	repository "wms-api/repository/inbound"
	"wms-api/requestscope"
	service "wms-api/services/inbound"
)

// Exercise the HTTP scope guard itself, not just the task service. The isolated
// normalized fixture deliberately has no owner/warehouse columns on rework_task.
func TestReworkHTTPScopePostgreSQL(t *testing.T) {
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
	schema := fmt.Sprintf("rework_scope_test_%d", time.Now().UnixNano())
	check(tx.Exec("CREATE SCHEMA " + schema).Error)
	check(tx.Exec("SET LOCAL search_path TO " + schema).Error)
	for _, statement := range []string{
		"CREATE TABLE rework_task (rework_task_id text PRIMARY KEY, quarantine_disposition_id text NOT NULL)",
		"CREATE TABLE quarantine_disposition (quarantine_disposition_id text PRIMARY KEY, quarantine_case_id text NOT NULL)",
		"CREATE TABLE quarantine_case (quarantine_case_id text PRIMARY KEY, owner_id uuid NOT NULL, warehouse_id uuid NOT NULL)",
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
	check(tx.Exec("INSERT INTO quarantine_case VALUES ('QCASE-1',?,?)", owner, warehouse).Error)
	check(tx.Exec("INSERT INTO quarantine_disposition VALUES ('QDIS-1','QCASE-1')").Error)
	check(tx.Exec("INSERT INTO rework_task VALUES ('RWK-1','QDIS-1')").Error)
	check(tx.Exec("INSERT INTO account_owner_access VALUES (?,?)", worker, owner).Error)
	check(tx.Exec("INSERT INTO account_warehouse_access VALUES (?,?)", worker, warehouse).Error)
	check(tx.Exec("INSERT INTO warehouse_owner VALUES (?,?,true)", owner, warehouse).Error)
	inboundService, err := service.NewService(repository.NewRepositories(tx), "Asia/Jakarta")
	check(err)
	controller := NewController(inboundService)
	gin.SetMode(gin.TestMode)
	for _, route := range []struct{ method, suffix string }{
		{http.MethodGet, ""}, {http.MethodPost, "/start"}, {http.MethodPost, "/complete"},
	} {
		for _, test := range []struct {
			name, actor, id, query string
			unrestricted           bool
			want                   int
		}{
			{"granted", worker, "RWK-1", "", false, http.StatusNoContent},
			{"denied", denied, "RWK-1", "", false, http.StatusForbidden},
			{"superadmin", denied, "RWK-1", "", true, http.StatusNoContent},
			{"forged query does not override resource scope", denied, "RWK-1", "?owner_id=" + denied + "&warehouse_id=" + denied, false, http.StatusForbidden},
			{"missing resource", worker, "RWK-MISSING", "", false, http.StatusNotFound},
		} {
			t.Run(route.method+route.suffix+"/"+test.name, func(t *testing.T) {
				router := gin.New()
				router.Use(func(c *gin.Context) {
					c.Set(middleware.ContextUserKey, authdto.UserResponse{AccountID: test.actor})
					c.Request = c.Request.WithContext(requestscope.WithPrincipal(c.Request.Context(), requestscope.Principal{AccountID: test.actor, Unrestricted: test.unrestricted}))
					c.Next()
				})
				reached := false
				router.Handle(route.method, "/api/v1/inbound/rework-tasks/:id"+route.suffix, controller.RequireScope(), func(c *gin.Context) {
					reached = true
					c.Status(http.StatusNoContent)
				})
				response := httptest.NewRecorder()
				router.ServeHTTP(response, httptest.NewRequest(route.method, "/api/v1/inbound/rework-tasks/"+test.id+route.suffix+test.query, nil))
				if response.Code != test.want || reached != (test.want == http.StatusNoContent) {
					t.Fatalf("status=%d reached=%t want=%d body=%s", response.Code, reached, test.want, response.Body.String())
				}
			})
		}
	}
}
