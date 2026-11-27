package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// changing request property required value to true
func TestRequestPropertyMarkedRequired(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Required = []string{""}
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyRequiredUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyBecameRequiredId,
		Args:        []any{"name"},
		Operation:   "POST",
		Path:        "/products",
		OperationId: "addProduct",
	}, errs)
}

// changing request property required value to false
func TestRequestPropertyMarkedOptional(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Required = []string{""}
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyRequiredUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyBecameOptionalId,
		Args:        []any{"name"},
		Operation:   "POST",
		Path:        "/products",
		OperationId: "addProduct",
	}, errs)
}

// making request property required, while also giving it a default value
func TestRequestPropertyWithDefaultMarkedRequired(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Required = []string{""}
	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value.Default = "default"
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyRequiredUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyBecameRequiredWithDefaultId,
		Args:        []any{"name"},
		Comment:     checker.RequiredRequestPropertyWithDefaultCommentId,
		Operation:   "POST",
		Path:        "/products",
		OperationId: "addProduct",
	}, errs)
}

func TestRequestPropertyRequiredAcrossAllOf(t *testing.T) {
	inherited := openapi3.NewObjectSchema().
		WithProperty("parts", openapi3.NewStringSchema()).
		WithProperty("size", openapi3.NewIntegerSchema().WithDefault(0)).
		WithProperty("id", openapi3.NewStringSchema()).
		WithProperty("always", openapi3.NewStringSchema())
	inherited.Properties["id"].Value.ReadOnly = true
	inherited.Required = []string{"always"}
	schema := func(required []string) *openapi3.Schema {
		// The first declaration of size has no default. Every allOf branch applies.
		result := openapi3.NewAllOfSchema(openapi3.NewObjectSchema().WithProperty("size", openapi3.NewIntegerSchema()), openapi3.NewAllOfSchema(inherited))
		result.AnyOf = openapi3.SchemaRefs{{Value: openapi3.NewObjectSchema().WithProperty("choice", openapi3.NewStringSchema())}, {Value: openapi3.NewObjectSchema()}}
		result.Required = required
		return result
	}

	for _, tc := range []struct {
		name           string
		base, revision []string
		wantID         string
	}{
		{"became required", nil, []string{"parts"}, checker.RequestPropertyBecameRequiredId},
		{"became optional", []string{"parts"}, nil, checker.RequestPropertyBecameOptionalId},
		{"inherited default", nil, []string{"size"}, checker.RequestPropertyBecameRequiredWithDefaultId},
		{"read-only", nil, []string{"id"}, ""},
		{"already required by a branch", nil, []string{"always"}, ""},
		{"still required by a branch", []string{"always"}, nil, ""},
		{"only in an alternative", nil, []string{"choice"}, ""},
		{"undeclared", nil, []string{"unknown"}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base := positionDoc("request-property", schema(tc.base))
			revision := positionDoc("request-property", schema(tc.revision))
			report, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
			require.NoError(t, err)
			changes := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyRequiredUpdatedCheck), report, sources, checker.INFO)
			if tc.wantID == "" {
				require.Empty(t, changes)
				return
			}
			requireSingleChange(t, changes, tc.wantID)
		})
	}
}

func TestRequestPropertyRequiredInParentAndBranch(t *testing.T) {
	// The outer list and a branch change together. Only a branch that also
	// declares the property can report the change at its own location.
	for _, tc := range []struct {
		name                 string
		declarationInSibling bool
	}{
		{"branch declares the property", false},
		{"sibling declares the property", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			schema := func(required bool) *openapi3.Schema {
				branch, sibling := openapi3.NewObjectSchema(), openapi3.NewObjectSchema()
				owner := branch
				if tc.declarationInSibling {
					owner = sibling
				}
				owner.WithProperty("parts", openapi3.NewStringSchema())
				result := openapi3.NewAllOfSchema(openapi3.NewAllOfSchema(branch), sibling)
				if required {
					branch.Required = []string{"parts"}
					result.Required = []string{"parts"}
				}
				return result
			}
			optional := positionDoc("request-body", schema(false))
			required := positionDoc("request-body", schema(true))
			for _, pair := range []struct {
				base, revision *load.SpecInfo
				wantID         string
			}{
				{optional, required, checker.RequestPropertyBecameRequiredId},
				{required, optional, checker.RequestPropertyBecameOptionalId},
			} {
				report, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), pair.base, pair.revision)
				require.NoError(t, err)
				changes := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyRequiredUpdatedCheck), report, sources, checker.INFO)
				requireSingleChange(t, changes, pair.wantID)
			}
		})
	}
}
