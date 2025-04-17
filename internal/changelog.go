package internal

import (
	"fmt"
	"io"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/load"
	"github.com/spf13/cobra"
)

const changelogCmd = "changelog"

func getChangelogCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:   "changelog base revision [flags]",
		Short: "Display changelog",
		Long:  "Display changes between base and revision specs." + specHelp,
		Args:  getParseArgs(),
		RunE:  getRun(runChangelog),
	}

	addCommonDiffFlags(&cmd)
	addCommonBreakingFlags(&cmd)
	enumWithOptions(&cmd, newEnumValue(GetSupportedLevels(), ""), "fail-on", "o", "exit with return code 1 when output includes errors with this level or higher")
	enumWithOptions(&cmd, newEnumValue(GetSupportedLevels(), LevelInfo), "level", "", "output errors with this level or higher")

	return &cmd
}

func enumWithOptions(cmd *cobra.Command, value enumVal, name, shorthand, usage string) {
	cmd.PersistentFlags().VarP(value, name, shorthand, usage+": "+value.listOf())
}

func runChangelog(flags *Flags, stdout io.Writer) (bool, *ReturnError) {

	level, err := checker.NewLevel(flags.getLevel())
	if err != nil {
		return false, getErrInvalidFlags(fmt.Errorf("invalid level value: %q", flags.getLevel()))
	}

	return getChangelog(flags, stdout, level)
}

func getChangelog(flags *Flags, stdout io.Writer, level checker.Level) (bool, *ReturnError) {

	diffResult, returnErr := calcDiff(flags)
	if returnErr != nil {
		return false, returnErr
	}

	severityLevels, returnErr := getCustomSeverityLevels(flags.getSeverityLevelsFile())
	if returnErr != nil {
		return false, returnErr
	}

	errs, returnErr := filterIgnored(
		checker.CheckBackwardCompatibilityUntilLevel(
			checker.NewConfig(checker.GetAllChecks()).WithOptionalChecks(flags.getIncludeChecks()).WithSeverityLevels(severityLevels).WithDeprecation(flags.getDeprecationDaysBeta(), flags.getDeprecationDaysStable()).WithAttributes(flags.getAttributes()),
			diffResult.diffReport,
			diffResult.operationsSources,
			level),
		flags.getWarnIgnoreFile(),
		flags.getErrIgnoreFile(),
		checker.NewLocalizer(flags.getLang()))

	if returnErr != nil {
		return false, returnErr
	}

	if returnErr := outputChangelog(flags, stdout, errs, diffResult.specInfoPair); returnErr != nil {
		return false, returnErr
	}

	if flags.getFailOn() != "" {
		level, err := checker.NewLevel(flags.getFailOn())
		if err != nil {
			return false, getErrInvalidFlags(fmt.Errorf("invalid fail-on value %s", flags.getFailOn()))
		}
		return errs.HasLevelOrHigher(level), nil
	}

	return false, nil
}

func filterIgnored(errs checker.Changes, warnIgnoreFile string, errIgnoreFile string, l checker.Localizer) (checker.Changes, *ReturnError) {

	if warnIgnoreFile != "" {
		var err error
		errs, err = checker.ProcessIgnoredBackwardCompatibilityErrors(checker.WARN, errs, warnIgnoreFile, l)
		if err != nil {
			return nil, getErrCantProcessIgnoreFile("warn", err)
		}
	}

	if errIgnoreFile != "" {
		var err error
		errs, err = checker.ProcessIgnoredBackwardCompatibilityErrors(checker.ERR, errs, errIgnoreFile, l)
		if err != nil {
			return nil, getErrCantProcessIgnoreFile("err", err)
		}
	}

	return errs, nil
}

func outputChangelog(flags *Flags, stdout io.Writer, errs checker.Changes, specInfoPair *load.SpecInfoPair) *ReturnError {

	// formatter lookup
	formatter, err := formatters.Lookup(flags.getFormat(), formatters.FormatterOpts{
		Language: flags.getLang(),
	})
	if err != nil {
		return getErrUnsupportedFormat(flags.getFormat(), changelogCmd)
	}

	// render
	colorMode, err := checker.NewColorMode(flags.getColor())
	if err != nil {
		return getErrInvalidColorMode(err)
	}

	bytes, err := formatter.RenderChangelog(errs, formatters.RenderOpts{ColorMode: colorMode}, specInfoPair.GetBaseVersion(), specInfoPair.GetRevisionVersion())
	if err != nil {
		return getErrFailedPrint(changelogCmd+" "+flags.getFormat(), err)
	}

	// print output
	_, _ = fmt.Fprintf(stdout, "%s\n", bytes)

	return nil
}

func getCustomSeverityLevels(severityLevelsFile string) (map[string]checker.Level, *ReturnError) {
	if severityLevelsFile == "" {
		return nil, nil
	}

	m, err := checker.ProcessSeverityLevels(severityLevelsFile)
	if err != nil {
		return nil, getErrFailedToLoadSeverityLevels(severityLevelsFile, err)
	}

	return m, nil
}
