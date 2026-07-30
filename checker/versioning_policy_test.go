package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// versioningPolicyIds runs the full pipeline over a pair of specs carrying one
// breaking change (a request property became required) and the given versions,
// and reports whether the versioning policy spoke.
func versioningPolicyIds(t *testing.T, baseVersion, revisionVersion string) checker.Changes {
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

	// Not singleCheckConfig: that switches the versioning policy off (see its
	// comment), which is exactly what these tests exercise.
	check := checker.RequestPropertyRequiredUpdatedCheck
	config := checker.NewConfig(checker.BackwardCompatibilityChecks{check}, checker.WithSingleCheck(check))

	return checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
}

// requireViolation asserts the policy reported exactly the given id.
func requireViolation(t *testing.T, baseVersion, revisionVersion, id string) {
	t.Helper()
	changes := versioningPolicyIds(t, baseVersion, revisionVersion)
	require.True(t, containsId(changes, id),
		"expected %s for %q -> %q, got %v", id, baseVersion, revisionVersion, ids(changes))
	for _, other := range versioningIds {
		if other != id {
			require.False(t, containsId(changes, other),
				"expected only %s for %q -> %q, also got %s", id, baseVersion, revisionVersion, other)
		}
	}
}

var versioningIds = []string{
	checker.APIVersionNotBumpedId,
	checker.APIVersionDecreasedId,
	checker.APIMajorVersionNotBumpedId,
}

func ids(changes checker.Changes) []string {
	out := make([]string, 0, len(changes))
	for _, c := range changes {
		out = append(out, c.GetId())
	}
	return out
}

func requireVersioningPolicySilent(t *testing.T, baseVersion, revisionVersion string) {
	t.Helper()
	changes := versioningPolicyIds(t, baseVersion, revisionVersion)
	for _, id := range versioningIds {
		require.False(t, containsId(changes, id),
			"expected silence for %q -> %q, got %s", baseVersion, revisionVersion, id)
	}
	// The breaking change itself must still be reported: the policy adds a
	// finding, it never suppresses one.
	require.True(t, containsId(changes, checker.RequestPropertyBecameRequiredId))
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

func TestVersioningPolicy_MajorBumpIsSilent(t *testing.T) {
	requireVersioningPolicySilent(t, "1.0.0", "2.0.0")
	requireVersioningPolicySilent(t, "1.9.3", "2.0.0")
}

// Anything that isn't semver leaves the policy with no definition of a major
// bump, so it stays out of the way rather than guessing.
func TestVersioningPolicy_NonSemverIsSilent(t *testing.T) {
	for _, versions := range [][2]string{
		{"1.0", "2.0"},
		{"2026-01-01", "2026-06-01"},
		{"v1", "v2"},
		{"1.0.0", "2.0"},
		{"", ""},
		{"1.0.0", ""},
	} {
		requireVersioningPolicySilent(t, versions[0], versions[1])
	}
}

// Below 1.0.0 semver gives the minor the major's role, so a minor bump carries
// a breaking change and a patch bump does not.
func TestVersioningPolicy_PreOneZero(t *testing.T) {
	requireVersioningPolicySilent(t, "0.1.0", "0.2.0")
	requireViolation(t, "0.1.0", "0.1.1", checker.APIMajorVersionNotBumpedId)
}

// Prerelease and build metadata don't decide which number was bumped.
func TestVersioningPolicy_PrereleaseAndBuildMetadata(t *testing.T) {
	requireViolation(t, "1.0.0", "1.1.0-rc.1", checker.APIMajorVersionNotBumpedId)
	requireVersioningPolicySilent(t, "1.0.0", "2.0.0+build.5")
}

// The finding is document-level: it names no endpoint, and carries both
// versions so the reader can see the bump that was made.
func TestVersioningPolicy_ChangeShape(t *testing.T) {
	change := findChange(versioningPolicyIds(t, "1.0.0", "1.1.0"), checker.APIMajorVersionNotBumpedId)
	require.NotNil(t, change)
	require.Equal(t, "info", change.GetSection())
	require.Empty(t, change.GetPath())
	require.Empty(t, change.GetOperation())
	require.Equal(t, checker.INFO, change.GetLevel())
	require.Equal(t, []any{"1.0.0", "1.1.0"}, change.GetArgs())
	require.Contains(t, change.GetUncolorizedText(checker.NewDefaultLocalizer()), "major version")
}

// Only a breaking change demands a major bump; a non-breaking release under a
// minor bump is exactly correct and must stay quiet.
func TestVersioningPolicy_NonBreakingChangeIsSilent(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)

	s2.Spec.Info.Description = "a description change is not breaking"
	s1.Spec.Info.Version = "1.0.0"
	s2.Spec.Info.Version = "1.1.0"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	// allChecksConfig would switch the policy off, making this vacuous.
	changes := checker.CheckBackwardCompatibilityUntilLevel(
		checker.NewConfig(checker.GetAllChecks()), d, osm, checker.INFO)
	for _, id := range versioningIds {
		require.False(t, containsId(changes, id))
	}
}
