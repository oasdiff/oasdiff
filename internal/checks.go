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
// The RunE, DisableFlagsInUseLine and Args below work around a cobra
// limitation; see noSubcommandArgs.
func getChecksCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:               "checks",
		Short:             "Display checks",
		Long:              `Display the rules oasdiff applies, one listing per rule set.`,
		ValidArgsFunction: cobra.NoFileCompletions,

		// These three go together and work around a cobra limitation;
		// see noSubcommandArgs for what to delete when it is fixed.
		Args:                  noSubcommandArgs,
		DisableFlagsInUseLine: true,
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
// WORKAROUND for a cobra limitation, to be removed when cobra fixes it.
//
// The problem: a command with subcommands and no Run/RunE is not Runnable, and
// cobra's execute() returns flag.ErrHelp at its `!c.Runnable()` check *before*
// it validates Args (command.go:956 vs 969 in v1.10.2). ExecuteC then turns
// flag.ErrHelp into a nil error, so `oasdiff checks <typo>` prints the help and
// exits 0. A script cannot tell that from success, and it disagrees with
// `oasdiff <typo>`, which exits non-zero.
//
// The workaround, in three parts:
//  1. RunE, which does nothing but print the help, so the command is Runnable
//     and cobra reaches Args at all.
//  2. noSubcommandArgs, which rejects a stray argument in cobra's own
//     unknown-command shape.
//  3. DisableFlagsInUseLine, because a Runnable command puts its UseLine in the
//     usage block, and without this it reads `oasdiff checks [flags]` while the
//     flags actually live on the subcommands.
//
// Upstream: reported since 2020 in https://github.com/spf13/cobra/issues/1156
// (also #706, #981, #2130). https://github.com/spf13/cobra/pull/2167 proposes
// an ErrorOnUnknownSubcommand field and has been awaiting review since 2024.
//
// To remove once cobra ships it: set ErrorOnUnknownSubcommand on this command,
// restore `Args: cobra.NoArgs`, and drop all three parts above plus
// noSubcommandArgs and its tests. The usage block then loses its
// `oasdiff checks` line and reads `oasdiff checks [command]` alone, which is
// what it should have said all along.
//
// Why the message is printed here rather than left to cobra: cobra only prints
// that form when it fails to *find* a command, and `checks <typo>` resolves to
// `checks` with a leftover argument, which otherwise prints the whole help. The
// two lines are emitted the way cobra emits them (ErrPrefix + the hint as its
// own line, not as part of the error string), so the error text stays free of
// trailing punctuation and the output is byte-identical.
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
