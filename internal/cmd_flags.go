package internal

import (
	"fmt"

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
	cmd.PersistentFlags().Bool("match-inline-refs", true, "match validation-equivalent inline/$ref subschemas as the same anyOf/oneOf branch")
	cmd.PersistentFlags().Bool("flatten-allof", false, "merge subschemas under allOf before diff")
	cmd.PersistentFlags().Bool("flatten-params", false, "merge common parameters at path level with operation parameters")
	cmd.PersistentFlags().Bool("case-insensitive-headers", true, "case-insensitive header name comparison (HTTP headers are case-insensitive per RFC 7230)")
	cmd.PersistentFlags().StringSlice("exclude-extensions", nil, "OpenAPI Extension names to exclude from diff (e.g., x-internal)")
	cmd.PersistentFlags().Bool("allow-external-refs", true, "allow external $refs in specs; disable to prevent SSRF when processing untrusted specs")
	cmd.PersistentFlags().Bool("auto-upgrade", false, "canonicalize both specs to the latest OpenAPI 3.x before diffing; useful for cross-version comparisons (e.g. 3.0 vs 3.1)")
	cmd.PersistentFlags().Bool("fetch", false, "fetch missing git revisions from the 'origin' remote (writes objects to your local repo)")

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

// deprecatedFlags maps a hidden, legacy flag to the guidance shown when it is
// used. These stay hidden and functional (so existing scripts don't break) but
// warn the user, via warnDeprecatedFlags, to move off them. We can't use cobra's
// MarkDeprecated because it writes the notice to stdout, which corrupts the
// json/yaml result (the same reason hideFlag exists); the notice goes to stderr
// instead. review-token / review-meta are intentionally not here: they are
// hidden because they are an automation interface, not because they are legacy.
var deprecatedFlags = map[string]string{
	"flatten":          "use --flatten-allof instead",
	"max-circular-dep": "it is no longer needed and is ignored",
	"include-checks":   "use --severity-levels instead",
}

// warnDeprecatedFlags writes a stderr notice for each deprecated flag the user
// actually set on cmd. Stderr, never stdout, so it can't corrupt machine output.
func warnDeprecatedFlags(cmd *cobra.Command) {
	for name, guidance := range deprecatedFlags {
		if f := cmd.Flags().Lookup(name); f != nil && f.Changed {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Flag --%s is deprecated: %s\n", name, guidance)
		}
	}
}

func addCommonBreakingFlags(cmd *cobra.Command) {
	enumWithOptions(cmd, newEnumValue(localizations.GetSupportedLanguages(), localizations.LangDefault), "lang", "l", "language for localized output")
	cmd.PersistentFlags().String("err-ignore", "", "configuration file for ignoring errors")
	cmd.PersistentFlags().String("warn-ignore", "", "configuration file for ignoring warnings")
	// Accepted for back-compat but ignored: the optional-checks mechanism it
	// drove was retired. Deprecated (see deprecatedFlags); use --severity-levels.
	cmd.PersistentFlags().StringSliceP("include-checks", "i", nil, "deprecated: use --severity-levels")
	hideFlag(cmd, "include-checks")
	cmd.PersistentFlags().Uint("deprecation-days-beta", checker.DefaultBetaDeprecationDays, "min days required between deprecating a beta resource and removing it")
	cmd.PersistentFlags().Uint("deprecation-days-stable", checker.DefaultStableDeprecationDays, "min days required between deprecating a stable resource and removing it")
	enumWithOptions(cmd, newEnumValue(checker.GetSupportedColorValues(), "auto"), "color", "", "when to colorize textual output")
	enumWithOptions(cmd, newEnumValue(formatters.SupportedFormatsByContentType(formatters.OutputChangelog), string(formatters.FormatText)), "format", "f", "output format")
	cmd.PersistentFlags().String("severity-levels", "", "configuration file for custom severity levels")
	cmd.PersistentFlags().StringSlice("attributes", nil, "OpenAPI Extensions to include in json or yaml output")
	cmd.PersistentFlags().String("template", "", "path to custom template file for changelog generation")
	enumWithOptions(cmd, newEnumValue(checker.GetSupportedStabilityLevels(), ""), "stability-level", "", "minimum stability level to include")
}

// addOpenFlags registers --open and its companion review-upload flags. Kept out
// of addCommonBreakingFlags so the git-diff driver (which shares that helper but
// has no --open) doesn't inherit them.
func addOpenFlags(cmd *cobra.Command, outputName string) {
	cmd.PersistentFlags().Bool("open", false, fmt.Sprintf("after printing the %s, encrypt the comparison and upload it to oasdiff.com, then open the side-by-side review in a browser", outputName))
	cmd.PersistentFlags().String("review-token", "", "with --open, upload an authenticated review using this token instead of the free anonymous one")
	cmd.PersistentFlags().StringSlice("review-meta", nil, "with --open and --review-token, attach repeatable key=value metadata to the authenticated review (opaque; not interpreted by the CLI)")

	// Hidden from --help: the authenticated review is assembled by the Action,
	// not hand-typed, so these are an automation interface. They still work.
	hideFlag(cmd, "review-token")
	hideFlag(cmd, "review-meta")
}
