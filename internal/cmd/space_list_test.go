package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestSpaceListCmd(t *testing.T) {
	successSetup := func(m *mock.MockClientWithResponsesInterface) {
		spaces := []api.Space{
			{ID: "test-space", Name: "Test Space", Role: new(api.MemberRoleOwner)},
			{ID: "other-proj", Name: "Other Space", Role: new(api.MemberRoleDeveloper)},
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
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error",
			args: []string{"space", "list"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListSpacesWithResponse(validCtx).
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to list spaces: connection refused",
		},
		{
			name: "API error",
			args: []string{"space", "list"},
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
			setup:      successSetup,
			wantStdout: "ID          NAME          ROLE       \ntest-space  Test Space *  owner      \nother-proj  Other Space   developer  \n",
		},
		{
			name:  "json output",
			args:  []string{"space", "list", "--json"},
			setup: successSetup,
			wantStdout: `[
  {
    "id": "test-space",
    "name": "Test Space",
    "role": "owner",
    "current": true
  },
  {
    "id": "other-proj",
    "name": "Other Space",
    "role": "developer",
    "current": false
  }
]
`,
		},
		{
			name:  "yaml output",
			args:  []string{"space", "list", "--yaml"},
			setup: successSetup,
			wantStdout: `- current: true
  id: test-space
  name: Test Space
  role: owner
- current: false
  id: other-proj
  name: Other Space
  role: developer
`,
		},
		{
			name:       "ls alias",
			args:       []string{"space", "ls"},
			setup:      successSetup,
			wantStdout: "ID          NAME          ROLE       \ntest-space  Test Space *  owner      \nother-proj  Other Space   developer  \n",
		},
	}

	runCmdTests(t, tests)
}
