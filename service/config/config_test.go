package config

import "testing"

func TestDebugConfiguration(t *testing.T) {
	for _, tt := range []struct {
		name      string
		value     string
		wantDebug bool
		wantError bool
	}{
		{name: "default production"},
		{name: "debug", value: "true", wantDebug: true},
		{name: "production", value: "false"},
		{name: "invalid", value: "development", wantError: true},
		{name: "numeric rejected", value: "1", wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DEBUG", tt.value)
			t.Setenv("APP_ENV", "development")
			t.Setenv("DATABASE_URL", "postgres://test@localhost/test")
			t.Setenv("JWT_SECRET", "test-secret")
			cfg, err := Load()
			if (err != nil) != tt.wantError {
				t.Fatalf("Load error = %v, wantError = %t", err, tt.wantError)
			}
			if err == nil && cfg.Debug != tt.wantDebug {
				t.Fatalf("Debug = %t, want %t", cfg.Debug, tt.wantDebug)
			}
		})
	}
}
