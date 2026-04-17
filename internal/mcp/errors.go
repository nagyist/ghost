package mcp

import (
	"errors"

	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/config"
)

var errReadOnly = errors.New("this operation is not allowed in read-only mode")

func checkReadOnly(cfg *config.Config) error {
	if cfg.ReadOnly {
		return errReadOnly
	}
	return nil
}

// handleDatabaseError translates common database errors into user-friendly
// MCP error messages. Returns the original error if it doesn't match any
// known error types.
func handleDatabaseError(err error) error {
	if errors.Is(err, common.ErrPaused) {
		return errors.New("database is currently paused — resume it with ghost_resume")
	}
	if errors.Is(err, common.ErrNotReady) {
		return errors.New("database is not yet ready — check status with ghost_list and try again")
	}
	if errors.Is(err, common.ErrPasswordNotFound) {
		return errors.New("password not found — reset the password with ghost_password, or add the entry to ~/.pgpass manually")
	}
	return err
}
