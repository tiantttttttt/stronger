package router

import (
	"encoding/json"
	standardhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"stronger/server/service/api/middleware"
)

func TestAuthenticationFlow(t *testing.T) {
	validTokens := middleware.NewTokenManager("test-secret", time.Hour)
	router := NewRouter(Dependencies{Database: fakePinger{}, Debug: true, Tokens: validTokens})

	loginRecorder := httptest.NewRecorder()
	loginRequest := httptest.NewRequest(standardhttp.MethodPost, "/api/v1/auth/dev/login", strings.NewReader(`{"user_id":"00000000-0000-0000-0000-000000000099"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(loginRecorder, loginRequest)
	if loginRecorder.Code != standardhttp.StatusOK {
		t.Fatalf("login status = %d, body = %s", loginRecorder.Code, loginRecorder.Body.String())
	}

	var loginResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(loginRecorder.Body.Bytes(), &loginResponse); err != nil {
		t.Fatal(err)
	}
	assertMeStatus(t, router, loginResponse.AccessToken, standardhttp.StatusOK)

	expiredTokens := middleware.NewTokenManager("test-secret", -time.Minute)
	expiredToken, err := expiredTokens.Issue("00000000-0000-0000-0000-000000000099")
	if err != nil {
		t.Fatal(err)
	}
	assertMeStatus(t, router, expiredToken, standardhttp.StatusUnauthorized)

	forgedTokens := middleware.NewTokenManager("different-secret", time.Hour)
	forgedToken, err := forgedTokens.Issue("00000000-0000-0000-0000-000000000099")
	if err != nil {
		t.Fatal(err)
	}
	assertMeStatus(t, router, forgedToken, standardhttp.StatusUnauthorized)
}

func assertMeStatus(t *testing.T, router standardhttp.Handler, accessToken string, wantStatus int) {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(standardhttp.MethodGet, "/api/v1/me", nil)
	request.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(recorder, request)
	if recorder.Code != wantStatus {
		t.Fatalf("/me status = %d, want %d; body = %s", recorder.Code, wantStatus, recorder.Body.String())
	}
}

func TestDevelopmentLoginDebugSwitch(t *testing.T) {
	for _, debug := range []bool{false, true} {
		dependencies := testDependencies(fakePinger{})
		dependencies.Debug = debug
		engine := NewRouter(dependencies)
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(standardhttp.MethodPost, "/api/v1/auth/dev/login", strings.NewReader("{}"))
		request.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(recorder, request)
		wantStatus := standardhttp.StatusNotFound
		if debug {
			wantStatus = standardhttp.StatusOK
		}
		if recorder.Code != wantStatus {
			t.Fatalf("debug=%t: login status = %d, want %d", debug, recorder.Code, wantStatus)
		}
	}
}
