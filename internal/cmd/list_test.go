package cmd

import (
	"errors"
	"net/http"
	"slices"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestListCmd(t *testing.T) {
	standardDbs := []api.DatabaseWithUsage{
		sampleDatabaseWithUsage(),
		sampleDatabaseWithUsage(func(db *api.DatabaseWithUsage) {
			db.Id = "def4567890"
			db.Name = "otherdb"
			db.Status = api.DatabaseStatusPaused
			db.ComputeMinutes = new(int64(0))
		}),
		sampleDatabaseWithUsage(func(db *api.DatabaseWithUsage) {
			db.Id = "ghi7890123"
			db.Name = "newdb"
			db.ComputeMinutes = nil
		}),
	}

	dedicatedDb := sampleDatabaseWithUsage(func(db *api.DatabaseWithUsage) {
		db.Id = "ded1234567"
		db.Name = "dedicateddb"
		db.Type = api.DatabaseTypeDedicated
		db.Size = new(api.DatabaseSizeN2x)
		db.ComputeMinutes = nil
	})

	mixedDbs := slices.Concat(standardDbs, []api.DatabaseWithUsage{dedicatedDb})

	setupList := func(dbs []api.DatabaseWithUsage) func(*mock.MockClientWithResponsesInterface) {
		return func(m *mock.MockClientWithResponsesInterface) {
			m.EXPECT().ListDatabasesWithResponse(validCtx, "test-project").
				Return(&api.ListDatabasesResponse{
					HTTPResponse: httpResponse(http.StatusOK),
					JSON200:      &dbs,
				}, nil)
		}
	}

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"list"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error",
			args: []string{"list"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListDatabasesWithResponse(validCtx, "test-project").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to list databases: connection refused",
		},
		{
			name: "API error",
			args: []string{"list"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListDatabasesWithResponse(validCtx, "test-project").
					Return(&api.ListDatabasesResponse{
						HTTPResponse: httpResponse(http.StatusInternalServerError),
						JSONDefault:  &api.Error{Message: "internal error"},
					}, nil)
			},
			wantErr: "internal error",
		},
		{
			name: "nil response body",
			args: []string{"list"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListDatabasesWithResponse(validCtx, "test-project").
					Return(&api.ListDatabasesResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name:  "text output",
			args:  []string{"list"},
			setup: setupList(standardDbs),
			wantStdout: `ID          NAME     STATUS   STORAGE  COMPUTE  
abc1234567  mydb     running  1GiB     1.5h     
def4567890  otherdb  paused   1GiB     0h       
ghi7890123  newdb    running  1GiB     -        
`,
		},
		{
			name:  "text output mixed",
			args:  []string{"list"},
			setup: setupList(mixedDbs),
			wantStdout: `ID          NAME     STATUS   STORAGE  COMPUTE  
abc1234567  mydb     running  1GiB     1.5h     
def4567890  otherdb  paused   1GiB     0h       
ghi7890123  newdb    running  1GiB     -        

Dedicated Databases
ID          NAME         SIZE  STATUS   STORAGE  
ded1234567  dedicateddb  2x    running  1GiB     
`,
		},
		{
			name:  "json output",
			args:  []string{"list", "--json"},
			setup: setupList(standardDbs),
			wantStdout: `[
  {
    "id": "abc1234567",
    "name": "mydb",
    "type": "standard",
    "status": "running",
    "storage_mib": 1024,
    "compute_minutes": 90
  },
  {
    "id": "def4567890",
    "name": "otherdb",
    "type": "standard",
    "status": "paused",
    "storage_mib": 1024,
    "compute_minutes": 0
  },
  {
    "id": "ghi7890123",
    "name": "newdb",
    "type": "standard",
    "status": "running",
    "storage_mib": 1024
  }
]
`,
		},
		{
			name:  "json output mixed",
			args:  []string{"list", "--json"},
			setup: setupList(mixedDbs),
			wantStdout: `[
  {
    "id": "abc1234567",
    "name": "mydb",
    "type": "standard",
    "status": "running",
    "storage_mib": 1024,
    "compute_minutes": 90
  },
  {
    "id": "def4567890",
    "name": "otherdb",
    "type": "standard",
    "status": "paused",
    "storage_mib": 1024,
    "compute_minutes": 0
  },
  {
    "id": "ghi7890123",
    "name": "newdb",
    "type": "standard",
    "status": "running",
    "storage_mib": 1024
  },
  {
    "id": "ded1234567",
    "name": "dedicateddb",
    "type": "dedicated",
    "size": "2x",
    "status": "running",
    "storage_mib": 1024
  }
]
`,
		},
		{
			name:  "yaml output",
			args:  []string{"list", "--yaml"},
			setup: setupList(standardDbs),
			wantStdout: `- compute_minutes: 90
  id: abc1234567
  name: mydb
  status: running
  storage_mib: 1024
  type: standard
- compute_minutes: 0
  id: def4567890
  name: otherdb
  status: paused
  storage_mib: 1024
  type: standard
- id: ghi7890123
  name: newdb
  status: running
  storage_mib: 1024
  type: standard
`,
		},
		{
			name:  "yaml output mixed",
			args:  []string{"list", "--yaml"},
			setup: setupList(mixedDbs),
			wantStdout: `- compute_minutes: 90
  id: abc1234567
  name: mydb
  status: running
  storage_mib: 1024
  type: standard
- compute_minutes: 0
  id: def4567890
  name: otherdb
  status: paused
  storage_mib: 1024
  type: standard
- id: ghi7890123
  name: newdb
  status: running
  storage_mib: 1024
  type: standard
- id: ded1234567
  name: dedicateddb
  size: 2x
  status: running
  storage_mib: 1024
  type: dedicated
`,
		},
		{
			name:       "empty list",
			args:       []string{"list"},
			setup:      setupList([]api.DatabaseWithUsage{}),
			wantStdout: "ID  NAME  STATUS  STORAGE  COMPUTE  \n",
		},
		{
			name:       "empty list json",
			args:       []string{"list", "--json"},
			setup:      setupList([]api.DatabaseWithUsage{}),
			wantStdout: "[]\n",
		},
		{
			name:  "ls alias",
			args:  []string{"ls"},
			setup: setupList(standardDbs),
			wantStdout: `ID          NAME     STATUS   STORAGE  COMPUTE  
abc1234567  mydb     running  1GiB     1.5h     
def4567890  otherdb  paused   1GiB     0h       
ghi7890123  newdb    running  1GiB     -        
`,
		},
	}

	runCmdTests(t, tests)
}
