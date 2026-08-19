package internal

import (
	"fmt"
	"io"
	"slices"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/spf13/cobra"
)

const checksCoverageCmd = "checks changelog coverage"

// getChecksCoverageCmd lists every possible edit of an OpenAPI document with
// the audit's disposition of it: the changelog checks that cover it, the
// waiver or non-contract reason when none do, and a suggested id for the
// missing check otherwise. It lives under `checks changelog` because only
// the changelog rules carry location claims; run it on demand to audit
// coverage and derive tasks (--tags uncovered lists the holes).
func getChecksCoverageCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:               "coverage",
		Short:             "Display the coverage of the changelog checks over every possible edit",
		Long:              `Display every possible edit of an OpenAPI document with the changelog checks that cover it, the reason when none are expected, and a suggested id for a missing check otherwise.`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE:              getRun(runChecksCoverage),
	}

	addChecksFormatFlags(&cmd)
	enumWithOptions(&cmd, newEnumSliceValue(getCoverageTags(), nil), "tags", "t", "include only edits with all specified tags")
	cmd.PersistentFlags().Bool("patterns", false, "list the waiver and non-contract patterns instead of the edits")

	return &cmd
}

// getCoverageTags returns the tags the coverage listing filters by: the
// audit statuses, the edit's polarity, and the edit's action.
func getCoverageTags() []string {
	return []string{
		// status
		"covered", "uncovered", "waived", "non-contract",
		// polarity
		"request", "response", "document", "shared",
		// action
		"add", "remove", "change", "increase", "decrease", "set", "unset",
	}
}

func matchCoverageTags(tags []string, edit checker.CoverageEdit) bool {
	for _, tag := range tags {
		if tag != string(edit.Status) && tag != edit.Polarity && tag != edit.Action {
			return false
		}
	}
	return true
}

func runChecksCoverage(flags *Flags, stdout io.Writer) (bool, *ReturnError) {

	format := flags.getFormat()

	formatter, err := formatters.Lookup(format, formatters.DefaultFormatterOpts())
	if err != nil {
		return false, getErrUnsupportedFormat(format, checksCoverageCmd)
	}

	var bytes []byte
	if flags.getViper().GetBool("patterns") {
		bytes, err = formatter.RenderCoveragePatterns(checker.CoveragePatterns(), formatters.NewRenderOpts())
	} else {
		edits := slices.DeleteFunc(checker.Coverage(), func(edit checker.CoverageEdit) bool {
			return !matchCoverageTags(flags.getTags(), edit)
		})
		bytes, err = formatter.RenderCoverage(edits, formatters.NewRenderOpts())
	}
	if err != nil {
		return false, getErrFailedPrint(checksCoverageCmd+" "+format, err)
	}

	_, _ = fmt.Fprintf(stdout, "%s\n", bytes)

	return false, nil
}
