package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// changing request's body schema type from string to number is breaking
func TestBreaking_ReqTypeStringToNumber(t *testing.T) {
	file := "../data/type-change/simple-request.yaml"

	s1, err := open(file)
	require.NoError(t, err)
	s1.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"string"}

	s2, err := open(file)
	require.NoError(t, err)
	s2.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"number"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, "the request's body `type` changed from `string` to `number`", requireChange(t, errs, checker.RequestBodyTypeChangedId).GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// changing request's body schema type from number to string is breaking
func TestBreaking_ReqTypeNumberToString(t *testing.T) {
	file := "../data/type-change/simple-request.yaml"

	s1, err := open(file)
	require.NoError(t, err)
	s1.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"number"}

	s2, err := open(file)
	require.NoError(t, err)
	s2.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"string"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, "the request's body `type` changed from `number` to `string`", requireChange(t, errs, checker.RequestBodyTypeChangedId).GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// changing request's body schema type from number to integer is breaking
func TestBreaking_ReqTypeNumberToInteger(t *testing.T) {
	file := "../data/type-change/simple-request.yaml"

	s1, err := open(file)
	require.NoError(t, err)
	s1.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"number"}

	s2, err := open(file)
	require.NoError(t, err)
	s2.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"integer"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, "the request's body `type` changed from `number` to `integer`", requireChange(t, errs, checker.RequestBodyTypeChangedId).GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// changing request's body schema type from integer to number is not breaking
func TestBreaking_ReqTypeIntegerToNumber(t *testing.T) {
	file := "../data/type-change/simple-request.yaml"

	s1, err := open(file)
	require.NoError(t, err)
	s1.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"integer"}

	s2, err := open(file)
	require.NoError(t, err)
	s2.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"number"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, "the request's body `type` was widened from `integer` to `number`", requireChange(t, errs, checker.RequestBodyTypeGeneralizedId).GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// narrowing a request's body schema union type is breaking (server rejects previously-valid values)
func TestBreaking_ReqTypeUnionNarrowed(t *testing.T) {
	file := "../data/type-change/simple-request.yaml"

	s1, err := open(file)
	require.NoError(t, err)
	s1.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"string", "integer"}

	s2, err := open(file)
	require.NoError(t, err)
	s2.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"string"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	requireSingleChange(t, errs, checker.RequestBodyTypeChangedId)
}

// removing request's body schema type is not breaking (server becomes more permissive)
func TestBreaking_ReqTypeStringDeleted(t *testing.T) {
	file := "../data/type-change/simple-request.yaml"

	s1, err := open(file)
	require.NoError(t, err)
	s1.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"string"}

	s2, err := open(file)
	require.NoError(t, err)
	s2.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value.Type = nil

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, "the request's body `type` was widened from `string` to `any`", requireChange(t, errs, checker.RequestBodyTypeGeneralizedId).GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// changing request's body schema type from number/none to integer/int32 is breaking
func TestBreaking_ReqTypeNumberToInt32(t *testing.T) {
	file := "../data/type-change/simple-request.yaml"

	s1, err := open(file)
	require.NoError(t, err)
	s1.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"number"}

	s2, err := open(file)
	require.NoError(t, err)
	s2.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"integer"}
	s2.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value.Format = "int32"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, "the request's body `type/format` changed from `number` to `integer/int32`", requireChange(t, errs, checker.RequestBodyTypeChangedId).GetUncolorizedText(checker.NewDefaultLocalizer()))
}
