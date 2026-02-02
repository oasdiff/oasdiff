package internal

import (
	"fmt"
	"io"
	"slices"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/checker/localizations"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/spf13/cobra"
)

const checksCmd = "checks"

func getChecksCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:               "checks [flags]",
		Short:             "Display checks",
		Long:              `Display a list of all supported checks.`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions, // see https://github.com/spf13/cobra/issues/1969
		RunE:              getRun(runChecks),
	}

	enumWithOptions(&cmd, newEnumValue(localizations.GetSupportedLanguages(), localizations.LangDefault), "lang", "l", "language for localized output")
	enumWithOptions(&cmd, newEnumValue(formatters.SupportedFormatsByContentType(formatters.OutputChecks), string(formatters.FormatText)), "format", "f", "output format")
	enumWithOptions(&cmd, newEnumSliceValue([]string{"info", "warn", "error"}, nil), "severity", "s", "include only checks with any of specified severities")
	enumWithOptions(&cmd, newEnumSliceValue(getAllTags(), nil), "tags", "t", "include only checks with all specified tags")

	return &cmd
}

func runChecks(flags *Flags, stdout io.Writer) (bool, *ReturnError) {
	return false, outputChecks(stdout, flags, checker.GetAllRules())
}

func outputChecks(stdout io.Writer, flags *Flags, rules []checker.BackwardCompatibilityRule) *ReturnError {

	format := flags.getFormat()

	// formatter lookup
	formatter, err := formatters.Lookup(format, formatters.FormatterOpts{
		Language: flags.getLang(),
	})
	if err != nil {
		return getErrUnsupportedFormat(format, checksCmd)
	}

	// filter rules
	severity := flags.getSeverity()
	checks := make(formatters.Checks, 0, len(rules))
	for _, rule := range rules {
		// severity
		if len(severity) > 0 {
			if rule.Level == checker.ERR && !slices.Contains(severity, "error") {
				continue
			}
			if rule.Level == checker.WARN && !slices.Contains(severity, "warn") {
				continue
			}
			if rule.Level == checker.INFO && !slices.Contains(severity, "info") {
				continue
			}
		}

		// tags
		if !matchTags(flags.getTags(), rule) {
			continue
		}

		checks = append(checks, formatters.Check{
			Id:          rule.Id,
			Level:       rule.Level.String(),
			Description: rule.Description,
		})
	}

	// render
	slices.SortFunc(checks, checks.SortFunc)
	bytes, err := formatter.RenderChecks(checks, formatters.NewRenderOpts())
	if err != nil {
		return getErrFailedPrint("checks "+format, err)
	}

	// print output
	_, _ = fmt.Fprintf(stdout, "%s\n", bytes)

	return nil
}
