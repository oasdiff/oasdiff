package internal_test

import (
	"bytes"
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

// kin errors not yet migrated to a typed cluster fall through to the
// generic kin-validation-error catchall. The mismatched-braces server
// URL check is one such site (kin still returns a plain errors.New
// for it). When kin migrates that site to a typed leaf, this test
// flips and gets a more specific rule ID.
func Test_ValidateCmd_UntypedKinErrorFallback(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate ../data/validate/server-url-mismatched-braces.yaml"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "kin-validation-error")
}

// YAML format produces a marshalled list of Finding records that round-
// trips through yaml.Unmarshal. Field names match changelog's:
// id, text, level, source.
func Test_ValidateCmd_YAMLFormat(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate -f yaml ../data/validate/missing-required-info.yaml"), &stdout, io.Discard))

	var findings []map[string]any
	require.NoError(t, yaml.Unmarshal(stdout.Bytes(), &findings))
	require.Len(t, findings, 1)
	require.Equal(t, "info-version-required", findings[0]["id"])
	require.Equal(t, "../data/validate/missing-required-info.yaml", findings[0]["source"])
	require.Equal(t, 3, findings[0]["level"]) // checker.ERR
	require.Contains(t, findings[0]["text"], "version")
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

// Origin tracking: line + column appear in YAML output for typed
// errors carrying an *Origin (info, license, server, schema). The
// loader is started with IncludeOrigin=true unconditionally so this
// always fires for converted call sites.
func Test_ValidateCmd_LineColumnFromOrigin(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate -f yaml ../data/validate/missing-required-info.yaml"), &stdout, io.Discard))

	var findings []map[string]any
	require.NoError(t, yaml.Unmarshal(stdout.Bytes(), &findings))
	require.Len(t, findings, 1)
	require.Equal(t, "info-version-required", findings[0]["id"])
	// info: starts at line 2 of the fixture; the cluster Origin points
	// at the info object's start.
	require.Equal(t, 2, findings[0]["line"])
	require.Equal(t, 1, findings[0]["column"])
}

// Document-root fields (openapi version, webhooks, jsonSchemaDialect)
// have no Origin since the loader doesn't track *T. Their findings
// have line/column omitted (not zero — yaml:"...,omitempty" elides
// them entirely).
func Test_ValidateCmd_DocRootFieldHasNoLineColumn(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate -f yaml ../data/validate/openapi-version-empty.yaml"), &stdout, io.Discard))

	var findings []map[string]any
	require.NoError(t, yaml.Unmarshal(stdout.Bytes(), &findings))
	require.Len(t, findings, 1)
	require.Equal(t, "openapi-required", findings[0]["id"])
	require.NotContains(t, findings[0], "line")
	require.NotContains(t, findings[0], "column")
}

// Text format renders source as <file>:<line>:<column> when origin is
// available, and as plain <file> otherwise (matching the changelog
// command's location formatting in MultiLineError).
func Test_ValidateCmd_TextFormatLocation(t *testing.T) {
	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff validate ../data/validate/missing-required-info.yaml"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "../data/validate/missing-required-info.yaml:2:1")
}

// Invalid file path → exit 102 (failed-to-load), not 1 (validation finding).
// Distinguishing these matters for CI: load failures and validation
// failures are different incidents.
func Test_ValidateCmd_LoadFailure(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 102, internal.Run(cmdToArgs("oasdiff validate ../data/validate/does-not-exist.yaml"), io.Discard, &stderr))
	require.Contains(t, stderr.String(), "failed to load")
}
