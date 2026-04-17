package cmd

import (
	"testing"
)

func TestConfigResetCmd(t *testing.T) {
	tests := []cmdTest{
		{
			name:       "reset config",
			args:       []string{"config", "reset"},
			wantStdout: "Configuration reset to defaults\n",
		},
	}

	runCmdTests(t, tests)
}
