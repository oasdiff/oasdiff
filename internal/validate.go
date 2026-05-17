package internal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/TwiN/go-color"
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
	formatJSON = "json"
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

	enumWithOptions(&cmd, newEnumValue([]string{formatText, formatYAML, formatJSON}, formatText), "format", "f", "output format")
	enumWithOptions(&cmd, newEnumValue(checker.GetSupportedColorValues(), "auto"), "color", "", "when to colorize textual output")
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

	verr := spec.Spec.Validate(context.Background(), openapi3.EnableMultiError())
	if verr == nil {
		return false, nil
	}

	findings := mapKinErrors(flags.getBase().String(), verr)
	colorMode, cerr := checker.NewColorMode(flags.getColor())
	if cerr != nil {
		return false, getErrFailedPrint(validateCmd+" color", cerr)
	}
	if err := writeFindings(stdout, findings, flags.getFormat(), colorMode); err != nil {
		return false, getErrFailedPrint(validateCmd+" "+flags.getFormat(), err)
	}

	return true, nil
}

// Finding is a single validation finding. The JSON / YAML shape mirrors
// the changelog command's output so consumers can parse both with the
// same data structure.
//
// Comment, Operation, and Path use omitempty because doc-root findings
// (e.g. info-version-required) have no operation/path scope.
//
// Source.Line and Source.Column are populated when kin-openapi tracks
// origins (Loader.IncludeOrigin = true) and the underlying typed error
// carries the offending element's Origin. Both are 0 when the element
// doesn't have origin metadata (doc-root fields like openapi version)
// or origin tracking is off.
//
// Fingerprint is a stable 12-char identifier that lets a downstream
// tool match the same logical finding across base/revision spec
// versions (used by the Pro PR-comment for new/pre-existing/fixed
// partitioning). Format mirrors formatters/changes.go:computeFingerprint.
type Finding struct {
	Id          string        `yaml:"id"                    json:"id"`
	Text        string        `yaml:"text"                  json:"text"`
	Comment     string        `yaml:"comment,omitempty"     json:"comment,omitempty"`
	Level       checker.Level `yaml:"level"                 json:"level"`
	Operation   string        `yaml:"operation,omitempty"   json:"operation,omitempty"`
	Path        string        `yaml:"path,omitempty"        json:"path,omitempty"`
	Section     string        `yaml:"section"               json:"section"`
	Source      Source        `yaml:"source"                json:"source"`
	Fingerprint string        `yaml:"fingerprint"           json:"fingerprint"`
}

// Source identifies the spec location of a finding. File is the spec
// path; Line and Column come from kin's Origin tracking and are 0 for
// doc-root findings that have no per-key origin.
type Source struct {
	File   string `yaml:"file"             json:"file"`
	Line   int    `yaml:"line,omitempty"   json:"line,omitempty"`
	Column int    `yaml:"column,omitempty" json:"column,omitempty"`
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
	id := ruleIDForKinError(err)
	path, operation := pathOperationForKinError(err)
	f := Finding{
		Id:        id,
		Text:      unwrapContext(err).Error(),
		Level:     checker.ERR,
		Operation: operation,
		Path:      path,
		Section:   sectionForKinError(err),
		Source: Source{
			File:   source,
			Line:   lineForKinError(err),
			Column: columnForKinError(err),
		},
	}
	// Fingerprint last so it sees the populated fields it hashes over.
	f.Fingerprint = fingerprintFor(f, argsForKinError(err))
	return []Finding{f}
}

// pathOperationForKinError extracts the path template and HTTP method
// from kin's typed context wrappers (PathValidationError and
// OperationValidationError, added in getkin/kin-openapi #1183). Either
// return value is "" when the error chain carries no such scope, e.g.
// doc-root findings like info-version-required.
func pathOperationForKinError(err error) (path, operation string) {
	var pve *openapi3.PathValidationError
	if errors.As(err, &pve) {
		path = pve.Path
	}
	var ove *openapi3.OperationValidationError
	if errors.As(err, &ove) {
		operation = ove.Method
	}
	return path, operation
}

