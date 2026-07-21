package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

func buildMemberRemoveCmd(app *common.App) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:     "remove <email>",
		Aliases: []string{"rm"},
		Short:   "Remove a member from the current space",
		Long: `Remove a member from the current space.

The space owner cannot be removed. By default, you will be prompted to
confirm the removal, unless you use the --confirm flag.`,
		Example: `  # Remove a member (with confirmation prompt)
  ghost member remove bob@example.com

  # Remove a member without confirmation prompt
  ghost member remove bob@example.com --confirm`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: memberEmailCompletion(app),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			email := args[0]

			client, spaceID, err := app.GetClient()
			if err != nil {
				return err
			}

			member, err := findMemberByEmail(cmd.Context(), client, spaceID, email)
			if err != nil {
				return err
			}

			// Prompt for confirmation unless --confirm is used
			if !confirm {
				if !util.IsTerminal(cmd.InOrStdin()) {
					return errors.New("cannot prompt for confirmation: stdin is not a terminal; use --confirm to skip")
				}
				cmd.PrintErrf("Remove %s (%s) from the space? [y/N] ", member.Name, member.Email)

				confirmation, err := util.ReadLine(cmd.Context(), cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}

				confirmation = strings.ToLower(confirmation)
				if confirmation != "y" && confirmation != "yes" {
					cmd.Println("Remove operation cancelled.")
					return nil
				}
			}

			resp, err := client.RemoveMemberWithResponse(
				cmd.Context(),
				api.SpaceID(spaceID),
				api.MemberUserID(member.UserID),
			)
			if err != nil {
				return fmt.Errorf("failed to remove member: %w", err)
			}
			if resp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}
			if resp.JSON200 == nil {
				return errors.New("empty response from API")
			}

			removed := resp.JSON200
			cmd.Printf("Removed %s (%s) from the space\n", removed.Name, removed.Email)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")

	return cmd
}
