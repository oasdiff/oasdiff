package checker_test

import (
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func verifyNonBreakingChangeIsChangelogEntry(t *testing.T, d *diff.Diff, osm *diff.OperationsSourcesMap, changeId string) {
	t.Helper()

	// Check no breaking change is detected
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
	// Check changelog captures the change
	errs = checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.INFO, errs[0].GetLevel())
	require.Equal(t, changeId, errs[0].GetId())
}

// BC: no change is not breaking
func TestBreaking_Same(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: adding an optional request body is not breaking
func TestBreaking_AddingOptionalRequestBody(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s2.Spec.Paths.Value(installCommandPath).Get.RequestBody = &openapi3.RequestBodyRef{
		Value: openapi3.NewRequestBody().WithRequired(false),
	}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// CL: changing an existing request body from required to optional
func TestBreaking_RequestBodyRequiredDisabled(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(installCommandPath).Get.RequestBody = &openapi3.RequestBodyRef{
		Value: openapi3.NewRequestBody().WithRequired(true),
	}

	s2.Spec.Paths.Value(installCommandPath).Get.RequestBody = &openapi3.RequestBodyRef{
		Value: openapi3.NewRequestBody().WithRequired(false),
	}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	verifyNonBreakingChangeIsChangelogEntry(t, d, osm, checker.RequestBodyBecameOptionalId)
}

// BC: deleting a tag is not breaking
func TestBreaking_DeletedTag(t *testing.T) {
	r := d(t, diff.NewConfig(), 1, 5)
	require.Len(t, r, 5)
	require.Equal(t, checker.ResponseSuccessStatusRemovedId, r[0].GetId())
	require.Equal(t, checker.APIPathRemovedWithoutDeprecationId, r[1].GetId())
	require.Equal(t, checker.APIPathRemovedWithoutDeprecationId, r[2].GetId())
	require.Equal(t, checker.OptionalResponseHeaderRemovedId, r[3].GetId())
	require.Equal(t, checker.RequestParameterRemovedId, r[4].GetId())
}

// BC: adding an enum value is not breaking
func TestBreaking_AddedEnum(t *testing.T) {
	r := d(t, diff.NewConfig(), 1, 3)
	require.Len(t, r, 6)
	require.Equal(t, checker.ResponseSuccessStatusRemovedId, r[0].GetId())
	require.Equal(t, checker.ResponseSuccessStatusRemovedId, r[1].GetId())
	require.Equal(t, checker.RequestParameterRemovedId, r[2].GetId())
	require.Equal(t, checker.RequestParameterRemovedId, r[3].GetId())
	require.Equal(t, checker.RequestParameterRemovedId, r[4].GetId())
	require.Equal(t, checker.RequestParameterRemovedId, r[5].GetId())
}

// BC: changing extensions is not breaking
func TestBreaking_ModifiedExtension(t *testing.T) {
	r := d(t, diff.NewConfig(), 1, 3)
	require.Len(t, r, 6)
	require.Equal(t, checker.ResponseSuccessStatusRemovedId, r[0].GetId())
	require.Equal(t, checker.ResponseSuccessStatusRemovedId, r[1].GetId())
	require.Equal(t, checker.RequestParameterRemovedId, r[2].GetId())
	require.Equal(t, checker.RequestParameterRemovedId, r[3].GetId())
	require.Equal(t, checker.RequestParameterRemovedId, r[4].GetId())
	require.Equal(t, checker.RequestParameterRemovedId, r[5].GetId())
}

// BC: changing comments is not breaking
func TestBreaking_Comments(t *testing.T) {
	r := d(t, diff.NewConfig(), 1, 3)
	require.Len(t, r, 6)
	require.Equal(t, checker.ResponseSuccessStatusRemovedId, r[0].GetId())
	require.Equal(t, checker.ResponseSuccessStatusRemovedId, r[1].GetId())
	require.Equal(t, checker.RequestParameterRemovedId, r[2].GetId())
	require.Equal(t, checker.RequestParameterRemovedId, r[3].GetId())
	require.Equal(t, checker.RequestParameterRemovedId, r[4].GetId())
	require.Equal(t, checker.RequestParameterRemovedId, r[5].GetId())
}

// BC: new optional header param is not breaking
func TestBreaking_NewOptionalHeaderParam(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	deleteParam(s1.Spec.Paths.Value(installCommandPath).Get, openapi3.ParameterInHeader, "network-policies")
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Required = false

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// CL: changing an existing header param to optional
func TestBreaking_HeaderParamRequiredDisabled(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Required = true
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Required = false

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)
	require.NotEmpty(t, changes)
	require.Equal(t, checker.RequestParameterBecomeOptionalId, changes[0].GetId())
	require.Len(t, changes, 1)
}

func deleteResponseHeader(response *openapi3.Response, name string) {
	delete(response.Headers, name)
}

// BC: new required response header param is not breaking
func TestBreaking_NewRequiredResponseHeader(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	deleteResponseHeader(s1.Spec.Paths.Value(installCommandPath).Get.Responses.Value("default").Value, "X-RateLimit-Limit")
	s2.Spec.Paths.Value(installCommandPath).Get.Responses.Value("default").Value.Headers["X-RateLimit-Limit"].Value.Required = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: changing operation ID is not breaking
func TestBreaking_OperationID(t *testing.T) {
	r := d(t, diff.NewConfig(), 3, 1)
	require.Len(t, r, 3)
	require.Equal(t, checker.RequestParameterMaxLengthDecreasedId, r[0].GetId())
	require.Equal(t, checker.RequestParameterPatternAddedId, r[1].GetId())
	require.Equal(t, checker.RequestParameterEnumValueRemovedId, r[2].GetId())
}

// BC: changing a link to operation ID is not breaking
func TestBreaking_LinkOperationID(t *testing.T) {
	r := d(t, diff.NewConfig(), 3, 1)
	require.Len(t, r, 3)
	require.Equal(t, checker.RequestParameterMaxLengthDecreasedId, r[0].GetId())
	require.Equal(t, checker.RequestParameterPatternAddedId, r[1].GetId())
	require.Equal(t, checker.RequestParameterEnumValueRemovedId, r[2].GetId())
}

// BC: adding a media-type to response is not breaking
func TestBreaking_ResponseAddMediaType(t *testing.T) {
	s1, err := open("../data/response-media-type-revision.yaml")
	require.NoError(t, err)

	s2, err := open("../data/response-media-type-base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// CL: deprecating an operation with sunset greater than min
func TestBreaking_DeprecatedOperation(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s2.Spec.Paths.Value(installCommandPath).Get.Deprecated = true
	s2.Spec.Paths.Value(installCommandPath).Get.Extensions = map[string]interface{}{diff.SunsetExtension: toJson(t, civil.DateOf(time.Now()).AddDays(180).String())}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, errs[0].GetLevel(), checker.INFO)
}

// BC: deprecating a parameter is not breaking
func TestBreaking_DeprecatedParameter(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Deprecated = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: deprecating a header is not breaking
func TestBreaking_DeprecatedHeader(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s2.Spec.Paths.Value(installCommandPath).Get.Responses.Value("default").Value.Headers["X-RateLimit-Limit"].Value.Deprecated = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: deprecating a schema is not breaking
func TestBreaking_DeprecatedSchema(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Schema.Value.Deprecated = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: changing servers is not breaking
func TestBreaking_Servers(t *testing.T) {
	s1, err := open("../data/servers/baseswagger.json")
	require.NoError(t, err)

	s2, err := open("../data/servers/revisionswagger.json")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: adding a tag is not breaking
func TestBreaking_TagAdded(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s2.Spec.Paths.Value(securityScorePath).Get.Tags = append(s2.Spec.Paths.Value(securityScorePath).Get.Tags, "newTag")
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	verifyNonBreakingChangeIsChangelogEntry(t, d, osm, checker.APITagAddedId)
}

// BC: adding an operation ID is not breaking
func TestBreaking_OperationIdAdded(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(securityScorePath).Get.OperationID = ""

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	verifyNonBreakingChangeIsChangelogEntry(t, d, osm, checker.APIOperationIdAddId)
}

// BC: adding a required property to response is not breaking
func TestBreaking_RequiredResponsePropertyAdded(t *testing.T) {
	s1, err := open("../data/checker/response_required_property_added_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/response_required_property_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	verifyNonBreakingChangeIsChangelogEntry(t, d, osm, "response-required-property-added")
}
