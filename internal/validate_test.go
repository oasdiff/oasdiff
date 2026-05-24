package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

// Happy path: a well-formed minimal spec produces no findings and exit 0.
func Test_ValidateCmd_NoFindings(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff validate ../data/validate/valid.yaml"), &stdout, io.Discard))
	require.Empty(t, stdout.String())
}

// Empty info.version → kin returns *RequiredFieldError{Field:"info.version"}
// wrapping *InfoVersionRequired. Our dispatch derives the rule ID from the
// cluster's Field: "info.version" → "info-version-required".
func Test_ValidateCmd_InfoVersionRequired(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate ../data/validate/missing-required-info.yaml"), &stdout, io.Discard))
	out := stdout.String()
	require.Contains(t, out, "info-version-required")
	require.Contains(t, out, "1 findings:")
}

// Empty doc.openapi → kin returns *RequiredFieldError{Field:"openapi"}.
// "openapi" → "openapi-required".
func Test_ValidateCmd_OpenAPIVersionRequired(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate ../data/validate/openapi-version-empty.yaml"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "openapi-required")
}

// 3.1-only license.identifier in a 3.0 doc → kin returns
// *FieldVersionMismatchError{Field:"identifier"} wrapping
// *LicenseIdentifierFieldFor31Plus. Rule ID: "identifier-field-for-3-1-plus".
//
// The fixture has `identifier: MIT` on line 7. This pin guards a real
// regression: until 2026-05-13 we used Origin.Key (the enclosing
// `license:` line, 5:3) instead of Origin.Fields["identifier"] (the
// offending line, 7:5).
func Test_ValidateCmd_FieldVersionMismatch_LicenseIdentifier(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate -f yaml --fail-on INFO ../data/validate/license-identifier-in-3-0.yaml"), &stdout, io.Discard))

	var findings []map[string]any
	require.NoError(t, yaml.Unmarshal(stdout.Bytes(), &findings))
	require.Len(t, findings, 1)
	require.Equal(t, "identifier-field-for-3-1-plus", findings[0]["id"])
	src := findings[0]["source"].(map[string]any)
	require.Equal(t, 7, src["line"], "should pin to the identifier: line, not the license: line")
	require.Equal(t, 5, src["column"])
}

// 3.1-only doc.webhooks in a 3.0 doc → kin returns
// *FieldVersionMismatchError{Field:"webhooks"}.
func Test_ValidateCmd_FieldVersionMismatch_Webhooks(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate --fail-on INFO ../data/validate/webhooks-in-3-0.yaml"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "webhooks-field-for-3-1-plus")
}

// Schema-keyword leaf reached through three layers of %w wrapping
// (paths → path → operation → schema). Pins that the typed dispatch
// works through arbitrarily nested error contexts — relies on kin
// using %w (not %v) at every wrap site.
func Test_ValidateCmd_FieldVersionMismatch_SchemaKeyword(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate --fail-on INFO ../data/validate/schema-3-1-keyword-in-3-0.yaml"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "const-field-for-3-1-plus")
}

// Mismatched-braces server URL → kin returns *ServerURLTemplateError
// wrapping *ServerURLMismatchedBraces. We dispatch the cluster to a
// single static rule ID for the whole template-error family rather
// than per-leaf, since the cluster carries only the offending URL and
// not a derivable field name.
func Test_ValidateCmd_ServerURLTemplate(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate ../data/validate/server-url-mismatched-braces.yaml"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "server-url-template-invalid")
}

// Path template declares variables not matched by parameters → kin
// returns *PathParametersError. Static rule ID for the whole cluster
// (the cluster carries Path + Method + Missing list, no single field
// to derive from).
func Test_ValidateCmd_PathParametersMismatch(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate ../data/openapi-test2.yaml"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "path-parameters-mismatch")
}

