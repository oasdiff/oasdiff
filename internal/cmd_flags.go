package internal

import (
	"github.com/spf13/cobra"
	"github.com/tufin/oasdiff/checker"
	"github.com/tufin/oasdiff/checker/localizations"
	"github.com/tufin/oasdiff/diff"
)

func addCommonDiffFlags(cmd *cobra.Command, flags Flags) {
	cmd.PersistentFlags().BoolVarP(flags.refComposed(), "composed", "c", false, "work in 'composed' mode, compare paths in all specs matching base and revision globs")
	enumWithOptions(cmd, newEnumSliceValue(diff.ExcludeDiffOptions, nil, flags.refExcludeElements()), "exclude-elements", "e", "comma-separated list of elements to exclude")
	cmd.PersistentFlags().StringVarP(flags.refMatchPath(), "match-path", "p", "", "include only paths that match this regular expression")
	cmd.PersistentFlags().StringVarP(flags.refFilterExtension(), "filter-extension", "", "", "exclude paths and operations with an OpenAPI Extension matching this regular expression")
	cmd.PersistentFlags().IntVarP(flags.refCircularReferenceCounter(), "max-circular-dep", "", 5, "maximum allowed number of circular dependencies between objects in OpenAPI specs")
	cmd.PersistentFlags().StringVarP(flags.refPrefixBase(), "prefix-base", "", "", "add this prefix to paths in base-spec before comparison")
	cmd.PersistentFlags().StringVarP(flags.refPrefixRevision(), "prefix-revision", "", "", "add this prefix to paths in revised-spec before comparison")
	cmd.PersistentFlags().StringVarP(flags.refStripPrefixBase(), "strip-prefix-base", "", "", "strip this prefix from paths in base-spec before comparison")
	cmd.PersistentFlags().StringVarP(flags.refStripPrefixRevision(), "strip-prefix-revision", "", "", "strip this prefix from paths in revised-spec before comparison")
	cmd.PersistentFlags().BoolVarP(flags.refIncludePathParams(), "include-path-params", "", false, "include path parameter names in endpoint matching")

	cmd.PersistentFlags().BoolVarP(flags.refFlattenAllOf(), "flatten", "", false, "merge subschemas under allOf before diff")
	_ = cmd.PersistentFlags().MarkDeprecated("flatten", "use flatten-allof instead")
	cmd.PersistentFlags().BoolVarP(flags.refFlattenAllOf(), "flatten-allof", "", false, "merge subschemas under allOf before diff")
	cmd.PersistentFlags().BoolVarP(flags.refFlattenParams(), "flatten-params", "", false, "merge common parameters at path level with operation parameters")
}

func addCommonBreakingFlags(cmd *cobra.Command, flags Flags) {
	enumWithOptions(cmd, newEnumValue(localizations.GetSupportedLanguages(), localizations.LangDefault, flags.refLang()), "lang", "l", "language for localized output")
	cmd.PersistentFlags().StringVarP(flags.refErrIgnoreFile(), "err-ignore", "", "", "configuration file for ignoring errors")
	cmd.PersistentFlags().StringVarP(flags.refWarnIgnoreFile(), "warn-ignore", "", "", "configuration file for ignoring warnings")
	cmd.PersistentFlags().VarP(newEnumSliceValue(checker.GetOptionalChecks(), nil, flags.refIncludeChecks()), "include-checks", "i", "comma-separated list of optional checks (run 'oasdiff checks --required false' to see options)")
	cmd.PersistentFlags().IntVarP(flags.refDeprecationDaysBeta(), "deprecation-days-beta", "", checker.BetaDeprecationDays, "min days required between deprecating a beta resource and removing it")
	cmd.PersistentFlags().IntVarP(flags.refDeprecationDaysStable(), "deprecation-days-stable", "", checker.StableDeprecationDays, "min days required between deprecating a stable resource and removing it")
	enumWithOptions(cmd, newEnumValue([]string{"auto", "always", "never"}, "auto", flags.refColor()), "color", "", "when to colorize textual output")
}
