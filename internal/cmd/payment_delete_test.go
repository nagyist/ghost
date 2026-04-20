package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestPaymentDeleteCmd(t *testing.T) {
	pm := api.PaymentMethod{
		Id: "pm_123", Brand: "Visa", Last4: "4242", ExpMonth: 12, ExpYear: 2025, Primary: true,
	}

	setupGet := func(m *mock.MockClientWithResponsesInterface) {
		m.EXPECT().GetPaymentMethodWithResponse(validCtx, "test-project", "pm_123").
			Return(&api.GetPaymentMethodResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &pm,
			}, nil)
	}

	setupDelete := func(m *mock.MockClientWithResponsesInterface) {
		m.EXPECT().DeletePaymentMethodWithResponse(validCtx, "test-project", "pm_123").
			Return(&api.DeletePaymentMethodResponse{
				HTTPResponse: httpResponse(http.StatusNoContent),
			}, nil)
	}

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"payment", "delete", "pm_123"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error on get",
			args: []string{"payment", "delete", "pm_123"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetPaymentMethodWithResponse(validCtx, "test-project", "pm_123").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to get payment method: connection refused",
		},
		{
			name: "API error on get",
			args: []string{"payment", "delete", "pm_123"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetPaymentMethodWithResponse(validCtx, "test-project", "pm_123").
					Return(&api.GetPaymentMethodResponse{
						HTTPResponse: httpResponse(http.StatusNotFound),
						JSONDefault:  &api.Error{Message: "payment method not found"},
					}, nil)
			},
			wantErr: "payment method not found",
		},
		{
			name: "nil response body on get",
			args: []string{"payment", "delete", "pm_123"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetPaymentMethodWithResponse(validCtx, "test-project", "pm_123").
					Return(&api.GetPaymentMethodResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name: "non-interactive stdin",
			args: []string{"payment", "delete", "pm_123"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGet(m)
			},
			wantErr: "cannot prompt for confirmation: stdin is not a terminal; use --confirm to skip",
		},
		{
			name: "confirmation declined",
			args: []string{"payment", "delete", "pm_123"},
			opts: []runOption{withStdin("n\n"), withIsTerminal(true)},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGet(m)
			},
			wantStderr: "Delete Visa ending in 4242? [y/N] ",
			wantStdout: "Delete cancelled.\n",
		},
		{
			name: "network error on delete",
			args: []string{"payment", "delete", "pm_123"},
			opts: []runOption{withStdin("y\n"), withIsTerminal(true)},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGet(m)
				m.EXPECT().DeletePaymentMethodWithResponse(validCtx, "test-project", "pm_123").
					Return(nil, errors.New("connection refused"))
			},
			wantErr:    "failed to delete payment method: connection refused",
			wantStderr: "Delete Visa ending in 4242? [y/N] Error: failed to delete payment method: connection refused\n",
		},
		{
			name: "API error on delete",
			args: []string{"payment", "delete", "pm_123", "--confirm"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGet(m)
				m.EXPECT().DeletePaymentMethodWithResponse(validCtx, "test-project", "pm_123").
					Return(&api.DeletePaymentMethodResponse{
						HTTPResponse: httpResponse(http.StatusInternalServerError),
						JSONDefault:  &api.Error{Message: "internal error"},
					}, nil)
			},
			wantErr: "internal error",
		},
		{
			name: "confirmation accepted",
			args: []string{"payment", "delete", "pm_123"},
			opts: []runOption{withStdin("y\n"), withIsTerminal(true)},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGet(m)
				setupDelete(m)
			},
			wantStderr: "Delete Visa ending in 4242? [y/N] ",
			wantStdout: "Deleted Visa ending in 4242.\n",
		},
		{
			name: "confirm flag",
			args: []string{"payment", "delete", "pm_123", "--confirm"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGet(m)
				setupDelete(m)
			},
			wantStdout: "Deleted Visa ending in 4242.\n",
		},
		{
			name: "rm alias",
			args: []string{"payment", "rm", "pm_123", "--confirm"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupGet(m)
				setupDelete(m)
			},
			wantStdout: "Deleted Visa ending in 4242.\n",
		},
	}

	runCmdTests(t, tests)
}
