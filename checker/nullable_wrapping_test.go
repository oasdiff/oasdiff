package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
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

// A wrap whose branch also narrows the enum is two changes: the property
// accepts null now, and the removal keeps being reported.
func TestNullableWrapping_NarrowedEnumStillReported(t *testing.T) {
	ids := nullableWrapChanges(t, "../data/checker/nullable_wrap_narrowed_base.yaml", "../data/checker/nullable_wrap_narrowed_revision.yaml")
	require.Contains(t, ids, checker.RequestPropertyEnumValueRemovedId)
	require.Contains(t, ids, checker.RequestPropertyBecomeNullableId)
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
		require.Equal(t, []string{checker.RequestParameterBecameNullableId}, ids)
	})
	t.Run("header parameter", func(t *testing.T) {
		ids := nullableWrapChanges(t, "../data/checker/nullable_wrap_header_property_base.yaml", "../data/checker/nullable_wrap_header_property_revision.yaml")
		require.ElementsMatch(t, []string{
			checker.RequestParameterBecameNullableId,         // X-Owner
			checker.RequestParameterBecameNullableId,         // X-Status
			checker.RequestParameterPropertyBecameNullableId, // X-Meta owner
			checker.RequestParameterPropertyBecameNullableId, // X-Meta status
		}, ids)
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
		require.Equal(t, []string{checker.RequestParameterBecameNotNullableId}, ids)
	})
	t.Run("header parameter", func(t *testing.T) {
		ids := nullableWrapChanges(t, "../data/checker/nullable_wrap_header_property_revision.yaml", "../data/checker/nullable_wrap_header_property_base.yaml")
		require.ElementsMatch(t, []string{
			checker.RequestParameterBecameNotNullableId,         // X-Owner
			checker.RequestParameterBecameNotNullableId,         // X-Status
			checker.RequestParameterPropertyBecameNotNullableId, // X-Meta owner
			checker.RequestParameterPropertyBecameNotNullableId, // X-Meta status
		}, ids)
	})
	t.Run("response property", func(t *testing.T) {
		ids := nullableWrapChanges(t, "../data/checker/nullable_wrap_response_revision.yaml", "../data/checker/nullable_wrap_response_base.yaml")
		require.Equal(t, []string{checker.ResponsePropertyBecameNotNullableId}, ids)
	})
}

// A parameter or parameter property losing the "null" type is breaking (the
// request no longer accepts null); gaining it is informational.
func TestParameterTypeArrayNullability(t *testing.T) {
	t.Run("became not nullable", func(t *testing.T) {
		ids := nullableWrapChanges(t, "../data/checker/nullable_type_array_param_base.yaml", "../data/checker/nullable_type_array_param_revision.yaml")
		require.ElementsMatch(t, []string{
			checker.RequestParameterBecameNotNullableId,         // q
			checker.RequestParameterPropertyBecameNotNullableId, // filter.tag
		}, ids)
	})
	t.Run("became nullable", func(t *testing.T) {
		ids := nullableWrapChanges(t, "../data/checker/nullable_type_array_param_revision.yaml", "../data/checker/nullable_type_array_param_base.yaml")
		require.ElementsMatch(t, []string{
			checker.RequestParameterBecameNullableId,
			checker.RequestParameterPropertyBecameNullableId,
		}, ids)
	})
}

// wrappedRefChange compares the wrapping fixtures after an edit to each side,
// and returns the nullability change reported for optionalRef, the property
// that refers to SomeObject in the base and wraps it in the revision.
func wrappedRefChange(t *testing.T, editBase, editRevision func(spec *openapi3.T)) *checker.ApiChange {
	t.Helper()
	s1, err := open("../data/checker/nullable_wrap_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/nullable_wrap_revision.yaml")
	require.NoError(t, err)
	editBase(s1.Spec)
	editRevision(s2.Spec)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyBecameNotNullableCheck), d, osm, checker.INFO)
	for _, e := range errs {
		if change := e.(checker.ApiChange); change.Args[0] == "optionalRef" {
			return &change
		}
	}
	return nil
}

func someObject(spec *openapi3.T) *openapi3.Schema {
	return spec.Components.Schemas["SomeObject"].Value
}

func untyped(spec *openapi3.T) { someObject(spec).Type = nil }

func unedited(*openapi3.T) {}

// A schema with no type already accepts null, so wrapping it makes null match
// both branches: the property stops accepting null, and the comment says why.
func TestNullableWrapping_UntypedSchemaStopsAcceptingNull(t *testing.T) {
	change := wrappedRefChange(t, untyped, untyped)
	require.NotNil(t, change)
	require.Equal(t, checker.RequestPropertyBecomeNotNullableId, change.Id)
	require.Equal(t, checker.ComposedNullabilityUntypedBranchCommentId, change.Comment)
}

// Giving the wrapped schema a type afterwards is what makes the property
// nullable: one branch accepts null now, where two did.
func TestNullableWrapping_TypingTheBranchMakesItNullable(t *testing.T) {
	wrapped := func(spec *openapi3.T) {
		revision, err := open("../data/checker/nullable_wrap_revision.yaml")
		require.NoError(t, err)
		request := spec.Components.Schemas["TestRequest"].Value
		request.Properties["optionalRef"] = &openapi3.SchemaRef{Value: &openapi3.Schema{OneOf: openapi3.SchemaRefs{
			revision.Spec.Components.Schemas["TestRequest"].Value.Properties["optionalRef"].Value.OneOf[0],
			{Ref: "#/components/schemas/SomeObject", Value: someObject(spec)},
		}}}
		untyped(spec)
	}
	change := wrappedRefChange(t, wrapped, unedited)
	require.NotNil(t, change)
	require.Equal(t, checker.RequestPropertyBecomeNullableId, change.Id)
	require.Equal(t, checker.ComposedNullabilityUntypedBranchCommentId, change.Comment)
}

// What a wrapped schema with an if accepts is not decided, and the wrapping
// itself is still recognized.
func TestNullableWrapping_UndecidedSchemaStillRecognized(t *testing.T) {
	withIf := func(spec *openapi3.T) {
		someObject(spec).If = &openapi3.SchemaRef{Value: openapi3.NewObjectSchema()}
	}
	change := wrappedRefChange(t, withIf, withIf)
	require.NotNil(t, change)
	require.Equal(t, checker.RequestPropertyBecomeNullableId, change.Id)
	require.Empty(t, change.Comment)
}

// The OpenAPI 3.0 way of saying a composition is nullable: nullable: true
// beside a oneOf whose branches do not accept null. A nullable property that
// becomes one keeps accepting null.
func TestNullableKeywordBesideOneOf_StaysNullable(t *testing.T) {
	nullableString := func(spec *openapi3.T) {
		spec.Components.Schemas["TestRequest"].Value.Properties["optionalRef"] = &openapi3.SchemaRef{
			Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Nullable: true},
		}
	}
	nullableOneOf := func(spec *openapi3.T) {
		spec.Components.Schemas["TestRequest"].Value.Properties["optionalRef"] = &openapi3.SchemaRef{
			Value: &openapi3.Schema{Nullable: true, OneOf: openapi3.SchemaRefs{
				{Value: openapi3.NewStringSchema()},
				{Value: openapi3.NewObjectSchema()},
			}},
		}
	}
	require.Nil(t, wrappedRefChange(t, nullableString, nullableOneOf))
}
