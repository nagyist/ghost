package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestPaymentPrimaryCmd(t *testing.T) {
	pm := api.PaymentMethod{
		Id: "pm_123", Brand: "Visa", Last4: "4242", ExpMonth: 12, ExpYear: 2025,
	}

	pmAlreadyPrimary := api.PaymentMethod{
		Id: "pm_123", Brand: "Visa", Last4: "4242", ExpMonth: 12, ExpYear: 2025, Primary: true,
	}

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"payment", "primary", "pm_123"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error on get",
			args: []string{"payment", "primary", "pm_123"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetPaymentMethodWithResponse(validCtx, "test-project", "pm_123").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to get payment method: connection refused",
		},
		{
			name: "API error on get",
			args: []string{"payment", "primary", "pm_123"},
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
			args: []string{"payment", "primary", "pm_123"},
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
			name: "already primary",
			args: []string{"payment", "primary", "pm_123"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetPaymentMethodWithResponse(validCtx, "test-project", "pm_123").
					Return(&api.GetPaymentMethodResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &pmAlreadyPrimary,
					}, nil)
			},
			wantErr: "payment method is already primary",
		},
		{
			name: "network error on set primary",
			args: []string{"payment", "primary", "pm_123"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetPaymentMethodWithResponse(validCtx, "test-project", "pm_123").
					Return(&api.GetPaymentMethodResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &pm,
					}, nil)
				m.EXPECT().SetPaymentMethodPrimaryWithResponse(validCtx, "test-project", "pm_123").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to set primary payment method: connection refused",
		},
		{
			name: "success",
			args: []string{"payment", "primary", "pm_123"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().GetPaymentMethodWithResponse(validCtx, "test-project", "pm_123").
					Return(&api.GetPaymentMethodResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &pm,
					}, nil)
				m.EXPECT().SetPaymentMethodPrimaryWithResponse(validCtx, "test-project", "pm_123").
					Return(&api.SetPaymentMethodPrimaryResponse{
						HTTPResponse: httpResponse(http.StatusNoContent),
					}, nil)
			},
			wantStdout: "Visa ending in 4242 is now your primary payment method.\n",
		},
	}

	runCmdTests(t, tests)
}
