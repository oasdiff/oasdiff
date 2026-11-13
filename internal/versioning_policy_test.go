package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

// runVersioning compares the versioning fixtures with the given extra flags
// and returns the exit code, stdout and stderr.
func runVersioning(t *testing.T, revision, flags string) (int, string, string) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	code := internal.Run(
		cmdToArgs("oasdiff breaking ../data/versioning/base.yaml ../data/versioning/"+revision+" "+flags),
		&stdout, &stderr)

	return code, stdout.String(), stderr.String()
}

// A breaking change released under a major version increase is what the policy
// allows, so the run passes while the report still lists it.
func Test_EnforceVersionBump_MajorBumpPasses(t *testing.T) {
	code, stdout, stderr := runVersioning(t, "revision_major_bump.yaml", "--enforce-version-bump --fail-on ERR")
	require.Zero(t, code)
	require.Contains(t, stdout, "api-path-removed-without-deprecation")
	require.Contains(t, stderr, "versioning policy: satisfied. info.version moved from 1.0.0 to 2.0.0, which allows the breaking changes.")
}

func Test_EnforceVersionBump_SameVersionFails(t *testing.T) {
	code, stdout, stderr := runVersioning(t, "revision_same_version.yaml", "--enforce-version-bump")
	require.Equal(t, 1, code)
	require.Contains(t, stdout, "api-path-removed-without-deprecation")
	require.Contains(t, stderr, "versioning policy: violated. info.version is still 1.0.0.")
}

func Test_EnforceVersionBump_MinorBumpFails(t *testing.T) {
	code, _, stderr := runVersioning(t, "revision_minor_bump.yaml", "--enforce-version-bump")
	require.Equal(t, 1, code)
	require.Contains(t, stderr, "versioning policy: violated. info.version moved from 1.0.0 to 1.1.0 without a major version increase.")
}

// Enforcement was asked for and the versions don't define a major increase, so
// the run fails rather than pass a release the policy never judged.
func Test_EnforceVersionBump_NonSemverFails(t *testing.T) {
	code, _, stderr := runVersioning(t, "revision_date_version.yaml", "--enforce-version-bump")
	require.Equal(t, 1, code)
	require.Contains(t, stderr, `versioning policy: cannot be evaluated. info.version must be a semantic version in both specs (base "1.0.0", revision "2026-01-01").`)
}

// A major version increase answers for the errors, not the warnings: --fail-on
// still answers for everything below error.
func Test_EnforceVersionBump_WarningStillFails(t *testing.T) {
	code, _, stderr := runVersioning(t, "revision_major_bump_with_warning.yaml", "--enforce-version-bump --fail-on WARN")
	require.Equal(t, 1, code)
	require.Contains(t, stderr, "versioning policy: satisfied")

	code, _, _ = runVersioning(t, "revision_major_bump_with_warning.yaml", "--enforce-version-bump --fail-on ERR")
	require.Zero(t, code)
}

// With no breaking change there is nothing for the policy to judge, and
// --fail-on behaves as it does without the flag.
func Test_EnforceVersionBump_NotApplicable(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := internal.Run(
		cmdToArgs("oasdiff breaking ../data/versioning/base.yaml ../data/versioning/base.yaml --enforce-version-bump --fail-on WARN"),
		&stdout, &stderr)
	require.Zero(t, code)
	require.Contains(t, stderr.String(), "versioning policy: not applicable. No breaking changes, so no version increase is required.")
}

// Without the flag, a major version increase changes nothing: this is the
// behavior --enforce-version-bump exists to change.
func Test_EnforceVersionBump_OffByDefault(t *testing.T) {
	code, _, stderr := runVersioning(t, "revision_major_bump.yaml", "--fail-on ERR")
	require.Equal(t, 1, code)
	require.Empty(t, stderr)
}

// The verdict is not a change, so a machine-readable report stays parseable
// and carries no entry for it.
func Test_EnforceVersionBump_JSONOutputIsUnaffected(t *testing.T) {
	code, stdout, stderr := runVersioning(t, "revision_same_version.yaml", "--enforce-version-bump --format json")
	require.Equal(t, 1, code)
	require.Contains(t, stderr, "versioning policy: violated")

	changes := formatters.Changes{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &changes))
	require.Len(t, changes, 1)
	require.Equal(t, "api-path-removed-without-deprecation", changes[0].Id)
}

// Without --fail-on there is nothing left to answer for once the policy is
// satisfied, so a warning in the report does not fail the run.
func Test_EnforceVersionBump_SatisfiedWithoutFailOn(t *testing.T) {
	code, stdout, stderr := runVersioning(t, "revision_major_bump_with_warning.yaml", "--enforce-version-bump")
	require.Zero(t, code)
	require.Contains(t, stdout, "api-path-removed-without-deprecation")
	require.Contains(t, stderr, "versioning policy: satisfied. info.version moved from 1.0.0 to 2.0.0, which allows the breaking changes.")
}

// More than one breaking change, and a version that moved without the major
// increase they need.
func Test_EnforceVersionBump_CountsEveryBreakingChange(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := internal.Run(
		cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml --enforce-version-bump"),
		&stdout, &stderr)

	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "versioning policy: violated. info.version moved from 1.0.0 to 1.0.1 without a major version increase.")
}

// The same pair the other way round, where info.version went backwards.
func Test_EnforceVersionBump_VersionDecreased(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := internal.Run(
		cmdToArgs("oasdiff breaking ../data/openapi-test3.yaml ../data/openapi-test1.yaml --enforce-version-bump"),
		&stdout, &stderr)

	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "versioning policy: violated. info.version decreased from 1.0.1 to 1.0.0.")
}

// The three ids are no longer checks, so a severity-levels file naming one of
// them is rejected like any other unknown id rather than silently having no
// effect.
func Test_RetiredSeverityLevelIsRejected(t *testing.T) {
	levels := filepath.Join(t.TempDir(), "levels.txt")
	require.NoError(t, os.WriteFile(levels, []byte("api-version-not-bumped err\n"), 0644))

	var stderr bytes.Buffer
	code := internal.Run(
		cmdToArgs("oasdiff breaking ../data/versioning/base.yaml ../data/versioning/revision_major_bump.yaml --severity-levels "+levels),
		io.Discard, &stderr)

	require.Equal(t, 106, code)
	require.Contains(t, stderr.String(), `invalid rule id "api-version-not-bumped" on line 1`)
}
