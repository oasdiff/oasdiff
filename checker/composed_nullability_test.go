package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

const (
	oneOfNullBranchBase     = "../data/checker/one_of_null_branch_base.yaml"
	oneOfNullBranchRevision = "../data/checker/one_of_null_branch_revision.yaml"
)

// oneOfNullBranchSpecs opens the fixture pair: a property whose schema is
// oneOf: [{type: "null"}, $ref], with the referenced schema plain in the base
// and nullable in the revision.
func oneOfNullBranchSpecs(t *testing.T) (*load.SpecInfo, *load.SpecInfo) {
	t.Helper()
	base, err := open(oneOfNullBranchBase)
	require.NoError(t, err)
	revision, err := open(oneOfNullBranchRevision)
	require.NoError(t, err)
	return base, revision
}

// addressSchemas returns the request and response property schemas holding the
// composition, in both specs.
func addressSchemas(specs ...*load.SpecInfo) []*openapi3.Schema {
	result := make([]*openapi3.Schema, 0, 2*len(specs))
	for _, spec := range specs {
		post := spec.Spec.Paths.Value("/products").Post
		result = append(result,
			post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["address"].Value,
			post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["address"].Value,
		)
	}
	return result
}

// A oneOf accepts null only while exactly one of its branches does, so making
// the second branch nullable takes null out of the request property: breaking,
// where reading the branch on its own says it became nullable.
func TestOneOfNullBranch_RequestPropertyStopsAcceptingNull(t *testing.T) {
	base, revision := oneOfNullBranchSpecs(t)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestPropertyBecameNotNullableCheck), d, osm)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyBecomeNotNullableId,
		Args:        []any{"address"},
		Comment:     checker.ComposedNullabilityOneOfCommentId,
		Operation:   "POST",
		Path:        "/products",
		OperationId: "addProduct",
	}, errs)
}

// The reverse edit: one of two nullable branches became non-nullable, so the
// request property accepts null again.
func TestOneOfNullBranch_RequestPropertyAcceptsNullAgain(t *testing.T) {
	base, revision := oneOfNullBranchSpecs(t)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), revision, base)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyBecameNotNullableCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyBecomeNullableId,
		Args:        []any{"address"},
		Comment:     checker.ComposedNullabilityOneOfCommentId,
		Operation:   "POST",
		Path:        "/products",
		OperationId: "addProduct",
	}, errs)
}

// On the response side the same edit takes null out of what clients receive,
// which is safe.
func TestOneOfNullBranch_ResponsePropertyStopsReturningNull(t *testing.T) {
	base, revision := oneOfNullBranchSpecs(t)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyBecameNullableCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponsePropertyBecameNotNullableId,
		Args:        []any{"address", "200"},
		Comment:     checker.ComposedNullabilityOneOfCommentId,
		Operation:   "POST",
		Path:        "/products",
		OperationId: "addProduct",
	}, errs)
}

// And the reverse puts null back in the response, which clients may not
// handle: breaking.
func TestOneOfNullBranch_ResponsePropertyReturnsNullAgain(t *testing.T) {
	base, revision := oneOfNullBranchSpecs(t)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), revision, base)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.ResponsePropertyBecameNullableCheck), d, osm)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponsePropertyBecameNullableId,
		Args:        []any{"address", "200"},
		Comment:     checker.ComposedNullabilityOneOfCommentId,
		Operation:   "POST",
		Path:        "/products",
		OperationId: "addProduct",
	}, errs)
}

// An anyOf accepts null when any branch does, so a null branch keeps the
// property nullable whatever the other branch does: nothing changed.
func TestAnyOfNullBranch_NullabilityUnchanged(t *testing.T) {
	base, revision := oneOfNullBranchSpecs(t)
	for _, schema := range addressSchemas(base, revision) {
		schema.AnyOf, schema.OneOf = schema.OneOf, nil
	}

	for _, checkFn := range []checker.BackwardCompatibilityCheck{
		checker.RequestPropertyBecameNotNullableCheck,
		checker.ResponsePropertyBecameNullableCheck,
	} {
		d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
		require.NoError(t, err)
		require.Empty(t, checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checkFn), d, osm, checker.INFO))

		d, osm, err = diff.GetWithOperationsSourcesMap(diff.NewConfig(), revision, base)
		require.NoError(t, err)
		require.Empty(t, checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checkFn), d, osm, checker.INFO))
	}
}

// With no other nullable branch, the oneOf follows the branch that changed: it
// now accepts null, which needs no explaining, so the change carries no
// comment.
func TestOneOfWithoutNullBranch_FollowsTheBranch(t *testing.T) {
	base, revision := oneOfNullBranchSpecs(t)
	for _, schema := range addressSchemas(base, revision) {
		schema.OneOf[0].Value.Type = &openapi3.Types{"string"}
	}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyBecameNotNullableCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyBecomeNullableId,
		Args:        []any{"address"},
		Operation:   "POST",
		Path:        "/products",
		OperationId: "addProduct",
	}, errs)
}

