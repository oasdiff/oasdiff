package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// increasing request body minProperties
func TestRequestBodyMinPropertiesIncreased(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.MinProps = uint64(1)
	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.MinProps = uint64(3)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMinPropertiesIncreasedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyMinPropertiesIncreasedId,
		Args:        []any{uint64(3)},
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}

// increasing request property minProperties
func TestRequestPropertyMinPropertiesIncreased(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["filter"].Value.MinProps = uint64(1)
	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["filter"].Value.MinProps = uint64(3)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMinPropertiesIncreasedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyMinPropertiesIncreasedId,
		Args:        []any{"filter", uint64(3)},
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}