// An error originating inside an operation (here: a missing responses
// object) is wrapped by kin in SectionValidationError + PathValidationError
// + OperationValidationError (getkin/kin-openapi #1183). validate lifts
// those into the typed section / path / operation fields instead of
// leaving them buried in the message text.
func Test_ValidateCmd_PathOperationScope(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate -f yaml ../data/validate/operation-missing-responses.yaml"), &stdout, io.Discard))

	var findings []map[string]any
	require.NoError(t, yaml.Unmarshal(stdout.Bytes(), &findings))
	require.Len(t, findings, 1)
	require.Equal(t, "operation-responses-required", findings[0]["id"])
	require.Equal(t, "GET", findings[0]["operation"])
	require.Equal(t, "/things", findings[0]["path"])
	require.Equal(t, "paths", findings[0]["section"])
	// text carries only the leaf message — the section/path/operation
	// prefixes are stripped, since that scope is in the typed fields above.
	require.Equal(t, "value of responses must be an object", findings[0]["text"])
}

// YAML format produces a marshalled list of findings that round-trips
// through yaml.Unmarshal. Field names match changelog's: id, text, level,
// source (object with file/line/column), fingerprint.
func Test_ValidateCmd_YAMLFormat(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate -f yaml ../data/validate/missing-required-info.yaml"), &stdout, io.Discard))

	var findings []map[string]any
	require.NoError(t, yaml.Unmarshal(stdout.Bytes(), &findings))
	require.Len(t, findings, 1)
	require.Equal(t, "info-version-required", findings[0]["id"])
	src, ok := findings[0]["source"].(map[string]any)
	require.True(t, ok, "source should be an object")
	require.Equal(t, "../data/validate/missing-required-info.yaml", src["file"])
	require.Equal(t, 3, findings[0]["level"]) // checker.ERR
	require.Contains(t, findings[0]["text"], "version")
	require.Equal(t, "info", findings[0]["section"])
	require.NotEmpty(t, findings[0]["fingerprint"])
}

// JSON format mirrors the changelog --format json shape: top-level
// bare array, lowercase field names, source as an object, numeric level.
func Test_ValidateCmd_JSONFormat(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate -f json ../data/validate/missing-required-info.yaml"), &stdout, io.Discard))

	var findings []map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &findings))
	require.Len(t, findings, 1)
	require.Equal(t, "info-version-required", findings[0]["id"])
	require.Equal(t, float64(3), findings[0]["level"]) // json numbers decode to float64
	require.Equal(t, "info", findings[0]["section"])
	src, ok := findings[0]["source"].(map[string]any)
	require.True(t, ok, "source should be an object")
	require.Equal(t, "../data/validate/missing-required-info.yaml", src["file"])
	require.NotEmpty(t, findings[0]["fingerprint"])
}

// Fingerprint is deterministic: running validate twice on the same
// input produces the same fingerprint for the same finding. This is
// the property Pro PR-comment partitioning depends on for matching
// findings across base/revision spec versions.
func Test_ValidateCmd_FingerprintStable(t *testing.T) {
	var out1, out2 bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate -f json ../data/validate/missing-required-info.yaml"), &out1, io.Discard))
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate -f json ../data/validate/missing-required-info.yaml"), &out2, io.Discard))

	var f1, f2 []map[string]any
	require.NoError(t, json.Unmarshal(out1.Bytes(), &f1))
	require.NoError(t, json.Unmarshal(out2.Bytes(), &f2))
	require.Equal(t, f1[0]["fingerprint"], f2[0]["fingerprint"])
}

// githubactions format emits one CI annotation per finding so the
// oasdiff-action validate wrapper can surface violations inline on the
// PR's Files Changed tab. The annotation carries the rule ID as title
// and the finding's file/line/column for anchoring.
func Test_ValidateCmd_GitHubActionsFormat(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate -f githubactions ../data/validate/missing-required-info.yaml"), &stdout, io.Discard))
	out := stdout.String()
	require.Contains(t, out, "::error ")
	require.Contains(t, out, "title=info-version-required")
	require.Contains(t, out, "file=../data/validate/missing-required-info.yaml")
	require.Contains(t, out, "line=4")
	require.Contains(t, out, "col=3")
	require.Contains(t, out, "value of version must be a non-empty string")
}

