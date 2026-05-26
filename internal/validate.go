package internal

import (
	"errors"
	"fmt"
	"io"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/load"
	"github.com/oasdiff/oasdiff/validate"
	"github.com/spf13/cobra"
)

const validateCmd = "validate"

func getValidateCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:   "validate spec",
		Short: "Validate an OpenAPI spec against the spec",
		Long: `Validate an OpenAPI spec, reporting per-RFC violations such as invalid
type values, missing required fields, malformed paths, and unresolved $refs.

Each finding has a stable rule ID, a severity (error, warning, or info), a
human-readable message, and a source location (file:line:column when the
loader tracks origins). Output format is selectable: text by default,
'-f yaml' or '-f json' for structured output, or '-f githubactions' to emit
CI annotations that surface each finding inline on the pull request.

Most findings are errors (the spec can't be reliably consumed). A few are
downgraded: version-portability issues (a 3.1 field in an older doc),
ambiguity or redundancy (conflicting paths, duplicate parameters), and a
default value that doesn't match its schema are warnings; an example that
doesn't match its schema is info. Use --fail-on to control which severities
fail the command.

Exit codes:
  0 — no findings at or above the --fail-on level
  1 — at least one finding at or above the --fail-on level
  102 — failed to load the spec

Spec can be a path to a file, a URL, a git ref (e.g. main:openapi.yaml), or '-' to read standard input.
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}
			// color only affects the text output; yaml/json/githubactions
			// ignore it, so flag the mismatch rather than silently dropping it.
			if cmd.Flags().Changed("color") {
				if format, _ := cmd.Flags().GetString("format"); format != string(formatters.FormatText) {
					return errors.New("--color is only relevant with the 'text' format")
				}
			}
			return nil
		},
		RunE: getRun(runValidate),
	}

	enumWithOptions(&cmd, newEnumValue(formatters.SupportedFormatsByContentType(formatters.OutputValidate), string(formatters.FormatText)), "format", "f", "output format")
	enumWithOptions(&cmd, newEnumValue(checker.GetSupportedColorValues(), "auto"), "color", "", "when to colorize textual output")
	enumWithOptions(&cmd, newEnumValue(GetSupportedLevels(), LevelErr), "fail-on", "o", "exit with code 1 when a finding has this severity or higher")
	cmd.PersistentFlags().Bool("allow-external-refs", true, "allow external $refs in specs; disable to prevent SSRF when processing untrusted specs")

	return &cmd
}

func runValidate(flags *Flags, stdout io.Writer) (bool, *ReturnError) {

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = flags.getAllowExternalRefs()
	// Origin tracking lets the typed cluster errors carry an *Origin
	// pointing at the offending element. Cheap to leave on; consumers
	// that don't surface line/column simply ignore the extra fields.
	loader.IncludeOrigin = true

	spec, err := load.NewSpecInfo(loader, flags.getBase())
	if err != nil {
		return false, getErrFailedToLoadSpec("original", flags.getBase(), err)
	}

	findings := validate.Validate(spec.Spec, flags.getBase().String())
	if len(findings) == 0 {
		return false, nil
	}

	if returnErr := outputFindings(flags, stdout, findings); returnErr != nil {
		return false, returnErr
	}

	// Findings are always printed; the --fail-on level decides whether they
	// also fail the command. Default is error, so warnings and info surface
	// without failing CI unless the caller lowers the threshold.
	failOn, err := checker.NewLevel(flags.getFailOn())
	if err != nil {
		return false, getErrInvalidFlags(fmt.Errorf("invalid fail-on value: %q", flags.getFailOn()))
	}

	return findings.HasLevelOrHigher(failOn), nil
}

// outputFindings renders the findings through the shared formatters, the
// same path changelog/breaking/flatten use: look up the formatter for the
// requested format and call its render method. RenderValidate is
// implemented only by the text, yaml, and json formatters, matching the
// formats the command advertises.
func outputFindings(flags *Flags, stdout io.Writer, findings formatters.Findings) *ReturnError {

	formatter, err := formatters.Lookup(flags.getFormat(), formatters.FormatterOpts{
		Language: flags.getLang(),
	})
	if err != nil {
		return getErrUnsupportedFormat(flags.getFormat(), validateCmd)
	}

	colorMode, err := checker.NewColorMode(flags.getColor())
	if err != nil {
		return getErrInvalidColorMode(err)
	}

	bytes, err := formatter.RenderValidate(findings, formatters.RenderOpts{ColorMode: colorMode})
	if err != nil {
		return getErrFailedPrint(validateCmd+" "+flags.getFormat(), err)
	}

	_, _ = fmt.Fprintf(stdout, "%s\n", bytes)

	return nil
}