// unwrapContext strips kin's structural context wrappers
// (SectionValidationError / PathValidationError / OperationValidationError,
// kin #1183) from the front of the chain. That section/path/operation scope
// is captured in the Finding's typed fields, so Text should carry only the
// underlying message, without the redundant "invalid <scope>:" prefixes
// those wrappers add to Error().
func unwrapContext(err error) error {
	for {
		switch err.(type) {
		case *openapi3.SectionValidationError,
			*openapi3.PathValidationError,
			*openapi3.OperationValidationError:
			u := errors.Unwrap(err)
			if u == nil {
				return err
			}
			err = u
		default:
			return err
		}
	}
}

// sectionForKinError maps a typed kin error to its logical doc section,
// matching the values used by ApiChange / ComponentChange / SecurityChange
// in the existing changelog output (`paths`, `components`, `security`).
//
// The mapping is per-cluster + a light Field-prefix check on the cluster
// types that carry one (RequiredFieldError, FieldVersionMismatchError).
// Doc-root findings without a natural section (e.g. *RequiredFieldError
// {Field: "openapi"}) get the empty string.
func sectionForKinError(err error) string {
	// SectionValidationError (kin #1183) names the section directly and
	// authoritatively; prefer it over the cluster heuristics below, which
	// predate it and only approximate (e.g. they miscount inline component
	// schemas as "paths"). The cluster logic remains the fallback for
	// doc-root errors that aren't wrapped in a section at all.
	var secErr *openapi3.SectionValidationError
	if errors.As(err, &secErr) {
		return secErr.Section
	}

	// Cluster types that have a structural section regardless of payload.
	var ppe *openapi3.PathParametersError
	if errors.As(err, &ppe) {
		return "paths"
	}
	var wne *openapi3.WebhookNilError
	if errors.As(err, &wne) {
		return "webhooks"
	}
	var sute *openapi3.ServerURLTemplateError
	if errors.As(err, &sute) {
		return "servers"
	}
	var pre *openapi3.PathParameterRequiredError
	if errors.As(err, &pre) {
		return "paths"
	}
	var doid *openapi3.DuplicateOperationIDError
	if errors.As(err, &doid) {
		return "paths"
	}
	var ipe *openapi3.InvalidParameterInError
	if errors.As(err, &ipe) {
		return "paths"
	}

	// Cluster types with a Field that hints at the section.
	var rfe *openapi3.RequiredFieldError
	if errors.As(err, &rfe) {
		return sectionFromField(rfe.Field)
	}
	var fvm *openapi3.FieldVersionMismatchError
	if errors.As(err, &fvm) {
		return sectionFromField(fvm.Field)
	}

	// Schema-deep clusters: lean toward "paths" since most kin
	// validation surfaces from request/response schemas inside
	// operations. Inline component schemas miscount here, but the
	// section is a navigational hint, not a hard claim.
	var sve *openapi3.SchemaValueError
	if errors.As(err, &sve) {
		return "paths"
	}
	var sbf *openapi3.SchemaBothFormsExclusive
	if errors.As(err, &sbf) {
		return "paths"
	}
	return ""
}

// sectionFromField returns the section a kin Field name lives in,
// based on the field's top-level prefix. Anything not recognised
// returns empty.
func sectionFromField(field string) string {
	switch {
	case strings.HasPrefix(field, "info"):
		return "info"
	case strings.HasPrefix(field, "paths"):
		return "paths"
	case strings.HasPrefix(field, "components"):
		return "components"
	case strings.HasPrefix(field, "webhooks"):
		return "webhooks"
	case strings.HasPrefix(field, "servers"):
		return "servers"
	case strings.HasPrefix(field, "security"):
		return "security"
	case strings.HasPrefix(field, "tags"):
		return "tags"
	default:
		return ""
	}
}

