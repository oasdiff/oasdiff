package internal

import (
	"io"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/spf13/cobra"
)

func getBreakingChangesCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:   "breaking base revision [flags]",
		Short: "Display breaking changes",
		Long:  "Display breaking changes between base and revision specs." + specHelp,
		Args:  getParseArgs(),
		RunE:  getRun(runBreakingChanges),
	}

	addCommonDiffFlags(&cmd)
	addCommonBreakingFlags(&cmd)
	enumWithOptions(&cmd, newEnumValue(GetBreakingLevels(), ""), "fail-on", "o", "exit with return code 1 when output includes errors with this level or higher")
	cmd.PersistentFlags().Bool("open", false, "after printing the breaking changes, encrypt the comparison and upload it to oasdiff.com, then open the side-by-side review in a browser")
	addReviewFlags(&cmd)

	return &cmd
}

func runBreakingChanges(flags *Flags, stdout io.Writer) (bool, *ReturnError) {
	return getChangelog(flags, stdout, checker.WARN, true)
}
