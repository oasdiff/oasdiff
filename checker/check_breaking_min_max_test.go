package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// BC: reducing max length in request is breaking
func TestBreaking_RequestMaxLengthSmaller(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	maxLengthFrom := uint64(13)
	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInPath, "domain").Schema.Value.MaxLength = &maxLengthFrom

	maxLengthTo := uint64(11)
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInPath, "domain").Schema.Value.MaxLength = &maxLengthTo

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterMaxLengthDecreasedId, errs[0].GetId())
}

// BC: reducing max length in response is not breaking
func TestBreaking_ResponseMaxLengthSmaller(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	maxLengthFrom := uint64(13)
	s1.Spec.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["application/xml"].Schema.Value.MaxLength = &maxLengthFrom

	maxLengthTo := uint64(11)
	s2.Spec.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["application/xml"].Schema.Value.MaxLength = &maxLengthTo

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: reducing min length in request is not breaking
func TestBreaking_RequestMinLengthSmaller(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInPath, "domain").Schema.Value.MinLength = uint64(13)
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInPath, "domain").Schema.Value.MinLength = uint64(11)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: reducing min length in response is breaking
func TestBreaking_MinLengthSmaller(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["application/xml"].Schema.Value.MinLength = uint64(13)
	s2.Spec.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["application/xml"].Schema.Value.MinLength = uint64(11)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Equal(t, checker.ResponseBodyMinLengthDecreasedId, errs[0].GetId())
}

// BC: increasing max length in request is not breaking
func TestBreaking_RequestMaxLengthGreater(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	maxLengthFrom := uint64(13)
	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInPath, "domain").Schema.Value.MaxLength = &maxLengthFrom

	maxLengthTo := uint64(14)
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInPath, "domain").Schema.Value.MaxLength = &maxLengthTo

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: increasing max length in response is breaking
func TestBreaking_ResponseMaxLengthGreater(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	maxLengthFrom := uint64(13)
	s1.Spec.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["application/xml"].Schema.Value.MaxLength = &maxLengthFrom

	maxLengthTo := uint64(14)
	s2.Spec.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["application/xml"].Schema.Value.MaxLength = &maxLengthTo

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
}

// BC: changing max length in request from nil to any value is breaking
func TestBreaking_MaxLengthFromNil(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInPath, "domain").Schema.Value.MaxLength = nil

	maxLengthTo := uint64(14)
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInPath, "domain").Schema.Value.MaxLength = &maxLengthTo

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterMaxLengthSetId, errs[0].GetId())
}

// BC: changing max length in response from nil to any value is not breaking
func TestBreaking_ResponseMaxLengthFromNil(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["application/xml"].Schema.Value.MaxLength = nil

	maxLengthTo := uint64(14)
	s2.Spec.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["application/xml"].Schema.Value.MaxLength = &maxLengthTo

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: changing max length in request from any value to nil is not breaking
func TestBreaking_RequestMaxLengthToNil(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	maxLengthFrom := uint64(13)
	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInPath, "domain").Schema.Value.MaxLength = &maxLengthFrom

	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInPath, "domain").Schema.Value.MaxLength = nil

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: changing max length in response from any value to nil is breaking
func TestBreaking_ResponseMaxLengthToNil(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	maxLengthFrom := uint64(13)
	s1.Spec.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["application/xml"].Schema.Value.MaxLength = &maxLengthFrom

	s2.Spec.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["application/xml"].Schema.Value.MaxLength = nil

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponseBodyMaxLengthUnsetId, errs[0].GetId())
}

// BC: both max lengths in request are nil is not breaking
func TestBreaking_MaxLengthBothNil(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInPath, "domain").Schema.Value.MaxLength = nil
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInPath, "domain").Schema.Value.MaxLength = nil

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: both max lengths in response are nil is not breaking
func TestBreaking_ResponseMaxLengthBothNil(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["application/xml"].Schema.Value.MaxLength = nil
	s2.Spec.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["application/xml"].Schema.Value.MaxLength = nil

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: reducing min items in request is not breaking
func TestBreaking_RequestMinItemsSmaller(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInPath, "domain").Schema.Value.MinItems = 13
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInPath, "domain").Schema.Value.MinItems = 11

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: reducing min items in response is breaking
func TestBreaking_ResponseMinItemsSmaller(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["application/xml"].Schema.Value.MinItems = 13
	s2.Spec.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["application/xml"].Schema.Value.MinItems = 11

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponseBodyMinItemsDecreasedId, errs[0].GetId())
}

// BC: increasing min items in request is breaking
func TestBreaking_RequeatMinItemsGreater(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInPath, "domain").Schema.Value.MinItems = 13
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInPath, "domain").Schema.Value.MinItems = 14

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
}

// BC: increasing min items in response is not breaking
func TestBreaking_ResponseMinItemsGreater(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["application/xml"].Schema.Value.MinItems = 13
	s2.Spec.Paths.Value(securityScorePath).Get.Responses.Value("201").Value.Content["application/xml"].Schema.Value.MinItems = 14

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: reducing max in request is breaking
func TestBreaking_MaxSmaller(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	maxFrom := float64(13)
	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInPath, "domain").Schema.Value.Max = &maxFrom

	maxTo := float64(11)
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInPath, "domain").Schema.Value.Max = &maxTo

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
}

// BC: reducing max in response is not breaking
func TestBreaking_MaxSmallerInResponse(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	maxFrom := float64(13)
	s1.Spec.Paths.Value(installCommandPath).Get.Responses.Value("default").Value.Headers["X-RateLimit-Limit"].Value.Schema.Value.Max = &maxFrom

	maxTo := float64(11)
	s2.Spec.Paths.Value(installCommandPath).Get.Responses.Value("default").Value.Headers["X-RateLimit-Limit"].Value.Schema.Value.Max = &maxTo

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}