// argsForKinError returns the disambiguating args used in fingerprint
// computation. For most validate clusters the args list is the
// cluster's structured Field (or Fields); for clusters that carry no
// per-finding field, the args are empty and identity is already
// captured by the rule ID + Source.
func argsForKinError(err error) []any {
	var rfe *openapi3.RequiredFieldError
	if errors.As(err, &rfe) {
		return []any{rfe.Field}
	}
	var fvm *openapi3.FieldVersionMismatchError
	if errors.As(err, &fvm) {
		return []any{fvm.Field}
	}
	var mef *openapi3.MutuallyExclusiveFieldsError
	if errors.As(err, &mef) {
		return []any{mef.Field1, mef.Field2}
	}
	var ffe *openapi3.ForbiddenFieldError
	if errors.As(err, &ffe) {
		return []any{ffe.Field}
	}
	var efr *openapi3.EitherFieldRequiredError
	if errors.As(err, &efr) {
		return []any{strings.Join(efr.Fields, "-or-")}
	}
	var sbf *openapi3.SchemaBothFormsExclusive
	if errors.As(err, &sbf) {
		return []any{sbf.Field}
	}
	var eofe *openapi3.ExactlyOneFieldError
	if errors.As(err, &eofe) {
		return []any{strings.Join(eofe.Fields, "-or-")}
	}
	var sec *openapi3.SingleEntryContentError
	if errors.As(err, &sec) {
		return []any{sec.Subject}
	}
	var sve *openapi3.SchemaValueError
	if errors.As(err, &sve) {
		return []any{sve.ValueKind}
	}
	var pre *openapi3.PathParameterRequiredError
	if errors.As(err, &pre) {
		return []any{pre.Param}
	}
	var doid *openapi3.DuplicateOperationIDError
	if errors.As(err, &doid) {
		return []any{doid.OperationID}
	}
	var esf *openapi3.ExtraSiblingFieldsError
	if errors.As(err, &esf) {
		return []any{strings.Join(esf.Fields, ",")}
	}
	var ste *openapi3.SchemaTypeError
	if errors.As(err, &ste) {
		return []any{ste.Type}
	}
	var ipe *openapi3.InvalidParameterInError
	if errors.As(err, &ipe) {
		return []any{ipe.Value}
	}
	var spre *openapi3.SchemaPatternRegexError
	if errors.As(err, &spre) {
		return []any{spre.Pattern}
	}
	return nil
}

// fingerprintFor produces a stable 12-char identifier for a finding.
// Format mirrors formatters/changes.go:computeFingerprint:
// sha256("{id}:{operation}:{path}:{args}") truncated to 12 hex chars.
// Inputs are structured rather than the rendered text so rendering
// changes (locale, message edits) don't invalidate stored fingerprints
// used for new/pre-existing/fixed partitioning across spec versions.
func fingerprintFor(f Finding, args []any) string {
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = fmt.Sprintf("%v", a)
	}
	h := fmt.Sprintf("%s:%s:%s:%s", f.Id, f.Operation, f.Path, strings.Join(parts, ";"))
	sum := sha256.Sum256([]byte(h))
	return hex.EncodeToString(sum[:])[:12]
}

// lineForKinError extracts the line number from the typed cluster
// errors' Origin. Returns 0 when origin metadata isn't available
// (untyped error, doc-root field, or loader.IncludeOrigin = false).
func lineForKinError(err error) int {
	if k := locationForKinError(err); k != nil {
		return k.Line
	}
	return 0
}

// columnForKinError extracts the column number from the typed cluster
// errors' Origin. Returns 0 when origin metadata isn't available.
func columnForKinError(err error) int {
	if k := locationForKinError(err); k != nil {
		return k.Column
	}
	return 0
}

// indentContinuation prefixes every non-empty continuation line of s
// with a tab. The first line is left as-is (the caller's format string
// already supplies its leading tab), and blank lines stay blank rather
// than becoming stray "\t" lines. Trailing whitespace is trimmed.
func indentContinuation(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if i == 0 || line == "" {
			continue
		}
		lines[i] = "\t" + line
	}
	return strings.TrimRight(strings.Join(lines, "\n"), " \t\n")
}

