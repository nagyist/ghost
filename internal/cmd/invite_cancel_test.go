package cmd

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestInviteCancelCmd(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)

	setupCancel := func(m *mock.MockClientWithResponsesInterface) {
		cancelled := api.Invite{Email: "bob@example.com", Role: api.MemberRoleDeveloper, CreatedAt: createdAt}
		m.EXPECT().CancelInviteWithResponse(validCtx, "test-space", api.InviteEmail("bob@example.com")).
			Return(&api.CancelInviteResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &cancelled,
			}, nil)
	}

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"invite", "cancel", "bob@example.com"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name:    "non-terminal stdin without confirm flag",
			args:    []string{"invite", "cancel", "bob@example.com"},
			wantErr: "cannot prompt for confirmation: stdin is not a terminal; use --confirm to skip",
		},
		{
			name:       "confirmation declined",
			args:       []string{"invite", "cancel", "bob@example.com"},
			opts:       []runOption{withStdin("n\n"), withIsTerminal(true)},
			wantStderr: "Cancel the invite to bob@example.com? [y/N] ",
			wantStdout: "Cancel operation aborted.\n",
		},
		{
			name: "network error on cancel",
			args: []string{"invite", "cancel", "bob@example.com", "--confirm"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().CancelInviteWithResponse(validCtx, "test-space", api.InviteEmail("bob@example.com")).
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to cancel invite: connection refused",
		},
		{
			name: "API error (not found)",
			args: []string{"invite", "cancel", "dave@example.com", "--confirm"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().CancelInviteWithResponse(validCtx, "test-space", api.InviteEmail("dave@example.com")).
					Return(&api.CancelInviteResponse{
						HTTPResponse: httpResponse(http.StatusNotFound),
						JSONDefault:  &api.Error{Message: "no pending invite to that email was found for this space"},
					}, nil)
			},
			wantErr: "no pending invite to that email was found for this space",
		},
		{
			name: "nil response body on cancel",
			args: []string{"invite", "cancel", "bob@example.com", "--confirm"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().CancelInviteWithResponse(validCtx, "test-space", api.InviteEmail("bob@example.com")).
					Return(&api.CancelInviteResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name:       "confirmation accepted",
			args:       []string{"invite", "cancel", "bob@example.com"},
			opts:       []runOption{withStdin("y\n"), withIsTerminal(true)},
			setup:      setupCancel,
			wantStderr: "Cancel the invite to bob@example.com? [y/N] ",
			wantStdout: "Cancelled the invite to bob@example.com\n",
		},
		{
			name:       "confirm flag",
			args:       []string{"invite", "cancel", "bob@example.com", "--confirm"},
			setup:      setupCancel,
			wantStdout: "Cancelled the invite to bob@example.com\n",
		},
		{
			name:       "email is lowercased",
			args:       []string{"invite", "cancel", "Bob@Example.com", "--confirm"},
			setup:      setupCancel,
			wantStdout: "Cancelled the invite to bob@example.com\n",
		},
		{
			name:       "revoke alias",
			args:       []string{"invite", "revoke", "bob@example.com", "--confirm"},
			setup:      setupCancel,
			wantStdout: "Cancelled the invite to bob@example.com\n",
		},
		{
			name:       "rm alias",
			args:       []string{"invite", "rm", "bob@example.com", "--confirm"},
			setup:      setupCancel,
			wantStdout: "Cancelled the invite to bob@example.com\n",
		},
	}

	runCmdTests(t, tests)
}
