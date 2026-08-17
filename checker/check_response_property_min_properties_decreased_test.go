package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// decreasing response body minProperties
func TestResponseBodyMinPropertiesDecreased(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	s1.Spec.Paths.Value("/products").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.MinProps = uint64(5)
	s2.Spec.Paths.Value("/products").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.MinProps = uint64(2)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyMinPropertiesDecreasedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponseBodyMinPropertiesDecreasedId,
		Args:        []any{uint64(5), uint64(2)},
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}

// decreasing response property minProperties
func TestResponsePropertyMinPropertiesDecreased(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	s1.Spec.Paths.Value("/products").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["filter"].Value.MinProps = uint64(5)
	s2.Spec.Paths.Value("/products").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["filter"].Value.MinProps = uint64(2)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyMinPropertiesDecreasedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponsePropertyMinPropertiesDecreasedId,
		Args:        []any{"filter", uint64(5), uint64(2), "200"},
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}
