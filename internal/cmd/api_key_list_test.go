package cmd

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestApiKeyListCmd(t *testing.T) {
	keys := []api.ApiKey{
		{Prefix: "gt_abc", Name: "My Key", CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)},
	}

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"api-key", "list"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error",
			args: []string{"api-key", "list"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListApiKeysWithResponse(validCtx, "test-project").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to list API keys: connection refused",
		},
		{
			name: "API error",
			args: []string{"api-key", "list"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListApiKeysWithResponse(validCtx, "test-project").
					Return(&api.ListApiKeysResponse{
						HTTPResponse: httpResponse(http.StatusInternalServerError),
						JSONDefault:  &api.Error{Message: "internal error"},
					}, nil)
			},
			wantErr: "internal error",
		},
		{
			name: "nil response body",
			args: []string{"api-key", "list"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListApiKeysWithResponse(validCtx, "test-project").
					Return(&api.ListApiKeysResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name: "empty list",
			args: []string{"api-key", "list"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				empty := []api.ApiKey{}
				m.EXPECT().ListApiKeysWithResponse(validCtx, "test-project").
					Return(&api.ListApiKeysResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &empty,
					}, nil)
			},
			wantStdout: "No API keys found.\nRun 'ghost api-key create' to create an API key.\n",
		},
		{
			name: "text output",
			args: []string{"api-key", "list"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListApiKeysWithResponse(validCtx, "test-project").
					Return(&api.ListApiKeysResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &keys,
					}, nil)
			},
			wantStdout: "PREFIX  NAME    CREATED AT                     \n" +
				"gt_abc  My Key  2024-01-15 10:30:00 +0000 UTC  \n",
		},
		{
			name: "json output",
			args: []string{"api-key", "list", "--json"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListApiKeysWithResponse(validCtx, "test-project").
					Return(&api.ListApiKeysResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &keys,
					}, nil)
			},
			wantStdout: `[
  {
    "prefix": "gt_abc",
    "name": "My Key",
    "created_at": "2024-01-15 10:30:00 +0000 UTC"
  }
]
`,
		},
		{
			name: "yaml output",
			args: []string{"api-key", "list", "--yaml"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListApiKeysWithResponse(validCtx, "test-project").
					Return(&api.ListApiKeysResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &keys,
					}, nil)
			},
			wantStdout: `- created_at: 2024-01-15 10:30:00 +0000 UTC
  name: My Key
  prefix: gt_abc
`,
		},
		{
			name: "ls alias",
			args: []string{"api-key", "ls"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListApiKeysWithResponse(validCtx, "test-project").
					Return(&api.ListApiKeysResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &keys,
					}, nil)
			},
			wantStdout: "PREFIX  NAME    CREATED AT                     \n" +
				"gt_abc  My Key  2024-01-15 10:30:00 +0000 UTC  \n",
		},
	}

	runCmdTests(t, tests)
}
