package cmd

import (
	"log"
	"log/slog"

	"github.com/spf13/cobra"
)

// newLogger configures the default log package to write to the command's stderr
// and returns the default slog logger.
func newLogger(cmd *cobra.Command) *slog.Logger {
	log.SetOutput(cmd.ErrOrStderr())
	return slog.Default()
}
