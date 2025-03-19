package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func cmdToArgs(cmd string) []string {
	return strings.Fields(cmd)
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func Test_InvalidCmd(t *testing.T) {
	require.Equal(t, 100, internal.Run(cmdToArgs("oasdiff invalid"), io.Discard, io.Discard))
}

func Test_NoRevision(t *testing.T) {
	require.Equal(t, 100, internal.Run(cmdToArgs("oasdiff diff base.yaml"), io.Discard, io.Discard))
}

func Test_ExtraRevision(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 100, internal.Run(cmdToArgs("oasdiff diff 1 2 3"), io.Discard, &stderr))
	require.Equal(t, "Error: invalid arguments after base and revision\n", stderr.String())
}

func Test_InvalidFlag(t *testing.T) {
	require.Equal(t, 100, internal.Run(cmdToArgs("oasdiff diff data/openapi-test1.yaml data/openapi-test1.yaml --invalid"), io.Discard, io.Discard))
}

func Test_BasicDiff(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff ../data/openapi-test1.yaml ../data/openapi-test3.yaml --exclude-elements endpoints"), &stdout, io.Discard))
	var bc interface{}
	require.Nil(t, yaml.Unmarshal(stdout.Bytes(), &bc))
}

func Test_DiffJson(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff ../data/openapi-test1.yaml ../data/openapi-test3.yaml -f json --exclude-elements endpoints"), &stdout, io.Discard))
	var bc interface{}
	require.Nil(t, json.Unmarshal(stdout.Bytes(), &bc))
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

func Test_Summary(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff summary ../data/openapi-test1.yaml ../data/openapi-test3.yaml"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), `diff: true`)
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

func Test_BreakingChangesIncludeChecks(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/run_test/breaking_changes_include_checks_base.yaml ../data/run_test/breaking_changes_include_checks_revision.yaml --include-checks response-non-success-status-removed,api-tag-removed --format json"), &stdout, io.Discard))
	bc := formatters.Changes{}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &bc))
	require.Len(t, bc, 2)
	for _, c := range bc {
		require.Equal(t, c.Level, checker.ERR)
	}
}

func Test_BasicBreakingChanges(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml"), io.Discard, io.Discard))
}

func Test_BreakingChangesYaml(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml --format yaml"), &stdout, io.Discard))
	var bc interface{}
	require.Nil(t, yaml.Unmarshal(stdout.Bytes(), &bc))
}

func Test_BreakingChangesJson(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml --format json"), &stdout, io.Discard))
	var bc interface{}
	require.Nil(t, json.Unmarshal(stdout.Bytes(), &bc))
}

func Test_BreakingChangesText(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml"), &stdout, io.Discard))
	var bc interface{}
	require.Error(t, json.Unmarshal(stdout.Bytes(), &bc))
	require.Error(t, yaml.Unmarshal(stdout.Bytes(), &bc))
}

func Test_BreakingChangesFailOnErr(t *testing.T) {
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml --fail-on ERR"), io.Discard, io.Discard))
}

func Test_BreakingChangesFailOnWarn(t *testing.T) {
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml --fail-on WARN"), io.Discard, io.Discard))
}

func Test_BreakingChangesFailOnWarnsErrsOnly(t *testing.T) {
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test2.yaml ../data/openapi-test4.yaml --fail-on WARN"), io.Discard, io.Discard))
}

func Test_BreakingChangesFailOnErrNoDiff(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test1.yaml --fail-on ERR"), io.Discard, io.Discard))
}

func Test_BreakingChangesFailOnWarnsNoDiff(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test1.yaml --fail-on WARN"), io.Discard, io.Discard))
}

func Test_ChangelogFailOnInfo(t *testing.T) {
	require.Equal(t, 1, internal.Run(cmdToArgs("oasdiff changelog ../data/simple3.yaml ../data/simple4.yaml --fail-on INFO"), io.Discard, io.Discard))
}

func Test_ChangelogFailOnWarnInfosOnly(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff changelog ../data/simple3.yaml ../data/simple4.yaml --fail-on WARN"), io.Discard, io.Discard))
}

func Test_BreakingChangesIgnoreErrs(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml --err-ignore ../data/ignore-err-example.txt --format json"), &stdout, io.Discard))
	bc := formatters.Changes{}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &bc))
	require.Len(t, bc, 5)
}

func Test_BreakingChangesIgnoreErrsAndWarns(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml --err-ignore ../data/ignore-err-example.txt --warn-ignore ../data/ignore-warn-example.txt --format json"), &stdout, io.Discard))
	bc := formatters.Changes{}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &bc))
	require.Len(t, bc, 4)
}

func Test_BreakingChangesIgnoreErrsApiSchemaOptional(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml --err-ignore ../data/ignore-err-example.txt --warn-ignore ../data/ignore-warn-example.txt --include-checks api-schema-removed --format json"), &stdout, io.Discard))
	bc := formatters.Changes{}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &bc))
	require.Len(t, bc, 4)
}

