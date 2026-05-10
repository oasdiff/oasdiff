package internal

import (
	"fmt"
	"io"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3conv"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/load"
	"github.com/spf13/cobra"
)

const upgradeCmd = "upgrade"

func getUpgradeCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:   "upgrade spec",
		Short: "Canonicalize an OpenAPI 3.x spec to the latest 3.x version",
		Long: `Convert an OpenAPI 3.x spec to the latest 3.x representation.

The walker rewrites schema-level constructs in place (nullable -> type array,
boolean exclusiveMinimum/Maximum -> numeric, example -> examples, and similar),
then updates the openapi version string. The transforms are idempotent: an
already-canonical spec is unchanged aside from a possible version-string bump.

Spec can be a path to a file, a URL or '-' to read standard input.
`,
		Args: cobra.ExactArgs(1),
		RunE: getRun(runUpgrade),
	}

	enumWithOptions(&cmd, newEnumValue(formatters.SupportedFormatsByContentType(formatters.OutputFlatten), string(formatters.FormatYAML)), "format", "f", "output format")
	cmd.PersistentFlags().Bool("allow-external-refs", true, "allow external $refs in specs; disable to prevent SSRF when processing untrusted specs")
	addHiddenCircularDepFlag(&cmd)

	return &cmd
}

func runUpgrade(flags *Flags, stdout io.Writer) (bool, *ReturnError) {

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = flags.getAllowExternalRefs()
	spec, err := load.NewSpecInfo(loader, flags.getBase())
	if err != nil {
		return false, getErrFailedToLoadSpec("original", flags.getBase(), err)
	}

	openapi3conv.Upgrade(spec.Spec)

	format := flags.getFormat()

	if returnErr := outputUpgradedSpec(stdout, spec.Spec, format); returnErr != nil {
		return false, returnErr
	}

	return false, nil
}

// autoUpgradeSpecs canonicalizes the given specs to the latest OpenAPI 3.x
// when --auto-upgrade is set. The walker is idempotent on already-canonical
// specs, so calling it on a same-version pair is a safe no-op (the version
// string bumps to the latest 3.x). Used by diff/breaking/changelog/summary
// to make cross-version comparisons (e.g. 3.0 vs 3.1) just work.
//
// Nil entries are skipped; callers don't have to filter.
func autoUpgradeSpecs(enabled bool, specs ...*load.SpecInfo) {
	if !enabled {
		return
	}
	for _, s := range specs {
		if s == nil || s.Spec == nil {
			continue
		}
		openapi3conv.Upgrade(s.Spec)
	}
}

func outputUpgradedSpec(stdout io.Writer, spec *openapi3.T, format string) *ReturnError {
	// Reuse the flatten output path: both subcommands serialize a full
	// *openapi3.T to JSON/YAML and the rendering is identical. If a future
	// upgrade-specific renderer is needed, split it then.
	formatter, err := formatters.Lookup(format, formatters.DefaultFormatterOpts())
	if err != nil {
		return getErrUnsupportedFormat(format, upgradeCmd)
	}

	bytes, err := formatter.RenderFlatten(spec, formatters.NewRenderOpts())
	if err != nil {
		return getErrFailedPrint("upgrade "+format, err)
	}

	_, _ = fmt.Fprintf(stdout, "%s\n", bytes)

	return nil
}
