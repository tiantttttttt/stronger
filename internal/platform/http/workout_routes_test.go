package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"stronger/server/internal/workout"
)

type recordingWorkoutRepository struct {
	completeCalls   int
	getSummaryCalls int
}

func (repository *recordingWorkoutRepository) Start(context.Context, string) (workout.Summary, error) {
	return workout.Summary{}, nil
}

func (repository *recordingWorkoutRepository) Complete(context.Context, string, string) (workout.Summary, error) {
	repository.completeCalls++
	return workout.Summary{}, nil
}

func (repository *recordingWorkoutRepository) GetSummary(context.Context, string, string) (workout.Summary, error) {
	repository.getSummaryCalls++
	return workout.Summary{}, nil
}

func TestWorkoutRoutesRejectInvalidUUID(t *testing.T) {
	repository := &recordingWorkoutRepository{}
	dependencies := testDependencies(fakePinger{})
	dependencies.Workouts = repository
	router := NewRouter(dependencies)

	accessToken, err := dependencies.Tokens.Issue("00000000-0000-0000-0000-000000000099")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "complete workout",
			method: http.MethodPost,
			path:   "/api/v1/workouts/not-a-uuid/complete",
		},
		{
			name:   "get workout summary",
			method: http.MethodGet,
			path:   "/api/v1/workouts/not-a-uuid/summary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(tt.method, tt.path, nil)
			request.Header.Set("Authorization", "Bearer "+accessToken)

			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
			}
			if recorder.Body.String() != `{"code":"invalid_path_parameter","message":"id must be a valid UUID","parameter":"id"}` {
				t.Fatalf("body = %s", recorder.Body.String())
			}
		})
	}

	if repository.completeCalls != 0 || repository.getSummaryCalls != 0 {
		t.Fatalf("repository calls = complete:%d summary:%d, want zero", repository.completeCalls, repository.getSummaryCalls)
	}
}

var _ workout.Repository = (*recordingWorkoutRepository)(nil)
