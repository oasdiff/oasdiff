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
func Test_FailOnMissingVersionBump_MajorBumpPasses(t *testing.T) {
	code, stdout, stderr := runVersioning(t, "revision_major_bump.yaml", "--fail-on MISSING-VERSION-BUMP")
	require.Zero(t, code)
	require.Contains(t, stdout, "api-path-removed-without-deprecation")
	require.Contains(t, stderr, "versioning policy: satisfied. info.version moved from 1.0.0 to 2.0.0, which allows the breaking changes.")
}

func Test_FailOnMissingVersionBump_SameVersionFails(t *testing.T) {
	code, stdout, stderr := runVersioning(t, "revision_same_version.yaml", "--fail-on MISSING-VERSION-BUMP")
	require.Equal(t, 1, code)
	require.Contains(t, stdout, "api-path-removed-without-deprecation")
	require.Contains(t, stderr, "versioning policy: violated. info.version is still 1.0.0.")
}

func Test_FailOnMissingVersionBump_MinorBumpFails(t *testing.T) {
	code, _, stderr := runVersioning(t, "revision_minor_bump.yaml", "--fail-on MISSING-VERSION-BUMP")
	require.Equal(t, 1, code)
	require.Contains(t, stderr, "versioning policy: violated. info.version moved from 1.0.0 to 1.1.0 without a major version increase.")
}

// Enforcement was asked for and the versions don't define a major increase, so
// the run fails rather than pass a release the policy never judged.
func Test_FailOnMissingVersionBump_NonSemverFails(t *testing.T) {
	code, _, stderr := runVersioning(t, "revision_date_version.yaml", "--fail-on MISSING-VERSION-BUMP")
	require.Equal(t, 1, code)
	require.Contains(t, stderr, `versioning policy: cannot be evaluated. info.version must be a semantic version in both specs (base "1.0.0", revision "2026-01-01").`)
}

// With no breaking change there is nothing for the policy to judge, so the run
// passes.
func Test_FailOnMissingVersionBump_NotApplicable(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := internal.Run(
		cmdToArgs("oasdiff breaking ../data/versioning/base.yaml ../data/versioning/base.yaml --fail-on MISSING-VERSION-BUMP"),
		&stdout, &stderr)
	require.Zero(t, code)
	require.Contains(t, stderr.String(), "versioning policy: not applicable. No breaking changes, so no version increase is required.")
}

// breaking reports a warning as a breaking change, so a warning on its own
// demands the major version increase.
func Test_FailOnMissingVersionBump_WarningFails(t *testing.T) {
	code, stdout, stderr := runVersioning(t, "revision_warning_only.yaml", "--fail-on MISSING-VERSION-BUMP")
	require.Equal(t, 1, code)
	require.Contains(t, stdout, "request-parameter-removed")
	require.Contains(t, stderr, "versioning policy: violated. info.version is still 1.0.0.")
}

// A severity asks a different question, so a major version increase changes
// nothing and the breaking change fails the run.
func Test_FailOnSeverity_IgnoresTheVersion(t *testing.T) {
	code, _, stderr := runVersioning(t, "revision_major_bump.yaml", "--fail-on ERR")
	require.Equal(t, 1, code)
	require.Empty(t, stderr)
}

// The verdict is not a change, so a machine-readable report stays parseable
// and carries no entry for it.
func Test_FailOnMissingVersionBump_JSONOutputIsUnaffected(t *testing.T) {
	code, stdout, stderr := runVersioning(t, "revision_same_version.yaml", "--fail-on MISSING-VERSION-BUMP --format json")
	require.Equal(t, 1, code)
	require.Contains(t, stderr, "versioning policy: violated")

	changes := formatters.Changes{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &changes))
	require.Len(t, changes, 1)
	require.Equal(t, "api-path-removed-without-deprecation", changes[0].Id)
}

// The policy is the whole answer, so a warning the release carries alongside
// its breaking changes does not fail the run.
func Test_FailOnMissingVersionBump_SatisfiedWithWarning(t *testing.T) {
	code, stdout, stderr := runVersioning(t, "revision_major_bump_with_warning.yaml", "--fail-on MISSING-VERSION-BUMP")
	require.Zero(t, code)
	require.Contains(t, stdout, "api-path-removed-without-deprecation")
	require.Contains(t, stderr, "versioning policy: satisfied. info.version moved from 1.0.0 to 2.0.0, which allows the breaking changes.")
}

// More than one breaking change, and a version that moved without the major
// increase they need.
func Test_FailOnMissingVersionBump_CountsEveryBreakingChange(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := internal.Run(
		cmdToArgs("oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test3.yaml --fail-on MISSING-VERSION-BUMP"),
		&stdout, &stderr)

	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "versioning policy: violated. info.version moved from 1.0.0 to 1.0.1 without a major version increase.")
}

// The same pair the other way round, where info.version went backwards.
func Test_FailOnMissingVersionBump_VersionDecreased(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := internal.Run(
		cmdToArgs("oasdiff breaking ../data/openapi-test3.yaml ../data/openapi-test1.yaml --fail-on MISSING-VERSION-BUMP"),
		&stdout, &stderr)

	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "versioning policy: violated. info.version decreased from 1.0.1 to 1.0.0.")
}

// validate has no versioning policy, so it offers the severities only.
func Test_FailOnMissingVersionBump_NotAvailableOnValidate(t *testing.T) {
	var stderr bytes.Buffer
	code := internal.Run(
		cmdToArgs("oasdiff validate ../data/versioning/base.yaml --fail-on MISSING-VERSION-BUMP"),
		io.Discard, &stderr)

	require.Equal(t, 100, code)
	require.Contains(t, stderr.String(), "MISSING-VERSION-BUMP is not one of the allowed values: ERR, WARN, or INFO")
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
