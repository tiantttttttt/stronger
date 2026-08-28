package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUUIDPathParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		value      string
		wantOK     bool
		wantStatus int
		wantUUID   string
	}{
		{
			name:       "valid UUID",
			value:      "01A0477E-0303-7AF8-BFCF-7092D1262E3D",
			wantOK:     true,
			wantStatus: http.StatusOK,
			wantUUID:   "01a0477e-0303-7af8-bfcf-7092d1262e3d",
		},
		{
			name:       "invalid UUID",
			value:      "not-a-uuid",
			wantOK:     false,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Params = gin.Params{{Key: "id", Value: tt.value}}

			value, ok := UUIDPathParam(context, "id")
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
			if tt.wantUUID != "" && value.String() != tt.wantUUID {
				t.Fatalf("UUID = %s, want %s", value, tt.wantUUID)
			}
		})
	}
}
