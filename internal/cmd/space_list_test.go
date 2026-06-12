package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestSpaceListCmd(t *testing.T) {
	experimental := withEnv("GHOST_EXPERIMENTAL", "true")

	successSetup := func(m *mock.MockClientWithResponsesInterface) {
		spaces := []api.Space{
			{Id: "test-project", Name: "Test Space"},
			{Id: "other-proj", Name: "Other Space"},
		}
		m.EXPECT().ListSpacesWithResponse(validCtx).
			Return(&api.ListSpacesResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &spaces,
			}, nil)
	}

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"space", "list"},
			opts:    []runOption{experimental, withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error",
			args: []string{"space", "list"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListSpacesWithResponse(validCtx).
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to list spaces: connection refused",
		},
		{
			name: "API error",
			args: []string{"space", "list"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListSpacesWithResponse(validCtx).
					Return(&api.ListSpacesResponse{
						HTTPResponse: httpResponse(http.StatusInternalServerError),
						JSONDefault:  &api.Error{Message: "internal error"},
					}, nil)
			},
			wantErr: "internal error",
		},
		{
			name: "nil response body",
			args: []string{"space", "list"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListSpacesWithResponse(validCtx).
					Return(&api.ListSpacesResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name:       "text output",
			args:       []string{"space", "list"},
			opts:       []runOption{experimental},
			setup:      successSetup,
			wantStdout: "ID            NAME          \ntest-project  Test Space *  \nother-proj    Other Space   \n",
		},
		{
			name:  "json output",
			args:  []string{"space", "list", "--json"},
			opts:  []runOption{experimental},
			setup: successSetup,
			wantStdout: `[
  {
    "id": "test-project",
    "name": "Test Space",
    "current": true
  },
  {
    "id": "other-proj",
    "name": "Other Space",
    "current": false
  }
]
`,
		},
		{
			name:  "yaml output",
			args:  []string{"space", "list", "--yaml"},
			opts:  []runOption{experimental},
			setup: successSetup,
			wantStdout: `- current: true
  id: test-project
  name: Test Space
- current: false
  id: other-proj
  name: Other Space
`,
		},
		{
			name:       "ls alias",
			args:       []string{"space", "ls"},
			opts:       []runOption{experimental},
			setup:      successSetup,
			wantStdout: "ID            NAME          \ntest-project  Test Space *  \nother-proj    Other Space   \n",
		},
	}

	runCmdTests(t, tests)
}
