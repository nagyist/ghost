package common

import (
	"errors"

	"github.com/timescale/ghost/internal/api"
)

var (
	// ErrPaused is returned when an operation is attempted against a paused database.
	ErrPaused = errors.New("database is paused")

	// ErrNotReady is returned when an operation is attempted against a database
	// that is still being provisioned (queued or configuring).
	ErrNotReady = errors.New("database is not ready")
)

// CheckReady returns ErrPaused if the database is paused or pausing, or
// ErrNotReady if the database is queued, configuring, or resuming.
func CheckReady(database api.Database) error {
	switch database.Status {
	case api.DatabaseStatusPaused, api.DatabaseStatusPausing:
		return ErrPaused
	case api.DatabaseStatusQueued, api.DatabaseStatusConfiguring, api.DatabaseStatusResuming:
		return ErrNotReady
	}
	return nil
}
