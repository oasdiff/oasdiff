package internal

import (
	"fmt"
	"io"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/spf13/cobra"
)

// getChecksCoverageCmd renders the coverage map: every possible edit of an
// OpenAPI document, the changelog checks that cover it, and the reasoned
// account of every edit without a check. It lives under `checks changelog`
// because only the changelog rules carry location claims; run it on demand
// to audit coverage and derive tasks.
func getChecksCoverageCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:               "coverage",
		Short:             "Display the coverage map of the changelog checks over every possible edit",
		Long:              `Display every possible edit of an OpenAPI document with the changelog checks that cover it, and the reason for every edit without a check, as markdown.`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE:              getRun(runChecksCoverage),
	}

	return &cmd
}

func runChecksCoverage(flags *Flags, stdout io.Writer) (bool, *ReturnError) {
	_, _ = fmt.Fprint(stdout, checker.CoverageDoc())
	return false, nil
}
