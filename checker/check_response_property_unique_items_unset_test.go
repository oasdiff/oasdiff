package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// unsetting response body uniqueItems
func TestResponseBodyUniqueItemsUnset(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	s1.Spec.Paths.Value("/tags").Put.Responses.Value("200").Value.Content["application/json"].Schema.Value.UniqueItems = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyUniqueItemsUnsetCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponseBodyUniqueItemsUnsetId,
		Args:        []any{},
		Operation:   "PUT",
		OperationId: "setTags",
		Path:        "/tags",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}

// unsetting response property uniqueItems
func TestResponsePropertyUniqueItemsUnset(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	s1.Spec.Paths.Value("/products").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["tags"].Value.UniqueItems = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyUniqueItemsUnsetCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponsePropertyUniqueItemsUnsetId,
		Args:        []any{"tags", "200"},
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}
