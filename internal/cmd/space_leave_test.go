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

func TestSpaceLeaveCmd(t *testing.T) {
	experimental := withEnv("GHOST_EXPERIMENTAL", "true")

	tokenCreds := config.Credentials{
		Token:   &oauth2.Token{AccessToken: "test-token"},
		SpaceID: "test-space",
	}

	// setupGetSpace mocks the current-space lookup used for the prompt/messages.
	setupGetSpace := func(m *mock.MockClientWithResponsesInterface) {
		m.EXPECT().GetSpaceWithResponse(validCtx, "test-space").
			Return(&api.GetSpaceResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &api.SpaceDetail{Id: "test-space", Name: "Test Space"},
			}, nil)
	}
	// setupLeave mocks a successful leave of the current space.
	setupLeave := func(m *mock.MockClientWithResponsesInterface) {
		m.EXPECT().LeaveSpaceWithResponse(validCtx, api.SpaceId("test-space")).
			Return(&api.LeaveSpaceResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &api.LeaveSpaceResult{SpaceId: "test-space", SpaceName: "Test Space"},
			}, nil)
	}
	// setupListOwned mocks the post-leave space list, returning the user's owned
	// space so the command switches the current space to it.
	setupListOwned := func(m *mock.MockClientWithResponsesInterface) {
		ownerRole := api.MemberRoleOwner
		spaces := []api.Space{{Id: "home-proj", Name: "My Space", Role: &ownerRole}}
		m.EXPECT().ListSpacesWithResponse(validCtx).
			Return(&api.ListSpacesResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &spaces,
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
			args:    []string{"space", "leave"},
			opts:    []runOption{experimental, withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error on get space",
			args: []string{"space", "leave", "--confirm"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetSpaceWithResponse(validCtx, "test-space").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to get space: connection refused",
		},
		{
			name: "API error on get space",
			args: []string{"space", "leave", "--confirm"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetSpaceWithResponse(validCtx, "test-space").
					Return(&api.GetSpaceResponse{
						HTTPResponse: httpResponse(http.StatusForbidden),
						JSONDefault:  &api.Error{Message: "user authentication required"},
					}, nil)
			},
			wantErr: "user authentication required",
		},
		{
			name: "nil response body on get space",
			args: []string{"space", "leave", "--confirm"},
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
			name:    "non-terminal stdin without confirm flag",
			args:    []string{"space", "leave"},
			opts:    []runOption{experimental},
			setup:   setupGetSpace,
			wantErr: "cannot prompt for confirmation: stdin is not a terminal; use --confirm to skip",
		},
		{
			name:       "confirmation declined",
			args:       []string{"space", "leave"},
			opts:       []runOption{experimental, withStdin("n\n"), withIsTerminal(true)},
			setup:      setupGetSpace,
			wantStderr: "Leave space 'Test Space' (test-space)? [y/N] ",
			wantStdout: "Leave operation cancelled.\n",
		},
		{
			name: "network error on leave",
			args: []string{"space", "leave", "--confirm"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGetSpace(m)
				m.EXPECT().LeaveSpaceWithResponse(validCtx, api.SpaceId("test-space")).
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to leave space: connection refused",
		},
		{
			name: "API error on leave (owner cannot leave)",
			args: []string{"space", "leave", "--confirm"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGetSpace(m)
				m.EXPECT().LeaveSpaceWithResponse(validCtx, api.SpaceId("test-space")).
					Return(&api.LeaveSpaceResponse{
						HTTPResponse: httpResponse(http.StatusBadRequest),
						JSONDefault:  &api.Error{Message: "the space owner cannot leave their own space"},
					}, nil)
			},
			wantErr: "the space owner cannot leave their own space",
		},
		{
			name: "nil response body on leave",
			args: []string{"space", "leave", "--confirm"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGetSpace(m)
				m.EXPECT().LeaveSpaceWithResponse(validCtx, api.SpaceId("test-space")).
					Return(&api.LeaveSpaceResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name: "switch best-effort: list spaces fails after leaving",
			args: []string{"space", "leave", "--confirm"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGetSpace(m)
				setupLeave(m)
				m.EXPECT().ListSpacesWithResponse(validCtx).
					Return(nil, errors.New("connection refused"))
			},
			wantStdout: "Left space 'Test Space' (test-space)\n" +
				"Run 'ghost space use <id>' to select a space for subsequent commands.\n",
		},
		{
			name: "confirmation accepted",
			args: []string{"space", "leave"},
			opts: []runOption{experimental, withStdin("y\n"), withIsTerminal(true), withStoredCredentials(tokenCreds)},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGetSpace(m)
				setupLeave(m)
				setupListOwned(m)
			},
			wantStderr: "Leave space 'Test Space' (test-space)? [y/N] ",
			wantStdout: "Left space 'Test Space' (test-space)\n" +
				"Switched to space 'My Space' (home-proj)\n",
			check: checkStoredSpaceID("home-proj"),
		},
		{
			name: "confirm flag",
			args: []string{"space", "leave", "--confirm"},
			opts: []runOption{experimental, withStoredCredentials(tokenCreds)},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGetSpace(m)
				setupLeave(m)
				setupListOwned(m)
			},
			wantStdout: "Left space 'Test Space' (test-space)\n" +
				"Switched to space 'My Space' (home-proj)\n",
			check: checkStoredSpaceID("home-proj"),
		},
	}

	runCmdTests(t, tests)
}
