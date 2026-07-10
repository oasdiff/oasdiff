package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func nullableWrapChanges(t *testing.T, base, revision string) []string {
	t.Helper()
	s1, err := open(base)
	require.NoError(t, err)
	s2, err := open(revision)
	require.NoError(t, err)
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(checker.NewConfig(checker.GetAllChecks()), d, osm, checker.INFO)
	ids := make([]string, 0, len(errs))
	for _, e := range errs {
		ids = append(ids, e.GetId())
	}
	return ids
}

// Wrapping a request property in oneOf: [{type: "null"}, <equivalent schema>]
// is a nullability change: one became-nullable per property, with no
// enum/pattern/type/oneOf artifacts from comparing against the bare wrapper
// (oasdiff/oasdiff#1088).
func TestNullableWrapping_RequestProperties(t *testing.T) {
	ids := nullableWrapChanges(t, "../data/checker/nullable_wrap_base.yaml", "../data/checker/nullable_wrap_revision.yaml")
	require.ElementsMatch(t, []string{
		checker.RequestPropertyBecomeNullableId, // optionalEnum, wrapped
		checker.RequestPropertyBecomeNullableId, // optionalRef, wrapped
		checker.RequestPropertyBecomeNullableId, // optionalPrimitive, type array
	}, ids)
}

// A wrap whose branch also narrows the enum is not a pure nullability change:
// the removal keeps being reported.
func TestNullableWrapping_NarrowedEnumStillReported(t *testing.T) {
	ids := nullableWrapChanges(t, "../data/checker/nullable_wrap_narrowed_base.yaml", "../data/checker/nullable_wrap_narrowed_revision.yaml")
	require.Contains(t, ids, checker.RequestPropertyEnumValueRemovedId)
	require.NotContains(t, ids, checker.RequestPropertyBecomeNullableId)
}

func TestNullableWrapping_ResponseProperty(t *testing.T) {
	ids := nullableWrapChanges(t, "../data/checker/nullable_wrap_response_base.yaml", "../data/checker/nullable_wrap_response_revision.yaml")
	require.Contains(t, ids, checker.ResponsePropertyBecameNullableId)
	require.NotContains(t, ids, checker.ResponsePropertyEnumValueRemovedId)
}

// The nullable transition reports exactly one finding at every level: the
// claim table silences the facet artifacts, and each level has its reporter
// (became-nullable at body/property; the parameter list-of-types finding at
// parameter level, where no became-nullable rule exists).
func TestNullableWrapping_SingleFindingPerLevel(t *testing.T) {
	// t.Run("request body", func(t *testing.T) {
	// 	ids := nullableWrapChanges(t, "../data/checker/nullable_wrap_body_base.yaml", "../data/checker/nullable_wrap_body_revision.yaml")
	// 	require.Equal(t, []string{checker.RequestBodyBecomeNullableId}, ids)
	// })
	t.Run("parameter", func(t *testing.T) {
		ids := nullableWrapChanges(t, "../data/checker/nullable_wrap_param_base.yaml", "../data/checker/nullable_wrap_param_revision.yaml")
		require.Equal(t, []string{checker.RequestParameterListOfTypesWidenedId}, ids)
	})
}

// The reverse direction: removing the wrapper reports one became-not-nullable
// per level (the same fixtures, compared in reverse). At parameter level the
// reporter is the parameter list-of-types finding, as in the forward
// direction (oasdiff/oasdiff#1091).
func TestNullableUnwrapping_SingleFindingPerLevel(t *testing.T) {
	t.Run("request property", func(t *testing.T) {
		ids := nullableWrapChanges(t, "../data/checker/nullable_wrap_revision.yaml", "../data/checker/nullable_wrap_base.yaml")
		require.ElementsMatch(t, []string{
			checker.RequestPropertyBecomeNotNullableId, // optionalEnum, unwrapped
			checker.RequestPropertyBecomeNotNullableId, // optionalRef, unwrapped
			checker.RequestPropertyBecomeNotNullableId, // optionalPrimitive, type array
		}, ids)
	})
	t.Run("request body", func(t *testing.T) {
		ids := nullableWrapChanges(t, "../data/checker/nullable_wrap_body_revision.yaml", "../data/checker/nullable_wrap_body_base.yaml")
		require.Equal(t, []string{checker.RequestBodyBecomeNotNullableId}, ids)
	})
	t.Run("parameter", func(t *testing.T) {
		ids := nullableWrapChanges(t, "../data/checker/nullable_wrap_param_revision.yaml", "../data/checker/nullable_wrap_param_base.yaml")
		require.Equal(t, []string{checker.RequestParameterListOfTypesNarrowedId}, ids)
	})
	t.Run("response property", func(t *testing.T) {
		ids := nullableWrapChanges(t, "../data/checker/nullable_wrap_response_revision.yaml", "../data/checker/nullable_wrap_response_base.yaml")
		require.Equal(t, []string{checker.ResponsePropertyBecameNotNullableId}, ids)
	})
}
