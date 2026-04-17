package cmd

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

// logsParamsMatcher returns a gomock.Matcher that checks DatabaseLogsParams
// fields. Page must match exactly. If until is non-zero, Until must match
// exactly; otherwise Until is only checked for non-nil (since FetchLogs
// defaults to time.Now()).
func logsParamsMatcher(page int, until time.Time) gomock.Matcher {
	return gomock.Cond(func(x any) bool {
		params, ok := x.(*api.DatabaseLogsParams)
		if !ok || params == nil {
			return false
		}
		if params.Page == nil || *params.Page != page {
			return false
		}
		if !until.IsZero() {
			return params.Until != nil && params.Until.Equal(until)
		}
		return params.Until != nil
	})
}

func TestLogsCmd(t *testing.T) {
	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"logs", "abc1234567"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error",
			args: []string{"logs", "abc1234567"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().DatabaseLogsWithResponse(validCtx, "test-project", "abc1234567", logsParamsMatcher(0, time.Time{})).
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to fetch logs: connection refused",
		},
		{
			name: "API error",
			args: []string{"logs", "abc1234567"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().DatabaseLogsWithResponse(validCtx, "test-project", "abc1234567", logsParamsMatcher(0, time.Time{})).
					Return(&api.DatabaseLogsResponse{
						HTTPResponse: httpResponse(http.StatusInternalServerError),
						JSONDefault:  &api.Error{Message: "internal error"},
					}, nil)
			},
			wantErr: "internal error",
		},
		{
			name: "nil response body",
			args: []string{"logs", "abc1234567"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().DatabaseLogsWithResponse(validCtx, "test-project", "abc1234567", logsParamsMatcher(0, time.Time{})).
					Return(&api.DatabaseLogsResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name: "text output",
			args: []string{"logs", "abc1234567"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				gomock.InOrder(
					m.EXPECT().DatabaseLogsWithResponse(validCtx, "test-project", "abc1234567", logsParamsMatcher(0, time.Time{})).
						Return(&api.DatabaseLogsResponse{
							HTTPResponse: httpResponse(http.StatusOK),
							JSON200:      &api.LogsResponse{Logs: []string{"2024-01-02 LOG: msg2", "2024-01-01 LOG: msg1"}},
						}, nil),
					m.EXPECT().DatabaseLogsWithResponse(validCtx, "test-project", "abc1234567", logsParamsMatcher(1, time.Time{})).
						Return(&api.DatabaseLogsResponse{
							HTTPResponse: httpResponse(http.StatusOK),
							JSON200:      &api.LogsResponse{Logs: []string{}},
						}, nil),
				)
			},
			wantStdout: "2024-01-01 LOG: msg1\n2024-01-02 LOG: msg2\n", // ANSI stripped by test helper
		},
		{
			name: "json output",
			args: []string{"logs", "abc1234567", "--json"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				gomock.InOrder(
					m.EXPECT().DatabaseLogsWithResponse(validCtx, "test-project", "abc1234567", logsParamsMatcher(0, time.Time{})).
						Return(&api.DatabaseLogsResponse{
							HTTPResponse: httpResponse(http.StatusOK),
							JSON200:      &api.LogsResponse{Logs: []string{"2024-01-02 LOG: msg2", "2024-01-01 LOG: msg1"}},
						}, nil),
					m.EXPECT().DatabaseLogsWithResponse(validCtx, "test-project", "abc1234567", logsParamsMatcher(1, time.Time{})).
						Return(&api.DatabaseLogsResponse{
							HTTPResponse: httpResponse(http.StatusOK),
							JSON200:      &api.LogsResponse{Logs: []string{}},
						}, nil),
				)
			},
			wantStdout: `[
  "2024-01-01 LOG: msg1",
  "2024-01-02 LOG: msg2"
]
`,
		},
		{
			name: "yaml output",
			args: []string{"logs", "abc1234567", "--yaml"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				gomock.InOrder(
					m.EXPECT().DatabaseLogsWithResponse(validCtx, "test-project", "abc1234567", logsParamsMatcher(0, time.Time{})).
						Return(&api.DatabaseLogsResponse{
							HTTPResponse: httpResponse(http.StatusOK),
							JSON200:      &api.LogsResponse{Logs: []string{"2024-01-02 LOG: msg2", "2024-01-01 LOG: msg1"}},
						}, nil),
					m.EXPECT().DatabaseLogsWithResponse(validCtx, "test-project", "abc1234567", logsParamsMatcher(1, time.Time{})).
						Return(&api.DatabaseLogsResponse{
							HTTPResponse: httpResponse(http.StatusOK),
							JSON200:      &api.LogsResponse{Logs: []string{}},
						}, nil),
				)
			},
			wantStdout: `- '2024-01-01 LOG: msg1'
- '2024-01-02 LOG: msg2'
`,
		},
		{
			name: "empty logs",
			args: []string{"logs", "abc1234567"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().DatabaseLogsWithResponse(validCtx, "test-project", "abc1234567", logsParamsMatcher(0, time.Time{})).
					Return(&api.DatabaseLogsResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &api.LogsResponse{Logs: []string{}},
					}, nil)
			},
			wantStdout: "",
		},
		{
			name: "tail flag",
			args: []string{"logs", "abc1234567", "--tail", "1"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().DatabaseLogsWithResponse(validCtx, "test-project", "abc1234567", logsParamsMatcher(0, time.Time{})).
					Return(&api.DatabaseLogsResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &api.LogsResponse{Logs: []string{"log2", "log1"}},
					}, nil)
			},
			wantStdout: "log2\n",
		},
		{
			name: "until flag",
			args: []string{"logs", "abc1234567", "--until", "2024-06-15T10:00:00Z"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				until := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
				gomock.InOrder(
					m.EXPECT().DatabaseLogsWithResponse(validCtx, "test-project", "abc1234567", logsParamsMatcher(0, until)).
						Return(&api.DatabaseLogsResponse{
							HTTPResponse: httpResponse(http.StatusOK),
							JSON200:      &api.LogsResponse{Logs: []string{"log before cutoff"}},
						}, nil),
					m.EXPECT().DatabaseLogsWithResponse(validCtx, "test-project", "abc1234567", logsParamsMatcher(1, until)).
						Return(&api.DatabaseLogsResponse{
							HTTPResponse: httpResponse(http.StatusOK),
							JSON200:      &api.LogsResponse{Logs: []string{}},
						}, nil),
				)
			},
			wantStdout: "log before cutoff\n",
		},
	}

	runCmdTests(t, tests)
}