// locationForKinError returns the most-specific *Location available
// for a typed kin error. kin's Origin model:
//
//   - Origin.Key       points at the start of the enclosing collection
//     (e.g. for a LicenseIdentifierFieldFor31Plus, Key is the line of
//     the parent "license:" key, not "identifier:").
//   - Origin.Fields[X] points at the specific scalar field X inside
//     that collection.
//
// For clusters that carry a Field, we want Fields[Field] (the
// offending line) rather than Key (the enclosing object's line).
// Falls back to Key when the per-field entry is missing, and finally
// to nil for clusters with no Origin at all (WebhookNilError).
func locationForKinError(err error) *openapi3.Location {
	var rfe *openapi3.RequiredFieldError
	if errors.As(err, &rfe) && rfe.Origin != nil {
		return fieldLoc(rfe.Origin, rfe.Field)
	}
	var fvm *openapi3.FieldVersionMismatchError
	if errors.As(err, &fvm) && fvm.Origin != nil {
		return fieldLoc(fvm.Origin, fvm.Field)
	}
	var sve *openapi3.SchemaValueError
	if errors.As(err, &sve) && sve.Origin != nil {
		// SchemaValueError carries ValueKind (e.g. "example", "default")
		// — the per-field key under the schema where the value lives.
		return fieldLoc(sve.Origin, sve.ValueKind)
	}
	var ppe *openapi3.PathParametersError
	if errors.As(err, &ppe) && ppe.Origin != nil {
		return ppe.Origin.Key
	}
	var mef *openapi3.MutuallyExclusiveFieldsError
	if errors.As(err, &mef) && mef.Origin != nil {
		// Prefer Field1's location; either offender pins the finding to
		// the right object. We don't carry both since a single Source
		// is the contract.
		return fieldLoc(mef.Origin, mef.Field1)
	}
	var ffe *openapi3.ForbiddenFieldError
	if errors.As(err, &ffe) && ffe.Origin != nil {
		return fieldLoc(ffe.Origin, ffe.Field)
	}
	var sute *openapi3.ServerURLTemplateError
	if errors.As(err, &sute) && sute.Origin != nil {
		return sute.Origin.Key
	}
	var efr *openapi3.EitherFieldRequiredError
	if errors.As(err, &efr) && efr.Origin != nil {
		// EitherFieldRequiredError fires when none of {Fields...} are
		// present, so per-field lookup wouldn't match anything — the
		// enclosing object's Key is the right pin.
		return efr.Origin.Key
	}
	var sbf *openapi3.SchemaBothFormsExclusive
	if errors.As(err, &sbf) && sbf.Origin != nil {
		return fieldLoc(sbf.Origin, sbf.Field)
	}
	var eofe *openapi3.ExactlyOneFieldError
	if errors.As(err, &eofe) && eofe.Origin != nil {
		// Same reasoning as EitherFieldRequiredError: cluster fires
		// when the constraint is violated across multiple fields; the
		// object Key is the cleanest pin.
		return eofe.Origin.Key
	}
	var sec *openapi3.SingleEntryContentError
	if errors.As(err, &sec) && sec.Origin != nil {
		return fieldLoc(sec.Origin, sec.Subject)
	}
	var pre *openapi3.PathParameterRequiredError
	if errors.As(err, &pre) && pre.Origin != nil {
		// PathParameterRequiredError fires on a parameter declared with
		// in: path but without required: true. The Key of the parameter
		// object pins the finding correctly; the `required` field would
		// be more precise but is absent (that's the whole bug).
		return pre.Origin.Key
	}
	var ste *openapi3.SchemaTypeError
	if errors.As(err, &ste) && ste.Origin != nil {
		// SchemaTypeError fires on the offending `type:` field of a
		// schema. Pin to the type field if the Origin tracks it,
		// otherwise to the schema's Key.
		return fieldLoc(ste.Origin, "type")
	}
	var doid *openapi3.DuplicateOperationIDError
	if errors.As(err, &doid) && doid.Origin != nil {
		// Pin to the offending operationId scalar inside the second
		// operation (not the operation's start), since the duplicate
		// is the field value, not the surrounding block. Falls back
		// to the operation Key if the loader didn't track the field.
		return fieldLoc(doid.Origin, "operationId")
	}
	var esf *openapi3.ExtraSiblingFieldsError
	if errors.As(err, &esf) && esf.Origin != nil {
		// Origin points at the parent object carrying the unexpected
		// sibling fields. The fields themselves may not have Origin
		// entries (Yaml parser tracks structural keys, not the
		// offending ones), so the object Key is the best pin.
		return esf.Origin.Key
	}
	var ipe *openapi3.InvalidParameterInError
	if errors.As(err, &ipe) && ipe.Origin != nil {
		// Pin to the parameter's `in` field if the loader tracks it,
		// otherwise the parameter object's Key.
		return fieldLoc(ipe.Origin, "in")
	}
	var spre *openapi3.SchemaPatternRegexError
	if errors.As(err, &spre) && spre.Origin != nil {
		// Pin to the schema's `pattern` field where the bad regex
		// lives, otherwise the schema's Key.
		return fieldLoc(spre.Origin, "pattern")
	}
	// WebhookNilError carries no Origin (the offending key is on the
	// document root, which the loader doesn't track per-key).
	return nil
}

