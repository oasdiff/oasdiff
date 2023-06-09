package internal

import "github.com/spf13/cobra"

func getBreakingChangesCmd() *cobra.Command {

	return &cobra.Command{
		Use:   `breaking-changes`,
		Short: "Display breaking-changes",
		Args:  cobra.ExactArgs(1),
		Run:   getBreakingChanges,
	}
}

func getBreakingChanges(cmd *cobra.Command, args []string) {

	failEmpty, returnErr := runDiffInternal(cmd, args)
	exit(failEmpty, returnErr, cmd.ErrOrStderr())
}
