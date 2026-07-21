package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

// InviteList is the combined view of invites you've sent and invitations
// you've received, used for `ghost invite list` output.
type InviteList struct {
	Sent     []api.Invite         `json:"sent"`
	Received []api.ReceivedInvite `json:"received"`
}

func buildInviteListCmd(app *common.App) *cobra.Command {
	var jsonOutput bool
	var yamlOutput bool

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List sent and received invites",
		Long: `List both the invites you've sent for the current space and the
invitations you've received across all spaces.`,
		Example: `  # List sent and received invites
  ghost invite list

  # Output as JSON
  ghost invite list --json`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, spaceID, err := app.GetClient()
			if err != nil {
				return err
			}

			// Fetch sent and received invites in parallel. errgroup cancels the
			// derived context and returns the first error if either call fails;
			// each goroutine writes to a disjoint variable, so the reads after
			// Wait are race-free.
			var sent []api.Invite
			var received []api.ReceivedInvite

			g, gctx := errgroup.WithContext(cmd.Context())
			g.Go(func() error {
				resp, err := client.ListInvitesWithResponse(gctx, api.SpaceID(spaceID))
				if err != nil {
					return fmt.Errorf("failed to list sent invites: %w", err)
				}
				if resp.StatusCode() != http.StatusOK {
					return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
				}
				if resp.JSON200 == nil {
					return errors.New("empty response from API")
				}
				sent = *resp.JSON200
				return nil
			})
			g.Go(func() error {
				resp, err := client.ListReceivedInvitesWithResponse(gctx)
				if err != nil {
					return fmt.Errorf("failed to list received invites: %w", err)
				}
				if resp.StatusCode() != http.StatusOK {
					return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
				}
				if resp.JSON200 == nil {
					return errors.New("empty response from API")
				}
				received = *resp.JSON200
				return nil
			})
			if err := g.Wait(); err != nil {
				return err
			}

			list := InviteList{
				Sent:     sent,
				Received: received,
			}

			switch {
			case jsonOutput:
				return util.SerializeToJSON(cmd.OutOrStdout(), list)
			case yamlOutput:
				return util.SerializeToYAML(cmd.OutOrStdout(), list)
			default:
				// Only resolve the space name when there are sent invites to
				// label with it.
				var spaceName string
				if len(list.Sent) > 0 {
					spaceName = currentSpaceName(cmd.Context(), client, spaceID)
				}
				return outputInviteList(cmd.OutOrStdout(), list, spaceID, spaceName)
			}
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&yamlOutput, "yaml", false, "Output in YAML format")
	cmd.MarkFlagsMutuallyExclusive("json", "yaml")

	return cmd
}

func outputInviteList(w io.Writer, list InviteList, spaceID, spaceName string) error {
	if len(list.Sent) == 0 && len(list.Received) == 0 {
		fmt.Fprintln(w, "No pending invites.")
		return nil
	}

	bold := lipgloss.NewStyle().Bold(true)

	if len(list.Sent) > 0 {
		// The sent invites are all for the current space, so label the section
		// with it ("name (id)", matching `ghost usage`); fall back to just the
		// ID when the name couldn't be resolved.
		space := spaceID
		if spaceName != "" {
			space = fmt.Sprintf("%s (%s)", spaceName, spaceID)
		}
		fmt.Fprintln(w, bold.Render("Sent for "+space))
		if err := outputInvites(w, list.Sent); err != nil {
			return err
		}
	}

	if len(list.Received) > 0 {
		// Separate the two tables with a blank line only when both are shown.
		if len(list.Sent) > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, bold.Render("Received"))
		if err := outputReceivedInvites(w, list.Received); err != nil {
			return err
		}
	}

	return nil
}
