package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func serializationChanges(t *testing.T, base, revision string) checker.Changes {
	t.Helper()
	s1, err := open(serializationFile(base + ".yaml"))
	require.NoError(t, err)
	s2, err := open(serializationFile(revision + ".yaml"))
	require.NoError(t, err)
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	return checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterSerializationUpdatedCheck), d, osm, checker.INFO)
}

func TestRequestParameterExplodeChanged(t *testing.T) {
	changes := serializationChanges(t, "base", "tags_explode_false")
	change := requireSingleChange(t, changes, checker.RequestParameterExplodeChangedId)
	require.Equal(t, checker.ERR, change.GetLevel())
	require.Equal(t, "for the `query` request parameter `tags`, explode was changed from `true` to `false`", change.GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// Writing out the default explode of a form parameter is not a change.
func TestRequestParameterExplodeDefaultWrittenOut(t *testing.T) {
	require.Empty(t, serializationChanges(t, "base", "tags_explode_true"))
}

// explode makes no difference to a single value.
func TestRequestParameterExplodeChangedOnScalar(t *testing.T) {
	require.Empty(t, serializationChanges(t, "base", "q_explode_false"))
}

// A schema without a type may be an array or an object.
func TestRequestParameterExplodeChangedOnUntypedSchema(t *testing.T) {
	requireSingleChange(t, serializationChanges(t, "base", "filter_explode_false"), checker.RequestParameterExplodeChangedId)
}

// A style change is reported once, even though the default explode changes with it.
func TestRequestParameterStyleChanged(t *testing.T) {
	change := requireSingleChange(t, serializationChanges(t, "base", "tags_pipe"), checker.RequestParameterStyleChangedId)
	require.Equal(t, checker.ERR, change.GetLevel())
	require.Equal(t, "for the `query` request parameter `tags`, the style was changed from `form` to `pipeDelimited`", change.GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// The default explode is false for every style but form.
func TestRequestParameterExplodeDefaultOfNonFormStyle(t *testing.T) {
	require.Empty(t, serializationChanges(t, "tags_pipe", "tags_pipe_explode_false"))
}

func TestRequestParameterStyleChangedOnPath(t *testing.T) {
	change := requireSingleChange(t, serializationChanges(t, "base", "id_label"), checker.RequestParameterStyleChangedId)
	require.Equal(t, "for the `path` request parameter `id`, the style was changed from `simple` to `label`", change.GetUncolorizedText(checker.NewDefaultLocalizer()))
}

func TestRequestParameterBecameAllowReserved(t *testing.T) {
	change := requireSingleChange(t, serializationChanges(t, "base", "q_allow_reserved"), checker.RequestParameterBecameAllowReservedId)
	require.Equal(t, checker.INFO, change.GetLevel())
}

func TestRequestParameterBecameNotAllowReserved(t *testing.T) {
	change := requireSingleChange(t, serializationChanges(t, "q_allow_reserved", "base"), checker.RequestParameterBecameNotAllowReservedId)
	require.Equal(t, checker.ERR, change.GetLevel())
	require.Equal(t, "the `query` request parameter `q` no longer accepts reserved characters without percent-encoding", change.GetUncolorizedText(checker.NewDefaultLocalizer()))
}

func TestRequestParameterSerializationSources(t *testing.T) {
	s1, err := open(serializationFile("base.yaml"), newLoaderWithOriginTracking())
	require.NoError(t, err)
	s2, err := open(serializationFile("tags_explode_false.yaml"), newLoaderWithOriginTracking())
	require.NoError(t, err)
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	change := requireSingleChange(t, checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterSerializationUpdatedCheck), d, osm, checker.INFO), checker.RequestParameterExplodeChangedId)
	require.NotNil(t, change.GetRevisionSource())
	require.Equal(t, 16, change.GetRevisionSource().Line)
}
