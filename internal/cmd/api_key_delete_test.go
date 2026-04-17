package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestApiKeyDeleteCmd(t *testing.T) {
	setupDelete := func(m *mock.MockClientWithResponsesInterface) {
		m.EXPECT().DeleteApiKeyWithResponse(validCtx, "test-project", "gt_abc").
			Return(&api.DeleteApiKeyResponse{
				HTTPResponse: httpResponse(http.StatusNoContent),
			}, nil)
	}

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"api-key", "delete", "gt_abc"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name:    "non-interactive stdin",
			args:    []string{"api-key", "delete", "gt_abc"},
			setup:   nil,
			wantErr: "cannot prompt for confirmation: stdin is not a terminal; use --confirm to skip",
		},
		{
			name:       "confirmation declined",
			args:       []string{"api-key", "delete", "gt_abc"},
			setup:      nil,
			opts:       []runOption{withStdin("n\n"), withIsTerminal(true)},
			wantStderr: "Delete API key with prefix 'gt_abc'? [y/N] ",
			wantStdout: "Delete cancelled.\n",
		},
		{
			name: "network error",
			args: []string{"api-key", "delete", "gt_abc", "--confirm"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().DeleteApiKeyWithResponse(validCtx, "test-project", "gt_abc").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to delete API key: connection refused",
		},
		{
			name: "API error",
			args: []string{"api-key", "delete", "gt_abc", "--confirm"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().DeleteApiKeyWithResponse(validCtx, "test-project", "gt_abc").
					Return(&api.DeleteApiKeyResponse{
						HTTPResponse: httpResponse(http.StatusNotFound),
						JSONDefault:  &api.Error{Message: "API key not found"},
					}, nil)
			},
			wantErr: "API key not found",
		},
		{
			name: "confirmation accepted",
			args: []string{"api-key", "delete", "gt_abc"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupDelete(m)
			},
			opts:       []runOption{withStdin("y\n"), withIsTerminal(true)},
			wantStderr: "Delete API key with prefix 'gt_abc'? [y/N] ",
			wantStdout: "Deleted API key with prefix 'gt_abc'.\n",
		},
		{
			name: "confirm flag",
			args: []string{"api-key", "delete", "gt_abc", "--confirm"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupDelete(m)
			},
			wantStdout: "Deleted API key with prefix 'gt_abc'.\n",
		},
		{
			name: "rm alias",
			args: []string{"api-key", "rm", "gt_abc", "--confirm"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupDelete(m)
			},
			wantStdout: "Deleted API key with prefix 'gt_abc'.\n",
		},
	}

	runCmdTests(t, tests)
}
