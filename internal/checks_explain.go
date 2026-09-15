package internal

import (
	"fmt"
	"io"
	"slices"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/checker/localizations"
	"github.com/oasdiff/oasdiff/checker/rules"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/validate"
	"github.com/spf13/cobra"
)

const checksExplainCmd = "checks explain"

// getChecksExplainCmd explains one check by id, resolving both the changelog
// and the validate rule sets, so there is a single place to ask what an id
// means wherever it was encountered.
func getChecksExplainCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:               "explain check-id",
		Short:             "Explain a check: what it reports and why it has its severity",
		Long:              `Explain one changelog or validate check: what change it reports, and, for changelog checks, the derivation that gives it its severity.`,
		Args:              getChecksExplainArgs(),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE:              runChecksExplain,
	}

	addChecksFormatFlags(&cmd)
	enumWithOptions(&cmd, newEnumValue(localizations.GetSupportedLanguages(), localizations.LangDefault), "lang", "l", "language for localized output")

	return &cmd
}

// getChecksExplainArgs requires exactly one argument naming a known check.
func getChecksExplainArgs() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		id := args[0]
		if findChangelogRule(id) == nil && !slices.Contains(validate.RuleIDs(), id) {
			return fmt.Errorf("unknown check id %q", id)
		}
		return nil
	}
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

	var explanation formatters.Explanation
	if rule := findChangelogRule(id); rule != nil {
		explanation = explainChangelogRule(*rule, checker.NewLocalizer(flags.getLang()))
	} else {
		explanation = explainValidateRule(id)
	}

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

// explainValidateRule explains a validate rule, which carries only an id,
// level, and description: its severity is set by the rule, not derived.
func explainValidateRule(id string) formatters.Explanation {
	return formatters.Explanation{
		Id:          id,
		Level:       validate.RuleLevel(id).String(),
		Description: validate.RuleDescription(id),
	}
}