// Severity: a 3.1-only field in an older doc is a portability warning, not a
// hard error. By default (--fail-on error) a warning-only spec passes (exit
// 0) while still printing the finding; --fail-on warning escalates the same
// finding to a failure (exit 1).
func Test_ValidateCmd_SeverityWarnAndFailOn(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 0, internal.Run(cmdToArgs("oasdiff validate -f yaml ../data/validate/webhooks-in-3-0.yaml"), &stdout, io.Discard))

	var findings []map[string]any
	require.NoError(t, yaml.Unmarshal(stdout.Bytes(), &findings))
	require.Len(t, findings, 1)
	require.Equal(t, "webhooks-field-for-3-1-plus", findings[0]["id"])
	require.Equal(t, 2, findings[0]["level"]) // checker.WARN

	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate --fail-on WARN ../data/validate/webhooks-in-3-0.yaml"), io.Discard, io.Discard))
}

// Severity: an example that violates its schema is doc-accuracy only, so it
// is info and doesn't fail the command by default.
func Test_ValidateCmd_SeverityInfo(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 0, internal.Run(cmdToArgs("oasdiff validate -f yaml ../data/validate/shared-schema-bad-example.yaml"), &stdout, io.Discard))

	var findings []map[string]any
	require.NoError(t, yaml.Unmarshal(stdout.Bytes(), &findings))
	require.Len(t, findings, 1)
	require.Equal(t, "example-violates-schema", findings[0]["id"])
	require.Equal(t, 1, findings[0]["level"]) // checker.INFO
}

// Text format: header summary line + changelog-style multi-line block.
// The block begins with "error\t[<id>]" so users grep'ing for severity
// can match across both validate and changelog output.
func Test_ValidateCmd_TextFormatHeaderAndBlock(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate ../data/validate/missing-required-info.yaml"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "1 findings: 1 error, 0 warning, 0 info")
	require.Contains(t, stdout.String(), "error\t[info-version-required] at ")
}

// Origin tracking: line + column appear inside the source object for
// typed errors carrying an *Origin (info, license, server, schema). The
// loader is started with IncludeOrigin=true unconditionally so this
// always fires for converted call sites.
func Test_ValidateCmd_LineColumnFromOrigin(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate -f yaml ../data/validate/missing-required-info.yaml"), &stdout, io.Discard))

	var findings []map[string]any
	require.NoError(t, yaml.Unmarshal(stdout.Bytes(), &findings))
	require.Len(t, findings, 1)
	require.Equal(t, "info-version-required", findings[0]["id"])
	src := findings[0]["source"].(map[string]any)
	// The version line is line 4 of the fixture; the Origin's per-field
	// location resolves to that exact line/column rather than the parent
	// info object's start.
	require.Equal(t, 4, src["line"])
	require.Equal(t, 3, src["column"])
}

// Document-root fields (openapi version, webhooks, jsonSchemaDialect) now
// carry source line/column via T.Origin (kin getkin/kin-openapi#1184). For
// scalar root fields the Origin resolves to the field key; the openapi
// version line is the first line of the fixture, so the finding points at
// line 1, column 1.
func Test_ValidateCmd_DocRootFieldHasLineColumn(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate -f yaml ../data/validate/openapi-version-empty.yaml"), &stdout, io.Discard))

	var findings []map[string]any
	require.NoError(t, yaml.Unmarshal(stdout.Bytes(), &findings))
	require.Len(t, findings, 1)
	require.Equal(t, "openapi-required", findings[0]["id"])
	src := findings[0]["source"].(map[string]any)
	require.Equal(t, 1, src["line"])
	require.Equal(t, 1, src["column"])
}

