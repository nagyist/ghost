package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestSpaceCmd(t *testing.T) {
	experimental := withEnv("GHOST_EXPERIMENTAL", "true")

	// userSetup mocks a space resolved via user auth: role and owner populated.
	userSetup := func(m *mock.MockClientWithResponsesInterface) {
		m.EXPECT().GetSpaceWithResponse(validCtx, "test-space").
			Return(&api.GetSpaceResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200: &api.SpaceDetail{
					ID:   "test-space",
					Name: "Test Space",
					Role: new(api.MemberRoleDeveloper),
					Owner: &api.Member{
						UserID: 42,
						Name:   "Jane Doe",
						Email:  "jane@example.com",
						Role:   api.MemberRoleOwner,
					},
				},
			}, nil)
	}

	// apiKeySetup mocks a space resolved via API-key auth: role and owner omitted.
	apiKeySetup := func(m *mock.MockClientWithResponsesInterface) {
		m.EXPECT().GetSpaceWithResponse(validCtx, "test-space").
			Return(&api.GetSpaceResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200: &api.SpaceDetail{
					ID:   "test-space",
					Name: "Test Space",
				},
			}, nil)
	}

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"space"},
			opts:    []runOption{experimental, withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error",
			args: []string{"space"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetSpaceWithResponse(validCtx, "test-space").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to get space: connection refused",
		},
		{
			name: "API error",
			args: []string{"space"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetSpaceWithResponse(validCtx, "test-space").
					Return(&api.GetSpaceResponse{
						HTTPResponse: httpResponse(http.StatusInternalServerError),
						JSONDefault:  &api.Error{Message: "internal error"},
					}, nil)
			},
			wantErr: "internal error",
		},
		{
			name: "nil response body",
			args: []string{"space"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetSpaceWithResponse(validCtx, "test-space").
					Return(&api.GetSpaceResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name:  "text output",
			args:  []string{"space"},
			opts:  []runOption{experimental},
			setup: userSetup,
			wantStdout: `Space: Test Space (test-space)
Owner: Jane Doe (jane@example.com)
Role: developer
`,
		},
		{
			name:  "json output",
			args:  []string{"space", "--json"},
			opts:  []runOption{experimental},
			setup: userSetup,
			wantStdout: `{
  "id": "test-space",
  "name": "Test Space",
  "owner": {
    "email": "jane@example.com",
    "name": "Jane Doe",
    "role": "owner",
    "user_id": 42
  },
  "role": "developer"
}
`,
		},
		{
			name:  "yaml output",
			args:  []string{"space", "--yaml"},
			opts:  []runOption{experimental},
			setup: userSetup,
			wantStdout: `id: test-space
name: Test Space
owner:
  email: jane@example.com
  name: Jane Doe
  role: owner
  user_id: 42
role: developer
`,
		},
		{
			name:  "api key omits owner and role",
			args:  []string{"space"},
			opts:  []runOption{experimental},
			setup: apiKeySetup,
			wantStdout: `Space: Test Space (test-space)
`,
		},
	}

	runCmdTests(t, tests)
}