// fieldLoc returns the location of a specific scalar field inside an
// Origin's collection if present; otherwise the collection's Key.
// Lookup is by the leaf field name (e.g. "identifier" for license
// errors, "version" for info errors).
func fieldLoc(origin *openapi3.Origin, field string) *openapi3.Location {
	if origin == nil {
		return nil
	}
	if loc, ok := origin.Fields[field]; ok {
		return &loc
	}
	// Cluster errors carry dotted Field names (e.g. "info.version") for
	// disambiguation in the rule ID, but kin's Origin.Fields is keyed by
	// the leaf name as it appears in the YAML mapping ("version"). When
	// the full name doesn't match, fall back to the suffix after the
	// last dot so we still resolve to the precise field location instead
	// of the parent object's Key.
	if i := strings.LastIndex(field, "."); i >= 0 {
		if loc, ok := origin.Fields[field[i+1:]]; ok {
			return &loc
		}
	}
	return origin.Key
}

// ruleIDForKinError dispatches a kin-openapi error to a stable
// kebab-case rule ID. Uses the typed cluster wrappers from kin's
// openapi3 package; one arm per cluster covers all the call-site
// leaves wrapped by that cluster.
//
// kin errors not yet migrated to a cluster fall through to
// kinUnknownID. Where a cluster carries field-name metadata, the rule
// ID is derived from that — the per-leaf type isn't consulted because
// (a) the cluster carries enough metadata, and (b) deriving from a
// single field keeps the dispatch stable as kin adds new leaves.
// Where a cluster has no useful field for derivation
// (ServerURLTemplateError carries only the offending URL,
// PathParametersError carries Path+Method+Missing), a static rule ID
// is returned for the whole cluster.
func ruleIDForKinError(err error) string {
	var rfe *openapi3.RequiredFieldError
	if errors.As(err, &rfe) {
		return ruleIDFromField(rfe.Field) + "-required"
	}

	var fvm *openapi3.FieldVersionMismatchError
	if errors.As(err, &fvm) {
		return ruleIDFromField(fvm.Field) + "-field-for-3-1-plus"
	}

	var sve *openapi3.SchemaValueError
	if errors.As(err, &sve) {
		return ruleIDFromField(sve.ValueKind) + "-violates-schema"
	}

	var ppe *openapi3.PathParametersError
	if errors.As(err, &ppe) {
		return "path-parameters-mismatch"
	}

	var mef *openapi3.MutuallyExclusiveFieldsError
	if errors.As(err, &mef) {
		return ruleIDFromField(mef.Field1) + "-" + ruleIDFromField(mef.Field2) + "-mutually-exclusive"
	}

	var ffe *openapi3.ForbiddenFieldError
	if errors.As(err, &ffe) {
		return ruleIDFromField(ffe.Field) + "-forbidden"
	}

	var sute *openapi3.ServerURLTemplateError
	if errors.As(err, &sute) {
		return "server-url-template-invalid"
	}

	var efr *openapi3.EitherFieldRequiredError
	if errors.As(err, &efr) {
		return joinFieldsForRuleID(efr.Fields) + "-required"
	}

	var sbf *openapi3.SchemaBothFormsExclusive
	if errors.As(err, &sbf) {
		return ruleIDFromField(sbf.Field) + "-both-forms-exclusive"
	}

	var eofe *openapi3.ExactlyOneFieldError
	if errors.As(err, &eofe) {
		return joinFieldsForRuleID(eofe.Fields) + "-exactly-one"
	}

	var sec *openapi3.SingleEntryContentError
	if errors.As(err, &sec) {
		return ruleIDFromField(sec.Subject) + "-content-single-entry"
	}

	var wne *openapi3.WebhookNilError
	if errors.As(err, &wne) {
		return "webhook-nil"
	}

	var pre *openapi3.PathParameterRequiredError
	if errors.As(err, &pre) {
		return "path-parameter-required"
	}

	var doid *openapi3.DuplicateOperationIDError
	if errors.As(err, &doid) {
		return "duplicate-operation-id"
	}

	var esf *openapi3.ExtraSiblingFieldsError
	if errors.As(err, &esf) {
		return "extra-sibling-fields"
	}

	var ste *openapi3.SchemaTypeError
	if errors.As(err, &ste) {
		return "schema-type-unsupported"
	}

	var ipe *openapi3.InvalidParameterInError
	if errors.As(err, &ipe) {
		return "parameter-in-invalid"
	}

	var spre *openapi3.SchemaPatternRegexError
	if errors.As(err, &spre) {
		return "schema-pattern-regex-invalid"
	}

	return kinUnknownID
}

