package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// increasing response body maxProperties
func TestResponseBodyMaxPropertiesIncreased(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := uint64(2)
	to := uint64(10)
	s1.Spec.Paths.Value("/products").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.MaxProps = &from
	s2.Spec.Paths.Value("/products").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.MaxProps = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyMaxPropertiesIncreasedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponseBodyMaxPropertiesIncreasedId,
		Args:        []any{uint64(2), uint64(10)},
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}

// increasing response property maxProperties
func TestResponsePropertyMaxPropertiesIncreased(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := uint64(2)
	to := uint64(10)
	s1.Spec.Paths.Value("/products").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["filter"].Value.MaxProps = &from
	s2.Spec.Paths.Value("/products").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["filter"].Value.MaxProps = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyMaxPropertiesIncreasedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponsePropertyMaxPropertiesIncreasedId,
		Args:        []any{"filter", uint64(2), uint64(10), "200"},
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}
