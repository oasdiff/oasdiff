package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// decreasing request body maxProperties
func TestRequestBodyMaxPropertiesDecreased(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := uint64(10)
	to := uint64(2)
	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.MaxProps = &from
	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.MaxProps = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMaxPropertiesUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyMaxPropertiesDecreasedId,
		Args:        []any{uint64(2)},
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}

// decreasing request property maxProperties
func TestRequestPropertyMaxPropertiesDecreased(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := uint64(10)
	to := uint64(2)
	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["filter"].Value.MaxProps = &from
	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["filter"].Value.MaxProps = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMaxPropertiesUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyMaxPropertiesDecreasedId,
		Args:        []any{"filter", uint64(2)},
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}

// decreasing read-only request property maxProperties
func TestRequestReadOnlyPropertyMaxPropertiesDecreased(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := uint64(10)
	to := uint64(2)
	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["filter"].Value.ReadOnly = true
	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["filter"].Value.ReadOnly = true
	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["filter"].Value.MaxProps = &from
	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["filter"].Value.MaxProps = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMaxPropertiesUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestReadOnlyPropertyMaxPropertiesDecreasedId,
		Args:        []any{"filter", uint64(2)},
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}

// increasing request property maxProperties
func TestRequestPropertyMaxPropertiesIncreased(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := uint64(2)
	to := uint64(10)
	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["filter"].Value.MaxProps = &from
	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["filter"].Value.MaxProps = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMaxPropertiesUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyMaxPropertiesIncreasedId,
		Args:        []any{"filter", uint64(2), uint64(10)},
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}
