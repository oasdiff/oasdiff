package internal

import (
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/checker/coverage"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/spf13/cobra"
)

const checksCoverageCmd = "checks changelog coverage"

// getChecksCoverageCmd lists what the audit decides about every possible
// edit of an OpenAPI document, the inverse of `checks changelog`, which
// lists the rules themselves.
func getChecksCoverageCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:               "coverage",
		Short:             "Display the coverage of the changelog checks over every possible edit",
		Long:              `Display every possible edit of an OpenAPI document with the changelog checks that cover it, the reason when none are expected, and a suggested id for a missing check otherwise.`,
		Args:              getChecksCoverageArgs(),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE:              getRun(runChecksCoverage),
	}

	addChecksFormatFlags(&cmd)
	enumWithOptions(&cmd, newEnumSliceValue(GetCoverageTags(), nil), "tags", "t", "include only edits matching the tags: values of the same dimension are ORed, dimensions are ANDed")
	cmd.PersistentFlags().Bool("patterns", false, "list the waiver and non-contract patterns instead of the edits")

	return &cmd
}

// getChecksCoverageArgs rejects arguments and the flag combination the
// command cannot honour.
func getChecksCoverageArgs() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := cobra.NoArgs(cmd, args); err != nil {
			return err
		}
		return checkPatternsWithoutTags(cmd)
	}
}

// checkPatternsWithoutTags rejects --tags with --patterns: the patterns
// listing has no edits to filter.
func checkPatternsWithoutTags(cmd *cobra.Command) error {
	patterns, err := cmd.Flags().GetBool("patterns")
	if err != nil || !patterns {
		return nil
	}
	if cmd.Flags().Changed("tags") {
		return errors.New("--tags cannot be used with --patterns: --tags filters edits, and --patterns lists patterns rather than edits")
	}
	return nil
}

func runChecksCoverage(flags *Flags, stdout io.Writer) (bool, *ReturnError) {

	format := flags.getFormat()

	formatter, err := formatters.Lookup(format, formatters.DefaultFormatterOpts())
	if err != nil {
		return false, getErrUnsupportedFormat(format, checksCoverageCmd)
	}

	var bytes []byte
	if flags.getViper().GetBool("patterns") {
		bytes, err = formatter.RenderCoveragePatterns(coverage.Patterns(), formatters.NewRenderOpts())
	} else {
		edits := slices.DeleteFunc(coverage.Analyze(checker.GetAllRules().Metadata()), func(edit coverage.Edit) bool {
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

// coverageTagDimensions is the tag vocabulary of `checks changelog coverage`:
// the audit status, the edit's polarity, and the edit's action.
var coverageTagDimensions = []tagDimension[coverage.Edit]{
	{
		values: []string{"covered", "uncovered", "waived", "non-contract"},
		match: func(value string, edit coverage.Edit) bool {
			return value == string(edit.Status)
		},
	},
	{
		values: []string{"request", "response", "document", "shared"},
		match: func(value string, edit coverage.Edit) bool {
			return value == edit.Polarity
		},
	},
	{
		values: []string{"add", "remove", "change", "increase", "decrease", "set", "unset"},
		match: func(value string, edit coverage.Edit) bool {
			return value == edit.Action
		},
	},
}

func GetCoverageTags() []string {
	return tagValues(coverageTagDimensions)
}

func matchCoverageTags(tags []string, edit coverage.Edit) bool {
	return matchTagDimensions(tags, coverageTagDimensions, edit)
}
