package internal

import (
	"fmt"
	"io"
	"slices"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/checker/localizations"
	"github.com/oasdiff/oasdiff/checker/metaschema"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/spf13/cobra"
)

const checksChangelogCmd = "checks changelog"

// getChecksChangelogCmd lists the breaking-change and changelog rules, the
// counterpart of `checks validate`.
func getChecksChangelogCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:               "changelog",
		Short:             "Display changelog and breaking-change checks",
		Long:              `Display a list of all supported changelog and breaking-change checks.`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE:              getRun(runChecksChangelog),
	}

	// Registered per command rather than inherited: viper binds a command's own
	// persistent flags (see bindFlags), so an inherited flag would parse but
	// never reach the config.
	addChecksChangelogFlags(&cmd)

	cmd.AddCommand(getChecksCoverageCmd())

	return &cmd
}

// addChecksChangelogFlags registers the flags for the changelog listing.
// --lang belongs here and not on the validate listing: these descriptions are
// localized, and --tags likewise, since only these rules carry tags.
func addChecksChangelogFlags(cmd *cobra.Command) {
	addChecksFormatFlags(cmd)
	addChecksSeverityFlag(cmd)
	enumWithOptions(cmd, newEnumSliceValue(GetChangelogTags(), nil), "tags", "t", "include only checks matching the tags: values of the same dimension are ORed, dimensions are ANDed")
	enumWithOptions(cmd, newEnumValue(localizations.GetSupportedLanguages(), localizations.LangDefault), "lang", "l", "language for localized output")
}

func runChecksChangelog(flags *Flags, stdout io.Writer) (bool, *ReturnError) {
	return false, outputChangelogRules(stdout, flags, checker.GetAllRules())
}

func outputChangelogRules(stdout io.Writer, flags *Flags, rules []checker.BackwardCompatibilityRule) *ReturnError {

	format := flags.getFormat()

	// formatter lookup
	formatter, err := formatters.Lookup(format, formatters.FormatterOpts{
		Language: flags.getLang(),
	})
	if err != nil {
		return getErrUnsupportedFormat(format, checksChangelogCmd)
	}

	localizer := checker.NewLocalizer(flags.getLang())

	// filter rules
	severity := flags.getSeverity()
	checks := make(formatters.Checks, 0, len(rules))
	for _, rule := range rules {
		if !matchSeverity(severity, rule.Level) {
			continue
		}

		// tags
		if !matchChangelogTags(flags.getTags(), rule) {
			continue
		}

		commentKey := rule.Id + "-comment"
		mitigation := localizer(commentKey)
		if mitigation == commentKey {
			mitigation = ""
		}

		checks = append(checks, formatters.Check{
			Id:          rule.Id,
			Level:       rule.Level.String(),
			Direction:   rule.Direction.String(),
			Area:        rule.Area.String(),
			Kind:        rule.Kind.String(),
			Actions:     actionStrings(rule.Actions()),
			Effect:      rule.Effect.String(),
			Locations:   rule.Locations,
			Description: localizer(rule.Description),
			Mitigation:  mitigation,
		})
	}

	// render
	slices.SortFunc(checks, checks.SortFunc)
	bytes, err := formatter.RenderChecks(checks, formatters.NewRenderOpts())
	if err != nil {
		return getErrFailedPrint(checksChangelogCmd+" "+format, err)
	}

	// print output
	_, _ = fmt.Fprintf(stdout, "%s\n", bytes)

	return nil
}

// changelogTagDimensions is the tag vocabulary of `checks changelog`:
// direction, action (the syntactic edits from the rule's location claims),
// effect (the rule's verdict), area, and kind.
var changelogTagDimensions = []tagDimension[checker.BackwardCompatibilityRule]{
	{
		values: []string{"request", "response"},
		match: func(value string, rule checker.BackwardCompatibilityRule) bool {
			return value == rule.Direction.String()
		},
	},
	{
		values: []string{"add", "remove", "change", "increase", "decrease", "set", "unset"},
		match: func(value string, rule checker.BackwardCompatibilityRule) bool {
			return slices.Contains(rule.Actions(), metaschema.Action(value))
		},
	},
	{
		values: []string{"widens", "narrows"},
		match: func(value string, rule checker.BackwardCompatibilityRule) bool {
			switch value {
			case "widens":
				return rule.Effect == checker.EffectWidens
			case "narrows":
				return rule.Effect == checker.EffectNarrows
			}
			return false
		},
	},
	{
		values: []string{"schema", "parameters", "requestBody", "responses", "paths", "headers", "security", "tags", "components"},
		match: func(value string, rule checker.BackwardCompatibilityRule) bool {
			return value == rule.Area.String()
		},
	},
	{
		values: []string{"existence", "requiredness", "mutability", "type", "constraints", "values", "structure", "lifecycle"},
		match: func(value string, rule checker.BackwardCompatibilityRule) bool {
			return value == rule.Kind.String()
		},
	},
}

func GetChangelogTags() []string {
	return tagValues(changelogTagDimensions)
}

func matchChangelogTags(tags []string, rule checker.BackwardCompatibilityRule) bool {
	return matchTagDimensions(tags, changelogTagDimensions, rule)
}

func actionStrings(actions []metaschema.Action) []string {
	strs := make([]string, len(actions))
	for i, a := range actions {
		strs[i] = string(a)
	}
	return strs
}
