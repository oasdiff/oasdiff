package checker_test

import (
	"testing"

	"github.com/oasdiff/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// BC: changing request's body schema type from string to number is breaking
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
	require.Equal(t, checker.RequestBodyTypeChangedId, errs[0].GetId())
	require.Equal(t, "the request's body type/format changed from `string`/`` to `number`/``", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: changing request's body schema type from number to string is breaking
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
	require.Equal(t, checker.RequestBodyTypeChangedId, errs[0].GetId())
	require.Equal(t, "the request's body type/format changed from `number`/`` to `string`/``", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: changing request's body schema type from number to integer is breaking
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
	require.Equal(t, checker.RequestBodyTypeChangedId, errs[0].GetId())
	require.Equal(t, "the request's body type/format changed from `number`/`` to `integer`/``", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: changing request's body schema type from integer to number is not breaking
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
	require.Equal(t, checker.RequestBodyTypeGeneralizedId, errs[0].GetId())
	require.Equal(t, "the request's body type/format was generalized from `integer`/`` to `number`/``", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: changing request's body schema type from number/none to integer/int32 is breaking
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
	require.Equal(t, checker.RequestBodyTypeChangedId, errs[0].GetId())
	require.Equal(t, "the request's body type/format changed from `number`/`` to `integer`/`int32`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}
