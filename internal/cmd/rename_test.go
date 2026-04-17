package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestRenameCmd(t *testing.T) {
	setupGet := func(m *mock.MockClientWithResponsesInterface) {
		db := sampleDatabase()
		m.EXPECT().GetDatabaseWithResponse(validCtx, "test-project", "abc1234567").
			Return(&api.GetDatabaseResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &db,
			}, nil)
	}

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"rename", "abc1234567", "new-name"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error on get",
			args: []string{"rename", "abc1234567", "new-name"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetDatabaseWithResponse(validCtx, "test-project", "abc1234567").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to get database details: connection refused",
		},
		{
			name: "API error on get",
			args: []string{"rename", "abc1234567", "new-name"},
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
			args: []string{"rename", "abc1234567", "new-name"},
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
			name: "network error on rename",
			args: []string{"rename", "abc1234567", "new-name"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGet(m)
				m.EXPECT().RenameDatabaseWithResponse(validCtx, "test-project", "abc1234567", api.RenameDatabaseRequest{Name: "new-name"}).
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to rename database: connection refused",
		},
		{
			name: "API error on rename",
			args: []string{"rename", "abc1234567", "new-name"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGet(m)
				m.EXPECT().RenameDatabaseWithResponse(validCtx, "test-project", "abc1234567", api.RenameDatabaseRequest{Name: "new-name"}).
					Return(&api.RenameDatabaseResponse{
						HTTPResponse: httpResponse(http.StatusNotFound),
						JSONDefault:  &api.Error{Message: "database not found"},
					}, nil)
			},
			wantErr: "database not found",
		},
		{
			name: "success",
			args: []string{"rename", "abc1234567", "new-name"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGet(m)
				m.EXPECT().RenameDatabaseWithResponse(validCtx, "test-project", "abc1234567", api.RenameDatabaseRequest{Name: "new-name"}).
					Return(&api.RenameDatabaseResponse{
						HTTPResponse: httpResponse(http.StatusNoContent),
					}, nil)
			},
			wantStdout: "Renamed 'mydb' (abc1234567) to 'new-name'\n",
		},
	}

	runCmdTests(t, tests)
}
