package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestUsageCmd(t *testing.T) {
	// setupListSpaces mocks the ListSpaces call used to resolve the space
	// name. AnyTimes allows error-path tests to fail fast on another call
	// without requiring this one to happen.
	setupListSpaces := func(m *mock.MockClientWithResponsesInterface, spaces []api.Space) {
		m.EXPECT().ListSpacesWithResponse(validCtx).
			Return(&api.ListSpacesResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &spaces,
			}, nil).AnyTimes()
	}
	testSpaces := []api.Space{{Id: "test-project", Name: "Test Space"}}

	successSetup := func(m *mock.MockClientWithResponsesInterface) {
		setupListSpaces(m, testSpaces)
		m.EXPECT().SpaceUsageWithResponse(validCtx, "test-project").
			Return(&api.SpaceUsageResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200: &api.SpaceUsage{
					ComputeMinutes:      120,
					FreeComputeMinutes:  6000,
					ComputeLimitMinutes: new(int64(600)),
					StorageMib:          512,
					StorageLimitMib:     1048576,
				},
			}, nil)
		databases := []api.DatabaseWithUsage{
			sampleDatabaseWithUsage(),
			sampleDatabaseWithUsage(func(db *api.DatabaseWithUsage) {
				db.Id = "def4567890"
				db.Name = "otherdb"
				db.Status = api.DatabaseStatusPaused
			}),
		}
		m.EXPECT().ListDatabasesWithResponse(validCtx, "test-project").
			Return(&api.ListDatabasesResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &databases,
			}, nil)
	}

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"usage"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error on space usage",
			args: []string{"usage"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupListSpaces(m, testSpaces)
				m.EXPECT().SpaceUsageWithResponse(validCtx, "test-project").
					Return(nil, errors.New("connection refused"))
				databases := []api.DatabaseWithUsage{}
				m.EXPECT().ListDatabasesWithResponse(validCtx, "test-project").
					Return(&api.ListDatabasesResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &databases,
					}, nil).AnyTimes()
			},
			wantErr: "failed to get space usage: connection refused",
		},
		{
			name: "API error on space usage",
			args: []string{"usage"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupListSpaces(m, testSpaces)
				m.EXPECT().SpaceUsageWithResponse(validCtx, "test-project").
					Return(&api.SpaceUsageResponse{
						HTTPResponse: httpResponse(http.StatusInternalServerError),
						JSONDefault:  &api.Error{Message: "internal error"},
					}, nil)
				databases := []api.DatabaseWithUsage{}
				m.EXPECT().ListDatabasesWithResponse(validCtx, "test-project").
					Return(&api.ListDatabasesResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &databases,
					}, nil).AnyTimes()
			},
			wantErr: "internal error",
		},
		{
			name: "nil space usage response body",
			args: []string{"usage"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupListSpaces(m, testSpaces)
				m.EXPECT().SpaceUsageWithResponse(validCtx, "test-project").
					Return(&api.SpaceUsageResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
				databases := []api.DatabaseWithUsage{}
				m.EXPECT().ListDatabasesWithResponse(validCtx, "test-project").
					Return(&api.ListDatabasesResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &databases,
					}, nil).AnyTimes()
			},
			wantErr: "empty response from API",
		},
		{
			name: "nil list databases response body",
			args: []string{"usage"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupListSpaces(m, testSpaces)
				m.EXPECT().SpaceUsageWithResponse(validCtx, "test-project").
					Return(&api.SpaceUsageResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200: &api.SpaceUsage{
							ComputeMinutes:      120,
							ComputeLimitMinutes: new(int64(600)),
							StorageMib:          512,
							StorageLimitMib:     1048576,
						},
					}, nil).AnyTimes()
				m.EXPECT().ListDatabasesWithResponse(validCtx, "test-project").
					Return(&api.ListDatabasesResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name: "network error on list spaces",
			args: []string{"usage"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListSpacesWithResponse(validCtx).
					Return(nil, errors.New("connection refused"))
				m.EXPECT().SpaceUsageWithResponse(validCtx, "test-project").
					Return(&api.SpaceUsageResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &api.SpaceUsage{},
					}, nil).AnyTimes()
				databases := []api.DatabaseWithUsage{}
				m.EXPECT().ListDatabasesWithResponse(validCtx, "test-project").
					Return(&api.ListDatabasesResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &databases,
					}, nil).AnyTimes()
			},
			wantErr: "failed to list spaces: connection refused",
		},
		{
			name: "nil list spaces response body",
			args: []string{"usage"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListSpacesWithResponse(validCtx).
					Return(&api.ListSpacesResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
				m.EXPECT().SpaceUsageWithResponse(validCtx, "test-project").
					Return(&api.SpaceUsageResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &api.SpaceUsage{},
					}, nil).AnyTimes()
				databases := []api.DatabaseWithUsage{}
				m.EXPECT().ListDatabasesWithResponse(validCtx, "test-project").
					Return(&api.ListDatabasesResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &databases,
					}, nil).AnyTimes()
			},
			wantErr: "empty response from API",
		},
		{
			name:  "text output",
			args:  []string{"usage"},
			setup: successSetup,
			wantStdout: `Space: Test Space (test-project)
Compute: 2/10 hours (20%)
Storage: 512MiB/1TiB (0%)
Databases: 2 (1 running, 1 paused)
`,
		},
		{
			name:  "json output",
			args:  []string{"usage", "--json"},
			setup: successSetup,
			wantStdout: `{
  "compute_minutes": 120,
  "free_compute_minutes": 6000,
  "compute_limit_minutes": 600,
  "overages_enabled": false,
  "storage_mib": 512,
  "storage_limit_mib": 1048576,
  "databases": {
    "running": 1,
    "paused": 1
  },
  "space_id": "test-project",
  "space_name": "Test Space"
}
`,
		},
		{
			name:  "yaml output",
			args:  []string{"usage", "--yaml"},
			setup: successSetup,
			wantStdout: `compute_limit_minutes: 600
compute_minutes: 120
databases:
  paused: 1
  running: 1
free_compute_minutes: 6000
overages_enabled: false
space_id: test-project
space_name: Test Space
storage_limit_mib: 1.048576e+06
storage_mib: 512
`,
		},
		{
			name:  "status alias",
			args:  []string{"status"},
			setup: successSetup,
			wantStdout: `Space: Test Space (test-project)
Compute: 2/10 hours (20%)
Storage: 512MiB/1TiB (0%)
Databases: 2 (1 running, 1 paused)
`,
		},
		{
			name: "text output with cost",
			args: []string{"usage"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupListSpaces(m, testSpaces)
				m.EXPECT().SpaceUsageWithResponse(validCtx, "test-project").
					Return(&api.SpaceUsageResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200: &api.SpaceUsage{
							ComputeMinutes:      120,
							ComputeLimitMinutes: new(int64(600)),
							StorageMib:          512,
							StorageLimitMib:     1048576,
							CostToDate:          new(12.34),
							EstimatedTotalCost:  new(27.50),
						},
					}, nil)
				databases := []api.DatabaseWithUsage{sampleDatabaseWithUsage()}
				m.EXPECT().ListDatabasesWithResponse(validCtx, "test-project").
					Return(&api.ListDatabasesResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &databases,
					}, nil)
			},
			wantStdout: `Space: Test Space (test-project)
Compute: 2/10 hours (20%)
Storage: 512MiB/1TiB (0%)
Databases: 1 (1 running)
Cost: $12.34 so far this cycle ($27.50 estimated total)
`,
		},
		{
			name: "text output with overages enabled",
			args: []string{"usage"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupListSpaces(m, testSpaces)
				m.EXPECT().SpaceUsageWithResponse(validCtx, "test-project").
					Return(&api.SpaceUsageResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200: &api.SpaceUsage{
							ComputeMinutes:      120,
							FreeComputeMinutes:  6000,
							ComputeLimitMinutes: new(int64(12000)),
							OveragesEnabled:     true,
							StorageMib:          512,
							StorageLimitMib:     1048576,
						},
					}, nil)
				databases := []api.DatabaseWithUsage{sampleDatabaseWithUsage()}
				m.EXPECT().ListDatabasesWithResponse(validCtx, "test-project").
					Return(&api.ListDatabasesResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &databases,
					}, nil)
			},
			wantStdout: `Space: Test Space (test-project)
Compute: 2/200 hours (1%)
Storage: 512MiB/1TiB (0%)
Databases: 1 (1 running)
Overages: enabled (billed for compute above 100 free hours)
`,
		},
		{
			name: "text output with zero cost is omitted",
			args: []string{"usage"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupListSpaces(m, testSpaces)
				m.EXPECT().SpaceUsageWithResponse(validCtx, "test-project").
					Return(&api.SpaceUsageResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200: &api.SpaceUsage{
							ComputeMinutes:      120,
							ComputeLimitMinutes: new(int64(600)),
							StorageMib:          512,
							StorageLimitMib:     1048576,
							CostToDate:          new(0.0),
							EstimatedTotalCost:  new(0.0),
						},
					}, nil)
				databases := []api.DatabaseWithUsage{sampleDatabaseWithUsage()}
				m.EXPECT().ListDatabasesWithResponse(validCtx, "test-project").
					Return(&api.ListDatabasesResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &databases,
					}, nil)
			},
			wantStdout: `Space: Test Space (test-project)
Compute: 2/10 hours (20%)
Storage: 512MiB/1TiB (0%)
Databases: 1 (1 running)
`,
		},
		{
			name: "space not in list falls back to ID only",
			args: []string{"usage"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupListSpaces(m, []api.Space{{Id: "other-project", Name: "Other Space"}})
				m.EXPECT().SpaceUsageWithResponse(validCtx, "test-project").
					Return(&api.SpaceUsageResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200: &api.SpaceUsage{
							ComputeMinutes:      120,
							ComputeLimitMinutes: new(int64(600)),
							StorageMib:          512,
							StorageLimitMib:     1048576,
						},
					}, nil)
				databases := []api.DatabaseWithUsage{sampleDatabaseWithUsage()}
				m.EXPECT().ListDatabasesWithResponse(validCtx, "test-project").
					Return(&api.ListDatabasesResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &databases,
					}, nil)
			},
			wantStdout: `Space: test-project
Compute: 2/10 hours (20%)
Storage: 512MiB/1TiB (0%)
Databases: 1 (1 running)
`,
		},
	}

	runCmdTests(t, tests)
}
