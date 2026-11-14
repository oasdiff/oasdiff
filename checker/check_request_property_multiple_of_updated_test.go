package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// setting request body multipleOf
func TestRequestBodyMultipleOfSet(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	to := 5.0
	s2.Spec.Paths.Value("/price").Put.RequestBody.Value.Content["application/json"].Schema.Value.MultipleOf = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMultipleOfUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyMultipleOfSetId,
		Args:        []any{5.0},
		Comment:     "bound-set-comment",
		Operation:   "PUT",
		OperationId: "setPrice",
		Path:        "/price",
	}, errs)
}

// setting request property multipleOf
func TestRequestPropertyMultipleOfSet(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	to := 5.0
	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["amount"].Value.MultipleOf = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMultipleOfUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyMultipleOfSetId,
		Args:        []any{"amount", 5.0},
		Comment:     "bound-set-comment",
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
	}, errs)
}

// unsetting request body multipleOf
func TestRequestBodyMultipleOfUnset(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := 5.0
	s1.Spec.Paths.Value("/price").Put.RequestBody.Value.Content["application/json"].Schema.Value.MultipleOf = &from

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMultipleOfUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyMultipleOfUnsetId,
		Args:        []any{5.0},
		Operation:   "PUT",
		OperationId: "setPrice",
		Path:        "/price",
	}, errs)
}

// unsetting request property multipleOf
func TestRequestPropertyMultipleOfUnset(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := 5.0
	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["amount"].Value.MultipleOf = &from

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMultipleOfUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyMultipleOfUnsetId,
		Args:        []any{"amount", 5.0},
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
	}, errs)
}

// changing request body multipleOf incompatibly
func TestRequestBodyMultipleOfChanged(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := 5.0
	to := 3.0
	s1.Spec.Paths.Value("/price").Put.RequestBody.Value.Content["application/json"].Schema.Value.MultipleOf = &from
	s2.Spec.Paths.Value("/price").Put.RequestBody.Value.Content["application/json"].Schema.Value.MultipleOf = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMultipleOfUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyMultipleOfChangedId,
		Args:        []any{5.0, 3.0},
		Operation:   "PUT",
		OperationId: "setPrice",
		Path:        "/price",
	}, errs)
}

// changing request property multipleOf incompatibly
func TestRequestPropertyMultipleOfChanged(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := 5.0
	to := 3.0
	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["amount"].Value.MultipleOf = &from
	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["amount"].Value.MultipleOf = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMultipleOfUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyMultipleOfChangedId,
		Args:        []any{"amount", 5.0, 3.0},
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
	}, errs)
}

// generalizing request body multipleOf
func TestRequestBodyMultipleOfGeneralized(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := 5.0
	to := 2.5
	s1.Spec.Paths.Value("/price").Put.RequestBody.Value.Content["application/json"].Schema.Value.MultipleOf = &from
	s2.Spec.Paths.Value("/price").Put.RequestBody.Value.Content["application/json"].Schema.Value.MultipleOf = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMultipleOfUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyMultipleOfGeneralizedId,
		Args:        []any{5.0, 2.5},
		Comment:     checker.RequestBodyMultipleOfGeneralizedId + "-comment",
		Operation:   "PUT",
		OperationId: "setPrice",
		Path:        "/price",
	}, errs)
}

// generalizing request property multipleOf
func TestRequestPropertyMultipleOfGeneralized(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := 5.0
	to := 2.5
	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["amount"].Value.MultipleOf = &from
	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["amount"].Value.MultipleOf = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMultipleOfUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyMultipleOfGeneralizedId,
		Args:        []any{"amount", 5.0, 2.5},
		Comment:     checker.RequestPropertyMultipleOfGeneralizedId + "-comment",
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
	}, errs)
}
