package router

import (
	"context"
	"errors"
	standardhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"stronger/server/service/api/handler"
	"stronger/server/service/api/middleware"
)

type fakePinger struct {
	err error
}

func (p fakePinger) Ping(context.Context) error {
	return p.err
}

func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name       string
		pingError  error
		wantStatus int
		wantBody   string
	}{
		{name: "database available", wantStatus: standardhttp.StatusOK, wantBody: `"status":"ok"`},
		{name: "database unavailable", pingError: errors.New("connection refused"), wantStatus: standardhttp.StatusServiceUnavailable, wantBody: `"code":"database_unavailable"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(standardhttp.MethodGet, "/health", nil)
			NewRouter(testDependencies(fakePinger{err: tt.pingError})).ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if !strings.Contains(recorder.Body.String(), tt.wantBody) {
				t.Fatalf("body = %s, want to contain %s", recorder.Body.String(), tt.wantBody)
			}
		})
	}
}

func testDependencies(database handler.DatabasePinger) Dependencies {
	return Dependencies{
		Database: database,
		Debug:    true,
		Tokens:   middleware.NewTokenManager("test-secret", time.Hour),
	}
}
