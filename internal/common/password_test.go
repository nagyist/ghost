package common

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/timescale/ghost/internal/api"
)

func TestGetPassword(t *testing.T) {
	tests := []struct {
		name          string
		database      api.Database
		role          string
		pgpassContent string // if non-empty, written to .pgpass in temp HOME
		wantPassword  string
		wantErr       error  // sentinel to check with errors.Is
		wantErrMsg    string // exact match for non-sentinel errors
	}{
		{
			name: "returns API password when present",
			database: api.Database{
				Host:     "host.example.com",
				Port:     5432,
				Password: new("api-secret"),
			},
			role:         "tsdbadmin",
			wantPassword: "api-secret",
		},
		{
			name: "empty role returns error even with API password",
			database: api.Database{
				Password: new("api-secret"),
			},
			role:       "",
			wantErrMsg: "role is required",
		},
		{
			name: "nil API password falls back to pgpass",
			database: api.Database{
				Host:     "host.example.com",
				Port:     5432,
				Password: nil,
			},
			role:          "tsdbadmin",
			pgpassContent: "host.example.com:5432:tsdb:tsdbadmin:pgpass-secret\n",
			wantPassword:  "pgpass-secret",
		},
		{
			name: "empty string API password falls back to pgpass",
			database: api.Database{
				Host:     "host.example.com",
				Port:     5432,
				Password: new(""),
			},
			role:          "tsdbadmin",
			pgpassContent: "host.example.com:5432:tsdb:tsdbadmin:pgpass-secret\n",
			wantPassword:  "pgpass-secret",
		},
		{
			name: "nil API password and no pgpass returns ErrPasswordNotFound",
			database: api.Database{
				Host:     "host.example.com",
				Port:     5432,
				Password: nil,
			},
			role:    "tsdbadmin",
			wantErr: ErrPasswordNotFound,
		},
		{
			name: "nil API password and no matching pgpass entry returns ErrPasswordNotFound",
			database: api.Database{
				Host:     "host.example.com",
				Port:     5432,
				Password: nil,
			},
			role:          "tsdbadmin",
			pgpassContent: "other-host:5432:tsdb:tsdbadmin:some-secret\n",
			wantErr:       ErrPasswordNotFound,
		},
		{
			name: "API password takes priority over pgpass",
			database: api.Database{
				Host:     "host.example.com",
				Port:     5432,
				Password: new("api-secret"),
			},
			role:          "tsdbadmin",
			pgpassContent: "host.example.com:5432:tsdb:tsdbadmin:pgpass-secret\n",
			wantPassword:  "api-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up a temp HOME so we control .pgpass
			tmpHome := t.TempDir()
			t.Setenv("HOME", tmpHome)

			if tt.pgpassContent != "" {
				writePgpass(t, tmpHome, tt.pgpassContent)
			}

			got, err := GetPassword(tt.database, tt.role)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("GetPassword() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if tt.wantErrMsg != "" {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Fatalf("GetPassword() error = %v, want %q", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetPassword() unexpected error: %v", err)
			}
			if got != tt.wantPassword {
				t.Errorf("GetPassword() = %q, want %q", got, tt.wantPassword)
			}
		})
	}
}

// writePgpass creates a .pgpass file in the given home directory.
func writePgpass(t *testing.T, home, content string) {
	t.Helper()
	path := filepath.Join(home, ".pgpass")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write .pgpass: %v", err)
	}
}
