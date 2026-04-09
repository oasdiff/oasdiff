package checker_test

import (
	"testing"

	"github.com/oasdiff/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// BC: changing response's body schema type from string to number is breaking
func TestBreaking_RespTypeStringToNumber(t *testing.T) {
	file := "../data/type-change/simple-response.yaml"

	s1, err := open(file)
	require.NoError(t, err)
	s1.Spec.Paths.Value("/test").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"string"}

	s2, err := open(file)
	require.NoError(t, err)
	s2.Spec.Paths.Value("/test").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"number"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponseBodyTypeChangedId, errs[0].GetId())
	require.Equal(t, "the response's body type/format changed from `string`/`` to `number`/`` for status `200`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: changing response's body schema type from number to string is breaking
func TestBreaking_RespTypeNumberToString(t *testing.T) {
	file := "../data/type-change/simple-response.yaml"

	s1, err := open(file)
	require.NoError(t, err)
	s1.Spec.Paths.Value("/test").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"number"}

	s2, err := open(file)
	require.NoError(t, err)
	s2.Spec.Paths.Value("/test").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"string"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponseBodyTypeChangedId, errs[0].GetId())
	require.Equal(t, "the response's body type/format changed from `number`/`` to `string`/`` for status `200`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: changing response's body schema type from number to integer is not breaking
func TestBreaking_RespTypeNumberToInteger(t *testing.T) {
	file := "../data/type-change/simple-response.yaml"

	s1, err := open(file)
	require.NoError(t, err)
	s1.Spec.Paths.Value("/test").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"number"}

	s2, err := open(file)
	require.NoError(t, err)
	s2.Spec.Paths.Value("/test").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"integer"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: changing response's body schema type from integer to number is breaking
func TestBreaking_RespTypeIntegerToNumber(t *testing.T) {
	file := "../data/type-change/simple-response.yaml"

	s1, err := open(file)
	require.NoError(t, err)
	s1.Spec.Paths.Value("/test").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"integer"}

	s2, err := open(file)
	require.NoError(t, err)
	s2.Spec.Paths.Value("/test").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"number"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponseBodyTypeChangedId, errs[0].GetId())
	require.Equal(t, "the response's body type/format changed from `integer`/`` to `number`/`` for status `200`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: changing response's body schema type from number/none to integer/int32 is not breaking
func TestBreaking_RespTypeNumberToInt32(t *testing.T) {
	file := "../data/type-change/simple-response.yaml"

	s1, err := open(file)
	require.NoError(t, err)
	s1.Spec.Paths.Value("/test").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"number"}

	s2, err := open(file)
	require.NoError(t, err)
	s2.Spec.Paths.Value("/test").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"integer"}
	s2.Spec.Paths.Value("/test").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value.Format = "int32"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: changing response's embedded property schema type from string/none to integer/int32 is breaking
func TestBreaking_RespTypeChanged(t *testing.T) {
	s1, err := open("../data/type-change/base-response.yaml")
	require.NoError(t, err)

	s2, err := open("../data/type-change/revision-response.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, "response-property-type-changed", errs[0].GetId())
	require.Equal(t, "the `/items/testField` response's property type/format changed from `string`/`` to `integer`/`int32` for status `200`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}
