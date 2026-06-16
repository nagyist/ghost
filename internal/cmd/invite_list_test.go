package cmd

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestInviteListCmd(t *testing.T) {
	experimental := withEnv("GHOST_EXPERIMENTAL", "true")
	createdAt := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)

	setupSent := func(m *mock.MockClientWithResponsesInterface) {
		invites := []api.Invite{
			{Email: "bob@example.com", Role: api.MemberRoleDeveloper, Status: api.InviteStatusPending, CreatedAt: createdAt},
		}
		m.EXPECT().ListInvitesWithResponse(validCtx, "test-project").
			Return(&api.ListInvitesResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &invites,
			}, nil)
	}
	setupReceived := func(m *mock.MockClientWithResponsesInterface) {
		invitations := []api.ReceivedInvite{
			{
				SpaceId:      "space-abc",
				SpaceName:    "Alice's space",
				InviterEmail: "alice@example.com",
				InviterName:  "Alice Smith",
				Role:         api.MemberRoleDeveloper,
				CreatedAt:    createdAt,
			},
		}
		m.EXPECT().ListReceivedInvitesWithResponse(validCtx).
			Return(&api.ListReceivedInvitesResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &invitations,
			}, nil)
	}

	setupSpaces := func(m *mock.MockClientWithResponsesInterface) {
		spaces := []api.Space{
			{Id: "test-project", Name: "Test Space"},
		}
		m.EXPECT().ListSpacesWithResponse(validCtx).
			Return(&api.ListSpacesResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &spaces,
			}, nil)
	}
	setupSentEmpty := func(m *mock.MockClientWithResponsesInterface) {
		invites := []api.Invite{}
		m.EXPECT().ListInvitesWithResponse(validCtx, "test-project").
			Return(&api.ListInvitesResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &invites,
			}, nil)
	}
	setupReceivedEmpty := func(m *mock.MockClientWithResponsesInterface) {
		invitations := []api.ReceivedInvite{}
		m.EXPECT().ListReceivedInvitesWithResponse(validCtx).
			Return(&api.ListReceivedInvitesResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &invitations,
			}, nil)
	}

	// successSetup mocks the two list calls (sufficient for JSON/YAML output).
	successSetup := func(m *mock.MockClientWithResponsesInterface) {
		setupSent(m)
		setupReceived(m)
	}

	// textSetup additionally mocks ListSpaces, which the text output uses to
	// resolve the current space name for the "Sent" header.
	textSetup := func(m *mock.MockClientWithResponsesInterface) {
		setupSent(m)
		setupReceived(m)
		setupSpaces(m)
	}

	sentTable := "Sent for Test Space (test-project)\n" +
		"EMAIL            ROLE       STATUS   INVITED               \n" +
		"bob@example.com  developer  pending  2026-01-02T15:04:05Z  \n"
	receivedTable := "Received\n" +
		"SPACE ID   SPACE          FROM               ROLE       INVITED               \n" +
		"space-abc  Alice's space  alice@example.com  developer  2026-01-02T15:04:05Z  \n"
	wantText := sentTable + "\n" + receivedTable

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"invite", "list"},
			opts:    []runOption{experimental, withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error on sent",
			args: []string{"invite", "list"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListInvitesWithResponse(validCtx, "test-project").
					Return(nil, errors.New("connection refused"))
				setupReceived(m)
			},
			wantErr: "failed to list sent invites: connection refused",
		},
		{
			name: "API error on sent",
			args: []string{"invite", "list"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListInvitesWithResponse(validCtx, "test-project").
					Return(&api.ListInvitesResponse{
						HTTPResponse: httpResponse(http.StatusForbidden),
						JSONDefault:  &api.Error{Message: "only the space owner or an admin can manage invites"},
					}, nil)
				setupReceived(m)
			},
			wantErr: "only the space owner or an admin can manage invites",
		},
		{
			name: "nil response body on sent",
			args: []string{"invite", "list"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListInvitesWithResponse(validCtx, "test-project").
					Return(&api.ListInvitesResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
				setupReceived(m)
			},
			wantErr: "empty response from API",
		},
		{
			name: "network error on received",
			args: []string{"invite", "list"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupSent(m)
				m.EXPECT().ListReceivedInvitesWithResponse(validCtx).
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to list received invites: connection refused",
		},
		{
			name: "API error on received",
			args: []string{"invite", "list"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupSent(m)
				m.EXPECT().ListReceivedInvitesWithResponse(validCtx).
					Return(&api.ListReceivedInvitesResponse{
						HTTPResponse: httpResponse(http.StatusForbidden),
						JSONDefault:  &api.Error{Message: "this endpoint requires user authentication"},
					}, nil)
			},
			wantErr: "this endpoint requires user authentication",
		},
		{
			name: "nil response body on received",
			args: []string{"invite", "list"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupSent(m)
				m.EXPECT().ListReceivedInvitesWithResponse(validCtx).
					Return(&api.ListReceivedInvitesResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name:       "text output",
			args:       []string{"invite", "list"},
			opts:       []runOption{experimental},
			setup:      textSetup,
			wantStdout: wantText,
		},
		{
			name:  "json output",
			args:  []string{"invite", "list", "--json"},
			opts:  []runOption{experimental},
			setup: successSetup,
			wantStdout: `{
  "sent": [
    {
      "created_at": "2026-01-02T15:04:05Z",
      "email": "bob@example.com",
      "role": "developer",
      "status": "pending"
    }
  ],
  "received": [
    {
      "created_at": "2026-01-02T15:04:05Z",
      "inviter_email": "alice@example.com",
      "inviter_name": "Alice Smith",
      "role": "developer",
      "space_id": "space-abc",
      "space_name": "Alice's space"
    }
  ]
}
`,
		},
		{
			name:  "yaml output",
			args:  []string{"invite", "list", "--yaml"},
			opts:  []runOption{experimental},
			setup: successSetup,
			wantStdout: `received:
  - created_at: "2026-01-02T15:04:05Z"
    inviter_email: alice@example.com
    inviter_name: Alice Smith
    role: developer
    space_id: space-abc
    space_name: Alice's space
sent:
  - created_at: "2026-01-02T15:04:05Z"
    email: bob@example.com
    role: developer
    status: pending
`,
		},
		{
			name:       "ls alias",
			args:       []string{"invite", "ls"},
			opts:       []runOption{experimental},
			setup:      textSetup,
			wantStdout: wantText,
		},
		{
			name: "only sent invites",
			args: []string{"invite", "list"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupSent(m)
				setupReceivedEmpty(m)
				setupSpaces(m)
			},
			wantStdout: sentTable,
		},
		{
			name: "only received invites",
			args: []string{"invite", "list"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupSentEmpty(m)
				setupReceived(m)
			},
			wantStdout: receivedTable,
		},
		{
			name: "no invites",
			args: []string{"invite", "list"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupSentEmpty(m)
				setupReceivedEmpty(m)
			},
			wantStdout: "No pending invites.\n",
		},
	}

	runCmdTests(t, tests)
}
