package internal

import (
	"io"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
	"github.com/tufin/oasdiff/diff"
)

func getDiffCmd() *cobra.Command {

	flags := DiffFlags{}

	cmd := cobra.Command{
		Use:   "diff",
		Short: "Generate a diff report",
		PreRun: func(cmd *cobra.Command, args []string) {
			if returnErr := flags.validate(); returnErr != nil {
				exit(false, returnErr, cmd.ErrOrStderr())
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			failEmpty, returnErr := runDiff(&flags, cmd.OutOrStdout())
			exit(failEmpty, returnErr, cmd.ErrOrStderr())
		},
	}

	cmd.PersistentFlags().BoolVarP(&flags.composed, "composed", "c", false, "work in 'composed' mode, compare paths in all specs matching base and revision globs")
	cmd.PersistentFlags().StringVarP(&flags.base, "base", "b", "", "path or URL (or a glob in Composed mode) of original OpenAPI spec in YAML or JSON format")
	cmd.PersistentFlags().StringVarP(&flags.revision, "revision", "r", "", "path or URL (or a glob in Composed mode) of revised OpenAPI spec in YAML or JSON format")
	cmd.PersistentFlags().StringVarP(&flags.format, "format", "f", "yaml", "output format: yaml, json, text or html")
	cmd.PersistentFlags().StringSliceVarP(&flags.excludeElements, "exclude-elements", "", nil, "comma-separated list of elements to exclude from diff")
	cmd.PersistentFlags().StringVarP(&flags.matchPath, "match-path", "", "", "include only paths that match this regular expression")
	cmd.PersistentFlags().StringVarP(&flags.filterExtension, "filter-extension", "", "", "exclude paths and operations with an OpenAPI Extension matching this regular expression")
	cmd.PersistentFlags().IntVarP(&flags.circularReferenceCounter, "max-circular-dep", "", 5, "maximum allowed number of circular dependencies between objects in OpenAPI specs")
	cmd.PersistentFlags().StringVarP(&flags.prefixBase, "prefix-base", "", "", "add this prefix to paths in 'base' spec before comparison")
	cmd.PersistentFlags().StringVarP(&flags.prefixRevision, "prefix-revision", "", "", "add this prefix to paths in 'revision' spec before comparison")
	cmd.PersistentFlags().StringVarP(&flags.stripPrefixBase, "strip-prefix-base", "", "", "strip this prefix from paths in 'base' spec before comparison")
	cmd.PersistentFlags().StringVarP(&flags.stripPrefixRevision, "strip-prefix-revision", "", "", "strip this prefix from paths in 'revision' spec before comparison")
	cmd.PersistentFlags().BoolVarP(&flags.matchPathParams, "match-path-params", "", false, "include path parameter names in endpoint matching")

	return &cmd
}

func runDiff(diffFlags *DiffFlags, stdout io.Writer) (bool, *ReturnError) {

	openapi3.CircularReferenceCounter = diffFlags.circularReferenceCounter

	config := diffFlags.toConfig()

	var diffReport *diff.Diff

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	if diffFlags.composed {
		var err *ReturnError
		if diffReport, _, err = composedDiff(loader, diffFlags.base, diffFlags.revision, config); err != nil {
			return false, err
		}
	} else {
		var err *ReturnError
		if diffReport, _, err = normalDiff(loader, diffFlags.base, diffFlags.revision, config); err != nil {
			return false, err
		}
	}

	if diffFlags.summary {
		if err := printYAML(stdout, diffReport.GetSummary()); err != nil {
			return false, getErrFailedPrint("summary", err)
		}
		return failEmpty(diffFlags.isFailOnDiff(), diffReport.Empty()), nil
	}

	return failEmpty(diffFlags.isFailOnDiff(), diffReport.Empty()), handleDiff(stdout, diffReport, diffFlags.format)
}
