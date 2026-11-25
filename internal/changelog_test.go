package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

func Test_ChangelogFailOnInfo(t *testing.T) {
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff changelog ../data/simple3.yaml ../data/simple4.yaml --fail-on INFO"), io.Discard, io.Discard))
}

func Test_ChangelogFailOnWarnInfosOnly(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff changelog ../data/simple3.yaml ../data/simple4.yaml --fail-on WARN"), io.Discard, io.Discard))
}

func Test_Changelog(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff changelog ../data/run_test/changelog_base.yaml ../data/run_test/changelog_revision.yaml --format json"), &stdout, io.Discard))
	cl := formatters.Changes{}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &cl))
	require.Len(t, cl, 1)
}

func Test_ChangelogWithAttributes(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff changelog ../data/openapi-test1.yaml ../data/openapi-test3.yaml --attributes x-beta,x-extension-test -f yaml"), &stdout, io.Discard))
	cl := formatters.Changes{}
	require.NoError(t, yaml.Unmarshal(stdout.Bytes(), &cl))
	require.Len(t, cl, 26)
	require.Equal(t, map[string]any{"x-beta": true, "x-extension-test": any(nil)}, cl[15].Attributes)
}

func Test_BreakingChangesChangelogOptionalCheckersAreInfoLevel(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff changelog ../data/run_test/changelog_include_checks_base.yaml ../data/run_test/changelog_include_checks_revision.yaml --format json"), &stdout, io.Discard))
	cl := formatters.Changes{}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &cl))
	require.Len(t, cl, 2)
	for _, c := range cl {
		require.Equal(t, c.Level, checker.INFO)
	}
}

// The checks compare header names ignoring case, since HTTP header names are
// case-insensitive: a header renamed only in case is not a change.
func Test_ChecksIgnoreHeaderCase(t *testing.T) {
	for _, command := range []string{"breaking", "changelog"} {
		var stdout bytes.Buffer
		require.Zero(t, internal.Run(cmdToArgs("oasdiff "+command+" ../data/header-case/base.yaml ../data/header-case/revision.yaml --fail-on WARN"), &stdout, io.Discard), command)
		require.NotContains(t, strings.ToLower(stdout.String()), "x-case", command)
	}
}

// The checks print a header name as written, with the flag or without it.
func Test_ChecksPrintHeaderNamesAsWritten(t *testing.T) {
	for _, flag := range []string{"", " --case-insensitive-headers"} {
		var stdout bytes.Buffer
		require.Zero(t, internal.Run(cmdToArgs("oasdiff changelog ../data/header-case/base.yaml ../data/header-case/required.yaml"+flag), &stdout, io.Discard), flag)
		require.Contains(t, stdout.String(), "X-Case", flag)
	}
}

func Test_ColorWithNonTextFormat(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 100, internal.Run(cmdToArgs("oasdiff changelog ../data/allof/simple.yaml ../data/allof/revision.yaml -f yaml --color always"), io.Discard, &stderr))
	require.Equal(t, "Error: --color flag is only relevant with 'text' or 'singleline' formats\n", stderr.String())
}

func Test_CustomSeverityLevels(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff changelog ../data/openapi-test1.yaml ../data/openapi-test3.yaml --severity-levels ../data/severity-levels.txt"), io.Discard, io.Discard))
}

func Test_CustomSeverityLevelsInvalidFile(t *testing.T) {
	require.Equal(t, 106, internal.Run(cmdToArgs("oasdiff changelog ../data/openapi-test1.yaml ../data/openapi-test3.yaml --severity-levels ../data/invalid.txt"), io.Discard, io.Discard))
}

func Test_Changelog_WithoutMatchPath(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff changelog ../data/path-filter/base.yaml ../data/path-filter/revision.yaml --format json"), &stdout, io.Discard))
	bc := formatters.Changes{}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &bc))
	require.Len(t, bc, 2)
}

func Test_Changelog_WithMatchPath(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff changelog ../data/path-filter/base.yaml ../data/path-filter/revision.yaml --format json -p a"), &stdout, io.Discard))
	bc := formatters.Changes{}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &bc))
	require.Len(t, bc, 1)
}

