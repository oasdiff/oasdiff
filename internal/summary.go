package internal

import (
	"fmt"
	"io"

	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/spf13/cobra"
)

const summaryCmd = "summary"

func getSummaryCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:   "summary base revision [flags]",
		Short: "Generate a diff summary",
		Long:  "Display a summary of changes between base and revision specs." + specHelp,
		Args:  getParseArgs(),
		RunE:  getRun(runSummary),
	}

	addCommonDiffFlags(&cmd)
	addComposedFlag(&cmd)
	enumWithOptions(&cmd, newEnumSliceValue(diff.GetExcludeDiffOptions(), nil), "exclude-elements", "e", "elements to exclude")
	enumWithOptions(&cmd, newEnumValue(formatters.SupportedFormatsByContentType(formatters.OutputSummary), string(formatters.FormatYAML)), "format", "f", "output format")
	cmd.PersistentFlags().BoolP("fail-on-diff", "", false, "exit with return code 1 when any change is found")

	return &cmd
}

func runSummary(flags *Flags, stdout io.Writer) (bool, *ReturnError) {

	diffResult, err := calcDiff(flags)
	if err != nil {
		return false, err
	}

	if err := outputSummary(stdout, diffResult.diffReport, flags.getFormat()); err != nil {
		return false, err
	}

	return flags.getFailOnDiff() && !diffResult.diffReport.Empty(), nil
}

func outputSummary(stdout io.Writer, diffReport *diff.Diff, format string) *ReturnError {
	// formatter lookup
	formatter, err := formatters.Lookup(format, formatters.DefaultFormatterOpts())
	if err != nil {
		return getErrUnsupportedFormat(format, summaryCmd)
	}

	// render
	bytes, err := formatter.RenderSummary(diffReport, formatters.NewRenderOpts())
	if err != nil {
		return getErrFailedPrint(summaryCmd+" "+format, err)
	}

	// print output
	_, _ = fmt.Fprintf(stdout, "%s\n", bytes)

	return nil
}
