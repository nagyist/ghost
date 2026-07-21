package cmd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
)

func buildSpaceRenameCmd(app *common.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename <new-name>",
		Short: "Rename the current space",
		Long:  `Rename the current space.`,
		Example: `  # Rename the current space
  ghost space rename "My Team's space"`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			newName := args[0]

			client, spaceID, err := app.GetClient()
			if err != nil {
				return err
			}

			resp, err := client.RenameSpaceWithResponse(
				cmd.Context(),
				api.SpaceID(spaceID),
				api.RenameSpaceRequest{Name: newName},
			)
			if err != nil {
				return fmt.Errorf("failed to rename space: %w", err)
			}
			if resp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}
			if resp.JSON200 == nil {
				return errors.New("empty response from API")
			}

			result := resp.JSON200
			cmd.Printf("Renamed space '%s' (%s) to '%s'\n", result.OldName, result.ID, result.NewName)
			return nil
		},
	}

	return cmd
}
