package internal

import (
	"fmt"
	"io"

	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/validate"
	"github.com/spf13/cobra"
)

const checksValidateCmd = "checks validate"

// getChecksValidateCmd lists the rules `oasdiff validate` can report, the
// counterpart of `oasdiff checks` for the breaking-change rules.
//
// The IDs are stable API, which is what a CI dashboard needs.
func getChecksValidateCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:   "validate",
		Short: "Display validate rule IDs",
		Long: `Display the rule IDs that 'oasdiff validate' can report.

IDs are stable: new ones may be added in later releases, but an existing ID is
never renamed or repurposed, so external tooling can depend on them.`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE:              getRun(runChecksValidate),
	}

	// No severity / tags filters: unlike a breaking-change rule, a validate rule
	// carries no tags, and there is no --severity-levels equivalent to reflect.
	addChecksFormatFlags(&cmd)

	return &cmd
}

func runChecksValidate(flags *Flags, stdout io.Writer) (bool, *ReturnError) {
	return false, outputValidateRuleIDs(stdout, flags)
}

func outputValidateRuleIDs(stdout io.Writer, flags *Flags) *ReturnError {

	format := flags.getFormat()

	formatter, err := formatters.Lookup(format, formatters.FormatterOpts{
		Language: flags.getLang(),
	})
	if err != nil {
		return getErrUnsupportedFormat(format, checksValidateCmd)
	}

	// RuleIDs() is already sorted. Level comes from the same mapping the
	// runtime classifier uses, so the listing cannot advertise a severity a
	// finding won't carry.
	ids := validate.RuleIDs()
	checks := make(formatters.Checks, 0, len(ids))
	for _, id := range ids {
		checks = append(checks, formatters.Check{
			Id:          id,
			Description: validate.RuleDescription(id),
			Level:       validate.RuleLevel(id).String(),
		})
	}

	bytes, err := formatter.RenderChecks(checks, formatters.NewRenderOpts())
	if err != nil {
		return getErrFailedPrint(checksValidateCmd+" "+format, err)
	}

	_, _ = fmt.Fprintf(stdout, "%s\n", bytes)

	return nil
}
