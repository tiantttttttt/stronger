package api

import (
	"context"
	"strings"
	"testing"
	"time"

	"stronger/server/service/config"
)

func TestRunReleasesResources(t *testing.T) {
	for _, test := range []struct {
		name      string
		addr      string
		canceled  bool
		wantError bool
	}{
		{name: "cancellation", addr: "127.0.0.1:0", canceled: true},
		{name: "listener failure", addr: "invalid address", wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			api, err := New(config.Config{
				Debug: true, HTTPAddr: test.addr,
				DatabaseURL: "postgres://test@127.0.0.1:1/test?sslmode=disable", JWTSecret: "test-secret",
			})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if test.canceled {
				cancel()
			}
			err = api.Run(ctx)
			if (err != nil) != test.wantError {
				t.Fatalf("Run error = %v, wantError = %v", err, test.wantError)
			}
			if err := api.pool.PingContext(context.Background()); err == nil || !strings.Contains(err.Error(), "database is closed") {
				t.Fatalf("database pool was not closed: %v", err)
			}
		})
	}
}
