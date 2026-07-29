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

const checksChangelogCmd = "checks changelog"

// getChecksCmd groups the per-rule-set listings and lists no rules itself:
// naming one rule set by omission stopped being tenable once `checks validate`
// existed, so the rule set is always explicit. Run bare it prints the help.
//
// It has a RunE only so that it stays Runnable, which is what makes cobra
// reach Args and reject `checks <typo>` with a non-zero exit. A non-runnable
// parent returns flag.ErrHelp before Args is ever evaluated, so a typo'd
// subcommand would print help and exit 0, which a script cannot detect.
// DisableFlagsInUseLine keeps `[flags]` out of the usage: the flags belong to
// the subcommands.
func getChecksCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:                   "checks",
		Short:                 "Display checks",
		Long:                  `Display the rules oasdiff applies, one listing per rule set.`,
		Args:                  noSubcommandArgs,
		DisableFlagsInUseLine: true,
		ValidArgsFunction:     cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(getChecksChangelogCmd(), getChecksValidateCmd())

	return &cmd
}

// getChecksChangelogCmd lists the breaking-change and changelog rules, the
// counterpart of `checks validate`.
func getChecksChangelogCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:               "changelog",
		Short:             "Display changelog and breaking-change checks",
		Long:              `Display a list of all supported changelog and breaking-change checks.`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE:              getRun(runChecks),
	}

	// Registered per command rather than inherited: viper binds a command's own
	// persistent flags (see bindFlags), so an inherited flag would parse but
	// never reach the config.
	addChecksChangelogFlags(&cmd)

	return &cmd
}

// noSubcommandArgs rejects a stray argument to `checks` as an unknown
// subcommand, reported exactly as cobra reports one at the top level, so the
// same mistake reads the same at either depth.
//
// Reporting it here rather than letting cobra do it: cobra only prints that
// form when it fails to *find* a command, and `checks <typo>` resolves to
// `checks` with a leftover argument, which otherwise prints the whole help.
// The two lines are emitted the way cobra emits them (ErrPrefix + the hint as
// its own line, not as part of the error string), so the error text stays free
// of trailing punctuation and the output is byte-identical.
func noSubcommandArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}

	err := fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
	cmd.PrintErrln(cmd.ErrPrefix(), err.Error())
	cmd.PrintErrf("Run '%v --help' for usage.\n", cmd.CommandPath())

	// Already reported, so cobra must not print it again or follow it with the
	// usage block.
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	return err
}

// addChecksChangelogFlags registers the flags for the changelog listing.
// --lang belongs here and not on the validate listing: these descriptions are
// localized, and --tags likewise, since only these rules carry tags.
func addChecksChangelogFlags(cmd *cobra.Command) {
	addChecksFormatFlags(cmd)
	addChecksSeverityFlag(cmd)
	enumWithOptions(cmd, newEnumSliceValue(getAllTags(), nil), "tags", "t", "include only checks with all specified tags")
	enumWithOptions(cmd, newEnumValue(localizations.GetSupportedLanguages(), localizations.LangDefault), "lang", "l", "language for localized output")
}

// addChecksValidateFlags registers the flags for the validate listing: output
// and severity only. The descriptions are plain English rather than localized
// (a finding's text comes from the parser at runtime), so --lang would have
// nothing to translate, and validate rules carry no tags.
func addChecksValidateFlags(cmd *cobra.Command) {
	addChecksFormatFlags(cmd)
	addChecksSeverityFlag(cmd)
}

// addChecksFormatFlags registers the output format flag every listing needs.
func addChecksFormatFlags(cmd *cobra.Command) {
	enumWithOptions(cmd, newEnumValue(formatters.SupportedFormatsByContentType(formatters.OutputChecks), string(formatters.FormatText)), "format", "f", "output format")
}

// addChecksSeverityFlag registers --severity, which both listings support now
// that a validate rule has a severity of its own.
func addChecksSeverityFlag(cmd *cobra.Command) {
	enumWithOptions(cmd, newEnumSliceValue([]string{"info", "warn", "error"}, nil), "severity", "s", "include only checks with any of specified severities")
}

// matchSeverity reports whether level passes the --severity filter. An empty
// filter matches every rule. The flag's values are the short forms, which do
// not all match Level.String(), so this switches on the level itself.
func matchSeverity(severity []string, level checker.Level) bool {
	if len(severity) == 0 {
		return true
	}
	switch level {
	case checker.ERR:
		return slices.Contains(severity, "error")
	case checker.WARN:
		return slices.Contains(severity, "warn")
	case checker.INFO:
		return slices.Contains(severity, "info")
	}
	return true
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
		if !matchTags(flags.getTags(), rule) {
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
			Action:      rule.Action.String(),
			Description: localizer(rule.Description),
			Mitigation:  mitigation,
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
