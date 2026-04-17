package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestDeleteCmd(t *testing.T) {
	setupGet := func(m *mock.MockClientWithResponsesInterface) {
		db := sampleDatabase()
		m.EXPECT().GetDatabaseWithResponse(validCtx, "test-project", "abc1234567").
			Return(&api.GetDatabaseResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &db,
			}, nil)
	}

	setupDelete := func(m *mock.MockClientWithResponsesInterface) {
		m.EXPECT().DeleteDatabaseWithResponse(validCtx, "test-project", "abc1234567").
			Return(&api.DeleteDatabaseResponse{
				HTTPResponse: httpResponse(http.StatusAccepted),
			}, nil)
	}

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"delete", "abc1234567"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error on get",
			args: []string{"delete", "abc1234567"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetDatabaseWithResponse(validCtx, "test-project", "abc1234567").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to get database details: connection refused",
		},
		{
			name: "API error on get",
			args: []string{"delete", "abc1234567"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetDatabaseWithResponse(validCtx, "test-project", "abc1234567").
					Return(&api.GetDatabaseResponse{
						HTTPResponse: httpResponse(http.StatusNotFound),
						JSONDefault:  &api.Error{Message: "database not found"},
					}, nil)
			},
			wantErr: "database not found",
		},
		{
			name: "nil response body on get",
			args: []string{"delete", "abc1234567"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetDatabaseWithResponse(validCtx, "test-project", "abc1234567").
					Return(&api.GetDatabaseResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name: "non-interactive stdin",
			args: []string{"delete", "abc1234567"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGet(m)
			},
			wantErr: "cannot prompt for confirmation: stdin is not a terminal; use --confirm to skip",
		},
		{
			name: "confirmation declined",
			args: []string{"delete", "abc1234567"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGet(m)
			},
			opts:       []runOption{withStdin("n\n"), withIsTerminal(true)},
			wantStderr: "Delete 'mydb' (abc1234567)? This cannot be undone. [y/N] ",
			wantStdout: "Delete operation cancelled.\n",
		},
		{
			name: "network error on delete",
			args: []string{"delete", "abc1234567"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGet(m)
				m.EXPECT().DeleteDatabaseWithResponse(validCtx, "test-project", "abc1234567").
					Return(nil, errors.New("connection refused"))
			},
			opts:       []runOption{withStdin("y\n"), withIsTerminal(true)},
			wantErr:    "failed to delete database: connection refused",
			wantStderr: "Delete 'mydb' (abc1234567)? This cannot be undone. [y/N] Error: failed to delete database: connection refused\n",
		},
		{
			name: "API error on delete",
			args: []string{"delete", "abc1234567"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGet(m)
				m.EXPECT().DeleteDatabaseWithResponse(validCtx, "test-project", "abc1234567").
					Return(&api.DeleteDatabaseResponse{
						HTTPResponse: httpResponse(http.StatusInternalServerError),
						JSONDefault:  &api.Error{Message: "internal server error"},
					}, nil)
			},
			opts:       []runOption{withStdin("y\n"), withIsTerminal(true)},
			wantErr:    "internal server error",
			wantStderr: "Delete 'mydb' (abc1234567)? This cannot be undone. [y/N] Error: internal server error\n",
		},
		{
			name: "confirmation accepted",
			args: []string{"delete", "abc1234567"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGet(m)
				setupDelete(m)
			},
			opts:       []runOption{withStdin("y\n"), withIsTerminal(true)},
			wantStderr: "Delete 'mydb' (abc1234567)? This cannot be undone. [y/N] ",
			wantStdout: "Deleted 'mydb' (abc1234567)\n",
		},
		{
			name: "confirm flag",
			args: []string{"delete", "abc1234567", "--confirm"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGet(m)
				setupDelete(m)
			},
			wantStdout: "Deleted 'mydb' (abc1234567)\n",
		},
		{
			name: "rm alias",
			args: []string{"rm", "abc1234567", "--confirm"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGet(m)
				setupDelete(m)
			},
			wantStdout: "Deleted 'mydb' (abc1234567)\n",
		},
	}

	runCmdTests(t, tests)
}
