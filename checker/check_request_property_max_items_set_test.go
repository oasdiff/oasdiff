package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// setting request body maxItems
func TestRequestBodyMaxItemsSet(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	to := uint64(10)
	s2.Spec.Paths.Value("/tags").Put.RequestBody.Value.Content["application/json"].Schema.Value.MaxItems = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMaxItemsSetCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyMaxItemsSetId,
		Args:        []any{uint64(10)},
		Comment:     checker.RequestBodyMaxItemsSetId + "-comment",
		Operation:   "PUT",
		OperationId: "setTags",
		Path:        "/tags",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}

// setting request property maxItems
func TestRequestPropertyMaxItemsSet(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	to := uint64(10)
	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["tags"].Value.MaxItems = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMaxItemsSetCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyMaxItemsSetId,
		Args:        []any{"tags", uint64(10)},
		Comment:     checker.RequestPropertyMaxItemsSetId + "-comment",
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}
