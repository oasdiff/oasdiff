package allof_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

func Test_MergeSpecOK(t *testing.T) {
	spec, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../../data/allof/simple.yaml"), load.WithFlattenAllOf())
	require.NoError(t, err)
	merged := spec.Spec
	require.NoError(t, err)
	require.True(t, merged.Components.Schemas["GroupView"].Value.Properties["created"].Value.Type.Is("string"))
	require.True(t, merged.Components.Parameters["groupId"].Value.Schema.Value.Properties["prop1"].Value.Type.Is("string"))
	require.True(t, merged.Components.Parameters["groupId"].Value.Schema.Value.Properties["prop2"].Value.Type.Is("boolean"))
	require.Empty(t, merged.Components.Parameters["groupId"].Value.Schema.Value.AllOf)
	require.True(t, merged.Paths.Value("/api/v1.0/groups").Patch.RequestBody.Value.Content["application/json"].Schema.Value.Properties["prop1"].Value.Type.Is("string"))
	require.True(t, merged.Paths.Value("/api/v1.0/groups").Patch.RequestBody.Value.Content["application/json"].Schema.Value.Properties["prop2"].Value.Type.Is("boolean"))
	require.Empty(t, merged.Paths.Value("/api/v1.0/groups").Patch.RequestBody.Value.Content["application/json"].Schema.Value.AllOf)
}

func Test_MergeSpecInvalid(t *testing.T) {
	_, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../../data/allof/invalid.yaml"), load.WithFlattenAllOf())
	require.EqualError(t, err, "failed to flatten allOf in \"../../data/allof/invalid.yaml\": unable to resolve Type conflict: all Type values must be identical")
}

// A 3.1 multi-type property inside an allOf branch flattens without a
// spurious Type conflict (https://github.com/oasdiff/oasdiff/issues/1078).
func Test_MergeSpecMultiType(t *testing.T) {
	spec, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("testdata/multi_type.yaml"), load.WithFlattenAllOf())
	require.NoError(t, err)

	merged := spec.Spec.Components.Schemas["LimitHitsByDateRequest"].Value
	require.Empty(t, merged.AllOf)
	require.True(t, merged.Properties["fromDate"].Value.Type.Is("string"))
	require.Equal(t, &openapi3.Types{"integer", "null"}, merged.Properties["nullablePrimitive"].Value.Type)
}

func TestMergeSpec_CircularAdditionalPropsWithoutAllOf(t *testing.T) {
	spec, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("testdata/circular_additional_props1.yaml"), load.WithFlattenAllOf())
	require.NoError(t, err)

	merged := spec.Spec
	require.True(t, merged.Components.Schemas["BaseSchema"].Value.Properties["prop1"].Value.Type.Is("object"))
	require.NotNil(t, merged.Components.Schemas["BaseSchema"].Value.Properties["prop1"].Value.AdditionalProperties.Schema)
	require.NotNil(t, merged.Components.Schemas["BaseSchema"].Value.Properties["prop1"].Value.AdditionalProperties.Schema.Value)

	baseSchema := merged.Components.Schemas["BaseSchema"].Value
	referencedAdditionalPropSchema := merged.Components.Schemas["BaseSchema"].Value.Properties["prop1"].Value.AdditionalProperties.Schema.Value
	require.Equal(t, baseSchema, referencedAdditionalPropSchema)
}

func TestMergeSpec_MergeCircularAdditionalPropsWithAllOf(t *testing.T) {
	spec, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("testdata/circular_additional_props2.yaml"), load.WithFlattenAllOf())
	require.NoError(t, err)

	merged := spec.Spec
	require.Nil(t, merged.Components.Schemas["BaseSchema"].Value.AllOf)
	require.True(t, merged.Components.Schemas["BaseSchema"].Value.Properties["fixedProperty"].Value.Type.Is("string"))
	require.True(t, merged.Components.Schemas["BaseSchema"].Value.Properties["prop1"].Value.Type.Is("object"))
	require.NotNil(t, merged.Components.Schemas["BaseSchema"].Value.Properties["prop1"].Value.AdditionalProperties.Schema)
	require.NotNil(t, merged.Components.Schemas["BaseSchema"].Value.Properties["prop1"].Value.AdditionalProperties.Schema.Value)

	baseSchema := merged.Components.Schemas["BaseSchema"].Value
	referencedAdditionalPropSchema := merged.Components.Schemas["BaseSchema"].Value.Properties["prop1"].Value.AdditionalProperties.Schema.Value
	require.Equal(t, baseSchema, referencedAdditionalPropSchema)
}

