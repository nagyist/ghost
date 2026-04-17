package cmd

import (
	"testing"
)

func TestConfigUnsetCmd(t *testing.T) {
	tests := []cmdTest{
		{
			name:       "unset analytics",
			args:       []string{"config", "unset", "analytics"},
			wantStdout: "Unset analytics\n",
		},
		{
			name:       "unset color",
			args:       []string{"config", "unset", "color"},
			wantStdout: "Unset color\n",
		},
	}

	runCmdTests(t, tests)
}