// Text format renders source as <file>:<line>:<column> when origin is
// available, and as plain <file> otherwise (matching the changelog
// command's location formatting in MultiLineError).
func Test_ValidateCmd_TextFormatLocation(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate ../data/validate/missing-required-info.yaml"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "../data/validate/missing-required-info.yaml:4:3")
}

// Multi-line error messages (e.g. kin's *SchemaError dumps Schema +
// Value blocks separated by newlines) render with every continuation
// line tab-indented, keeping the finding visually grouped. Also pins
// the SchemaValueError dispatch (parameter example violates schema)
// and Origin extraction (line/column from the parameter).
func Test_ValidateCmd_SchemaValueError(t *testing.T) {
	// data/openapi-test1.yaml's "token" parameter has example "26734...8"
	// (36 chars) but maxLength: 29. kin's SchemaError dumps Schema +
	// Value; SchemaValueError wraps it and supplies parameter.Origin.
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate --fail-on INFO ../data/openapi-test1.yaml"), &stdout, io.Discard))
	out := stdout.String()
	// Cluster dispatch: ValueKind "example" → "example-violates-schema".
	require.Contains(t, out, "[example-violates-schema]")
	// Origin from parameter.Origin produces a line:column suffix.
	require.Regexp(t, `at \S+:\d+:\d+`, out)
	// Multi-line content stays indented. This finding carries operation
	// context so the message body and its continuation lines indent at
	// "\t\t" (one level deeper than the "in API ..." line).
	require.Contains(t, out, "\n\t\tSchema:")
	require.Contains(t, out, "\n\t\tValue:")
}

// A defect in a shared components schema referenced from multiple
// operations is reported once, not once per reference. The components-
// rooted finding wins; the path-rooted cascade is dropped. Pins the
// dedupePreferringComponents behavior — without it, this fixture
// produces 4 identical findings (one from components/schemas/Status,
// three from each operation that $refs it).
func Test_ValidateCmd_DedupePreferringComponents(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate --fail-on INFO ../data/validate/shared-schema-bad-example.yaml"), &stdout, io.Discard))
	out := stdout.String()
	// Exactly one finding (not four).
	require.Contains(t, out, "1 findings:")
	// The surviving finding's location is the components-schema definition
	// (line 38 — the `example:` line inside components/schemas/Status),
	// not any of the operation-side $ref sites.
	require.Contains(t, out, "[example-violates-schema] at ../data/validate/shared-schema-bad-example.yaml:38:7")
	require.Contains(t, out, "invalid example: value must be a number")
}

// --color only affects the text output, so pairing it with a non-text
// format is rejected rather than silently ignored (matching changelog).
func Test_ValidateCmd_ColorWithNonTextFormatRejected(t *testing.T) {
	var stderr bytes.Buffer
	require.NotZero(t, internal.Run(cmdToArgs("oasdiff validate --color always -f json ../data/validate/valid.yaml"), io.Discard, &stderr))
	require.Contains(t, stderr.String(), "--color is only relevant with the 'text' format")
}

// Invalid file path → exit 102 (failed-to-load), not 1 (validation finding).
// Distinguishing these matters for CI: load failures and validation
// failures are different incidents.
func Test_ValidateCmd_LoadFailure(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 102, internal.Run(cmdToArgs("oasdiff validate ../data/validate/does-not-exist.yaml"), io.Discard, &stderr))
	require.Contains(t, stderr.String(), "failed to load")
}

// Path parameter declared with in: path but without required: true →
// kin returns *PathParameterRequiredError{Param:"id"} (added in
// getkin/kin-openapi #1187). Rule ID: "path-parameter-required".
// Origin pins the finding to the parameter object.
func Test_ValidateCmd_PathParameterRequired(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate ../data/validate/path-parameter-not-required.yaml"), &stdout, io.Discard))
	out := stdout.String()
	require.Contains(t, out, "[path-parameter-required]")
	require.Contains(t, out, `path parameter "id" must be required`)
	// Origin should resolve to the parameter's line:col, not file-only.
	require.Regexp(t, `at \S+:\d+:\d+`, out)
}

