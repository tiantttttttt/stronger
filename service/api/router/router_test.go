package router

import (
	"errors"
	standardhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"stronger/server/consts"
	"stronger/server/service/api/middleware"
)

func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name       string
		pingError  error
		pingDelay  time.Duration
		wantStatus int
		wantBody   string
	}{
		{name: "database available", wantStatus: standardhttp.StatusOK, wantBody: `"status":"ok"`},
		{name: "database unavailable", pingError: errors.New("connection refused"), wantStatus: standardhttp.StatusServiceUnavailable, wantBody: `"code":"database_unavailable"`},
		{name: "database timeout", pingDelay: 3 * time.Second, wantStatus: standardhttp.StatusServiceUnavailable, wantBody: `"code":"database_unavailable"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			if err != nil {
				t.Fatal(err)
			}
			previous := consts.DB
			// 全局连接串行替换，请求结束后恢复。
			t.Cleanup(func() {
				consts.DB = previous
				if err := pool.Close(); err != nil {
					t.Error(err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Error(err)
				}
			})
			consts.DB, err = gorm.Open(postgres.New(postgres.Config{Conn: pool}), &gorm.Config{DisableAutomaticPing: true})
			if err != nil {
				t.Fatal(err)
			}
			mock.ExpectPing().WillDelayFor(tt.pingDelay).WillReturnError(tt.pingError)
			mock.ExpectClose()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(standardhttp.MethodGet, "/health", nil)
			started := time.Now()
			NewRouter(false, middleware.NewTokenManager("test-secret", time.Hour)).ServeHTTP(recorder, request)
			if tt.pingDelay > 0 && time.Since(started) >= tt.pingDelay {
				t.Fatal("health check did not enforce its two-second timeout")
			}

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if !strings.Contains(recorder.Body.String(), tt.wantBody) {
				t.Fatalf("body = %s, want to contain %s", recorder.Body.String(), tt.wantBody)
			}
		})
	}
}
