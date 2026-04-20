package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestPaymentAddCmd(t *testing.T) {
	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"payment", "add"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error",
			args: []string{"payment", "add"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().CreatePaymentMethodSetupWithResponse(validCtx, "test-project").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to create payment setup: connection refused",
		},
		{
			name: "API error",
			args: []string{"payment", "add"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().CreatePaymentMethodSetupWithResponse(validCtx, "test-project").
					Return(&api.CreatePaymentMethodSetupResponse{
						HTTPResponse: httpResponse(http.StatusInternalServerError),
						JSONDefault:  &api.Error{Message: "internal error"},
					}, nil)
			},
			wantErr: "internal error",
		},
		{
			name: "nil response body",
			args: []string{"payment", "add"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().CreatePaymentMethodSetupWithResponse(validCtx, "test-project").
					Return(&api.CreatePaymentMethodSetupResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name: "browser opens successfully",
			args: []string{"payment", "add"},
			opts: []runOption{withOpenBrowser(func(string) error { return nil })},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().CreatePaymentMethodSetupWithResponse(validCtx, "test-project").
					Return(&api.CreatePaymentMethodSetupResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &api.PaymentSetupResponse{PaymentUrl: "/pay/123"},
					}, nil)
			},
			wantStdout: "Opening browser to add payment method...\nComplete the payment form in your browser.\n",
			wantStderr: "",
		},
		{
			name: "browser fails to open",
			args: []string{"payment", "add"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().CreatePaymentMethodSetupWithResponse(validCtx, "test-project").
					Return(&api.CreatePaymentMethodSetupResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &api.PaymentSetupResponse{PaymentUrl: "/pay/123"},
					}, nil)
			},
			wantStdout: "Opening browser to add payment method...\nComplete the payment form in your browser.\n",
			wantStderr: "Could not open browser automatically.\nPlease open this URL in your browser:\n\n  https://api.ghost.build/v0/pay/123\n\n",
		},
	}

	runCmdTests(t, tests)
}
