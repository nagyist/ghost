package cmd

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestInviteCmd(t *testing.T) {
	experimental := withEnv("GHOST_EXPERIMENTAL", "true")
	createdAt := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)

	setupCreate := func(role api.MemberRole) func(m *mock.MockClientWithResponsesInterface) {
		return func(m *mock.MockClientWithResponsesInterface) {
			created := api.Invite{
				Email:     "bob@example.com",
				Role:      role,
				Status:    api.InviteStatusPending,
				CreatedAt: createdAt,
			}
			m.EXPECT().CreateInviteWithResponse(validCtx, "test-project", api.CreateInviteJSONRequestBody{
				Email: "bob@example.com",
				Role:  new(role),
			}).
				Return(&api.CreateInviteResponse{
					HTTPResponse: httpResponse(http.StatusCreated),
					JSON201:      &created,
				}, nil)
		}
	}
	setupListSpaces := func(m *mock.MockClientWithResponsesInterface) {
		spaces := []api.Space{
			{Id: "test-project", Name: "Test Space"},
		}
		m.EXPECT().ListSpacesWithResponse(validCtx).
			Return(&api.ListSpacesResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &spaces,
			}, nil)
	}

	wantText := "Invited bob@example.com to space Test Space (test-project) as a developer.\n\n" +
		"If they're new to Ghost, the invitation will be accepted automatically\n" +
		"at signup. If they already use Ghost, they can accept the invite with\n" +
		"'ghost invite accept test-project'.\n\n" +
		"Note: the invitation is tied to bob@example.com; they must log in to\n" +
		"Ghost with a GitHub account with that primary email.\n"

	wantTextAdmin := "Invited bob@example.com to space Test Space (test-project) as an admin.\n\n" +
		"If they're new to Ghost, the invitation will be accepted automatically\n" +
		"at signup. If they already use Ghost, they can accept the invite with\n" +
		"'ghost invite accept test-project'.\n\n" +
		"Note: the invitation is tied to bob@example.com; they must log in to\n" +
		"Ghost with a GitHub account with that primary email.\n"

	tests := []cmdTest{
		{
			name:    "invalid role",
			args:    []string{"invite", "bob@example.com", "--role", "superuser"},
			opts:    []runOption{experimental},
			wantErr: "invalid role 'superuser'; must be one of admin, developer, or viewer",
		},
		{
			name:    "owner role rejected",
			args:    []string{"invite", "bob@example.com", "--role", "owner"},
			opts:    []runOption{experimental},
			wantErr: "the owner role cannot be granted; every space has exactly one owner",
		},
		{
			name:    "not logged in",
			args:    []string{"invite", "bob@example.com"},
			opts:    []runOption{experimental, withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error",
			args: []string{"invite", "bob@example.com"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().CreateInviteWithResponse(validCtx, "test-project", api.CreateInviteJSONRequestBody{
					Email: "bob@example.com",
					Role:  new(api.MemberRoleDeveloper),
				}).
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to create invite: connection refused",
		},
		{
			name: "API error",
			args: []string{"invite", "bob@example.com"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().CreateInviteWithResponse(validCtx, "test-project", api.CreateInviteJSONRequestBody{
					Email: "bob@example.com",
					Role:  new(api.MemberRoleDeveloper),
				}).
					Return(&api.CreateInviteResponse{
						HTTPResponse: httpResponse(http.StatusConflict),
						JSONDefault:  &api.Error{Message: "that email already belongs to a member of this space"},
					}, nil)
			},
			wantErr: "that email already belongs to a member of this space",
		},
		{
			name: "nil response body",
			args: []string{"invite", "bob@example.com"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().CreateInviteWithResponse(validCtx, "test-project", api.CreateInviteJSONRequestBody{
					Email: "bob@example.com",
					Role:  new(api.MemberRoleDeveloper),
				}).
					Return(&api.CreateInviteResponse{
						HTTPResponse: httpResponse(http.StatusCreated),
						JSON201:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name: "text output",
			args: []string{"invite", "bob@example.com"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupCreate(api.MemberRoleDeveloper)(m)
				setupListSpaces(m)
			},
			wantStdout: wantText,
		},
		{
			name: "email is lowercased",
			args: []string{"invite", "Bob@Example.com"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupCreate(api.MemberRoleDeveloper)(m)
				setupListSpaces(m)
			},
			wantStdout: wantText,
		},
		{
			name:  "json output",
			args:  []string{"invite", "bob@example.com", "--json"},
			opts:  []runOption{experimental},
			setup: setupCreate(api.MemberRoleDeveloper),
			wantStdout: `{
  "created_at": "2026-01-02T15:04:05Z",
  "email": "bob@example.com",
  "role": "developer",
  "status": "pending"
}
`,
		},
		{
			name:  "yaml output",
			args:  []string{"invite", "bob@example.com", "--yaml"},
			opts:  []runOption{experimental},
			setup: setupCreate(api.MemberRoleDeveloper),
			wantStdout: `created_at: "2026-01-02T15:04:05Z"
email: bob@example.com
role: developer
status: pending
`,
		},
		{
			name:  "admin role json",
			args:  []string{"invite", "bob@example.com", "--role", "admin", "--json"},
			opts:  []runOption{experimental},
			setup: setupCreate(api.MemberRoleAdmin),
			wantStdout: `{
  "created_at": "2026-01-02T15:04:05Z",
  "email": "bob@example.com",
  "role": "admin",
  "status": "pending"
}
`,
		},
		{
			name: "admin role text output uses 'an'",
			args: []string{"invite", "bob@example.com", "--role", "admin"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupCreate(api.MemberRoleAdmin)(m)
				setupListSpaces(m)
			},
			wantStdout: wantTextAdmin,
		},
	}

	runCmdTests(t, tests)
}