func TestMergeSpec_MergeCircularAdditionalPropsNestedWithinAllOf(t *testing.T) {
	spec, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("testdata/circular_additional_props3.yaml"), load.WithFlattenAllOf())
	require.NoError(t, err)

	merged := spec.Spec
	require.Nil(t, merged.Components.Schemas["BaseSchema"].Value.AllOf)
	require.True(t, merged.Components.Schemas["BaseSchema"].Value.Properties["prop1"].Value.Type.Is("object"))
	require.NotNil(t, merged.Components.Schemas["BaseSchema"].Value.Properties["prop1"].Value.AdditionalProperties.Schema)
	require.NotNil(t, merged.Components.Schemas["BaseSchema"].Value.Properties["prop1"].Value.AdditionalProperties.Schema.Value)

	baseSchemaReferencedAdditionalPropSchema := merged.Components.Schemas["BaseSchema"].Value.Properties["prop1"].Value.AdditionalProperties.Schema.Value
	NestedSelfReferentialSchema := merged.Components.Schemas["NestedSelfReferentialSchema"].Value
	require.Equal(t, baseSchemaReferencedAdditionalPropSchema, NestedSelfReferentialSchema)
}

// allOf under a webhook is merged like everywhere else; the walk covers every
// schema attachment point in the document.
func Test_MergeSpecWebhook(t *testing.T) {
	spec, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../../data/allof/webhook.yaml"), load.WithFlattenAllOf())
	require.NoError(t, err)
	schema := spec.Spec.Webhooks["newPet"].Post.RequestBody.Value.Content["application/json"].Schema.Value
	require.Empty(t, schema.AllOf)
	require.True(t, schema.Properties["a"].Value.Type.Is("string"))
	require.True(t, schema.Properties["b"].Value.Type.Is("integer"))
}

// A schema used through a $ref must be merged where it is used, not only where
// it is defined. Each $ref is a separate SchemaRef sharing one Value, so a
// merge that repointed the definition's SchemaRef left every use unmerged, and
// the allOf reached the diff even though --flatten-allof was given.
func TestMergeSpec_SchemaReachedThroughRef(t *testing.T) {
	spec, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../../data/allof/ref-to-allof.yaml"), load.WithFlattenAllOf())
	require.NoError(t, err)

	used := spec.Spec.Paths.Value("/pets").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value
	require.Empty(t, used.AllOf, "the allOf survived at the point of use")
	require.ElementsMatch(t, []string{"id", "name"}, used.Required)
	require.Contains(t, used.Properties, "id")
	require.Contains(t, used.Properties, "name")
}

// Merging `allOf: [$ref Node, <inline overlay whose items is $ref Node>]`,
// where Node is recursive: both branches' items merge to the same schema,
// so the merge must reuse it, $ref intact, and the flattened spec must
// marshal.
func Test_MergeSpec_RecursiveOverlaySerializes(t *testing.T) {
	spec, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../../data/allof/circular-overlay.yaml"), load.WithFlattenAllOf())
	require.NoError(t, err)

	merged := spec.Spec.Paths.Value("/x").Post.RequestBody.Value.Content["application/json"].
		Schema.Value.Properties["filters"].Value.Items.Value
	require.Empty(t, merged.AllOf)
	require.Equal(t, "#/components/schemas/Node",
		merged.Properties["filters"].Value.Items.Ref)

	_, err = spec.Spec.MarshalJSON()
	require.NoError(t, err)
}

// An allOf over two distinct recursive components merges to a node whose
// recursion the in-flight guard anchored without a $ref; the anchor's value
// is right but a ref-less cycle has no serialized form. MergeSpec names the
// anchored target, hoisting it into components, so the merged spec marshals
// and the name is stable across runs.
func Test_MergeSpec_TwoRecursiveBranchesSerializes(t *testing.T) {
	loadMerged := func() *openapi3.T {
		spec, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../../data/allof/two-recursive-branches.yaml"), load.WithFlattenAllOf())
		require.NoError(t, err)
		return spec.Spec
	}
	merged := loadMerged()

	tree := merged.Paths.Value("/x").Post.RequestBody.Value.Content["application/json"].
		Schema.Value.Properties["tree"]
	require.Empty(t, tree.Value.AllOf)

	child := tree.Value.Properties["child"]
	require.Equal(t, "#/components/schemas/AllOfMerged1", child.Ref)
	require.Same(t, tree.Value, child.Value)
	require.Same(t, tree.Value, merged.Components.Schemas["AllOfMerged1"].Value)

	first, err := merged.MarshalJSON()
	require.NoError(t, err)
	second, err := loadMerged().MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, string(first), string(second))
}
