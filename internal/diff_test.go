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

func Test_BasicDiff(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff ../data/openapi-test1.yaml ../data/openapi-test3.yaml --exclude-elements endpoints"), &stdout, io.Discard))
	var bc any
	require.Nil(t, yaml.Unmarshal(stdout.Bytes(), &bc))
}

func Test_DiffJson(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff ../data/openapi-test1.yaml ../data/openapi-test3.yaml -f json --exclude-elements endpoints"), &stdout, io.Discard))
	var bc any
	require.Nil(t, json.Unmarshal(stdout.Bytes(), &bc))
}

// Identical specs produce an empty diff. json/yaml must still emit a valid
// empty document (`{}`), not empty bytes — an empty string is not valid JSON.
func Test_DiffEmptyJson(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff ../data/openapi-test1.yaml ../data/openapi-test1.yaml -f json"), &stdout, io.Discard))
	require.Equal(t, "{}", strings.TrimSpace(stdout.String()))
	var bc any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &bc))
}

func Test_DiffEmptyYaml(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff ../data/openapi-test1.yaml ../data/openapi-test1.yaml -f yaml"), &stdout, io.Discard))
	require.Equal(t, "{}", strings.TrimSpace(stdout.String()))
	var bc any
	require.NoError(t, yaml.Unmarshal(stdout.Bytes(), &bc))
}

func Test_DiffHtml(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff ../data/openapi-test1.yaml ../data/openapi-test3.yaml -f html"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), `<h3 id="new-endpoints-none">New Endpoints: None</h3>`)
}

func Test_DiffText(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff ../data/openapi-test1.yaml ../data/openapi-test3.yaml -f text"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), `### New Endpoints: None`)
}

func Test_InvalidGlob(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 103, internal.Run(cmdToArgs(`oasdiff diff -c "a[" ../data/openapi-test3.yaml`), io.Discard, &stderr))
	require.Equal(t, "Error: failed to load base specs from glob \"\\\"a[\\\"\": syntax error in pattern\n", stderr.String())
}

func Test_GlobNoFiles(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 103, internal.Run(cmdToArgs("oasdiff diff -c no-file ../data/openapi-test3.yaml"), io.Discard, &stderr))
	require.Equal(t, "Error: failed to load base specs from glob \"no-file\": no matching files\n", stderr.String())
}

func Test_GlobWithUrl(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 103, internal.Run(cmdToArgs("oasdiff diff -c ../data/openapi-test1.yaml https://"), io.Discard, &stderr))
	require.Equal(t, "Error: failed to load revision specs from glob \"https://\": no matching files (should be a glob, not a URL)\n", stderr.String())
}

func Test_DiffInvalidFormat(t *testing.T) {
	require.Equal(t, 100, internal.Run(cmdToArgs("oasdiff diff ../data/openapi-test1.yaml ../data/openapi-test3.yaml --format xxx"), io.Discard, io.Discard))
}

func Test_ComposedMode(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff ../data/composed/base/*.yaml ../data/composed/revision/*.yaml --composed --exclude-elements endpoints,extensions"), &stdout, io.Discard))
	var bc any
	require.NoError(t, yaml.Unmarshal(stdout.Bytes(), &bc))
	require.Equal(t, map[string]any{"paths": map[string]any{"deleted": []any{"/api/old-test"}}}, bc)
}

func Test_ComposedModeStdin(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 100, internal.Run(cmdToArgs("oasdiff diff - - --composed"), io.Discard, &stderr))
	require.Equal(t, "Error: can't read from stdin in composed mode\n", stderr.String())
}

// diff and summary report the documents as written, so a header renamed only
// in case is a deleted and an added header.
func Test_DiffReportsHeaderCase(t *testing.T) {
	var stdout bytes.Buffer
	require.NotZero(t, internal.Run(cmdToArgs("oasdiff diff ../data/header-case/base.yaml ../data/header-case/revision.yaml --fail-on-diff"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "X-Case")
	require.Contains(t, stdout.String(), "x-case")
	require.Contains(t, stdout.String(), "X-RateLimit-Limit")

	stdout.Reset()
	require.Zero(t, internal.Run(cmdToArgs("oasdiff summary ../data/header-case/base.yaml ../data/header-case/revision.yaml"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "diff: true")
}

// --case-insensitive-headers lowercases header names before the diff
func Test_DiffCaseInsensitiveHeaders(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff ../data/header-case/base.yaml ../data/header-case/revision.yaml --case-insensitive-headers --fail-on-diff"), io.Discard, io.Discard))

	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff summary ../data/header-case/base.yaml ../data/header-case/revision.yaml --case-insensitive-headers"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "diff: false")
}

// Test_JsonWithExcludeElements replicates https://github.com/oasdiff/oasdiff/issues/674
func Test_JsonWithExcludeElements(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff --format json --exclude-elements=description,title,summary ../data/description/spec1.yml ../data/description/spec2.yml --fail-on-diff"), io.Discard, io.Discard))
}

func Test_ExcludeServers(t *testing.T) {
	const specs = "../data/servers/exclude-base.yaml ../data/servers/exclude-revision.yaml"

	var stdout bytes.Buffer
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff diff "+specs+" --fail-on-diff"), &stdout, io.Discard))
	for _, url := range []string{"https://www.example.com", "https://www.example.com/users", "https://www.example.com/users/list", "https://www.example.com/next"} {
		require.Contains(t, stdout.String(), url+"\n")
	}

	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff "+specs+" --exclude-elements servers --fail-on-diff"), io.Discard, io.Discard))
}

// Test_ExcludeExtensions tests the --exclude-extensions flag (issue #765)
func Test_ExcludeExtensions(t *testing.T) {
	// With --exclude-extensions flag, the x-amazon-apigateway-integration changes should be excluded
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff ../data/extensions/base.yaml ../data/extensions/revision.yaml --exclude-extensions x-amazon-apigateway-integration --fail-on-diff"), io.Discard, io.Discard))
}

// Test_ExcludeExtensionsMultiple tests excluding multiple extension names
func Test_ExcludeExtensionsMultiple(t *testing.T) {
	// Should work with comma-separated values
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff ../data/extensions/base.yaml ../data/extensions/revision.yaml --exclude-extensions x-amazon-apigateway-integration,x-other-extension --fail-on-diff"), io.Discard, io.Discard))
}

func Test_EmptyComponentsOmitted(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff ../data/version/base.yaml ../data/version/revision.yaml -f json"), &stdout, io.Discard))
	require.Equal(t, "{\"info\":{\"version\":{\"from\":\"0.0.0\",\"to\":\"0.0.1\"}}}\n", stdout.String())
}
