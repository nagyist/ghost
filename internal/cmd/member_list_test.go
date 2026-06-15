package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestMemberListCmd(t *testing.T) {
	experimental := withEnv("GHOST_EXPERIMENTAL", "true")

	successSetup := func(m *mock.MockClientWithResponsesInterface) {
		members := []api.Member{
			{UserId: 101, Name: "Alice Smith", Email: "alice@example.com", Role: api.MemberRoleOwner},
			{UserId: 102, Name: "Bob Jones", Email: "bob@example.com", Role: api.MemberRoleDeveloper},
		}
		m.EXPECT().ListMembersWithResponse(validCtx, "test-project").
			Return(&api.ListMembersResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &members,
			}, nil)
	}

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"member", "list"},
			opts:    []runOption{experimental, withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error",
			args: []string{"member", "list"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListMembersWithResponse(validCtx, "test-project").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to list members: connection refused",
		},
		{
			name: "API error",
			args: []string{"member", "list"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListMembersWithResponse(validCtx, "test-project").
					Return(&api.ListMembersResponse{
						HTTPResponse: httpResponse(http.StatusForbidden),
						JSONDefault:  &api.Error{Message: "this endpoint requires user authentication"},
					}, nil)
			},
			wantErr: "this endpoint requires user authentication",
		},
		{
			name: "nil response body",
			args: []string{"member", "list"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListMembersWithResponse(validCtx, "test-project").
					Return(&api.ListMembersResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name:       "text output",
			args:       []string{"member", "list"},
			opts:       []runOption{experimental},
			setup:      successSetup,
			wantStdout: "NAME         EMAIL              ROLE       \nAlice Smith  alice@example.com  owner      \nBob Jones    bob@example.com    developer  \n",
		},
		{
			name:  "json output",
			args:  []string{"member", "list", "--json"},
			opts:  []runOption{experimental},
			setup: successSetup,
			wantStdout: `[
  {
    "user_id": 101,
    "name": "Alice Smith",
    "email": "alice@example.com",
    "role": "owner"
  },
  {
    "user_id": 102,
    "name": "Bob Jones",
    "email": "bob@example.com",
    "role": "developer"
  }
]
`,
		},
		{
			name:  "yaml output",
			args:  []string{"member", "list", "--yaml"},
			opts:  []runOption{experimental},
			setup: successSetup,
			wantStdout: `- email: alice@example.com
  name: Alice Smith
  role: owner
  user_id: 101
- email: bob@example.com
  name: Bob Jones
  role: developer
  user_id: 102
`,
		},
		{
			name:       "ls alias",
			args:       []string{"member", "ls"},
			opts:       []runOption{experimental},
			setup:      successSetup,
			wantStdout: "NAME         EMAIL              ROLE       \nAlice Smith  alice@example.com  owner      \nBob Jones    bob@example.com    developer  \n",
		},
	}

	runCmdTests(t, tests)
}
