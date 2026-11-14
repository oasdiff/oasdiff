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
	addComposedFlag(&cmd)
	addCommonBreakingFlags(&cmd)
	enumWithOptions(&cmd, newEnumValue(GetBreakingFailOnValues(), ""), "fail-on", "o", "exit with return code 1 on a change at this severity or higher, or with MISSING-VERSION-BUMP on breaking changes that have no major version increase in info.version")
	addOpenFlags(&cmd, "breaking changes")

	return &cmd
}

func runBreakingChanges(flags *Flags, stdout, stderr io.Writer) (bool, *ReturnError) {
	return getChangelog(flags, stdout, stderr, checker.WARN, true)
}
