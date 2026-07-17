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

func TestSpaceUseCmd(t *testing.T) {
	experimental := withEnv("GHOST_EXPERIMENTAL", "true")

	tokenCreds := config.Credentials{
		Token:   &oauth2.Token{AccessToken: "test-token"},
		SpaceID: "test-space",
	}

	// setupGetSpace mocks a successful lookup of the requested space.
	setupGetSpace := func(m *mock.MockClientWithResponsesInterface, id, name string) {
		m.EXPECT().GetSpaceWithResponse(validCtx, id).
			Return(&api.GetSpaceResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &api.SpaceDetail{Id: id, Name: name},
			}, nil)
	}
	// setupGetSpaceNotFound mocks a 404 for the requested space.
	setupGetSpaceNotFound := func(m *mock.MockClientWithResponsesInterface, id string) {
		m.EXPECT().GetSpaceWithResponse(validCtx, id).
			Return(&api.GetSpaceResponse{
				HTTPResponse: httpResponse(http.StatusNotFound),
				JSONDefault:  &api.Error{Message: "space not found"},
			}, nil)
	}

	// checkStoredSpaceID verifies the space ID written to the stored credentials.
	checkStoredSpaceID := func(want string) func(t *testing.T, result cmdResult) {
		return func(t *testing.T, result cmdResult) {
			creds := readStoredCredentials(t, result.configDir)
			if creds.SpaceID != want {
				t.Errorf("stored space ID = %q, want %q", creds.SpaceID, want)
			}
		}
	}

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"space", "use", "other-proj"},
			opts:    []runOption{experimental, withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name:    "API key env var",
			args:    []string{"space", "use", "other-proj"},
			opts:    []runOption{experimental, withEnv("GHOST_API_KEY", "gt_abc123")},
			wantErr: "cannot switch spaces when authenticated with an API key; unset GHOST_API_KEY and run 'ghost login'",
		},
		{
			name:    "no stored credentials",
			args:    []string{"space", "use", "other-proj"},
			opts:    []runOption{experimental, withEnv("GHOST_KEYRING", "false")},
			wantErr: "failed to read credentials: not logged in",
		},
		{
			name: "legacy API key credentials",
			args: []string{"space", "use", "other-proj"},
			opts: []runOption{experimental, withStoredCredentials(config.Credentials{
				APIKey:  "legacy-key",
				SpaceID: "test-space",
			})},
			wantErr: "cannot switch spaces when authenticated with an API key; run 'ghost login'",
		},
		{
			name: "network error",
			args: []string{"space", "use", "other-proj"},
			opts: []runOption{experimental, withStoredCredentials(tokenCreds)},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetSpaceWithResponse(validCtx, "other-proj").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to get space: connection refused",
		},
		{
			name: "API error",
			args: []string{"space", "use", "other-proj"},
			opts: []runOption{experimental, withStoredCredentials(tokenCreds)},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetSpaceWithResponse(validCtx, "other-proj").
					Return(&api.GetSpaceResponse{
						HTTPResponse: httpResponse(http.StatusInternalServerError),
						JSONDefault:  &api.Error{Message: "internal error"},
					}, nil)
			},
			wantErr: "internal error",
		},
		{
			name: "nil response body",
			args: []string{"space", "use", "other-proj"},
			opts: []runOption{experimental, withStoredCredentials(tokenCreds)},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetSpaceWithResponse(validCtx, "other-proj").
					Return(&api.GetSpaceResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name: "space not found",
			args: []string{"space", "use", "nonexistent"},
			opts: []runOption{experimental, withStoredCredentials(tokenCreds)},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGetSpaceNotFound(m, "nonexistent")
			},
			wantErr: "space 'nonexistent' not found; run 'ghost space list' to see your spaces",
		},
		{
			name: "space name is not accepted",
			args: []string{"space", "use", "Other Space"},
			opts: []runOption{experimental, withStoredCredentials(tokenCreds)},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGetSpaceNotFound(m, "Other Space")
			},
			wantErr: "space 'Other Space' not found; run 'ghost space list' to see your spaces",
		},
		{
			name: "switch by ID",
			args: []string{"space", "use", "other-proj"},
			opts: []runOption{experimental, withStoredCredentials(tokenCreds)},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGetSpace(m, "other-proj", "Other Space")
			},
			wantStdout: "Switched to space 'Other Space' (other-proj)\n",
			check:      checkStoredSpaceID("other-proj"),
		},
		{
			name: "switch alias",
			args: []string{"space", "switch", "other-proj"},
			opts: []runOption{experimental, withStoredCredentials(tokenCreds)},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGetSpace(m, "other-proj", "Other Space")
			},
			wantStdout: "Switched to space 'Other Space' (other-proj)\n",
			check:      checkStoredSpaceID("other-proj"),
		},
	}

	runCmdTests(t, tests)
}