// An allOf alongside the oneOf constrains the instance too, and what it
// accepts is not decided here, so the branch keeps reporting its own
// nullability rather than nothing being reported at all.
func TestOneOfNullBranch_AllOfLeavesTheBranchReporting(t *testing.T) {
	base, revision := oneOfNullBranchSpecs(t)
	for _, schema := range addressSchemas(base, revision) {
		schema.AllOf = openapi3.SchemaRefs{{Value: openapi3.NewObjectSchema()}}
	}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyBecameNotNullableCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertyBecomeNullableId, errs[0].GetId())
	require.Equal(t, "address/oneOf[#/components/schemas/RequestAddress]", errs[0].(checker.ApiChange).Args[0])
}

// requestNullabilityAfter compares the base fixture with a copy of it whose
// second oneOf branch, the referenced schema, was edited, and returns what
// the request nullability check reports.
func requestNullabilityAfter(t *testing.T, edit func(branch *openapi3.Schema)) checker.Changes {
	t.Helper()
	base, err := open(oneOfNullBranchBase)
	require.NoError(t, err)
	revision, err := open(oneOfNullBranchBase)
	require.NoError(t, err)
	edit(addressSchemas(revision)[0].OneOf[1].Value)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	return checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyBecameNotNullableCheck), d, osm, checker.INFO)
}

var requestPropertyStoppedAcceptingNull = checker.ApiChange{
	Id:          checker.RequestPropertyBecomeNotNullableId,
	Args:        []any{"address"},
	Comment:     checker.ComposedNullabilityOneOfCommentId,
	Operation:   "POST",
	Path:        "/products",
	OperationId: "addProduct",
}

// untypedBranch is the change carrying the comment for a oneOf that rejects
// null because a branch has no type.
func untypedBranch(change checker.ApiChange) checker.ApiChange {
	change.Comment = checker.ComposedNullabilityUntypedBranchCommentId
	return change
}

// A branch with no type accepts any value, null included, so dropping the
// type from the second branch makes null match both branches.
func TestOneOfNullBranch_UntypedBranchAcceptsNull(t *testing.T) {
	errs := requestNullabilityAfter(t, func(branch *openapi3.Schema) {
		branch.Type = nil
	})
	requireSingleApiChange(t, untypedBranch(requestPropertyStoppedAcceptingNull), errs)
}

// An enum decides for an untyped branch: null matches it only when null is
// one of the values.
func TestOneOfNullBranch_UntypedEnumBranch(t *testing.T) {
	require.Empty(t, requestNullabilityAfter(t, func(branch *openapi3.Schema) {
		branch.Type = nil
		branch.Enum = []any{"home"}
	}))

	errs := requestNullabilityAfter(t, func(branch *openapi3.Schema) {
		branch.Type = nil
		branch.Enum = []any{"home", nil}
	})
	requireSingleApiChange(t, untypedBranch(requestPropertyStoppedAcceptingNull), errs)
}

// A const other than null keeps null out of an untyped branch.
func TestOneOfNullBranch_UntypedConstBranch(t *testing.T) {
	require.Empty(t, requestNullabilityAfter(t, func(branch *openapi3.Schema) {
		branch.Type = nil
		branch.Const = "home"
	}))
}

// An enum without null keeps null out of a branch whose type allows it.
func TestOneOfNullBranch_EnumOverridesNullableType(t *testing.T) {
	require.Empty(t, requestNullabilityAfter(t, func(branch *openapi3.Schema) {
		branch.Type = &openapi3.Types{"string", "null"}
		branch.Enum = []any{"home"}
	}))
}

// What a branch with if/then/else accepts is not decided here, so the oneOf
// reports nothing about the property.
func TestOneOfNullBranch_ConditionalBranchIsNotDecided(t *testing.T) {
	errs := requestNullabilityAfter(t, func(branch *openapi3.Schema) {
		branch.Type = nil
		branch.If = &openapi3.SchemaRef{Value: openapi3.NewObjectSchema()}
	})
	for _, err := range errs {
		require.NotEqual(t, "address", err.(checker.ApiChange).Args[0])
	}
}

// In an OpenAPI 3.0 document null is accepted only where nullable: true says
// so, so a branch with no type does not accept it. A oneOf with a nullable
// branch and a branch with no type accepts null, and stops when the first
// loses its nullable.
func TestOneOfNullableBranch_OpenAPI30UntypedBranchRejectsNull(t *testing.T) {
	base, err := open("../data/checker/one_of_nullable_branch_30_base.yaml")
	require.NoError(t, err)
	revision, err := open("../data/checker/one_of_nullable_branch_30_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestPropertyBecameNotNullableCheck), d, osm)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyBecomeNotNullableId,
		Args:        []any{"address"},
		Operation:   "POST",
		Path:        "/products",
		OperationId: "addProduct",
	}, errs)

	d, osm, err = diff.GetWithOperationsSourcesMap(diff.NewConfig(), revision, base)
	require.NoError(t, err)
	errs = checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyBecameNotNullableCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyBecomeNullableId,
		Args:        []any{"address"},
		Operation:   "POST",
		Path:        "/products",
		OperationId: "addProduct",
	}, errs)
}
