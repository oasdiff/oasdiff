package versioning_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/oasdiff/oasdiff/versioning"
	"github.com/stretchr/testify/require"
)

func open(file string) (*load.SpecInfo, error) {
	return load.NewSpecInfo(openapi3.NewLoader(), load.NewSource(file))
}

func singleCheckConfig(c checker.BackwardCompatibilityCheck) *checker.Config {
	return checker.NewConfig(checker.BackwardCompatibilityChecks{c}, checker.WithSingleCheck(c))
}

func allChecksConfig() *checker.Config {
	return checker.NewConfig(checker.GetAllChecks())
}

func requireChange(t *testing.T, changes checker.Changes, id string) checker.Change {
	t.Helper()
	for _, change := range changes {
		if change.GetId() == id {
			return change
		}
	}
	require.FailNowf(t, "change not found", "expected a change with id %q", id)
	return nil
}

func requireNoChange(t *testing.T, changes checker.Changes, id string) {
	t.Helper()
	for _, change := range changes {
		require.NotEqual(t, id, change.GetId(), "unexpected change with id %q", id)
	}
}

// versioningVerdict compares a pair of specs carrying one breaking change (a
// request property became required) under the given versions, and returns the
// policy's verdict on them.
func versioningVerdict(t *testing.T, baseVersion, revisionVersion string) versioning.Verdict {
	t.Helper()

	s1, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Required = []string{""}
	s1.Spec.Info.Version = baseVersion
	s2.Spec.Info.Version = revisionVersion

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	check := checker.RequestPropertyRequiredUpdatedCheck
	changes := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(check), d, osm, checker.INFO)
	requireChange(t, changes, checker.RequestPropertyBecameRequiredId)

	return versioning.Check(d, changes)
}

func requireViolation(t *testing.T, baseVersion, revisionVersion string, reason versioning.Reason) {
	t.Helper()
	verdict := versioningVerdict(t, baseVersion, revisionVersion)
	require.Equal(t, versioning.Violated, verdict.Status,
		"expected a violation for %q -> %q", baseVersion, revisionVersion)
	require.Equal(t, reason, verdict.Reason)
	require.Equal(t, baseVersion, verdict.BaseVersion)
	require.Equal(t, revisionVersion, verdict.RevisionVersion)
}

func requireSatisfied(t *testing.T, baseVersion, revisionVersion string) {
	t.Helper()
	verdict := versioningVerdict(t, baseVersion, revisionVersion)
	require.Equal(t, versioning.Satisfied, verdict.Status,
		"expected a pass for %q -> %q, got %v", baseVersion, revisionVersion, verdict)
	require.Equal(t, versioning.NoReason, verdict.Reason)
}

func requireUndecidable(t *testing.T, baseVersion, revisionVersion string) {
	t.Helper()
	verdict := versioningVerdict(t, baseVersion, revisionVersion)
	require.Equal(t, versioning.Undecidable, verdict.Status,
		"expected no verdict for %q -> %q", baseVersion, revisionVersion)
	require.Equal(t, versioning.NoReason, verdict.Reason)
}

func TestVersioningPolicy_BreakingChangeWithoutMajorBump(t *testing.T) {
	requireViolation(t, "1.0.0", "1.1.0", versioning.MajorNotBumped)
	requireViolation(t, "1.0.0", "1.0.1", versioning.MajorNotBumped)
	requireViolation(t, "v1.0.0", "v1.1.0", versioning.MajorNotBumped)
}

// The worst violation of the policy: shipping a breaking change under the same
// version the previous release used.
func TestVersioningPolicy_BreakingChangeWithoutAnyBump(t *testing.T) {
	requireViolation(t, "1.0.0", "1.0.0", versioning.NotBumped)
	requireViolation(t, "0.3.1", "0.3.1", versioning.NotBumped)
}

func TestVersioningPolicy_BreakingChangeWithVersionDecrease(t *testing.T) {
	requireViolation(t, "2.0.0", "1.0.0", versioning.Decreased)
	requireViolation(t, "1.2.0", "1.1.0", versioning.Decreased)
	requireViolation(t, "1.0.1", "1.0.0", versioning.Decreased)
}

