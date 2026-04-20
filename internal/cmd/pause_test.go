package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestPauseCmd(t *testing.T) {
	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"pause", "abc1234567"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error",
			args: []string{"pause", "abc1234567"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().PauseDatabaseWithResponse(validCtx, "test-project", "abc1234567").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to pause database: connection refused",
		},
		{
			name: "API error",
			args: []string{"pause", "abc1234567"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().PauseDatabaseWithResponse(validCtx, "test-project", "abc1234567").
					Return(&api.PauseDatabaseResponse{
						HTTPResponse: httpResponse(http.StatusNotFound),
						JSONDefault:  &api.Error{Message: "database not found"},
					}, nil)
			},
			wantErr: "database not found",
		},
		{
			name: "nil response body",
			args: []string{"pause", "abc1234567"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().PauseDatabaseWithResponse(validCtx, "test-project", "abc1234567").
					Return(&api.PauseDatabaseResponse{
						HTTPResponse: httpResponse(http.StatusAccepted),
						JSON202:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name: "text output",
			args: []string{"pause", "abc1234567"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				db := sampleDatabase()
				m.EXPECT().PauseDatabaseWithResponse(validCtx, "test-project", "abc1234567").
					Return(&api.PauseDatabaseResponse{
						HTTPResponse: httpResponse(http.StatusAccepted),
						JSON202:      &db,
					}, nil)
			},
			wantStdout: "Pausing 'mydb' (abc1234567)...\n",
		},
	}

	runCmdTests(t, tests)
}
