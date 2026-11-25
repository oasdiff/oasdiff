package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

// --include-checks is accepted for back-compat but ignored (the optional-checks
// mechanism was retired). The named checks stay informational, so breaking
// reports nothing and the run does not error on the flag.
func Test_IncludeChecksAcceptedButIgnored(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/run_test/breaking_changes_include_checks_base.yaml ../data/run_test/breaking_changes_include_checks_revision.yaml --include-checks response-non-success-status-removed,api-tag-removed --format json"), &stdout, io.Discard))
	bc := formatters.Changes{}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &bc))
	require.Empty(t, bc)
}

func Test_BasicBreakingChanges(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml"), io.Discard, io.Discard))
}

func Test_BreakingChangesYaml(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml --format yaml"), &stdout, io.Discard))
	var bc any
	require.Nil(t, yaml.Unmarshal(stdout.Bytes(), &bc))
}

func Test_BreakingChangesJson(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml --format json"), &stdout, io.Discard))
	var bc any
	require.Nil(t, json.Unmarshal(stdout.Bytes(), &bc))
}

func Test_BreakingChangesText(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml"), &stdout, io.Discard))
	var bc any
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

func Test_BreakingChangesIgnoreErrs(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml --err-ignore ../data/ignore-err-example.txt --format json"), &stdout, io.Discard))
	bc := formatters.Changes{}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &bc))
	require.Len(t, bc, 7)
}

func Test_BreakingChangesIgnoreErrsAndWarns(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml --err-ignore ../data/ignore-err-example.txt --warn-ignore ../data/ignore-warn-example.txt --format json"), &stdout, io.Discard))
	bc := formatters.Changes{}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &bc))
	require.Len(t, bc, 6)
}

func Test_BreakingChangesIgnoreErrsApiSchemaOptional(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml --err-ignore ../data/ignore-err-example.txt --warn-ignore ../data/ignore-warn-example.txt --include-checks api-schema-removed --format json"), &stdout, io.Discard))
	bc := formatters.Changes{}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &bc))
	require.Len(t, bc, 6)
}

func Test_BreakingChangesInvalidIgnoreFile(t *testing.T) {
	require.Equal(t, 121, internal.Run(cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml --err-ignore no-file"), io.Discard, io.Discard))
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

func Test_BreakingChangesFlattenDeprecated(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/allof/simple.yaml ../data/allof/revision.yaml --flatten --fail-on ERR"), io.Discard, io.Discard))
}

func Test_BreakingChangesFlattenAllOf(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/allof/simple.yaml ../data/allof/revision.yaml --flatten-allof --fail-on ERR"), io.Discard, io.Discard))
}

func Test_BreakingChangesInvalidDeprecationDays(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 100, internal.Run(cmdToArgs("oasdiff breaking ../data/deprecation/base.yaml ../data/deprecation/deprecated-with-sunset.yaml --deprecation-days-stable=-1"), io.Discard, &stderr))
	require.Equal(t, "Error: invalid argument \"-1\" for \"--deprecation-days-stable\" flag: must be a non-negative integer\n", stderr.String())
}

func Test_BreakingChangesFlattenCommonParams(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/common-params/params_in_path.yaml ../data/common-params/params_in_op.yaml --flatten-params --fail-on ERR"), io.Discard, io.Discard))
}

func Test_Color(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking ../data/allof/simple.yaml ../data/allof/revision.yaml --color always"), io.Discard, io.Discard))
}
