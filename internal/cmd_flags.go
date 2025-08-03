package internal

import (
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/checker/localizations"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/spf13/cobra"
)

func addCommonDiffFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolP("composed", "c", false, "work in 'composed' mode, compare paths in all specs matching base and revision globs")
	cmd.PersistentFlags().StringP("match-path", "p", "", "include only paths that match this regular expression")
	cmd.PersistentFlags().StringP("unmatch-path", "q", "", "exclude paths that match this regular expression")
	cmd.PersistentFlags().String("filter-extension", "", "exclude paths and operations with an OpenAPI Extension matching this regular expression")
	cmd.PersistentFlags().String("prefix-base", "", "add this prefix to paths in base-spec before comparison")
	cmd.PersistentFlags().String("prefix-revision", "", "add this prefix to paths in revised-spec before comparison")
	cmd.PersistentFlags().String("strip-prefix-base", "", "strip this prefix from paths in base-spec before comparison")
	cmd.PersistentFlags().String("strip-prefix-revision", "", "strip this prefix from paths in revised-spec before comparison")
	cmd.PersistentFlags().Bool("include-path-params", false, "include path parameter names in endpoint matching")
	cmd.PersistentFlags().Bool("flatten-allof", false, "merge subschemas under allOf before diff")
	cmd.PersistentFlags().Bool("flatten-params", false, "merge common parameters at path level with operation parameters")
	cmd.PersistentFlags().Bool("case-insensitive-headers", false, "case-insensitive header name comparison")

	addHiddenFlattenFlag(cmd)
	addHiddenCircularDepFlag(cmd)
}

// addHiddenFlattenFlag adds --flatten as a hidden flag
// --flatten was replaced by --flatten-allof
// we still accept --flatten as a synonym for --flatten-allof to avoid breaking existing scripts
func addHiddenFlattenFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool("flatten", false, "merge subschemas under allOf before diff")
	hideFlag(cmd, "flatten")
}

// addHiddenCircularDepFlag adds --max-circular-dep as a hidden flag
// --max-circular-dep is no longer needed because kin-openapi3 handles circular references automatically since https://github.com/getkin/kin-openapi/pull/970
// we still accept --max-circular-dep to avoid breaking existing scripts, but we ignore this flag
func addHiddenCircularDepFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().Int("max-circular-dep", 5, "maximum allowed number of circular dependencies between objects in OpenAPI specs")
	hideFlag(cmd, "max-circular-dep")
}

// hideFlag hides a flag from the help
// this is an alternative to marking the flag as deprecated
// marking the flag as deprecated is problematic because it causes cobra to write an error message to stdout which messes up the json and yaml output
func hideFlag(cmd *cobra.Command, flag string) {
	if err := cmd.PersistentFlags().MarkHidden(flag); err != nil {
		// we can ignore this error safely
		_ = err
	}
}

func addCommonBreakingFlags(cmd *cobra.Command) {
	enumWithOptions(cmd, newEnumValue(localizations.GetSupportedLanguages(), localizations.LangDefault), "lang", "l", "language for localized output")
	cmd.PersistentFlags().String("err-ignore", "", "configuration file for ignoring errors")
	cmd.PersistentFlags().String("warn-ignore", "", "configuration file for ignoring warnings")
	cmd.PersistentFlags().VarPF(newEnumSliceValue(checker.GetOptionalRuleIds(), nil), "include-checks", "i", "optional checks")
	hideFlag(cmd, "include-checks")
	cmd.PersistentFlags().Uint("deprecation-days-beta", checker.DefaultBetaDeprecationDays, "min days required between deprecating a beta resource and removing it")
	cmd.PersistentFlags().Uint("deprecation-days-stable", checker.DefaultStableDeprecationDays, "min days required between deprecating a stable resource and removing it")
	enumWithOptions(cmd, newEnumValue(checker.GetSupportedColorValues(), "auto"), "color", "", "when to colorize textual output")
	enumWithOptions(cmd, newEnumValue(formatters.SupportedFormatsByContentType(formatters.OutputChangelog), string(formatters.FormatText)), "format", "f", "output format")
	cmd.PersistentFlags().String("severity-levels", "", "configuration file for custom severity levels")
	cmd.PersistentFlags().StringSlice("attributes", nil, "OpenAPI Extensions to include in json or yaml output")
	cmd.PersistentFlags().String("template", "", "path to custom template file for changelog generation")
}
