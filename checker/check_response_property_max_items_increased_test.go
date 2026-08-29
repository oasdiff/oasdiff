package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// increasing response body maxItems
func TestResponseBodyMaxItemsIncreased(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := uint64(10)
	to := uint64(100)
	s1.Spec.Paths.Value("/tags").Put.Responses.Value("200").Value.Content["application/json"].Schema.Value.MaxItems = &from
	s2.Spec.Paths.Value("/tags").Put.Responses.Value("200").Value.Content["application/json"].Schema.Value.MaxItems = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyMaxItemsIncreasedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponseBodyMaxItemsIncreasedId,
		Args:        []any{uint64(10), uint64(100)},
		Operation:   "PUT",
		OperationId: "setTags",
		Path:        "/tags",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}

// increasing response property maxItems
func TestResponsePropertyMaxItemsIncreased(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := uint64(10)
	to := uint64(100)
	s1.Spec.Paths.Value("/products").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["tags"].Value.MaxItems = &from
	s2.Spec.Paths.Value("/products").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["tags"].Value.MaxItems = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyMaxItemsIncreasedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponsePropertyMaxItemsIncreasedId,
		Args:        []any{"tags", uint64(10), uint64(100), "200"},
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}
