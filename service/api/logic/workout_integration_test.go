package logic_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"stronger/server/service/api"
	"stronger/server/service/api/logic"
	"stronger/server/service/api/logic/responses"
	"stronger/server/service/api/middleware"
	"stronger/server/service/api/router"
	"stronger/server/service/model"
)

// Each run creates an isolated schema and applies the real migration, never AutoMigrate.
func integrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx := context.Background()
	base, err := api.NewDB(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := base.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	schema := "gorm_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := base.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := base.Exec("DROP SCHEMA " + schema + " CASCADE").Error; err != nil {
			t.Error(err)
		}
	})
	parsed, err := url.Parse(dsn)
	if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") {
		t.Fatal("TEST_DATABASE_URL must be a PostgreSQL URL")
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	db, err := api.NewDB(ctx, parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	migration, err := os.ReadFile("../../../db/migrations/000001_phase1_init.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sqlDB.ExecContext(ctx, string(migration)); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestPostgresWorkoutFlow(t *testing.T) {
	db := integrationDB(t)
	ctx := context.Background()
	auth := logic.NewAuth(db)
	workouts := logic.NewWorkouts(db)
	userID, otherUser := uuid.NewString(), uuid.NewString()
	for _, id := range []string{userID, otherUser} {
		if err := auth.EnsureDevUser(ctx, id); err != nil {
			t.Fatal(err)
		}
	}
	// Repeated login must update timestamps without overwriting the profile.
	if err := db.Model(&model.User{}).Where("id = ?", userID).Update("nickname", "Preserved").Error; err != nil {
		t.Fatal(err)
	}
	if err := auth.EnsureDevUser(ctx, userID); err != nil {
		t.Fatal(err)
	}
	var user model.User
	if err := db.Where("id = ?", userID).Take(&user).Error; err != nil {
		t.Fatal(err)
	}
	if user.Nickname != "Preserved" {
		t.Fatal("development login overwrote the profile")
	}

	t.Run("concurrent start returns one active workout", func(t *testing.T) {
		type outcome struct {
			summary responses.WorkoutSummary
			err     error
		}
		ch := make(chan outcome, 8)
		for i := 0; i < cap(ch); i++ {
			go func() { summary, err := workouts.Start(ctx, userID); ch <- outcome{summary, err} }()
		}
		var workoutID string
		for i := 0; i < cap(ch); i++ {
			result := <-ch
			if result.err != nil {
				t.Fatal(result.err)
			}
			if result.summary.Status != int16(model.WorkoutActive) || result.summary.StartedAt.IsZero() {
				t.Fatalf("database defaults missing: %+v", result.summary)
			}
			if workoutID == "" {
				workoutID = result.summary.ID
			}
			if workoutID != result.summary.ID {
				t.Fatal("concurrent start created multiple workouts")
			}
		}
		var count int64
		if err := db.Model(&model.WorkoutSession{}).Where("user_id = ? AND status = ?", userID, model.WorkoutActive).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("active workouts = %d", count)
		}
	})

	active, err := workouts.Start(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := workouts.GetSummary(ctx, userID, active.ID); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("active summary error = %v", err)
	}
	if _, err := workouts.Complete(ctx, otherUser, active.ID); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("cross-user complete error = %v", err)
	}

	t.Run("concurrent completion returns frozen zero values", func(t *testing.T) {
		type outcome struct {
			summary responses.WorkoutSummary
			err     error
		}
		ch := make(chan outcome, 8)
		for i := 0; i < cap(ch); i++ {
			go func() { summary, err := workouts.Complete(ctx, userID, active.ID); ch <- outcome{summary, err} }()
		}
		var frozen *responses.WorkoutSummary
		for i := 0; i < cap(ch); i++ {
			result := <-ch
			if result.err != nil {
				t.Fatal(result.err)
			}
			summary := result.summary
			if summary.Status != int16(model.WorkoutCompleted) || summary.CompletedAt == nil || summary.DurationSeconds == nil || summary.CompletedSetCount == nil || summary.CompletedExerciseCount == nil || summary.TotalReps == nil || summary.TotalVolumeKG == nil {
				t.Fatalf("incomplete frozen response: %+v", summary)
			}
			if *summary.TotalVolumeKG != "0.00" || *summary.TotalReps != 0 || *summary.CompletedSetCount != 0 || *summary.CompletedExerciseCount != 0 {
				t.Fatalf("incorrect empty summary: %+v", summary)
			}
			if frozen == nil {
				frozen = &summary
			} else if !reflect.DeepEqual(*frozen, summary) {
				t.Fatal("completion changed frozen response")
			}
		}
		summary, err := workouts.GetSummary(ctx, userID, active.ID)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(*frozen, summary) {
			t.Fatal("summary differs from completion response")
		}
		if _, err := workouts.GetSummary(ctx, otherUser, active.ID); !errors.Is(err, model.ErrNotFound) {
			t.Fatalf("cross-user summary error = %v", err)
		}
	})

	t.Run("incomplete profile rolls back without a workout", func(t *testing.T) {
		if err := db.Model(&model.User{}).Where("id = ?", otherUser).Update("height_cm", nil).Error; err != nil {
			t.Fatal(err)
		}
		if _, err := workouts.Start(ctx, otherUser); !errors.Is(err, model.ErrProfileIncomplete) {
			t.Fatalf("profile error = %v", err)
		}
		var count int64
		if err := db.Model(&model.WorkoutSession{}).Where("user_id = ?", otherUser).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatal("profile rejection created a workout")
		}
	})

	t.Run("discarded workout is terminal", func(t *testing.T) {
		summary, err := workouts.Start(ctx, userID)
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Model(&model.WorkoutSession{}).Where("id = ?", summary.ID).Updates(map[string]interface{}{"status": model.WorkoutDiscarded, "discarded_at": gorm.Expr("now()")}).Error; err != nil {
			t.Fatal(err)
		}
		if _, err := workouts.Complete(ctx, userID, summary.ID); !errors.Is(err, model.ErrTerminalState) {
			t.Fatalf("terminal error = %v", err)
		}
	})

	t.Run("canceled context propagates", func(t *testing.T) {
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, err := workouts.Start(canceled, userID); !errors.Is(err, context.Canceled) {
			t.Fatalf("canceled start error = %v", err)
		}
	})
}

func TestPostgresHTTPEmptyWorkoutFlow(t *testing.T) {
	db := integrationDB(t)
	engine := router.NewRouter(router.Dependencies{
		Debug: true, Tokens: middleware.NewTokenManager("test-secret", time.Hour),
		DevUsers: logic.NewAuth(db), Workouts: logic.NewWorkouts(db),
	})
	request := func(method, path, body, token string) []byte {
		t.Helper()
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		engine.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s %s: status=%d body=%s", method, path, recorder.Code, recorder.Body.String())
		}
		return recorder.Body.Bytes()
	}
	var login responses.Login
	if err := json.Unmarshal(request("POST", "/api/v1/auth/dev/login", "{}", ""), &login); err != nil {
		t.Fatal(err)
	}
	if login.AccessToken == "" || login.ExpiresIn != 3600 {
		t.Fatalf("invalid login: %+v", login)
	}
	var started, completed, summary responses.WorkoutSummary
	if err := json.Unmarshal(request("POST", "/api/v1/workouts", "", login.AccessToken), &started); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(request("POST", "/api/v1/workouts/"+started.ID+"/complete", "", login.AccessToken), &completed); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(request("GET", "/api/v1/workouts/"+started.ID+"/summary", "", login.AccessToken), &summary); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(completed, summary) {
		t.Fatal("HTTP summary differs from completion")
	}
}
