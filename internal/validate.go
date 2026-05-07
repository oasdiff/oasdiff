package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/load"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v3"
)

const validateCmd = "validate"

// kinUnknownID is the fallback rule ID for any kin-openapi error that
// hasn't yet been migrated to the typed RequiredFieldError /
// FieldVersionMismatchError clusters in kin's openapi3 package. The
// intentionally awkward name signals that the finding wants typing on
// the kin side; leaving findings stuck under this catchall is a smell,
// not a steady state. Each new kin release converts more sites.
const kinUnknownID = "kin-validation-error"

const (
	formatText = "text"
	formatYAML = "yaml"
)

func getValidateCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:   "validate spec",
		Short: "Validate an OpenAPI spec against the spec",
		Long: `Validate an OpenAPI spec, reporting per-RFC violations such as invalid
type values, missing required fields, malformed paths, and unresolved $refs.

Phase 1 wraps kin-openapi's Validate() walker and dispatches each typed
error to a stable kebab-case rule ID via errors.As against kin's
RequiredFieldError and FieldVersionMismatchError clusters. Findings are
emitted in the chosen format (text by default; -f yaml for structured
output). Field names match the changelog command's output
(id, text, level, source) so a single CI script can parse both.

Exit code is 0 if no findings, 1 if any finding is reported.

Spec can be a path to a file, a URL or '-' to read standard input.
`,
		Args: cobra.ExactArgs(1),
		RunE: getRun(runValidate),
	}

	enumWithOptions(&cmd, newEnumValue([]string{formatText, formatYAML}, formatText), "format", "f", "output format")
	cmd.PersistentFlags().Bool("allow-external-refs", true, "allow external $refs in specs; disable to prevent SSRF when processing untrusted specs")

	return &cmd
}

func runValidate(flags *Flags, stdout io.Writer) (bool, *ReturnError) {

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = flags.getAllowExternalRefs()

	spec, err := load.NewSpecInfo(loader, flags.getBase())
	if err != nil {
		return false, getErrFailedToLoadSpec("original", flags.getBase(), err)
	}

	verr := spec.Spec.Validate(context.Background())
	if verr == nil {
		return false, nil
	}

	findings := mapKinErrors(flags.getBase().String(), verr)
	if err := writeFindings(stdout, findings, flags.getFormat()); err != nil {
		return false, getErrFailedPrint("validate "+flags.getFormat(), err)
	}

	return true, nil
}

// Finding is a single validation finding. Field names mirror the
// changelog command's output (id/text/level/source) so consumers can
// parse both with the same data structure. Id is a descriptive
// kebab-case identifier matching the changelog rule-ID convention.
type Finding struct {
	Id     string        `yaml:"id"`
	Text   string        `yaml:"text"`
	Level  checker.Level `yaml:"level"`
	Source string        `yaml:"source"`
}

// mapKinErrors flattens kin-openapi's MultiError tree into a list of
// Findings. kin can return either a single error or a MultiError; the
// MultiError can itself contain MultiErrors, so we recurse.
func mapKinErrors(source string, err error) []Finding {
	if err == nil {
		return nil
	}
	if me, ok := err.(openapi3.MultiError); ok {
		var out []Finding
		for _, sub := range me {
			out = append(out, mapKinErrors(source, sub)...)
		}
		return out
	}
	return []Finding{{
		Id:     ruleIDForKinError(err),
		Text:   err.Error(),
		Level:  checker.ERR,
		Source: source,
	}}
}

// ruleIDForKinError dispatches a kin-openapi error to a stable
// kebab-case rule ID. Uses the typed cluster wrappers from kin's
// openapi3 package — *RequiredFieldError covers "X must be non-empty"
// failures, *FieldVersionMismatchError covers "X is for OpenAPI >=Y"
// failures. Both clusters carry the offending field name, which we
// transform into the canonical rule ID.
//
// kin errors not yet migrated to a cluster fall through to
// kinUnknownID. The rule ID for converted leaves is derived from the
// cluster's .Field — the per-leaf type isn't consulted here because
// (a) the cluster carries enough metadata, and (b) deriving from a
// single field keeps the dispatch stable as kin adds new leaves.
func ruleIDForKinError(err error) string {
	var rfe *openapi3.RequiredFieldError
	if errors.As(err, &rfe) {
		return ruleIDFromField(rfe.Field) + "-required"
	}

	var fvm *openapi3.FieldVersionMismatchError
	if errors.As(err, &fvm) {
		return ruleIDFromField(fvm.Field) + "-field-for-3-1-plus"
	}

	return kinUnknownID
}

// ruleIDFromField normalises a field path into a kebab-case identifier.
// Strips a leading "$" (for JSON Schema keywords like "$defs"),
// replaces "." with "-" (for paths like "info.version"), and inserts
// "-" before each uppercase letter (for camelCase like "prefixItems").
func ruleIDFromField(field string) string {
	field = strings.TrimPrefix(field, "$")
	field = strings.ReplaceAll(field, ".", "-")
	var b strings.Builder
	for i, r := range field {
		if i > 0 && unicode.IsUpper(r) {
			b.WriteByte('-')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

func writeFindings(w io.Writer, findings []Finding, format string) error {
	switch format {
	case formatYAML:
		bytes, err := yaml.Marshal(findings)
		if err != nil {
			return err
		}
		_, err = w.Write(bytes)
		return err
	case formatText, "":
		writeFindingsText(w, findings)
		return nil
	default:
		return fmt.Errorf("unsupported format %q", format)
	}
}

// writeFindingsText emits the changelog-style header summary line
// followed by one block per finding. The block format mirrors
// checker.ApiChange's MultiLineError output minus the operation/path
// line that doesn't apply to validate findings:
//
//	error	[<rule-id>] at <source>
//		<text>
func writeFindingsText(w io.Writer, findings []Finding) {
	var nErr, nWarn, nInfo int
	for _, f := range findings {
		switch f.Level {
		case checker.ERR:
			nErr++
		case checker.WARN:
			nWarn++
		case checker.INFO:
			nInfo++
		}
	}
	fmt.Fprintf(w, "%d findings: %d error, %d warning, %d info\n", len(findings), nErr, nWarn, nInfo)

	for _, f := range findings {
		fmt.Fprintf(w, "%s\t[%s] at %s\n\t%s\n\n", f.Level.String(), f.Id, f.Source, f.Text)
	}
}
