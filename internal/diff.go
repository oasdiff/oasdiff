package internal

import (
	"fmt"
	"io"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/load"
	"github.com/spf13/cobra"
)

const diffCmd = "diff"

func getDiffCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:   "diff base revision [flags]",
		Short: "Generate a diff report",
		Long:  "Generate a diff report between base and revision specs." + specHelp,
		Args:  getParseArgs(),
		RunE:  getRun(runDiff),
	}

	addCommonDiffFlags(&cmd)
	enumWithOptions(&cmd, newEnumSliceValue(diff.GetExcludeDiffOptions(), nil), "exclude-elements", "e", "elements to exclude")
	enumWithOptions(&cmd, newEnumValue(formatters.SupportedFormatsByContentType(formatters.OutputDiff), string(formatters.FormatYAML)), "format", "f", "output format")
	cmd.PersistentFlags().BoolP("fail-on-diff", "o", false, "exit with return code 1 when any change is found")

	return &cmd
}

func runDiff(flags *Flags, stdout io.Writer) (bool, *ReturnError) {

	if flags.getFormat() == string(formatters.FormatJSON) {
		flags.addExcludeElements(diff.ExcludeEndpointsOption)
	}

	diffResult, err := calcDiff(flags)
	if err != nil {
		return false, err
	}

	if err := outputDiff(stdout, diffResult.diffReport, flags.getFormat()); err != nil {
		return false, err
	}

	return flags.getFailOnDiff() && !diffResult.diffReport.Empty(), nil
}

func outputDiff(stdout io.Writer, diffReport *diff.Diff, format string) *ReturnError {
	// formatter lookup
	formatter, err := formatters.Lookup(format, formatters.DefaultFormatterOpts())
	if err != nil {
		return getErrUnsupportedFormat(format, diffCmd)
	}

	// render
	bytes, err := formatter.RenderDiff(diffReport, formatters.NewRenderOpts())
	if err != nil {
		return getErrFailedPrint("diff "+format, err)
	}

	// print output
	_, _ = fmt.Fprintf(stdout, "%s\n", bytes)

	return nil
}

func calcDiff(flags *Flags) (*diffResult, *ReturnError) {

	openapi3.IncludeOrigin = true
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	if flags.getComposed() {
		return composedDiff(loader, flags)
	}

	return normalDiff(loader, flags)
}

type diffResult struct {
	diffReport        *diff.Diff
	operationsSources *diff.OperationsSourcesMap
	specInfoPair      *load.SpecInfoPair
}

func newDiffResult(d *diff.Diff, o *diff.OperationsSourcesMap, s *load.SpecInfoPair) *diffResult {
	return &diffResult{
		diffReport:        d,
		operationsSources: o,
		specInfoPair:      s,
	}
}

func normalDiff(loader load.Loader, flags *Flags) (*diffResult, *ReturnError) {

	flattenAllOf := load.GetOption(load.WithFlattenAllOf(), flags.getFlattenAllOf())
	flattenParams := load.GetOption(load.WithFlattenParams(), flags.getFlattenParams())
	lowerHeaderNames := load.GetOption(load.WithLowercaseHeaders(), flags.getCaseInsensitiveHeaders())

	s1, err := load.NewSpecInfo(loader, flags.getBase(), flattenAllOf, flattenParams, lowerHeaderNames)
	if err != nil {
		return nil, getErrFailedToLoadSpec("base", flags.getBase(), err)
	}

	s2, err := load.NewSpecInfo(loader, flags.getRevision(), flattenAllOf, flattenParams, lowerHeaderNames)
	if err != nil {
		return nil, getErrFailedToLoadSpec("revision", flags.getRevision(), err)
	}

	if flags.getBase().IsStdin() && flags.getRevision().IsStdin() {
		// io.ReadAll can only read stdin once, so in this edge case, we copy base into revision
		s2.Spec = s1.Spec
	}

	diffReport, operationsSources, err := diff.GetWithOperationsSourcesMap(flags.toConfig(), s1, s2)
	if err != nil {
		return nil, getErrDiffFailed(err)
	}

	return newDiffResult(diffReport, operationsSources, load.NewSpecInfoPair(s1, s2)), nil
}

func composedDiff(loader load.Loader, flags *Flags) (*diffResult, *ReturnError) {

	flattenAllOf := load.GetOption(load.WithFlattenAllOf(), flags.getFlattenAllOf())
	flattenParams := load.GetOption(load.WithFlattenParams(), flags.getFlattenParams())
	lowerHeaderNames := load.GetOption(load.WithLowercaseHeaders(), flags.getCaseInsensitiveHeaders())

	s1, err := load.NewSpecInfoFromGlob(loader, flags.getBase().Path, flattenAllOf, flattenParams, lowerHeaderNames)
	if err != nil {
		return nil, getErrFailedToLoadSpecs("base", flags.getBase().Path, err)
	}

	s2, err := load.NewSpecInfoFromGlob(loader, flags.getRevision().Path, flattenAllOf, flattenParams, lowerHeaderNames)
	if err != nil {
		return nil, getErrFailedToLoadSpecs("revision", flags.getRevision().Path, err)
	}

	diffReport, operationsSources, err := diff.GetPathsDiff(flags.toConfig(), s1, s2)
	if err != nil {
		return nil, getErrDiffFailed(err)
	}

	return newDiffResult(diffReport, operationsSources, nil), nil
}