// Two operations sharing the same operationId across paths → kin
// returns *DuplicateOperationIDError (added in #1187). Rule ID:
// "duplicate-operation-id". Origin (also added in #1187) pins to the
// second operation's operationId field, not the operation block start.
func Test_ValidateCmd_DuplicateOperationID(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate ../data/validate/duplicate-operation-id.yaml"), &stdout, io.Discard))
	out := stdout.String()
	require.Contains(t, out, "[duplicate-operation-id]")
	require.Contains(t, out, `operations "GET /a" and "GET /b" have the same operation id "shared"`)
	require.Regexp(t, `at \S+:\d+:\d+`, out)
}

// Fields appearing as siblings of $ref where the spec disallows them →
// kin returns *ExtraSiblingFieldsError{Fields:[...]} (added in #1187).
// Rule ID: "extra-sibling-fields". Origin (also added in #1187) pins
// to the parent object holding the unexpected siblings.
func Test_ValidateCmd_ExtraSiblingFields(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate --fail-on INFO ../data/validate/extra-sibling-fields.yaml"), &stdout, io.Discard))
	out := stdout.String()
	require.Contains(t, out, "[extra-sibling-fields]")
	require.Contains(t, out, "extra sibling fields: [description]")
	require.Regexp(t, `at \S+:\d+:\d+`, out)
}

// Schema with an unsupported 'type' value (e.g. "foobar" instead of
// "string"/"integer"/...) → kin returns *SchemaTypeError{Type:"foobar"}
// (added in #1187). Rule ID: "schema-type-unsupported". Origin pins to
// the offending `type:` field.
func Test_ValidateCmd_SchemaTypeUnsupported(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate ../data/validate/schema-type-unsupported.yaml"), &stdout, io.Discard))
	out := stdout.String()
	require.Contains(t, out, "[schema-type-unsupported]")
	require.Contains(t, out, `unsupported 'type' value "foobar"`)
	require.Regexp(t, `at \S+:\d+:\d+`, out)
}

// Parameter with `in:` set to a non-3.x value (e.g. "body", a Swagger
// 2.0 leftover) → kin returns *InvalidParameterInError{Value:"body"}
// (added in #1187). Rule ID: "parameter-in-invalid". Origin pins to
// the parameter's `in` field.
func Test_ValidateCmd_InvalidParameterIn(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate ../data/validate/parameter-in-invalid.yaml"), &stdout, io.Discard))
	out := stdout.String()
	require.Contains(t, out, "[parameter-in-invalid]")
	require.Contains(t, out, `parameter can't have 'in' value "body"`)
	require.Regexp(t, `at \S+:\d+:\d+`, out)
}

// Schema with a Perl-only regex feature in `pattern:` (e.g. lookahead
// `(?!...)`) → kin's RE2 rejects compilation; cluster
// *SchemaPatternRegexError carries the offending Pattern and chains
// through to the legacy *SchemaError via Unwrap (added in #1187).
// Rule ID: "schema-pattern-regex-invalid". Origin pins to the
// schema's `pattern` field. The rendered message is byte-identical
// to the legacy *SchemaError's, so this test asserts the typed
// dispatch only, not a message change.
func Test_ValidateCmd_SchemaPatternRegexInvalid(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate ../data/validate/schema-pattern-regex-invalid.yaml"), &stdout, io.Discard))
	out := stdout.String()
	require.Contains(t, out, "[schema-pattern-regex-invalid]")
	require.Contains(t, out, "error parsing regexp")
	require.Regexp(t, `at \S+:\d+:\d+`, out)
}
