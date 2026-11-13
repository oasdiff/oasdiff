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

Some rules only break once allOf combines subschemas, and validate checks each
schema as written. --flatten-allof merges the subschemas first, so the rules see
the schema a payload is validated against. Subschemas that cannot be merged are
reported as a finding, and the spec is validated as written.

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
	cmd.PersistentFlags().Bool("flatten-allof", false, "merge subschemas under allOf before validating")
	addHiddenFlattenFlag(&cmd)

	return &cmd
}

func runValidate(flags *Flags, stdout io.Writer) (bool, *ReturnError) {

	spec, findings, returnErr := loadSpecToValidate(flags)
	if returnErr != nil {
		return false, returnErr
	}

	// Render zero findings through the formatter too: the empty representation
	// is format-specific, so it's the formatter's call, not an early return's.
	//
	// A merge finding goes first: Validate sorts its findings by line, and a
	// merge finding has none.
	findings = append(findings, validate.Validate(spec.Spec, flags.getBase().String())...)

	if returnErr := outputFindings(flags, stdout, findings); returnErr != nil {
		return false, returnErr
	}

	// Findings are always printed; the --fail-on level decides whether they
	// also fail the command. Default is error, so warnings and info surface
	// without failing CI unless the caller lowers the threshold.
	failOn, returnErr := flags.failOnLevel()
	if returnErr != nil {
		return false, returnErr
	}

	return findings.HasLevelOrHigher(failOn), nil
}

// loadSpecToValidate loads the spec to validate, and the findings the load
// itself produced.
//
// With --flatten-allof the subschemas under allOf are merged first, so the
// rules see the schema a payload is validated against rather than each
// subschema on its own. A merge failure is a finding rather than a failure to
// load: the spec parsed, and the rules still have the spec as written to check.
func loadSpecToValidate(flags *Flags) (*load.SpecInfo, formatters.Findings, *ReturnError) {

	source := flags.getBase()

	if !flags.getFlattenAllOf() {
		spec, err := load.NewSpecInfo(newValidateLoader(flags), source)
		if err != nil {
			return nil, nil, getErrFailedToLoadSpec("original", source, err)
		}
		return spec, nil, nil
	}

	spec, err := load.NewSpecInfo(newValidateLoader(flags), source, load.WithFlattenAllOf())
	if err == nil {
		return spec, nil, nil
	}

	flattenErr, ok := errors.AsType[*load.FlattenError](err)
	if !ok {
		return nil, nil, getErrFailedToLoadSpec("original", source, err)
	}

	// The merge writes into the document it was given, so a failed merge
	// leaves it half merged. Load it again for the spec as written.
	spec, err = load.NewSpecInfo(newValidateLoader(flags), source)
	if err != nil {
		return nil, nil, getErrFailedToLoadSpec("original", source, err)
	}

	return spec, formatters.Findings{validate.AllOfConflict(flattenErr.Err, source.String())}, nil
}

func newValidateLoader(flags *Flags) *openapi3.Loader {

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = flags.getAllowExternalRefs()
	// Origin tracking lets the typed cluster errors carry an *Origin
	// pointing at the offending element. Cheap to leave on; consumers
	// that don't surface line/column simply ignore the extra fields.
	loader.IncludeOrigin = true

	return loader
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
