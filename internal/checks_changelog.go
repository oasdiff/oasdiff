package internal

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/checker/localizations"
	"github.com/oasdiff/oasdiff/checker/metaschema"
	"github.com/oasdiff/oasdiff/checker/rules"
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
	addCheckIdFlag(cmd, "display only the check with this id")
	cmd.PersistentFlags().String("location", "", "include only checks with a location containing this string")
}

// addCheckIdFlag registers an --id flag that completes to the changelog
// check ids.
func addCheckIdFlag(cmd *cobra.Command, usage string) {
	cmd.PersistentFlags().String("id", "", usage)
	_ = cmd.RegisterFlagCompletionFunc("id", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		rules := checker.GetAllRules()
		ids := make([]string, len(rules))
		for i, rule := range rules {
			ids[i] = rule.Id
		}
		return ids, cobra.ShellCompDirectiveNoFileComp
	})
}

// checkKnownId rejects an --id that names no changelog check.
func checkKnownId(id string) *ReturnError {
	if id == "" {
		return nil
	}
	if !slices.ContainsFunc(checker.GetAllRules(), func(rule checker.BackwardCompatibilityRule) bool {
		return rule.Id == id
	}) {
		return getErrInvalidFlags(fmt.Errorf("unknown check id %q", id))
	}
	return nil
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

	id := flags.getId()
	if returnErr := checkKnownId(id); returnErr != nil {
		return returnErr
	}

	// filter rules
	severity := flags.getSeverity()
	location := flags.getLocation()
	checks := make(formatters.Checks, 0, len(rules))
	for _, rule := range rules {
		if id != "" && rule.Id != id {
			continue
		}

		if location != "" && !slices.ContainsFunc(rule.Locations, func(loc string) bool {
			return strings.Contains(loc, location)
		}) {
			continue
		}

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
			Guards:      guardStrings(rule.Guards),
			Locations:   rule.Locations,
			Generated:   rule.Generated,
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
		// the two wire directions; DirectionNone is deliberately not offered,
		// since its rules are selected by their kind (lifecycle) or area
		values: []string{checker.DirectionRequest.String(), checker.DirectionResponse.String()},
		match: func(value string, rule checker.BackwardCompatibilityRule) bool {
			return value == rule.Direction.String()
		},
	},
	{
		values: actionStrings(metaschema.Actions),
		match: func(value string, rule checker.BackwardCompatibilityRule) bool {
			return slices.Contains(rule.Actions(), metaschema.Action(value))
		},
	},
	{
		// only the two ordered effects; the others (incomparable, unknown,
		// none, violation) would collide with other dimensions' vocabulary or
		// add little as filters
		values: []string{checker.EffectWidens.String(), checker.EffectNarrows.String()},
		match: func(value string, rule checker.BackwardCompatibilityRule) bool {
			return value == rule.Effect.String()
		},
	},
	{
		// the areas with rules; AreaInfo, AreaServers and AreaNone are
		// deliberately not offered
		values: areaStrings(rules.AreaSchema, rules.AreaParameters, rules.AreaRequestBody, rules.AreaResponses,
			rules.AreaPaths, rules.AreaHeaders, rules.AreaSecurity, rules.AreaTags, rules.AreaComponents),
		match: func(value string, rule checker.BackwardCompatibilityRule) bool {
			return value == rule.Area.String()
		},
	},
	{
		values: kindStrings(rules.Kinds),
		match: func(value string, rule checker.BackwardCompatibilityRule) bool {
			return value == rule.Kind.String()
		},
	},
	{
		values: guardStrings(rules.Guards),
		match: func(value string, rule checker.BackwardCompatibilityRule) bool {
			return slices.Contains(rule.Guards, checker.Guard(value))
		},
	},
	{
		values: []string{"generated", "hand-written"},
		match: func(value string, rule checker.BackwardCompatibilityRule) bool {
			return (value == "generated") == rule.Generated
		},
	},
}

func areaStrings(areas ...rules.Area) []string {
	strs := make([]string, len(areas))
	for i, a := range areas {
		strs[i] = a.String()
	}
	return strs
}

func kindStrings(kinds []rules.Kind) []string {
	strs := make([]string, len(kinds))
	for i, k := range kinds {
		strs[i] = k.String()
	}
	return strs
}

func GetChangelogTags() []string {
	return tagValues(changelogTagDimensions)
}

func matchChangelogTags(tags []string, rule checker.BackwardCompatibilityRule) bool {
	return matchTagDimensions(tags, changelogTagDimensions, rule)
}

func guardStrings(guards []checker.Guard) []string {
	strs := make([]string, len(guards))
	for i, g := range guards {
		strs[i] = string(g)
	}
	return strs
}

func actionStrings(actions []metaschema.Action) []string {
	strs := make([]string, len(actions))
	for i, a := range actions {
		strs[i] = string(a)
	}
	return strs
}
