package cmd

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestInviteDeclineCmd(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)

	setupList := func(m *mock.MockClientWithResponsesInterface) {
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
	setupDecline := func(m *mock.MockClientWithResponsesInterface) {
		result := api.InviteActionResult{SpaceID: "space-abc", SpaceName: "Alice's space"}
		m.EXPECT().DeclineInviteWithResponse(validCtx, api.SpaceID("space-abc")).
			Return(&api.DeclineInviteResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &result,
			}, nil)
	}

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"invite", "decline", "space-abc"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error on list",
			args: []string{"invite", "decline", "space-abc"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListReceivedInvitesWithResponse(validCtx).
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to list invitations: connection refused",
		},
		{
			name:    "invitation not found",
			args:    []string{"invite", "decline", "nonexistent"},
			setup:   setupList,
			wantErr: "no pending invitation to space 'nonexistent' found; run 'ghost invite received' to see your invitations",
		},
		{
			name:    "non-terminal stdin without confirm flag",
			args:    []string{"invite", "decline", "space-abc"},
			setup:   setupList,
			wantErr: "cannot prompt for confirmation: stdin is not a terminal; use --confirm to skip",
		},
		{
			name:       "confirmation declined",
			args:       []string{"invite", "decline", "space-abc"},
			opts:       []runOption{withStdin("n\n"), withIsTerminal(true)},
			setup:      setupList,
			wantStderr: "Decline the invitation to 'Alice's space'? [y/N] ",
			wantStdout: "Decline operation aborted.\n",
		},
		{
			name: "network error on decline",
			args: []string{"invite", "decline", "space-abc", "--confirm"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupList(m)
				m.EXPECT().DeclineInviteWithResponse(validCtx, api.SpaceID("space-abc")).
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to decline invitation: connection refused",
		},
		{
			name: "nil response body on decline",
			args: []string{"invite", "decline", "space-abc", "--confirm"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupList(m)
				m.EXPECT().DeclineInviteWithResponse(validCtx, api.SpaceID("space-abc")).
					Return(&api.DeclineInviteResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name: "confirmation accepted",
			args: []string{"invite", "decline", "space-abc"},
			opts: []runOption{withStdin("y\n"), withIsTerminal(true)},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupList(m)
				setupDecline(m)
			},
			wantStderr: "Decline the invitation to 'Alice's space'? [y/N] ",
			wantStdout: "Declined the invitation to 'Alice's space'\n",
		},
		{
			name: "confirm flag",
			args: []string{"invite", "decline", "space-abc", "--confirm"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupList(m)
				setupDecline(m)
			},
			wantStdout: "Declined the invitation to 'Alice's space'\n",
		},
	}

	runCmdTests(t, tests)
}