func TestVersioningPolicy_MajorBumpIsSatisfied(t *testing.T) {
	requireSatisfied(t, "1.0.0", "2.0.0")
	requireSatisfied(t, "1.9.3", "2.0.0")
}

// Anything that isn't semver leaves the policy with no definition of a major
// increase, so it reaches no verdict rather than guessing.
func TestVersioningPolicy_NonSemverHasNoVerdict(t *testing.T) {
	for _, versions := range [][2]string{
		{"1.0", "2.0"},
		{"2026-01-01", "2026-06-01"},
		{"v1", "v2"},
		{"1.0.0", "2.0"},
		{"", ""},
		{"1.0.0", ""},
	} {
		requireUndecidable(t, versions[0], versions[1])
	}
}

// Below 1.0.0 semver gives the minor the major's role, so a minor bump carries
// a breaking change and a patch bump does not.
func TestVersioningPolicy_PreOneZero(t *testing.T) {
	requireSatisfied(t, "0.1.0", "0.2.0")
	requireViolation(t, "0.1.0", "0.1.1", versioning.MajorNotBumped)
}

// Prerelease and build metadata don't decide which number was bumped.
func TestVersioningPolicy_PrereleaseAndBuildMetadata(t *testing.T) {
	requireViolation(t, "1.0.0", "1.1.0-rc.1", versioning.MajorNotBumped)
	requireSatisfied(t, "1.0.0", "2.0.0+build.5")
}

// The verdict is about the release, so it is not reported as a change and the
// three outcome names are not rules.
func TestVersioningPolicy_IsNotAChange(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Required = []string{""}
	s1.Spec.Info.Version = "1.0.0"
	s2.Spec.Info.Version = "1.0.0"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	changes := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)
	require.Equal(t, versioning.Violated, versioning.Check(d, changes).Status)

	for _, id := range retiredIds {
		requireNoChange(t, changes, id)
		require.NotContains(t, checker.GetAllRuleIds(), id)
	}
}

// retiredIds are the ids the three version checks reported under, kept here so
// that re-registering one of them as a rule fails this test.
var retiredIds = []string{
	"api-version-not-bumped",
	"api-version-decreased",
	"api-major-version-not-bumped",
}

// Only a breaking change demands a major bump; a non-breaking release under a
// minor bump is exactly correct, and the policy has nothing to judge.
func TestVersioningPolicy_NonBreakingChangeIsNotApplicable(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)

	s2.Spec.Info.Description = "a description change is not breaking"
	s1.Spec.Info.Version = "1.0.0"
	s2.Spec.Info.Version = "1.1.0"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	changes := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)
	verdict := versioning.Check(d, changes)
	require.Equal(t, versioning.NotApplicable, verdict.Status)
}

// A warning demands the bump too: the commands that offer the policy report
// warnings as breaking changes, so a warning that did not count would leave
// the verdict contradicting the report above it. This one was lowered to a
// warning by a disclaimer, which is the level a change is reported at rather
// than the level its rule declares.
func TestVersioningPolicy_CountsAWarning(t *testing.T) {
	s1, err := open("../data/checker/disclaimer_allof_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/disclaimer_allof_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)

	change := requireChange(t, changes, checker.ResponsePropertyBecameOptionalId)
	require.Equal(t, checker.WARN, change.GetLevel(), "the disclaimer lowered it, which is what makes this case interesting")

	// Both specs are 1.0.0, so the warning is the only thing that can demand
	// a bump.
	verdict := versioning.Check(d, changes)
	require.Equal(t, versioning.Violated, verdict.Status)
	require.Equal(t, versioning.NotBumped, verdict.Reason)
}

// Below a warning there is nothing to carry, so an info change leaves the
// policy with nothing to judge.
func TestVersioningPolicy_IgnoresAnInfoChange(t *testing.T) {
	changes := checker.Changes{checker.ApiChange{Id: "any", Level: checker.INFO}}
	require.Equal(t, versioning.NotApplicable, versioning.Check(nil, changes).Status)
}

// A spec with no info section at all, which is the other way the versions
// cannot be compared.
func TestVersioningPolicy_NoDiffReport(t *testing.T) {
	changes := checker.Changes{checker.ApiChange{Id: "any", Level: checker.ERR}}
	verdict := versioning.Check(nil, changes)
	require.Equal(t, versioning.Undecidable, verdict.Status)
}
