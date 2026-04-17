package cmd

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"

	"github.com/timescale/ghost/internal/config"
)

func TestVersionCmd(t *testing.T) {
	goVersion := runtime.Version()
	platform := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

	// Build expected text output by calling the same renderer the command uses
	vo := VersionOutput{
		Version:   config.Version,
		BuildTime: config.BuildTime,
		GitCommit: config.GitCommit,
		GoVersion: goVersion,
		Platform:  platform,
	}
	var textBuf bytes.Buffer
	if err := outputVersion(&textBuf, vo); err != nil {
		t.Fatalf("outputVersion: %v", err)
	}
	wantText := textBuf.String()

	tests := []cmdTest{
		{
			name:       "bare output",
			args:       []string{"version", "--bare"},
			wantStdout: config.Version + "\n",
		},
		{
			name: "json output",
			args: []string{"version", "--json"},
			wantStdout: fmt.Sprintf(`{
  "version": "%s",
  "build_time": "%s",
  "git_commit": "%s",
  "go_version": "%s",
  "platform": "%s"
}
`, config.Version, config.BuildTime, config.GitCommit, goVersion, platform),
		},
		{
			name: "yaml output",
			args: []string{"version", "--yaml"},
			wantStdout: fmt.Sprintf(`build_time: %s
git_commit: %s
go_version: %s
platform: %s
version: %s
`, config.BuildTime, config.GitCommit, goVersion, platform, config.Version),
		},
		{
			name:       "text output",
			args:       []string{"version"},
			wantStdout: wantText,
		},
	}

	runCmdTests(t, tests)
}
