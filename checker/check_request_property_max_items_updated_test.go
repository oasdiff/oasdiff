package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

const constraintKeywordsBase = "../data/checker/constraint_keywords_base.yaml"

// decreasing request body maxItems
func TestRequestBodyMaxItemsDecreased(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := uint64(100)
	to := uint64(10)
	s1.Spec.Paths.Value("/tags").Put.RequestBody.Value.Content["application/json"].Schema.Value.MaxItems = &from
	s2.Spec.Paths.Value("/tags").Put.RequestBody.Value.Content["application/json"].Schema.Value.MaxItems = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMaxItemsUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyMaxItemsDecreasedId,
		Args:        []any{uint64(10)},
		Operation:   "PUT",
		OperationId: "setTags",
		Path:        "/tags",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}

// decreasing request property maxItems
func TestRequestPropertyMaxItemsDecreased(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := uint64(100)
	to := uint64(10)
	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["tags"].Value.MaxItems = &from
	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["tags"].Value.MaxItems = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMaxItemsUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyMaxItemsDecreasedId,
		Args:        []any{"tags", uint64(10)},
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}

// decreasing read-only request property maxItems
func TestRequestReadOnlyPropertyMaxItemsDecreased(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := uint64(100)
	to := uint64(10)
	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["tags"].Value.ReadOnly = true
	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["tags"].Value.ReadOnly = true
	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["tags"].Value.MaxItems = &from
	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["tags"].Value.MaxItems = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMaxItemsUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestReadOnlyPropertyMaxItemsDecreasedId,
		Args:        []any{"tags", uint64(10)},
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}

// increasing request property maxItems
func TestRequestPropertyMaxItemsIncreased(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := uint64(10)
	to := uint64(100)
	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["tags"].Value.MaxItems = &from
	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["tags"].Value.MaxItems = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMaxItemsUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyMaxItemsIncreasedId,
		Args:        []any{"tags", uint64(10), uint64(100)},
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}
