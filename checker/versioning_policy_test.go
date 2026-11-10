package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// versioningVerdict compares a pair of specs carrying one breaking change (a
// request property became required) under the given versions, and returns the
// policy's verdict on them.
func versioningVerdict(t *testing.T, baseVersion, revisionVersion string) checker.VersioningVerdict {
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

	return checker.CheckVersioningPolicy(d, changes)
}

func requireViolation(t *testing.T, baseVersion, revisionVersion, reason string) {
	t.Helper()
	verdict := versioningVerdict(t, baseVersion, revisionVersion)
	require.Equal(t, checker.VersioningPolicyViolated, verdict.Status,
		"expected a violation for %q -> %q", baseVersion, revisionVersion)
	require.Equal(t, reason, verdict.Reason)
	require.Equal(t, baseVersion, verdict.BaseVersion)
	require.Equal(t, revisionVersion, verdict.RevisionVersion)
	require.Equal(t, 1, verdict.BreakingChanges)
}

func requireSatisfied(t *testing.T, baseVersion, revisionVersion string) {
	t.Helper()
	verdict := versioningVerdict(t, baseVersion, revisionVersion)
	require.Equal(t, checker.VersioningPolicySatisfied, verdict.Status,
		"expected a pass for %q -> %q, got %v", baseVersion, revisionVersion, verdict)
	require.Empty(t, verdict.Reason)
	require.Equal(t, 1, verdict.BreakingChanges)
}

func requireUndecidable(t *testing.T, baseVersion, revisionVersion string) {
	t.Helper()
	verdict := versioningVerdict(t, baseVersion, revisionVersion)
	require.Equal(t, checker.VersioningPolicyUndecidable, verdict.Status,
		"expected no verdict for %q -> %q", baseVersion, revisionVersion)
	require.Empty(t, verdict.Reason)
}

func TestVersioningPolicy_BreakingChangeWithoutMajorBump(t *testing.T) {
	requireViolation(t, "1.0.0", "1.1.0", checker.APIMajorVersionNotBumpedId)
	requireViolation(t, "1.0.0", "1.0.1", checker.APIMajorVersionNotBumpedId)
	requireViolation(t, "v1.0.0", "v1.1.0", checker.APIMajorVersionNotBumpedId)
}

// The worst violation of the policy: shipping a breaking change under the same
// version the previous release used.
func TestVersioningPolicy_BreakingChangeWithoutAnyBump(t *testing.T) {
	requireViolation(t, "1.0.0", "1.0.0", checker.APIVersionNotBumpedId)
	requireViolation(t, "0.3.1", "0.3.1", checker.APIVersionNotBumpedId)
}

func TestVersioningPolicy_BreakingChangeWithVersionDecrease(t *testing.T) {
	requireViolation(t, "2.0.0", "1.0.0", checker.APIVersionDecreasedId)
	requireViolation(t, "1.2.0", "1.1.0", checker.APIVersionDecreasedId)
	requireViolation(t, "1.0.1", "1.0.0", checker.APIVersionDecreasedId)
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
	requireViolation(t, "0.1.0", "0.1.1", checker.APIMajorVersionNotBumpedId)
}

// Prerelease and build metadata don't decide which number was bumped.
func TestVersioningPolicy_PrereleaseAndBuildMetadata(t *testing.T) {
	requireViolation(t, "1.0.0", "1.1.0-rc.1", checker.APIMajorVersionNotBumpedId)
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
	require.Equal(t, checker.VersioningPolicyViolated, checker.CheckVersioningPolicy(d, changes).Status)

	for _, id := range versioningIds {
		requireNoChange(t, changes, id)
		require.NotContains(t, checker.GetAllRuleIds(), id)
	}
}

var versioningIds = []string{
	checker.APIVersionNotBumpedId,
	checker.APIVersionDecreasedId,
	checker.APIMajorVersionNotBumpedId,
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
	verdict := checker.CheckVersioningPolicy(d, changes)
	require.Equal(t, checker.VersioningPolicyNotApplicable, verdict.Status)
	require.Zero(t, verdict.BreakingChanges)
}

// The policy counts a change at the level it is reported at, not the level its
// rule declares. A disclaimer that lowers a change below error therefore also
// stops it demanding a bump: reporting that oasdiff cannot tell whether the
// change is breaking, and then requiring the major version to be raised for
// it, would contradict itself.
//
// Both specs here are 1.0.0, so an uncounted change is the only reason the
// policy has nothing to judge.
func TestVersioningPolicy_IgnoresAChangeADisclaimerLowered(t *testing.T) {
	s1, err := open("../data/checker/disclaimer_allof_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/disclaimer_allof_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)

	change := requireChange(t, changes, checker.ResponsePropertyBecameOptionalId)
	require.Equal(t, checker.WARN, change.GetLevel(), "the disclaimer lowered it, which is what makes this case interesting")

	require.Equal(t, checker.VersioningPolicyNotApplicable, checker.CheckVersioningPolicy(d, changes).Status)
}

// The other half: a change that reaches error still demands the bump, so the
// test above cannot be satisfied by a policy that never speaks.
func TestVersioningPolicy_CountsAChangeNoDisclaimerLowered(t *testing.T) {
	s1, err := open("../data/checker/disclaimer_allof_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/disclaimer_allof_revision.yaml")
	require.NoError(t, err)

	// A removed path is an error no disclaimer lowers.
	s2.Spec.Paths.Delete("/x")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)

	verdict := checker.CheckVersioningPolicy(d, changes)
	require.Equal(t, checker.VersioningPolicyViolated, verdict.Status)
	require.Equal(t, checker.APIVersionNotBumpedId, verdict.Reason)
}

// A spec with no info section at all, which is the other way the versions
// cannot be compared.
func TestVersioningPolicy_NoDiffReport(t *testing.T) {
	changes := checker.Changes{checker.ApiChange{Id: "any", Level: checker.ERR}}
	verdict := checker.CheckVersioningPolicy(nil, changes)
	require.Equal(t, checker.VersioningPolicyUndecidable, verdict.Status)
	require.Equal(t, 1, verdict.BreakingChanges)
}
