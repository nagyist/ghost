package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestApiKeyCreateCmd(t *testing.T) {
	setupCreate := func(name string) func(m *mock.MockClientWithResponsesInterface) {
		return func(m *mock.MockClientWithResponsesInterface) {
			m.EXPECT().CreateApiKeyWithResponse(validCtx, "test-project", api.CreateApiKeyJSONRequestBody{Name: name}).
				Return(&api.CreateApiKeyResponse{
					HTTPResponse: httpResponse(http.StatusCreated),
					JSON201:      &api.ApiKeyCredentials{ApiKey: "gt_test_key_123"},
				}, nil)
		}
	}

	setupAuthInfo := func(m *mock.MockClientWithResponsesInterface) {
		m.EXPECT().AuthInfoWithResponse(validCtx).
			Return(&api.AuthInfoResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &api.AuthInfo{Type: api.AuthInfoTypeUser, User: &api.UserInfo{Name: "Alice"}},
			}, nil)
	}

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"api-key", "create"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error on create",
			args: []string{"api-key", "create", "--name", "CI Key"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().CreateApiKeyWithResponse(validCtx, "test-project", api.CreateApiKeyJSONRequestBody{Name: "CI Key"}).
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to create API key: connection refused",
		},
		{
			name: "API error on create",
			args: []string{"api-key", "create", "--name", "CI Key"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().CreateApiKeyWithResponse(validCtx, "test-project", api.CreateApiKeyJSONRequestBody{Name: "CI Key"}).
					Return(&api.CreateApiKeyResponse{
						HTTPResponse: httpResponse(http.StatusInternalServerError),
						JSONDefault:  &api.Error{Message: "internal error"},
					}, nil)
			},
			wantErr: "internal error",
		},
		{
			name: "nil response body on create",
			args: []string{"api-key", "create", "--name", "CI Key"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().CreateApiKeyWithResponse(validCtx, "test-project", api.CreateApiKeyJSONRequestBody{Name: "CI Key"}).
					Return(&api.CreateApiKeyResponse{
						HTTPResponse: httpResponse(http.StatusCreated),
						JSON201:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name: "with --name flag",
			args: []string{"api-key", "create", "--name", "CI Key"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupCreate("CI Key")(m)
			},
			wantStdout: "Created API key 'CI Key'\nAPI key: gt_test_key_123\n\nThis key will not be shown again. Make sure to save it.\n",
		},
		{
			name: "auto-generated name",
			args: []string{"api-key", "create"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupAuthInfo(m)
				setupCreate("Alice's API Key")(m)
			},
			wantStdout: "Created API key 'Alice's API Key'\nAPI key: gt_test_key_123\n\nThis key will not be shown again. Make sure to save it.\n",
		},
		{
			name: "json output",
			args: []string{"api-key", "create", "--name", "CI Key", "--json"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupCreate("CI Key")(m)
			},
			wantStdout: `{
  "api_key": "gt_test_key_123"
}
`,
		},
		{
			name: "yaml output",
			args: []string{"api-key", "create", "--name", "CI Key", "--yaml"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupCreate("CI Key")(m)
			},
			wantStdout: "api_key: gt_test_key_123\n",
		},
		{
			name: "env output",
			args: []string{"api-key", "create", "--name", "CI Key", "--env"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupCreate("CI Key")(m)
			},
			wantStdout: "GHOST_API_KEY=gt_test_key_123\n",
		},
	}

	runCmdTests(t, tests)
}
