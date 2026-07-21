package cmd

import (
	"errors"
	"net/http"
	"testing"

	"golang.org/x/oauth2"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
	"github.com/timescale/ghost/internal/config"
)

func TestInviteAcceptCmd(t *testing.T) {
	experimental := withEnv("GHOST_EXPERIMENTAL", "true")

	tokenCreds := config.Credentials{
		Token:   &oauth2.Token{AccessToken: "test-token"},
		SpaceID: "test-space",
	}

	setupAccept := func(m *mock.MockClientWithResponsesInterface) {
		result := api.InviteActionResult{SpaceID: "space-abc", SpaceName: "New Space"}
		m.EXPECT().AcceptInviteWithResponse(validCtx, api.SpaceID("space-abc")).
			Return(&api.AcceptInviteResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &result,
			}, nil)
	}

	checkStoredSpaceID := func(want string) func(t *testing.T, result cmdResult) {
		return func(t *testing.T, result cmdResult) {
			creds := readStoredCredentials(t, result.configDir)
			if creds.SpaceID != want {
				t.Errorf("stored space ID = %q, want %q", creds.SpaceID, want)
			}
		}
	}

	joinedAndHint := "Joined space 'New Space' (space-abc)\n" +
		"Run 'ghost space use space-abc' to switch to it.\n"
	joinedAndSwitched := "Joined space 'New Space' (space-abc)\n" +
		"Switched to space 'New Space' (space-abc)\n"

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"invite", "accept", "space-abc"},
			opts:    []runOption{experimental, withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error",
			args: []string{"invite", "accept", "space-abc"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().AcceptInviteWithResponse(validCtx, api.SpaceID("space-abc")).
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to accept invitation: connection refused",
		},
		{
			name: "API error",
			args: []string{"invite", "accept", "space-abc"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().AcceptInviteWithResponse(validCtx, api.SpaceID("space-abc")).
					Return(&api.AcceptInviteResponse{
						HTTPResponse: httpResponse(http.StatusNotFound),
						JSONDefault:  &api.Error{Message: "no pending invitation with that ID was found"},
					}, nil)
			},
			wantErr: "no pending invitation with that ID was found",
		},
		{
			name: "nil response body",
			args: []string{"invite", "accept", "space-abc"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().AcceptInviteWithResponse(validCtx, api.SpaceID("space-abc")).
					Return(&api.AcceptInviteResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name:       "non-terminal without switch flag prints hint",
			args:       []string{"invite", "accept", "space-abc"},
			opts:       []runOption{experimental},
			setup:      setupAccept,
			wantStdout: joinedAndHint,
		},
		{
			name:       "switch=false prints hint",
			args:       []string{"invite", "accept", "space-abc", "--switch=false"},
			opts:       []runOption{experimental},
			setup:      setupAccept,
			wantStdout: joinedAndHint,
		},
		{
			name:       "switch flag switches space",
			args:       []string{"invite", "accept", "space-abc", "--switch"},
			opts:       []runOption{experimental, withStoredCredentials(tokenCreds)},
			setup:      setupAccept,
			wantStdout: joinedAndSwitched,
			check:      checkStoredSpaceID("space-abc"),
		},
		{
			name:       "switch flag with API key env",
			args:       []string{"invite", "accept", "space-abc", "--switch"},
			opts:       []runOption{experimental, withEnv("GHOST_API_KEY", "gt_abc123")},
			setup:      setupAccept,
			wantStdout: "Joined space 'New Space' (space-abc)\n",
			wantErr:    "cannot switch spaces when authenticated with an API key; unset GHOST_API_KEY and run 'ghost login'",
		},
		{
			name:       "prompt accepted switches space",
			args:       []string{"invite", "accept", "space-abc"},
			opts:       []runOption{experimental, withStoredCredentials(tokenCreds), withStdin("y\n"), withIsTerminal(true)},
			setup:      setupAccept,
			wantStderr: "Switch to this space now? [Y/n] ",
			wantStdout: joinedAndSwitched,
			check:      checkStoredSpaceID("space-abc"),
		},
		{
			name:       "prompt empty defaults to yes",
			args:       []string{"invite", "accept", "space-abc"},
			opts:       []runOption{experimental, withStoredCredentials(tokenCreds), withStdin("\n"), withIsTerminal(true)},
			setup:      setupAccept,
			wantStderr: "Switch to this space now? [Y/n] ",
			wantStdout: joinedAndSwitched,
			check:      checkStoredSpaceID("space-abc"),
		},
		{
			name:       "prompt declined prints hint",
			args:       []string{"invite", "accept", "space-abc"},
			opts:       []runOption{experimental, withStdin("n\n"), withIsTerminal(true)},
			setup:      setupAccept,
			wantStderr: "Switch to this space now? [Y/n] ",
			wantStdout: joinedAndHint,
		},
	}

	runCmdTests(t, tests)
}
