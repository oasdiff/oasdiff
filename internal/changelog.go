package internal

import (
	"fmt"
	"io"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/load"
	"github.com/oasdiff/oasdiff/versioning"
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
	addComposedFlag(&cmd)
	addCommonBreakingFlags(&cmd)
	enumWithOptions(&cmd, newEnumValue(GetFailOnValues(), ""), "fail-on", "o", "exit with return code 1 on a change at this severity or higher, or with MISSING-VERSION-BUMP on breaking changes that have no major version increase in info.version")
	enumWithOptions(&cmd, newEnumValue(GetSupportedLevels(), LevelInfo), "level", "", "output errors with this level or higher")
	addOpenFlags(&cmd, "changelog")

	return &cmd
}

func enumWithOptions(cmd *cobra.Command, value enumVal, name, shorthand, usage string) {
	cmd.PersistentFlags().VarP(value, name, shorthand, usage+": "+value.listOf())
}

func runChangelog(flags *Flags, stdout, stderr io.Writer) (bool, *ReturnError) {

	level, err := checker.NewLevel(flags.getLevel())
	if err != nil {
		return false, getErrInvalidFlags(fmt.Errorf("invalid level value: %q", flags.getLevel()))
	}

	return getChangelog(flags, stdout, stderr, level, false)
}

func getChangelog(flags *Flags, stdout, stderr io.Writer, level checker.Level, isBreaking bool) (bool, *ReturnError) {

	diffResult, returnErr := calcDiff(flags)
	if returnErr != nil {
		return false, returnErr
	}

	severityLevels, returnErr := getCustomSeverityLevels(flags.getSeverityLevelsFile())
	if returnErr != nil {
		return false, returnErr
	}

	bcConfig := checker.NewConfig(
		checker.GetAllChecks(),
		checker.WithSeverityLevels(severityLevels),
		checker.WithDeprecation(flags.getDeprecationDaysBeta(), flags.getDeprecationDaysStable()),
		checker.WithAttributes(flags.getAttributes()),
		checker.WithStabilityLevel(flags.getStabilityLevel()),
	)

	errs, returnErr := filterIgnored(
		checker.CheckBackwardCompatibilityUntilLevel(
			bcConfig,
			diffResult.diffReport,
			diffResult.operationsSources,
			level),
		flags.getWarnIgnoreFile(),
		flags.getErrIgnoreFile(),
		checker.NewLocalizer(flags.getLang()))

	if returnErr != nil {
		return false, returnErr
	}

	if returnErr := outputChangelog(flags, stdout, errs, diffResult.specInfoPair, diffResult.diffReport.Empty(), isBreaking); returnErr != nil {
		return false, returnErr
	}

	if flags.getOpen() {
		if err := uploadAndOpen(flags, stderr, isBreaking, errs, diffResult.baseSpecs, diffResult.revSpecs, diffResult.diffReport.Empty()); err != nil {
			// --open is additive: an upload error, unsupported source, or
			// composed mode must not change the exit code or pre-empt --fail-on.
			// Warn to stderr (not stdout, so it never corrupts piped --format
			// json/yaml output) and continue.
			_, _ = fmt.Fprintf(stderr, "warning: could not open the side-by-side review: %v\n", err)
		}
	}

	// the policy and a severity are two values of --fail-on, so the caller asked
	// for one or the other, never both
	if flags.getFailOn() == FailOnMissingVersionBump {
		verdict := versioning.Check(diffResult.diffReport, errs)
		// stderr, not stdout: the verdict is not a change, and stdout carries
		// the report, which may be a json or yaml document.
		_, _ = fmt.Fprintln(stderr, versioningPolicyLine(verdict))
		return failOnVersionBump(verdict), nil
	}

	failOn, returnErr := flags.failOnLevel()
	if returnErr != nil {
		return false, returnErr
	}

	if failOn == checker.NONE {
		return false, nil
	}

	return errs.HasLevelOrHigher(failOn), nil
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

func outputChangelog(flags *Flags, stdout io.Writer, errs checker.Changes, specInfoPair *load.SpecInfoPair, diffEmpty, isBreaking bool) *ReturnError {

	// formatter lookup
	formatter, err := formatters.Lookup(flags.getFormat(), formatters.FormatterOpts{
		Language:        flags.getLang(),
		BaseVersion:     specInfoPair.GetBaseVersion(),
		RevisionVersion: specInfoPair.GetRevisionVersion(),
	})
	if err != nil {
		return getErrUnsupportedFormat(flags.getFormat(), changelogCmd)
	}

	// render
	colorMode, err := checker.NewColorMode(flags.getColor())
	if err != nil {
		return getErrInvalidColorMode(err)
	}

	bytes, err := formatter.RenderChangelog(errs, formatters.RenderOpts{ColorMode: colorMode, TemplatePath: flags.getTemplate(), DiffEmpty: diffEmpty, IsBreaking: isBreaking})
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
