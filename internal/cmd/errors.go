package cmd

import (
	"errors"
	"fmt"

	"github.com/timescale/ghost/internal/common"
)

// handleDatabaseError translates common database errors into user-friendly
// CLI error messages. Returns the original error if it doesn't match any
// known error types.
func handleDatabaseError(err error, databaseRef string) error {
	if errors.Is(err, common.ErrPaused) {
		return fmt.Errorf("database is currently paused — resume it with 'ghost resume %s'", databaseRef)
	}
	if errors.Is(err, common.ErrNotReady) {
		return fmt.Errorf("database is not yet ready — check status with 'ghost list' and try again")
	}
	if errors.Is(err, common.ErrPasswordNotFound) {
		return fmt.Errorf("password not found\n\nRun 'ghost password %s' to reset the password, or add the entry to ~/.pgpass manually", databaseRef)
	}
	return err
}
