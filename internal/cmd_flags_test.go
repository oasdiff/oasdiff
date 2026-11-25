package internal_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

// --flatten-params is deprecated in the commands that run the checks, and a
// regular flag in diff.
func Test_FlattenParamsDeprecatedInChecksOnly(t *testing.T) {
	const specs = "../data/common-params/params_in_path.yaml ../data/common-params/params_in_op.yaml --flatten-params"

	var stdout, stderr bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking "+specs), &stdout, &stderr))
	require.Contains(t, stderr.String(), "Flag --flatten-params is deprecated")

	stdout.Reset()
	stderr.Reset()
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking --help"), &stdout, &stderr))
	require.NotContains(t, stdout.String(), "--flatten-params")

	stdout.Reset()
	stderr.Reset()
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff "+specs), &stdout, &stderr))
	require.Empty(t, stderr.String())

	stdout.Reset()
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff --help"), &stdout, &stderr))
	require.Contains(t, stdout.String(), "--flatten-params")
}

// A deprecated hidden flag (e.g. --flatten) still works but warns on stderr,
// never stdout, so json/yaml output stays clean.
func Test_DeprecatedFlagWarnsOnStderrOnly(t *testing.T) {
	var stdout, stderr bytes.Buffer
	internal.Run(cmdToArgs("oasdiff diff data/openapi-test1.yaml data/openapi-test1.yaml --flatten -f json"), &stdout, &stderr)

	require.Contains(t, stderr.String(), "Flag --flatten is deprecated: use --flatten-allof instead")
	require.NotContains(t, stdout.String(), "deprecated")
	var parsed any
	require.NoError(t, yaml.Unmarshal(stdout.Bytes(), &parsed), "stdout must stay valid json")
}

func Test_NoDeprecatedFlagNoWarning(t *testing.T) {
	var stderr bytes.Buffer
	internal.Run(cmdToArgs("oasdiff diff data/openapi-test1.yaml data/openapi-test1.yaml -f json"), io.Discard, &stderr)
	require.NotContains(t, stderr.String(), "deprecated")
}

// --case-insensitive-headers is deprecated and ignored in the commands that
// run the checks, and a regular flag in diff.
func Test_CaseInsensitiveHeadersDeprecatedInChecksOnly(t *testing.T) {
	const specs = "../data/header-case/base.yaml ../data/header-case/revision.yaml"

	for _, value := range []string{"", "=false"} {
		var stdout, stderr bytes.Buffer
		require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking "+specs+" --fail-on WARN --case-insensitive-headers"+value), &stdout, &stderr), value)
		require.Contains(t, stderr.String(), "Flag --case-insensitive-headers is deprecated")
	}

	var stdout, stderr bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff breaking --help"), &stdout, &stderr))
	require.NotContains(t, stdout.String(), "--case-insensitive-headers")

	stdout.Reset()
	stderr.Reset()
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff "+specs+" --case-insensitive-headers"), &stdout, &stderr))
	require.Empty(t, stderr.String())

	stdout.Reset()
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff --help"), &stdout, &stderr))
	require.Contains(t, stdout.String(), "--case-insensitive-headers")
}

// --exclude-extensions is deprecated and ignored in the commands that run the
// checks, and a regular flag in diff.
func Test_ExcludeExtensionsDeprecatedInChecksOnly(t *testing.T) {
	const specs = "../data/checker/excluded_elements_base.yaml ../data/checker/excluded_elements_revision.yaml"

	var stdout, stderr bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff changelog "+specs+" --exclude-extensions x-sunset"), &stdout, &stderr))
	require.Contains(t, stderr.String(), "Flag --exclude-extensions is deprecated")

	stdout.Reset()
	stderr.Reset()
	require.Zero(t, internal.Run(cmdToArgs("oasdiff changelog --help"), &stdout, &stderr))
	require.NotContains(t, stdout.String(), "--exclude-extensions")

	stdout.Reset()
	stderr.Reset()
	require.Zero(t, internal.Run(cmdToArgs("oasdiff diff "+specs+" --exclude-extensions x-sunset"), &stdout, &stderr))
	require.Empty(t, stderr.String())
	require.NotContains(t, stdout.String(), "x-sunset")
	require.Contains(t, stdout.String(), "x-stability-level")
}
