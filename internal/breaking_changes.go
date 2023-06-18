package internal

import (
	"io"

	"github.com/spf13/cobra"
	"github.com/tufin/oasdiff/checker"
)

func getBreakingChangesCmd() *cobra.Command {

	flags := ChangelogFlags{}

	cmd := cobra.Command{
		Use:   "breaking original-spec revised-spec [flags]",
		Short: "Display breaking-changes",
		RunE: func(cmd *cobra.Command, args []string) error {

			flags.base = args[0]
			flags.revision = args[1]

			// by now flags have been parsed successfully so we don't need to show usage on any errors
			cmd.Root().SilenceUsage = true

			failEmpty, err := runBreakingChanges(&flags, cmd.OutOrStdout())
			if err != nil {
				setReturnValue(cmd, err.Code)
				return err
			}

			if failEmpty {
				setReturnValue(cmd, 1)
			}

			return nil
		},
	}

	cmd.PersistentFlags().BoolVarP(&flags.composed, "composed", "c", false, "work in 'composed' mode, compare paths in all specs matching base and revision globs")
	cmd.PersistentFlags().StringVarP(&flags.format, "format", "f", "text", "output format: yaml, json, text")
	cmd.PersistentFlags().StringSliceVarP(&flags.excludeElements, "exclude-elements", "", nil, "comma-separated list of elements to exclude from diff")
	cmd.PersistentFlags().StringVarP(&flags.matchPath, "match-path", "", "", "include only paths that match this regular expression")
	cmd.PersistentFlags().StringVarP(&flags.filterExtension, "filter-extension", "", "", "exclude paths and operations with an OpenAPI Extension matching this regular expression")
	cmd.PersistentFlags().IntVarP(&flags.circularReferenceCounter, "max-circular-dep", "", 5, "maximum allowed number of circular dependencies between objects in OpenAPI specs")
	cmd.PersistentFlags().StringVarP(&flags.prefixBase, "prefix-base", "", "", "add this prefix to paths in 'base' spec before comparison")
	cmd.PersistentFlags().StringVarP(&flags.prefixRevision, "prefix-revision", "", "", "add this prefix to paths in 'revision' spec before comparison")
	cmd.PersistentFlags().StringVarP(&flags.stripPrefixBase, "strip-prefix-base", "", "", "strip this prefix from paths in 'base' spec before comparison")
	cmd.PersistentFlags().StringVarP(&flags.stripPrefixRevision, "strip-prefix-revision", "", "", "strip this prefix from paths in 'revision' spec before comparison")
	cmd.PersistentFlags().BoolVarP(&flags.includePathParams, "include-path-params", "", false, "include path parameter names in endpoint matching")

	cmd.PersistentFlags().VarP(&flags.failOn, "fail-on", "", "exit with return code 1 when output includes errors with this level or higher")
	cmd.PersistentFlags().VarP(&flags.lang, "lang", "", "language for localized output")
	cmd.PersistentFlags().StringVarP(&flags.errIgnoreFile, "err-ignore", "", "", "configuration file for ignoring errors")
	cmd.PersistentFlags().StringVarP(&flags.warnIgnoreFile, "warn-ignore", "", "", "configuration file for ignoring warnings")
	// level
	// info-ignore
	// deprecation-days
	// lang
	cmd.PersistentFlags().StringSliceVarP(&flags.includeChecks, "include-checks", "", nil, "comma-separated list of optional breaking-changes checks")
	return &cmd
}

func runBreakingChanges(flags *ChangelogFlags, stdout io.Writer) (bool, *ReturnError) {
	return getChangelog(flags, stdout, checker.WARN)
}
