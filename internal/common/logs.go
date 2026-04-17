package common

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/timescale/ghost/internal/api"
)

type FetchLogsArgs struct {
	Client      api.ClientWithResponsesInterface
	ProjectID   string
	DatabaseRef string
	Tail        int
	Until       time.Time
}

// FetchLogs fetches database logs with pagination up to the specified tail
// limit. Returns logs in ascending chronological order (oldest first).
func FetchLogs(ctx context.Context, args FetchLogsArgs) ([]string, error) {
	params := &api.DatabaseLogsParams{
		Page:  new(0),
		Until: &args.Until,
	}

	// Set until to current time if not provided, so pagination is consistent
	// across multiple page requests.
	if params.Until.IsZero() {
		params.Until = new(time.Now())
	}

	var logs []string
	for {
		resp, err := args.Client.DatabaseLogsWithResponse(ctx, args.ProjectID, args.DatabaseRef, params)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch logs: %w", err)
		}

		if resp.StatusCode() != http.StatusOK {
			return nil, ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
		}

		if resp.JSON200 == nil {
			return nil, errors.New("empty response from API")
		}

		pageLogs := resp.JSON200.Logs
		logs = append(logs, pageLogs...)

		// Stop if page is empty or we have enough logs
		if len(pageLogs) == 0 || len(logs) >= args.Tail {
			break
		}

		*params.Page++
	}

	// Trim to tail limit
	if len(logs) > args.Tail {
		logs = logs[:args.Tail]
	}

	// Reverse so oldest logs appear first (API returns newest first)
	slices.Reverse(logs)

	return logs, nil
}
