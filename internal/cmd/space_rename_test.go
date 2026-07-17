package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestSpaceRenameCmd(t *testing.T) {
	experimental := withEnv("GHOST_EXPERIMENTAL", "true")

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"space", "rename", "New Space"},
			opts:    []runOption{experimental, withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error",
			args: []string{"space", "rename", "New Space"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().RenameSpaceWithResponse(validCtx, api.SpaceId("test-space"), api.RenameSpaceRequest{Name: "New Space"}).
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to rename space: connection refused",
		},
		{
			name: "API error",
			args: []string{"space", "rename", "New Space"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().RenameSpaceWithResponse(validCtx, api.SpaceId("test-space"), api.RenameSpaceRequest{Name: "New Space"}).
					Return(&api.RenameSpaceResponse{
						HTTPResponse: httpResponse(http.StatusForbidden),
						JSONDefault:  &api.Error{Message: "user authentication required"},
					}, nil)
			},
			wantErr: "user authentication required",
		},
		{
			name: "nil response body",
			args: []string{"space", "rename", "New Space"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().RenameSpaceWithResponse(validCtx, api.SpaceId("test-space"), api.RenameSpaceRequest{Name: "New Space"}).
					Return(&api.RenameSpaceResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name: "success",
			args: []string{"space", "rename", "New Space"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().RenameSpaceWithResponse(validCtx, api.SpaceId("test-space"), api.RenameSpaceRequest{Name: "New Space"}).
					Return(&api.RenameSpaceResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200: &api.RenameSpaceResult{
							Id:      "test-space",
							OldName: "Test Space",
							NewName: "New Space",
						},
					}, nil)
			},
			wantStdout: "Renamed space 'Test Space' (test-space) to 'New Space'\n",
		},
	}

	runCmdTests(t, tests)
}
