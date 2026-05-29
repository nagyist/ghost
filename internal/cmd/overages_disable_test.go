package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestOveragesDisableCmd(t *testing.T) {
	setupDisable := func(m *mock.MockClientWithResponsesInterface) {
		m.EXPECT().UpdateOveragesWithResponse(validCtx, "test-project", api.UpdateOverageSettingsRequest{
			Enabled: false,
		}).Return(&api.UpdateOveragesResponse{
			HTTPResponse: httpResponse(http.StatusNoContent),
		}, nil)
	}

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"overages", "disable", "--confirm"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name:    "non-interactive stdin",
			args:    []string{"overages", "disable"},
			setup:   func(m *mock.MockClientWithResponsesInterface) {},
			wantErr: "cannot prompt for confirmation: stdin is not a terminal; use --confirm to skip",
		},
		{
			name:       "confirmation declined",
			args:       []string{"overages", "disable"},
			opts:       []runOption{withStdin("n\n"), withIsTerminal(true)},
			setup:      func(m *mock.MockClientWithResponsesInterface) {},
			wantStderr: "Disable compute overages? Standard databases will pause once usage reaches the included free allowance, or pause immediately if you are already above that. [y/N] ",
			wantStdout: "Disable cancelled.\n",
		},
		{
			name: "network error",
			args: []string{"overages", "disable", "--confirm"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().UpdateOveragesWithResponse(validCtx, "test-project", api.UpdateOverageSettingsRequest{
					Enabled: false,
				}).Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to disable overages: connection refused",
		},
		{
			name: "API error",
			args: []string{"overages", "disable", "--confirm"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().UpdateOveragesWithResponse(validCtx, "test-project", api.UpdateOverageSettingsRequest{
					Enabled: false,
				}).Return(&api.UpdateOveragesResponse{
					HTTPResponse: httpResponse(http.StatusBadRequest),
					JSONDefault:  &api.Error{Message: "delete primary payment method first"},
				}, nil)
			},
			wantErr: "delete primary payment method first",
		},
		{
			name:       "confirmation accepted",
			args:       []string{"overages", "disable"},
			opts:       []runOption{withStdin("y\n"), withIsTerminal(true)},
			setup:      setupDisable,
			wantStderr: "Disable compute overages? Standard databases will pause once usage reaches the included free allowance, or pause immediately if you are already above that. [y/N] ",
			wantStdout: "Overages disabled.\n",
		},
		{
			name:       "confirm flag skips prompt",
			args:       []string{"overages", "disable", "--confirm"},
			setup:      setupDisable,
			wantStdout: "Overages disabled.\n",
		},
	}

	runCmdTests(t, tests)
}
