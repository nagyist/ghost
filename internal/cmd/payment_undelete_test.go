package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestPaymentUndeleteCmd(t *testing.T) {
	experimental := withEnv("GHOST_EXPERIMENTAL", "true")

	pmPending := api.PaymentMethod{
		Id: "pm_123", Brand: "Visa", Last4: "4242", ExpMonth: 12, ExpYear: 2025, PendingDeletion: true,
	}

	pmNotPending := api.PaymentMethod{
		Id: "pm_123", Brand: "Visa", Last4: "4242", ExpMonth: 12, ExpYear: 2025, PendingDeletion: false,
	}

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"payment", "undelete", "pm_123"},
			opts:    []runOption{experimental, withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error on get",
			args: []string{"payment", "undelete", "pm_123"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetPaymentMethodWithResponse(validCtx, "test-project", "pm_123").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to get payment method: connection refused",
		},
		{
			name: "API error on get",
			args: []string{"payment", "undelete", "pm_123"},
			opts: []runOption{experimental},
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
			args: []string{"payment", "undelete", "pm_123"},
			opts: []runOption{experimental},
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
			name: "not pending deletion",
			args: []string{"payment", "undelete", "pm_123"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetPaymentMethodWithResponse(validCtx, "test-project", "pm_123").
					Return(&api.GetPaymentMethodResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &pmNotPending,
					}, nil)
			},
			wantErr: "payment method does not have a pending deletion",
		},
		{
			name: "network error on cancel",
			args: []string{"payment", "undelete", "pm_123"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetPaymentMethodWithResponse(validCtx, "test-project", "pm_123").
					Return(&api.GetPaymentMethodResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &pmPending,
					}, nil)
				m.EXPECT().CancelPaymentMethodDeletionWithResponse(validCtx, "test-project", "pm_123").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to cancel payment method deletion: connection refused",
		},
		{
			name: "success",
			args: []string{"payment", "undelete", "pm_123"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetPaymentMethodWithResponse(validCtx, "test-project", "pm_123").
					Return(&api.GetPaymentMethodResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &pmPending,
					}, nil)
				m.EXPECT().CancelPaymentMethodDeletionWithResponse(validCtx, "test-project", "pm_123").
					Return(&api.CancelPaymentMethodDeletionResponse{
						HTTPResponse: httpResponse(http.StatusNoContent),
					}, nil)
			},
			wantStdout: "Cancelled pending deletion for Visa ending in 4242.\n",
		},
	}

	runCmdTests(t, tests)
}
