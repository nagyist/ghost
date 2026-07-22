package cmd

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestInviteReceivedCmd(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)

	successSetup := func(m *mock.MockClientWithResponsesInterface) {
		invitations := []api.ReceivedInvite{
			{
				SpaceID:      "space-abc",
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

	wantText := "SPACE ID   SPACE          FROM               ROLE       INVITED               \n" +
		"space-abc  Alice's space  alice@example.com  developer  2026-01-02T15:04:05Z  \n"

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"invite", "received"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error",
			args: []string{"invite", "received"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListReceivedInvitesWithResponse(validCtx).
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to list invitations: connection refused",
		},
		{
			name: "API error",
			args: []string{"invite", "received"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListReceivedInvitesWithResponse(validCtx).
					Return(&api.ListReceivedInvitesResponse{
						HTTPResponse: httpResponse(http.StatusForbidden),
						JSONDefault:  &api.Error{Message: "this endpoint requires user authentication"},
					}, nil)
			},
			wantErr: "this endpoint requires user authentication",
		},
		{
			name: "nil response body",
			args: []string{"invite", "received"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
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
			args:       []string{"invite", "received"},
			setup:      successSetup,
			wantStdout: wantText,
		},
		{
			name:  "json output",
			args:  []string{"invite", "received", "--json"},
			setup: successSetup,
			wantStdout: `[
  {
    "created_at": "2026-01-02T15:04:05Z",
    "inviter_email": "alice@example.com",
    "inviter_name": "Alice Smith",
    "role": "developer",
    "space_id": "space-abc",
    "space_name": "Alice's space"
  }
]
`,
		},
		{
			name:  "yaml output",
			args:  []string{"invite", "received", "--yaml"},
			setup: successSetup,
			wantStdout: `- created_at: "2026-01-02T15:04:05Z"
  inviter_email: alice@example.com
  inviter_name: Alice Smith
  role: developer
  space_id: space-abc
  space_name: Alice's space
`,
		},
	}

	runCmdTests(t, tests)
}
