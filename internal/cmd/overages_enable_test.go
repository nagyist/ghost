package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestOveragesEnableCmd(t *testing.T) {
	setupUpdate := func(m *mock.MockClientWithResponsesInterface, req api.UpdateOverageSettingsRequest) {
		m.EXPECT().UpdateOveragesWithResponse(validCtx, "test-project", req).
			Return(&api.UpdateOveragesResponse{
				HTTPResponse: httpResponse(http.StatusNoContent),
			}, nil)
	}

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"overages", "enable", "--limit", "200"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error on update",
			args: []string{"overages", "enable", "--limit", "200"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().UpdateOveragesWithResponse(validCtx, "test-project", api.UpdateOverageSettingsRequest{
					Enabled:             true,
					ComputeLimitMinutes: new(int64(12000)),
				}).Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to enable overages: connection refused",
		},
		{
			name: "API error on update",
			args: []string{"overages", "enable", "--limit", "200"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().UpdateOveragesWithResponse(validCtx, "test-project", api.UpdateOverageSettingsRequest{
					Enabled:             true,
					ComputeLimitMinutes: new(int64(12000)),
				}).Return(&api.UpdateOveragesResponse{
					HTTPResponse: httpResponse(http.StatusBadRequest),
					JSONDefault:  &api.Error{Message: "payment method required"},
				}, nil)
			},
			wantErr: "payment method required",
		},
		{
			name: "enable with limit",
			args: []string{"overages", "enable", "--limit", "200"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupUpdate(m, api.UpdateOverageSettingsRequest{
					Enabled:             true,
					ComputeLimitMinutes: new(int64(12000)),
				})
			},
			wantStdout: "Overages enabled. You will be charged for compute beyond the included free allowance, up to 200 hours/month. See 'ghost pricing' for current rates.\n",
		},
		{
			name:    "no-limit non-interactive stdin",
			args:    []string{"overages", "enable"},
			setup:   func(m *mock.MockClientWithResponsesInterface) {},
			wantErr: "cannot prompt for confirmation: stdin is not a terminal; pass --limit <hours> or --confirm to skip",
		},
		{
			name:       "no-limit confirmation declined",
			args:       []string{"overages", "enable"},
			opts:       []runOption{withStdin("n\n"), withIsTerminal(true)},
			setup:      func(m *mock.MockClientWithResponsesInterface) {},
			wantStderr: "You are enabling overages with no monthly limit. Your databases will never be auto-paused for hitting a compute limit, and you will be billed for all overage usage with no upper bound. Continue? [y/N] ",
			wantStdout: "Enable cancelled.\n",
		},
		{
			name: "no-limit confirmation accepted",
			args: []string{"overages", "enable"},
			opts: []runOption{withStdin("y\n"), withIsTerminal(true)},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupUpdate(m, api.UpdateOverageSettingsRequest{Enabled: true})
			},
			wantStderr: "You are enabling overages with no monthly limit. Your databases will never be auto-paused for hitting a compute limit, and you will be billed for all overage usage with no upper bound. Continue? [y/N] ",
			wantStdout: "Overages enabled with no monthly limit. You will be charged for ALL compute usage beyond the included free allowance, with no upper bound. See 'ghost pricing' for current rates.\n",
		},
		{
			name: "no-limit with --confirm skips prompt",
			args: []string{"overages", "enable", "--confirm"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupUpdate(m, api.UpdateOverageSettingsRequest{Enabled: true})
			},
			wantStdout: "Overages enabled with no monthly limit. You will be charged for ALL compute usage beyond the included free allowance, with no upper bound. See 'ghost pricing' for current rates.\n",
		},
	}

	runCmdTests(t, tests)
}
