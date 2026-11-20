package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/versioning"
	"github.com/stretchr/testify/require"
)

const versionSpec = "../data/checker/request_property_became_required_base.yaml"

// versionChanges compares a pair of specs that differ only in info.version.
// Each spec gets its own loader, as the loader holds the state of one decode.
func versionChanges(t *testing.T, baseVersion, revisionVersion string, newLoaders ...func() *openapi3.Loader) checker.Changes {
	t.Helper()

	newLoader := openapi3.NewLoader
	if len(newLoaders) > 0 {
		newLoader = newLoaders[0]
	}

	s1, err := open(versionSpec, newLoader())
	require.NoError(t, err)
	s2, err := open(versionSpec, newLoader())
	require.NoError(t, err)

	s1.Spec.Info.Version = baseVersion
	s2.Spec.Info.Version = revisionVersion

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	return checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)
}

// Any change to info.version is reported, whether or not anything breaking
// came with it, and whatever the versioning scheme is.
func TestAPIVersionChanged(t *testing.T) {
	for _, versions := range [][2]string{
		{"1.0.0", "2.0.0"},
		{"1.0.0", "1.0.1"},
		{"2.0.0", "1.0.0"},
		{"2026-01-01", "2026-06-01"},
		{"v1", "v2"},
	} {
		changes := versionChanges(t, versions[0], versions[1])
		change := requireChange(t, changes, checker.APIVersionChangedId)
		require.Equal(t, checker.INFO, change.GetLevel())
		require.Equal(t, "info", change.GetSection())
		require.Equal(t, []any{versions[0], versions[1]}, change.GetArgs())
		require.Equal(t,
			"the API version changed from `"+versions[0]+"` to `"+versions[1]+"`",
			change.GetUncolorizedText(checker.NewDefaultLocalizer()))
	}
}

// The change points at the version field of each spec.
func TestAPIVersionChanged_WithSources(t *testing.T) {
	change := requireChange(t,
		versionChanges(t, "1.0.0", "2.0.0", newLoaderWithOriginTracking),
		checker.APIVersionChangedId)

	for _, source := range []*checker.Source{change.GetBaseSource(), change.GetRevisionSource()} {
		require.NotNil(t, source)
		require.Equal(t, versionSpec, source.File)
		require.Equal(t, 4, source.Line)
		require.Equal(t, 3, source.Column)
	}
}

func TestAPIVersionChanged_SameVersionIsSilent(t *testing.T) {
	requireNoChange(t, versionChanges(t, "1.0.0", "1.0.0"), checker.APIVersionChangedId)
}

// The change says what happened, the verdict says whether it was allowed, and
// the two are independent: the change is reported at INFO either way.
func TestAPIVersionChanged_IsNotTheVerdict(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Required = []string{""}
	s1.Spec.Info.Version = "1.0.0"
	s2.Spec.Info.Version = "1.1.0"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)

	requireChange(t, changes, checker.APIVersionChangedId)
	verdict := versioning.Check(d, changes)
	require.Equal(t, versioning.Violated, verdict.Status)
	require.Equal(t, versioning.MajorNotBumped, verdict.Reason)
}