// joinFieldsForRuleID renders an N-field "any/exactly one of" rule ID
// fragment as kebab-case fields joined by "-or-" (e.g. ["value",
// "externalValue"] → "value-or-external-value"). The caller appends
// the cluster-specific suffix ("-required", "-exactly-one", ...).
func joinFieldsForRuleID(fields []string) string {
	parts := make([]string, len(fields))
	for i, f := range fields {
		parts[i] = ruleIDFromField(f)
	}
	return strings.Join(parts, "-or-")
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

func writeFindings(w io.Writer, findings []Finding, format string, colorMode checker.ColorMode) error {
	switch format {
	case formatYAML:
		bytes, err := yaml.Marshal(findings)
		if err != nil {
			return err
		}
		_, err = w.Write(bytes)
		return err
	case formatJSON:
		bytes, err := json.Marshal(findings)
		if err != nil {
			return err
		}
		_, err = w.Write(bytes)
		return err
	case formatText, "":
		writeFindingsText(w, findings, colorMode)
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
//
// When color is enabled, level is rendered red/purple/cyan via
// Level.StringCond and the rule ID is rendered yellow, matching
// changelog / breaking. The color decision honours the --color flag
// (auto/always/never) via checker.IsColorEnabled.
func writeFindingsText(w io.Writer, findings []Finding, colorMode checker.ColorMode) {
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

	useColor := checker.IsColorEnabled(colorMode)
	for _, f := range findings {
		loc := f.Source.File
		if f.Source.Line > 0 {
			loc = fmt.Sprintf("%s:%d:%d", f.Source.File, f.Source.Line, f.Source.Column)
		}
		id := f.Id
		if useColor {
			id = color.InYellow(f.Id)
		}
		// Some kin errors (notably *SchemaError) embed newlines in the
		// message — Schema:\n... + Value:\n... blocks. Indent every
		// non-empty continuation line so the finding stays visually
		// grouped, while leaving blank lines blank (not "\t").
		fmt.Fprintf(w, "%s\t[%s] at %s\n\t%s\n\n", f.Level.StringCond(colorMode), id, loc, indentContinuation(f.Text))
	}
}
