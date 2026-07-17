package cmd

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestInviteSentCmd(t *testing.T) {
	experimental := withEnv("GHOST_EXPERIMENTAL", "true")
	createdAt := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)

	successSetup := func(m *mock.MockClientWithResponsesInterface) {
		invites := []api.Invite{
			{Email: "bob@example.com", Role: api.MemberRoleDeveloper, Status: api.InviteStatusPending, CreatedAt: createdAt},
			{Email: "carol@example.com", Role: api.MemberRoleViewer, Status: api.InviteStatusDeclined, CreatedAt: createdAt},
		}
		m.EXPECT().ListInvitesWithResponse(validCtx, "test-space").
			Return(&api.ListInvitesResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &invites,
			}, nil)
	}

	wantText := "EMAIL              ROLE       STATUS    INVITED               \n" +
		"bob@example.com    developer  pending   2026-01-02T15:04:05Z  \n" +
		"carol@example.com  viewer     declined  2026-01-02T15:04:05Z  \n"

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"invite", "sent"},
			opts:    []runOption{experimental, withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error",
			args: []string{"invite", "sent"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListInvitesWithResponse(validCtx, "test-space").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to list invites: connection refused",
		},
		{
			name: "API error",
			args: []string{"invite", "sent"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListInvitesWithResponse(validCtx, "test-space").
					Return(&api.ListInvitesResponse{
						HTTPResponse: httpResponse(http.StatusForbidden),
						JSONDefault:  &api.Error{Message: "only the space owner or an admin can manage invites"},
					}, nil)
			},
			wantErr: "only the space owner or an admin can manage invites",
		},
		{
			name: "nil response body",
			args: []string{"invite", "sent"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListInvitesWithResponse(validCtx, "test-space").
					Return(&api.ListInvitesResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name:       "text output",
			args:       []string{"invite", "sent"},
			opts:       []runOption{experimental},
			setup:      successSetup,
			wantStdout: wantText,
		},
		{
			name:  "json output",
			args:  []string{"invite", "sent", "--json"},
			opts:  []runOption{experimental},
			setup: successSetup,
			wantStdout: `[
  {
    "created_at": "2026-01-02T15:04:05Z",
    "email": "bob@example.com",
    "role": "developer",
    "status": "pending"
  },
  {
    "created_at": "2026-01-02T15:04:05Z",
    "email": "carol@example.com",
    "role": "viewer",
    "status": "declined"
  }
]
`,
		},
		{
			name:  "yaml output",
			args:  []string{"invite", "sent", "--yaml"},
			opts:  []runOption{experimental},
			setup: successSetup,
			wantStdout: `- created_at: "2026-01-02T15:04:05Z"
  email: bob@example.com
  role: developer
  status: pending
- created_at: "2026-01-02T15:04:05Z"
  email: carol@example.com
  role: viewer
  status: declined
`,
		},
	}

	runCmdTests(t, tests)
}
