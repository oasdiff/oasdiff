package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// unsetting response body multipleOf
func TestResponseBodyMultipleOfUnset(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := 5.0
	s1.Spec.Paths.Value("/price").Put.Responses.Value("200").Value.Content["application/json"].Schema.Value.MultipleOf = &from

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyMultipleOfUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponseBodyMultipleOfUnsetId,
		Args:        []any{5.0},
		Operation:   "PUT",
		OperationId: "setPrice",
		Path:        "/price",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}

// unsetting response property multipleOf
func TestResponsePropertyMultipleOfUnset(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := 5.0
	s1.Spec.Paths.Value("/products").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["amount"].Value.MultipleOf = &from

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyMultipleOfUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponsePropertyMultipleOfUnsetId,
		Args:        []any{"amount", 5.0, "200"},
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}

// changing response body multipleOf incompatibly
func TestResponseBodyMultipleOfChanged(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := 5.0
	to := 3.0
	s1.Spec.Paths.Value("/price").Put.Responses.Value("200").Value.Content["application/json"].Schema.Value.MultipleOf = &from
	s2.Spec.Paths.Value("/price").Put.Responses.Value("200").Value.Content["application/json"].Schema.Value.MultipleOf = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyMultipleOfUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponseBodyMultipleOfChangedId,
		Args:        []any{5.0, 3.0},
		Operation:   "PUT",
		OperationId: "setPrice",
		Path:        "/price",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}

// changing response property multipleOf incompatibly
func TestResponsePropertyMultipleOfChanged(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := 5.0
	to := 3.0
	s1.Spec.Paths.Value("/products").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["amount"].Value.MultipleOf = &from
	s2.Spec.Paths.Value("/products").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["amount"].Value.MultipleOf = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyMultipleOfUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponsePropertyMultipleOfChangedId,
		Args:        []any{"amount", 5.0, 3.0, "200"},
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}

// specializing response body multipleOf
func TestResponseBodyMultipleOfSpecialized(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := 5.0
	to := 10.0
	s1.Spec.Paths.Value("/price").Put.Responses.Value("200").Value.Content["application/json"].Schema.Value.MultipleOf = &from
	s2.Spec.Paths.Value("/price").Put.Responses.Value("200").Value.Content["application/json"].Schema.Value.MultipleOf = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyMultipleOfUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponseBodyMultipleOfSpecializedId,
		Args:        []any{5.0, 10.0},
		Comment:     checker.ResponseBodyMultipleOfSpecializedId + "-comment",
		Operation:   "PUT",
		OperationId: "setPrice",
		Path:        "/price",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}

// specializing response property multipleOf
func TestResponsePropertyMultipleOfSpecialized(t *testing.T) {
	s1, err := open(constraintKeywordsBase)
	require.NoError(t, err)
	s2, err := open(constraintKeywordsBase)
	require.NoError(t, err)

	from := 5.0
	to := 10.0
	s1.Spec.Paths.Value("/products").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["amount"].Value.MultipleOf = &from
	s2.Spec.Paths.Value("/products").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["amount"].Value.MultipleOf = &to

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyMultipleOfUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponsePropertyMultipleOfSpecializedId,
		Args:        []any{"amount", 5.0, 10.0, "200"},
		Comment:     checker.ResponsePropertyMultipleOfSpecializedId + "-comment",
		Operation:   "POST",
		OperationId: "createProduct",
		Path:        "/products",
		Source:      load.NewSource(constraintKeywordsBase),
	}, errs)
}
