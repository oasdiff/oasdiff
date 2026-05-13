package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
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
func Test_ValidateCmd_FieldVersionMismatch_LicenseIdentifier(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate ../data/validate/license-identifier-in-3-0.yaml"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "identifier-field-for-3-1-plus")
}

// 3.1-only doc.webhooks in a 3.0 doc → kin returns
// *FieldVersionMismatchError{Field:"webhooks"}.
func Test_ValidateCmd_FieldVersionMismatch_Webhooks(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate ../data/validate/webhooks-in-3-0.yaml"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "webhooks-field-for-3-1-plus")
}

// Schema-keyword leaf reached through three layers of %w wrapping
// (paths → path → operation → schema). Pins that the typed dispatch
// works through arbitrarily nested error contexts — relies on kin
// using %w (not %v) at every wrap site.
func Test_ValidateCmd_FieldVersionMismatch_SchemaKeyword(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate ../data/validate/schema-3-1-keyword-in-3-0.yaml"), &stdout, io.Discard))
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

// YAML format produces a marshalled list of Finding records that round-
// trips through yaml.Unmarshal. Field names match changelog's:
// id, text, level, source (object with file/line/column), fingerprint.
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

// Text format: header summary line + changelog-style multi-line block.
// The block begins with "error\t[<id>]" so users grep'ing for severity
// can match across both validate and changelog output.
func Test_ValidateCmd_TextFormatHeaderAndBlock(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate ../data/validate/missing-required-info.yaml"), &stdout, io.Discard))
	lines := strings.Split(strings.TrimRight(stdout.String(), "\n"), "\n")
	require.Equal(t, "1 findings: 1 error, 0 warning, 0 info", lines[0])
	require.Contains(t, lines[1], "error\t[info-version-required] at ")
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
	// info: starts at line 2 of the fixture; the cluster Origin points
	// at the info object's start.
	require.Equal(t, 2, src["line"])
	require.Equal(t, 1, src["column"])
}

// Document-root fields (openapi version, webhooks, jsonSchemaDialect)
// have no Origin since the loader doesn't track *T. Their findings
// have source.line / source.column omitted (yaml:"...,omitempty" on
// Source's Line/Column elides them entirely).
func Test_ValidateCmd_DocRootFieldHasNoLineColumn(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate -f yaml ../data/validate/openapi-version-empty.yaml"), &stdout, io.Discard))

	var findings []map[string]any
	require.NoError(t, yaml.Unmarshal(stdout.Bytes(), &findings))
	require.Len(t, findings, 1)
	require.Equal(t, "openapi-required", findings[0]["id"])
	src := findings[0]["source"].(map[string]any)
	require.NotContains(t, src, "line")
	require.NotContains(t, src, "column")
}

// Text format renders source as <file>:<line>:<column> when origin is
// available, and as plain <file> otherwise (matching the changelog
// command's location formatting in MultiLineError).
func Test_ValidateCmd_TextFormatLocation(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate ../data/validate/missing-required-info.yaml"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "../data/validate/missing-required-info.yaml:2:1")
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
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate ../data/openapi-test1.yaml"), &stdout, io.Discard))
	out := stdout.String()
	// Cluster dispatch: ValueKind "example" → "example-violates-schema".
	require.Contains(t, out, "[example-violates-schema]")
	// Origin from parameter.Origin produces a line:column suffix.
	require.Regexp(t, `at \S+:\d+:\d+`, out)
	// Multi-line content stays indented.
	require.Contains(t, out, "\n\tSchema:")
	require.Contains(t, out, "\n\tValue:")
}

// Invalid file path → exit 102 (failed-to-load), not 1 (validation finding).
// Distinguishing these matters for CI: load failures and validation
// failures are different incidents.
func Test_ValidateCmd_LoadFailure(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 102, internal.Run(cmdToArgs("oasdiff validate ../data/validate/does-not-exist.yaml"), io.Discard, &stderr))
	require.Contains(t, stderr.String(), "failed to load")
}
