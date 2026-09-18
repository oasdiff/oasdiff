package internal

import (
	"fmt"
	"io"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/checker/localizations"
	"github.com/oasdiff/oasdiff/checker/rules"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/spf13/cobra"
)

const checksExplainCmd = "checks changelog explain"

// getChecksExplainCmd explains one changelog check by id: what it reports and
// the derivation that gives it its severity.
func getChecksExplainCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:               "explain check-id",
		Short:             "Explain a check: what it reports and why it has its severity",
		Long:              `Explain one changelog check: what change it reports, and the derivation that gives it its severity.`,
		Args:              getChecksExplainArgs(),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE:              runChecksExplain,
	}

	addChecksFormatFlags(&cmd)
	enumWithOptions(&cmd, newEnumValue(localizations.GetSupportedLanguages(), localizations.LangDefault), "lang", "l", "language for localized output")

	return &cmd
}

// getChecksExplainArgs requires exactly one argument naming a changelog check,
// and rejects the listing filters explain inherits from `checks changelog`.
func getChecksExplainArgs() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if err := checkNoListingFilters(cmd); err != nil {
			return err
		}
		return checkChangelogId(args[0])
	}
}

// checkNoListingFilters rejects the `checks changelog` filters: they select
// rows of the listing, and explain is given its check as an argument.
func checkNoListingFilters(cmd *cobra.Command) error {
	for _, name := range []string{"id", "location", "tags", "severity"} {
		if cmd.Flags().Changed(name) {
			return fmt.Errorf("--%s cannot be used with explain: it filters the listing, and explain takes the check id as its argument", name)
		}
	}
	return nil
}

func findChangelogRule(id string) *checker.BackwardCompatibilityRule {
	for _, rule := range checker.GetAllRules() {
		if rule.Id == id {
			return &rule
		}
	}
	return nil
}

// runChecksExplain is a plain cobra handler rather than a getRun runner: the
// positional argument is a check id, which getRun would misread as a base
// spec source.
func runChecksExplain(cmd *cobra.Command, args []string) error {
	flags := NewFlags()

	if err := RunViper(cmd, flags.getViper()); err != nil {
		setReturnValue(cmd, err.Code)
		return err
	}

	cmd.Root().SilenceUsage = true

	if err := outputExplanation(cmd.OutOrStdout(), flags, args[0]); err != nil {
		setReturnValue(cmd, err.Code)
		return err
	}
	return nil
}

func outputExplanation(stdout io.Writer, flags *Flags, id string) *ReturnError {

	format := flags.getFormat()

	formatter, err := formatters.Lookup(format, formatters.FormatterOpts{
		Language: flags.getLang(),
	})
	if err != nil {
		return getErrUnsupportedFormat(format, checksExplainCmd)
	}

	explanation := explainChangelogRule(*findChangelogRule(id), checker.NewLocalizer(flags.getLang()))

	bytes, err := formatter.RenderExplain(explanation, formatters.NewRenderOpts())
	if err != nil {
		return getErrFailedPrint(checksExplainCmd+" "+format, err)
	}

	_, _ = fmt.Fprintf(stdout, "%s\n", bytes)

	return nil
}

func explainChangelogRule(rule checker.BackwardCompatibilityRule, localizer checker.Localizer) formatters.Explanation {

	derived, reasoning := rules.ExplainLevel(rule.Effect, rule.Direction, rule.Guards...)
	if derived != rule.Level {
		// cannot happen while the severity-deviations ledger is empty
		// (TestSeverityLaw); if a deviation is ever recorded, its reason
		// should be surfaced here
		reasoning = []string{fmt.Sprintf("The stored level deviates from the severity law, which derives %s.", derived.String())}
	}

	mitigation := ""
	if commentKey := rule.Id + "-comment"; localizer(commentKey) != commentKey {
		mitigation = localizer(commentKey)
	}

	return formatters.Explanation{
		Id:          rule.Id,
		Level:       rule.Level.String(),
		Description: localizer(rule.Description),
		Derived:     derived == rule.Level,
		Reasoning:   reasoning,
		Direction:   rule.Direction.String(),
		Area:        rule.Area.String(),
		Kind:        rule.Kind.String(),
		Effect:      rule.Effect.String(),
		Guards:      guardStrings(rule.Guards),
		Generated:   rule.Generated,
		Locations:   rule.Locations,
		Mitigation:  mitigation,
		Override:    fmt.Sprintf("a --severity-levels file line %q overrides the level", rule.Id+" "+overrideLevel(rule.Level)),
	}
}

// overrideLevel picks the alternative level for the override hint: the point
// is to show the syntax with a level that differs from the current one.
func overrideLevel(level checker.Level) string {
	if level == checker.WARN {
		return "info"
	}
	return "warn"
}
