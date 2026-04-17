package cmd

import (
	"errors"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestPasswordCmd(t *testing.T) {
	successfulGet := func(m *mock.MockClientWithResponsesInterface) {
		db := sampleDatabase()
		m.EXPECT().GetDatabaseWithResponse(validCtx, "test-project", "abc1234567").
			Return(&api.GetDatabaseResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &db,
			}, nil)
	}

	// successfulUpdate uses gomock.Any() for the request body because the
	// --generate flag produces a random password that can't be predicted.
	successfulUpdate := func(m *mock.MockClientWithResponsesInterface) {
		m.EXPECT().UpdatePasswordWithResponse(validCtx, "test-project", "abc1234567", gomock.Any()).
			Return(&api.UpdatePasswordResponse{
				HTTPResponse: httpResponse(http.StatusNoContent),
			}, nil)
	}

	tests := []cmdTest{
		{
			name:    "generate and password conflict",
			args:    []string{"password", "abc1234567", "mypass", "--generate"},
			wantErr: "cannot use --generate when password is provided as an argument",
		},
		{
			name:    "not logged in",
			args:    []string{"password", "abc1234567", "--generate"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error on get",
			args: []string{"password", "abc1234567", "--generate"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetDatabaseWithResponse(validCtx, "test-project", "abc1234567").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to get database details: connection refused",
		},
		{
			name: "API error on get",
			args: []string{"password", "abc1234567", "--generate"},
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
			args: []string{"password", "abc1234567", "--generate"},
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
			name: "database paused",
			args: []string{"password", "abc1234567", "--generate"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				db := sampleDatabase(func(db *api.Database) {
					db.Status = api.DatabaseStatusPaused
				})
				m.EXPECT().GetDatabaseWithResponse(validCtx, "test-project", "abc1234567").
					Return(&api.GetDatabaseResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &db,
					}, nil)
			},
			wantErr: "database is currently paused — resume it with 'ghost resume abc1234567'",
		},
		{
			name: "database not ready",
			args: []string{"password", "abc1234567", "--generate"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				db := sampleDatabase(func(db *api.Database) {
					db.Status = api.DatabaseStatusConfiguring
				})
				m.EXPECT().GetDatabaseWithResponse(validCtx, "test-project", "abc1234567").
					Return(&api.GetDatabaseResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &db,
					}, nil)
			},
			wantErr: "database is not yet ready — check status with 'ghost list' and try again",
		},
		{
			name:    "no password non-tty stdin",
			args:    []string{"password", "abc1234567"},
			setup:   successfulGet,
			wantErr: "no password provided and stdin is not a terminal; use --generate or provide password as argument",
		},
		{
			name: "network error on update",
			args: []string{"password", "abc1234567", "newpass123"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				successfulGet(m)
				m.EXPECT().UpdatePasswordWithResponse(validCtx, "test-project", "abc1234567", api.UpdatePasswordRequest{Password: "newpass123"}).
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to update password: connection refused",
		},
		{
			name: "API error on update",
			args: []string{"password", "abc1234567", "newpass123"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				successfulGet(m)
				m.EXPECT().UpdatePasswordWithResponse(validCtx, "test-project", "abc1234567", api.UpdatePasswordRequest{Password: "newpass123"}).
					Return(&api.UpdatePasswordResponse{
						HTTPResponse: httpResponse(http.StatusInternalServerError),
						JSONDefault:  &api.Error{Message: "internal error"},
					}, nil)
			},
			wantErr: "internal error",
		},
		{
			name: "password from argument",
			args: []string{"password", "abc1234567", "newpass123"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				successfulGet(m)
				m.EXPECT().UpdatePasswordWithResponse(validCtx, "test-project", "abc1234567", api.UpdatePasswordRequest{Password: "newpass123"}).
					Return(&api.UpdatePasswordResponse{
						HTTPResponse: httpResponse(http.StatusNoContent),
					}, nil)
			},
			wantStdout: "Password updated for 'mydb'\n",
		},
		{
			name: "generate flag",
			args: []string{"password", "abc1234567", "--generate"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				successfulGet(m)
				successfulUpdate(m)
			},
			wantStdout: "Password updated for 'mydb'\n",
		},
	}

	runCmdTests(t, tests)
}