func Test_BreakingChangesInvalidIgnoreFile(t *testing.T) {
	require.Equal(t, 121, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml --err-ignore no-file"), io.Discard, io.Discard))
}

func Test_ComposedMode(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff ../data/composed/base/*.yaml ../data/composed/revision/*.yaml --composed --exclude-elements endpoints,extensions"), &stdout, io.Discard))
	var bc interface{}
	require.NoError(t, yaml.Unmarshal(stdout.Bytes(), &bc))
	require.Equal(t, map[string]interface{}{"paths": map[string]interface{}{"deleted": []interface{}{"/api/old-test"}}}, bc)
}

func Test_ComposedModeStdin(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 100, internal.Run(cmdToArgs("oasdiff diff - - --composed"), io.Discard, &stderr))
	require.Equal(t, "Error: can't read from stdin in composed mode\n", stderr.String())
}

func Test_Help(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff --help"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "Usage:")
}

func Test_HelpShortcut(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff -h"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "Usage:")
}

func Test_Version(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff -v"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "oasdiff version")
}

func Test_StripPrefixBase(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/simple.yaml ../data/simple.yaml --strip-prefix-base /partner-api"), io.Discard, io.Discard))
}

func Test_DuplicatePathsFail(t *testing.T) {
	require.NotZero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/duplicate_endpoints/base.yaml ../data/duplicate_endpoints/revision.yaml"), io.Discard, io.Discard))
}

func Test_DuplicatePathsOK(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/duplicate_endpoints/base.yaml ../data/duplicate_endpoints/revision.yaml --include-path-params"), io.Discard, io.Discard))
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
	require.Len(t, cl, 21)
	require.Equal(t, map[string]interface{}{"x-beta": true, "x-extension-test": interface{}(nil)}, cl[12].Attributes)
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

func Test_BreakingChangesChangelogOptionalCheckersAreErrorLevelWhenSpecified(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff changelog ../data/run_test/changelog_include_checks_base.yaml ../data/run_test/changelog_include_checks_revision.yaml --format json --include-checks api-tag-removed,response-non-success-status-removed"), &stdout, io.Discard))
	cl := formatters.Changes{}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &cl))
	require.Len(t, cl, 2)
	for _, c := range cl {
		require.Equal(t, c.Level, checker.ERR)
	}
}

func Test_BreakingChangesFlattenDeprecated(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/allof/simple.yaml ../data/allof/revision.yaml --flatten --fail-on ERR"), io.Discard, io.Discard))
}

func Test_BreakingChangesFlattenAllOf(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/allof/simple.yaml ../data/allof/revision.yaml --flatten-allof --fail-on ERR"), io.Discard, io.Discard))
}

func Test_BreakingChangesInvalidDeprecationDays(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 100, internal.Run(cmdToArgs("oasdiff breaking ../data/deprecation/base.yaml ../data/deprecation/deprecated-with-sunset.yaml --deprecation-days-stable=-1"), io.Discard, &stderr))
	require.Equal(t, "Error: invalid argument \"-1\" for \"--deprecation-days-stable\" flag: strconv.ParseUint: parsing \"-1\": invalid syntax\n", stderr.String())
}

func Test_BreakingChangesFlattenCommonParams(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/common-params/params_in_path.yaml ../data/common-params/params_in_op.yaml --flatten-params --fail-on ERR"), io.Discard, io.Discard))
}

func Test_BreakingChangesCaseInsensitiveHeaders(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff ../data/header-case/base.yaml ../data/header-case/revision.yaml --case-insensitive-headers --fail-on-diff"), io.Discard, io.Discard))
}

func Test_FlattenCmdOK(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff flatten ../data/allof/simple.yaml"), io.Discard, io.Discard))
}

func Test_FlattenCmdInvalid(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 102, internal.Run(cmdToArgs("oasdiff flatten ../data/allof/invalid.yaml"), io.Discard, &stderr))
	require.Equal(t, `Error: failed to load original spec from "../data/allof/invalid.yaml": failed to flatten allOf in "../data/allof/invalid.yaml": unable to resolve Type conflict: all Type values must be identical
`, stderr.String())
}

func Test_Checks(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks -l ru --tags decrease,parameters --severity info,warn,error"), io.Discard, io.Discard))
}

func Test_Color(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/allof/simple.yaml ../data/allof/revision.yaml --color always"), io.Discard, io.Discard))
}

func Test_ColorWithNonTextFormat(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 100, internal.Run(cmdToArgs("oasdiff changelog ../data/allof/simple.yaml ../data/allof/revision.yaml -f yaml --color always"), io.Discard, &stderr))
	require.Equal(t, "Error: --color flag is only relevant with 'text' or 'singleline' formats\n", stderr.String())
}

func Test_QR(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff qr"), io.Discard, io.Discard))
}

func Test_InvalidEnumValue(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 100, internal.Run(cmdToArgs("oasdiff diff ../data/openapi-test1.yaml ../data/openapi-test3.yaml --exclude-elements xxx"), io.Discard, &stderr))
	require.Equal(t, `Error: invalid argument "xxx" for "-e, --exclude-elements" flag: xxx is not one of the allowed values: description, endpoints, examples, extensions, summary, or title
`, stderr.String())
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

// Test_JsonWithExcludeElements replicates https://github.com/oasdiff/oasdiff/issues/674
func Test_JsonWithExcludeElements(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff --format json --exclude-elements=description,title,summary ../data/description/spec1.yml ../data/description/spec2.yml --fail-on-diff"), io.Discard, io.Discard))
}