func Test_Changelog_WithUnmatchPath(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff changelog ../data/path-filter/base.yaml ../data/path-filter/revision.yaml --format json -q a"), &stdout, io.Discard))
	bc := formatters.Changes{}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &bc))
	require.Len(t, bc, 1)
}

func Test_Changelog_WithCustomTemplate(t *testing.T) {
	// Create a custom template file
	customTemplate := `### API Changes {{ .GetVersionTitle }}
{{ range $endpoint, $changes := .APIChanges }}
#### {{ $endpoint.Operation }} {{ $endpoint.Path }}
{{ range $changes }}* {{ if .IsBreaking }}[BREAKING] {{ end }}{{ .Text }}
{{ end }}
{{ end }}`

	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "custom-template.md")
	err := os.WriteFile(templatePath, []byte(customTemplate), 0644)
	require.NoError(t, err)

	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff changelog ../data/run_test/changelog_base.yaml ../data/run_test/changelog_revision.yaml --format markdown --template "+templatePath), &stdout, io.Discard))

	result := stdout.String()
	require.Contains(t, result, "### API Changes")
	require.Contains(t, result, "#### ")
	require.Contains(t, result, "* ")
}

func Test_Changelog_WithInvalidCustomTemplate(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 105, internal.Run(cmdToArgs("oasdiff changelog ../data/run_test/changelog_base.yaml ../data/run_test/changelog_revision.yaml --format markdown --template /nonexistent/template.md"), io.Discard, &stderr))
	require.Contains(t, stderr.String(), "failed to load custom template")
}

// changelogIds runs changelog on the excluded_elements fixtures with args
// appended, and returns the ids it reports.
func changelogIds(t *testing.T, args string) []string {
	t.Helper()
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff changelog ../data/checker/excluded_elements_base.yaml ../data/checker/excluded_elements_revision.yaml --format json"+args), &stdout, io.Discard), args)
	changes := formatters.Changes{}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &changes))
	ids := []string{}
	for _, change := range changes {
		ids = append(ids, change.Id)
	}
	return ids
}

var excludedElementsIds = []string{
	"api-stability-decreased",
	"api-sunset-date-changed-earlier",
	"request-property-x-extensible-enum-value-removed",
	"api-tag-added",
	"api-tag-removed",
}

// Excluding an extension from the diff does not turn off the checks that read
// it.
func Test_ChecksIgnoreExcludeExtensions(t *testing.T) {
	for _, exclude := range []string{
		"",
		" --exclude-extensions x-stability-level,x-sunset,x-extensible-enum",
		" --exclude-extensions x-internal,x-stability-level,x-sunset,x-extensible-enum",
	} {
		require.ElementsMatch(t, excludedElementsIds, changelogIds(t, exclude), exclude)
	}
}

// exclude-elements is an option of diff and summary. In a configuration file
// it leaves the checks as they are.
func Test_ChecksIgnoreExcludeElementsInConfigFile(t *testing.T) {
	config := filepath.Join(t.TempDir(), "oasdiff.yaml")
	require.NoError(t, os.WriteFile(config, []byte("exclude-elements:\n  - extensions\n  - tags\nexclude-extensions:\n  - x-sunset\n"), 0o600))

	require.ElementsMatch(t, excludedElementsIds, changelogIds(t, " --config "+config))

	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff ../data/checker/excluded_elements_base.yaml ../data/checker/excluded_elements_revision.yaml --config "+config), &stdout, io.Discard))
	require.NotContains(t, stdout.String(), "x-stability-level")
	require.NotContains(t, stdout.String(), "billing")
}

// --flatten-params is ignored by the checks, so an operation header still
// overrides the path-level one whose name differs only in case.
func Test_ChecksIgnoreFlattenParams(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff changelog ../data/header-case/override-base.yaml ../data/header-case/override-revision.yaml --flatten-params"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "request-parameter-became-required")
	require.NotContains(t, stdout.String(), "request-parameter-removed")
}
