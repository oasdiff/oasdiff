package checker_test

import (
	"fmt"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

const (
	securityScorePath      = "/api/{domain}/{project}/badges/security-score"
	securityScorePathSlash = securityScorePath + "/"
	installCommandPath     = "/api/{domain}/{project}/install-command"
)

func l(t *testing.T, v int) *load.SpecInfo {
	t.Helper()
	specInfo, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource(fmt.Sprintf("../data/openapi-test%d.yaml", v)))
	require.NoError(t, err)
	return specInfo
}

func d(t *testing.T, config *diff.Config, v1, v2 int) checker.Changes {
	t.Helper()
	l1 := l(t, v1)
	l2 := l(t, v2)
	d, osm, err := diff.GetWithOperationsSourcesMap(config, l1, l2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	return errs
}

// BC: adding a required request body is breaking
func TestBreaking_AddingRequiredRequestBody(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s2.Spec.Paths.Value(installCommandPath).Get.RequestBody = &openapi3.RequestBodyRef{
		Value: openapi3.NewRequestBody().WithRequired(true),
	}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.AddedRequiredRequestBodyId, errs[0].GetId())
	require.Equal(t, "added required request body", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: changing an existing request body from optional to required is breaking
func TestBreaking_RequestBodyRequiredEnabled(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(installCommandPath).Get.RequestBody = &openapi3.RequestBodyRef{
		Value: openapi3.NewRequestBody().WithRequired(false),
	}

	s2.Spec.Paths.Value(installCommandPath).Get.RequestBody = &openapi3.RequestBodyRef{
		Value: openapi3.NewRequestBody().WithRequired(true),
	}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestBodyBecameRequiredId, errs[0].GetId())
	require.Equal(t, "request body became required", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: deleting an enum value is breaking
func TestBreaking_DeletedEnum(t *testing.T) {
	errs := d(t, diff.NewConfig(), 702, 1)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterEnumValueRemovedId, errs[0].GetId())
	require.Equal(t, "removed the enum value 'removed-value' from the 'path' request parameter 'domain'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: added an enum value to response breaking
func TestBreaking_AddedResponseEnum(t *testing.T) {
	errs := d(t, diff.NewConfig(), 703, 704)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 2)
	require.Equal(t, checker.ResponsePropertyEnumValueAddedId, errs[0].GetId())
	require.Equal(t, "added the new 'QWE' enum value to the 'respenum' response property for the response status 'default'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, checker.ResponsePropertyEnumValueAddedId, errs[1].GetId())
	require.Equal(t, "added the new 'TER2' enum value to the 'respenum2/respenum3' response property for the response status 'default'", errs[1].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

func deleteParam(op *openapi3.Operation, in string, name string) {

	result := openapi3.NewParameters()

	for _, item := range op.Parameters {
		if v := item.Value; v != nil {
			if v.Name == name && v.In == in {
				continue
			}
			result = append(result, item)
		}
	}
	op.Parameters = result
}

// BC: renaming a path parameter is not breaking
func TestBreaking_PathParamRename(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile("../data/param-rename/method-base.yaml")
	require.NoError(t, err)

	s2, err := loader.LoadFromFile("../data/param-rename/method-revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(),
		&load.SpecInfo{Spec: s1},
		&load.SpecInfo{Spec: s2},
	)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)

	require.Empty(t, errs)
}

// BC: new required path param is breaking
func TestBreaking_NewPathParam(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	deleteParam(s1.Spec.Paths.Value(installCommandPath).Get, openapi3.ParameterInPath, "project")
	// note: path params are always required

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)

	require.Len(t, errs, 1)
	require.Equal(t, checker.NewRequestPathParameterId, errs[0].GetId())
	require.Equal(t, "added the new path request parameter 'project'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: new required header param is breaking
func TestBreaking_NewRequiredHeaderParam(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	deleteParam(s1.Spec.Paths.Value(installCommandPath).Get, openapi3.ParameterInHeader, "network-policies")
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Required = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.NewRequiredRequestParameterId, errs[0].GetId())
	require.Equal(t, "added the new required 'header' request parameter 'network-policies'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: changing an existing header param from optional to required is breaking
func TestBreaking_HeaderParamRequiredEnabled(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Required = false
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Required = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t,
		checker.ApiChange{
			Id:        checker.RequestParameterBecomeRequiredId,
			Args:      []any{"header", "network-policies"},
			Level:     checker.ERR,
			Operation: "GET",
			Path:      "/api/{domain}/{project}/install-command",
			Source:    load.NewSource("../data/openapi-test1.yaml"),
		}, errs[0])
	require.Equal(t, "the 'header' request parameter 'network-policies' became required", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: changing an existing response header from required to optional is breaking
func TestBreaking_ResponseHeaderParamRequiredDisabled(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(installCommandPath).Get.Responses.Value("default").Value.Headers["X-RateLimit-Limit"].Value.Required = true
	s2.Spec.Paths.Value(installCommandPath).Get.Responses.Value("default").Value.Headers["X-RateLimit-Limit"].Value.Required = false

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponseHeaderBecameOptionalId, errs[0].GetId())
	require.Equal(t, "the response header 'X-RateLimit-Limit' became optional for the status 'default'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: removing an existing required response header is breaking as error
func TestBreaking_ResponseHeaderRemoved(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(installCommandPath).Get.Responses.Value("default").Value.Headers["X-RateLimit-Limit"].Value.Required = true
	delete(s2.Spec.Paths.Value(installCommandPath).Get.Responses.Value("default").Value.Headers, "X-RateLimit-Limit")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	for _, err := range errs {
		require.Equal(t, checker.ERR, err.GetLevel())
	}
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequiredResponseHeaderRemovedId, errs[0].GetId())
	require.Equal(t, "the mandatory response header 'X-RateLimit-Limit' removed for the status 'default'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: removing an existing response with successful status is breaking
func TestBreaking_ResponseSuccessStatusUpdated(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s2.Spec.Paths.Value(securityScorePath).Get.Responses.Delete("200")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	for _, err := range errs {
		require.Equal(t, checker.ERR, err.GetLevel())
	}
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponseSuccessStatusRemovedId, errs[0].GetId())
	require.Equal(t, "removed the success response with the status '200'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: removing an existing response with non-successful status is breaking (optional)
func TestBreaking_ResponseNonSuccessStatusUpdated(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s2.Spec.Paths.Value(securityScorePath).Get.Responses.Delete("400")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig().WithOptionalCheck(checker.ResponseNonSuccessStatusRemovedId), d, osm)
	for _, err := range errs {
		require.Equal(t, checker.ERR, err.GetLevel())
	}
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponseNonSuccessStatusRemovedId, errs[0].GetId())
	require.Equal(t, "removed the non-success response with the status '400'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: removing/updating an operation id is breaking (optional)
func TestBreaking_OperationIdRemoved(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s2.Spec.Paths.Value(securityScorePath).Get.OperationID = "newOperationId"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibility(allChecksConfig().WithOptionalCheck(checker.APIOperationIdRemovedId), d, osm)
	for _, err := range errs {
		require.Equal(t, checker.ERR, err.GetLevel())
	}
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIOperationIdRemovedId, errs[0].GetId())
	require.Equal(t, "api operation id 'GetSecurityScores' removed and replaced with 'newOperationId'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	verifyNonBreakingChangeIsChangelogEntry(t, d, osm, checker.APIOperationIdRemovedId)
}

// BC: removing/updating an enum in request body is breaking (optional)
func TestBreaking_RequestBodyEnumRemoved(t *testing.T) {
	s1, err := open("../data/enums/request-body-enum.yaml")
	require.NoError(t, err)

	s2, err := open("../data/enums/request-body-enum.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v2/changeOfRequestFieldValueTiedToEnumTest").Get.RequestBody.Value.Content["application/json"].Schema.Value.Enum = []interface{}{}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibility(allChecksConfig().WithOptionalCheck(checker.RequestBodyEnumValueRemovedId), d, osm)
	for _, err := range errs {
		require.Equal(t, checker.ERR, err.GetLevel())
	}

	require.Len(t, errs, 3)
	require.Equal(t, checker.RequestBodyEnumValueRemovedId, errs[0].GetId())
	require.Equal(t, "request body enum value removed 'VALUE_1'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: removing/updating a property enum in response is breaking (optional)
func TestBreaking_ResponsePropertyEnumRemoved(t *testing.T) {
	s1 := l(t, 704)
	s2 := l(t, 703)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibility(allChecksConfig().WithOptionalCheck(checker.ResponsePropertyEnumValueRemovedId), d, osm)
	for _, err := range errs {
		require.Equal(t, checker.ERR, err.GetLevel())
	}
	require.NotEmpty(t, errs)
	require.Len(t, errs, 2)
	require.Equal(t, checker.ResponsePropertyEnumValueRemovedId, errs[0].GetId())
	require.Equal(t, "removed the 'QWE' enum value from the 'respenum' response property for the response status 'default'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: removing/updating a tag is breaking (optional)
func TestBreaking_TagRemoved(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s2.Spec.Paths.Value(securityScorePath).Get.Tags[0] = "newTag"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig().WithOptionalCheck(checker.APITagRemovedId), d, osm)
	for _, err := range errs {
		require.Equal(t, checker.ERR, err.GetLevel())
	}
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APITagRemovedId, errs[0].GetId())
	require.Equal(t, "api tag 'security' removed", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: removing/updating a media type enum in response (optional)
func TestBreaking_ResponseMediaTypeEnumRemoved(t *testing.T) {
	s1, err := open("../data/enums/response-enum.yaml")
	require.NoError(t, err)

	s2, err := open("../data/enums/response-enum-2.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig().WithOptionalCheck(checker.ResponseMediaTypeEnumValueRemovedId), d, osm)
	for _, err := range errs {
		require.Equal(t, checker.ERR, err.GetLevel())
	}
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponseMediaTypeEnumValueRemovedId, errs[0].GetId())
	require.Equal(t, "response schema 'application/json' enum value removed 'VALUE_3'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: removing an existing response with unparseable status is not breaking
func TestBreaking_ResponseUnparseableStatusRemoved(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s2.Spec.Paths.Value(installCommandPath).Get.Responses.Delete("default")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	for _, err := range errs {
		require.Equal(t, checker.ERR, err.GetLevel())
	}
	require.Empty(t, errs)
}

// BC: removing an existing response with error status is not breaking
func TestBreaking_ResponseErrorStatusRemoved(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s2.Spec.Paths.Value(securityScorePath).Get.Responses.Delete("400")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	for _, err := range errs {
		require.Equal(t, checker.ERR, err.GetLevel())
	}
	require.Empty(t, errs)
}

// BC: removing an existing optional response header is breaking as warn
func TestBreaking_OptionalResponseHeaderRemoved(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(installCommandPath).Get.Responses.Value("default").Value.Headers["X-RateLimit-Limit"].Value.Required = false
	delete(s2.Spec.Paths.Value(installCommandPath).Get.Responses.Value("default").Value.Headers, "X-RateLimit-Limit")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	for _, err := range errs {
		require.Equal(t, checker.WARN, err.GetLevel())
	}
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.OptionalResponseHeaderRemovedId, errs[0].GetId())
	require.Equal(t, "the optional response header 'X-RateLimit-Limit' removed for the status 'default'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: deleting a media-type from response is breaking
func TestBreaking_ResponseDeleteMediaType(t *testing.T) {
	s1, err := open("../data/response-media-type-base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/response-media-type-revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponseMediaTypeRemovedId, errs[0].GetId())
	require.Equal(t, "removed the media type 'application/json' for the response with the status '200'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: deleting a pattern from a schema is not breaking
func TestBreaking_DeletePatten(t *testing.T) {
	s1, err := open("../data/pattern-base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/pattern-revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: adding a pattern to a schema is breaking
func TestBreaking_AddPattern(t *testing.T) {
	s1, err := open("../data/pattern-revision.yaml")
	require.NoError(t, err)

	s2, err := open("../data/pattern-base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, "request-property-pattern-added", errs[0].GetId())
	require.Equal(t, "added the pattern '^[a-z]+$' to the request property 'created'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: adding a pattern to a schema is breaking for recursive properties
func TestBreaking_AddPatternRecursive(t *testing.T) {
	s1, err := open("../data/pattern-revision-recursive.yaml")
	require.NoError(t, err)

	s2, err := open("../data/pattern-base-recursive.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, "request-property-pattern-added", errs[0].GetId())
	require.Equal(t, "added the pattern '^[a-z]+$' to the request property 'data/created'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: modifying a pattern in a schema is breaking
func TestBreaking_ModifyPattern(t *testing.T) {
	s1, err := open("../data/pattern-base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/pattern-modified-not-anystring.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertyPatternChangedId, errs[0].GetId())
	require.Equal(t, "changed the pattern of the request property 'created' from '^[a-z]+$' to '.+'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, checker.WARN, errs[0].GetLevel())
}

// BC: modifying a pattern to .* in a schema is not breaking
func TestBreaking_GeneralizedPattern(t *testing.T) {
	s1, err := open("../data/pattern-base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/pattern-modified.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: modifying a pattern in request parameter is breaking
func TestBreaking_ModifyParameterPattern(t *testing.T) {
	s1, err := open("../data/pattern-parameter-base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/pattern-parameter-modified-not-anystring.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterPatternChangedId, errs[0].GetId())
	require.Equal(t, "changed the pattern of the 'path' request parameter 'groupId' from '[0-9a-f]+' to '[0-9]+'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: modifying a pattern to ".*" in a schema is not breaking
func TestBreaking_ModifyPatternToAnyString(t *testing.T) {
	s1, err := open("../data/pattern-base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/pattern-modified.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: modifying the default value of an optional request parameter is breaking
func TestBreaking_ModifyRequiredOptionalParamDefaultValue(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Schema.Value.Default = "X"
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Schema.Value.Default = "Y"

	// By default, OpenAPI treats all request parameters as optional

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterDefaultValueChangedId, errs[0].GetId())
	require.Equal(t, "for the 'header' request parameter 'network-policies', default value was changed from 'X' to 'Y'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: setting the default value of an optional request parameter is breaking
func TestBreaking_SettingOptionalParamDefaultValue(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Schema.Value.Default = nil
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Schema.Value.Default = "Y"

	// By default, OpenAPI treats all request parameters as optional

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterDefaultValueAddedId, errs[0].GetId())
	require.Equal(t, "for the 'header' request parameter 'network-policies', default value 'Y' was added", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: removing the default value of an optional request parameter is breaking
func TestBreaking_RemovingOptionalParamDefaultValue(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Schema.Value.Default = "Y"
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Schema.Value.Default = nil

	// By default, OpenAPI treats all request parameters as optional

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, "request-parameter-default-value-removed", errs[0].GetId())
	require.Equal(t, "for the 'header' request parameter 'network-policies', default value 'Y' was removed", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: modifying the default value of a required request parameter is not breaking
func TestBreaking_ModifyRequiredRequiredParamDefaultValue(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	paramBase := s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies")
	paramBase.Required = true
	paramBase.Schema.Value.Default = "X"

	paramRevision := s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies")
	paramRevision.Required = true
	paramRevision.Schema.Value.Default = "Y"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: removing an schema object from components is breaking (optional)
func TestBreaking_SchemaRemoved(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)
	s1.Spec.Paths = openapi3.NewPaths()
	s2.Spec.Paths = openapi3.NewPaths()

	for k := range s2.Spec.Components.Schemas {
		delete(s2.Spec.Components.Schemas, k)
	}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	checks := allChecksConfig().WithOptionalCheck(checker.APISchemasRemovedId)
	errs := checker.CheckBackwardCompatibility(checks, d, osm)
	for _, err := range errs {
		require.Equal(t, checker.ERR, err.GetLevel())
	}
	require.NotEmpty(t, errs)
	require.Equal(t, checker.APISchemasRemovedId, errs[0].GetId())
	require.Equal(t, "removed the schema 'network-policies'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, checker.APISchemasRemovedId, errs[1].GetId())
	require.Equal(t, "removed the schema 'rules'", errs[1].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: removing a media type from request body is breaking
func TestBreaking_RequestBodyMediaTypeRemoved(t *testing.T) {
	s1, err := open("../data/checker/request_body_media_type_updated_revision.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/request_body_media_type_updated_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Equal(t, "request-body-media-type-removed", errs[0].GetId())
	require.Equal(t, "removed the media type 'application/json' from the request body", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: removing 'anyOf' schema from the request body or request body property is breaking
func TestBreaking_RequestPropertyAnyOfRemoved(t *testing.T) {
	s1, err := open("../data/checker/request_property_any_of_removed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_any_of_removed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)

	require.Len(t, errs, 2)

	require.Equal(t, checker.RequestBodyAnyOfRemovedId, errs[0].GetId())
	require.Equal(t, checker.ERR, errs[0].GetLevel())
	require.Equal(t, "removed '#/components/schemas/Rabbit' from the request body 'anyOf' list", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))

	require.Equal(t, checker.RequestPropertyAnyOfRemovedId, errs[1].GetId())
	require.Equal(t, checker.ERR, errs[1].GetLevel())
	require.Equal(t, "removed '#/components/schemas/Breed3' from the '/anyOf[#/components/schemas/Dog]/breed' request property 'anyOf' list", errs[1].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: removing 'oneOf' schema from the request body or request body property is breaking
func TestBreaking_RequestPropertyOneOfRemoved(t *testing.T) {
	s1, err := open("../data/checker/request_property_one_of_removed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_one_of_removed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)

	require.Len(t, errs, 2)
	require.Equal(t, checker.RequestBodyOneOfRemovedId, errs[0].GetId())
	require.Equal(t, checker.ERR, errs[0].GetLevel())
	require.Equal(t, "removed '#/components/schemas/Rabbit' from the request body 'oneOf' list", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))

	require.Equal(t, checker.RequestPropertyOneOfRemovedId, errs[1].GetId())
	require.Equal(t, checker.ERR, errs[1].GetLevel())
	require.Equal(t, "removed '#/components/schemas/Breed3' from the '/oneOf[#/components/schemas/Dog]/breed' request property 'oneOf' list", errs[1].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: adding 'allOf' subschema to the request body or request body property is breaking
func TestBreaking_RequestPropertyAllOfAdded(t *testing.T) {
	s1, err := open("../data/checker/request_property_all_of_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_all_of_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)

	require.Len(t, errs, 2)

	require.Equal(t, checker.RequestBodyAllOfAddedId, errs[0].GetId())
	require.Equal(t, checker.ERR, errs[0].GetLevel())
	require.Equal(t, "added '#/components/schemas/Rabbit' to the request body 'allOf' list", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))

	require.Equal(t, checker.RequestPropertyAllOfAddedId, errs[1].GetId())
	require.Equal(t, checker.ERR, errs[1].GetLevel())
	require.Equal(t, "added '#/components/schemas/Breed3' to the '/allOf[#/components/schemas/Dog]/breed' request property 'allOf' list", errs[1].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: removing 'allOf' subschema from the request body or request body property is breaking with warn
func TestBreaking_RequestPropertyAllOfRemoved(t *testing.T) {
	s1, err := open("../data/checker/request_property_all_of_removed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_all_of_removed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)

	require.Len(t, errs, 2)

	require.Equal(t, checker.RequestBodyAllOfRemovedId, errs[0].GetId())
	require.Equal(t, checker.WARN, errs[0].GetLevel())
	require.Equal(t, "removed '#/components/schemas/Rabbit' from the request body 'allOf' list", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))

	require.Equal(t, checker.RequestPropertyAllOfRemovedId, errs[1].GetId())
	require.Equal(t, checker.WARN, errs[1].GetLevel())
	require.Equal(t, "removed '#/components/schemas/Breed3' from the '/allOf[#/components/schemas/Dog]/breed' request property 'allOf' list", errs[1].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: removing an optional field from the request body is breaking (optional)
func TestBreaking_RequestPropertyRemoved(t *testing.T) {
	s1, err := open("../data/checker/request_property_removed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_removed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	checks := allChecksConfig().WithOptionalCheck(checker.RequestPropertyRemovedId)
	errs := checker.CheckBackwardCompatibility(checks, d, osm)

	require.Len(t, errs, 1)

	require.Equal(t, checker.RequestPropertyRemovedId, errs[0].GetId())
	require.Equal(t, checker.ERR, errs[0].GetLevel())
	require.Equal(t, "removed the request property 'breed'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))

	d, osm, err = diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	checks = allChecksConfig()
	errs = checker.CheckBackwardCompatibility(checks, d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.WARN, errs[0].GetLevel())
}

// BC: removing an optional field from the response body is breaking (optional)
func TestBreaking_ResponsePropertyRemoved(t *testing.T) {
	s1, err := open("../data/checker/response_property_removed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_removed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	checks := allChecksConfig().WithOptionalCheck(checker.ResponseOptionalPropertyRemovedId)
	errs := checker.CheckBackwardCompatibility(checks, d, osm)

	require.Len(t, errs, 1)

	require.Equal(t, checker.ResponseOptionalPropertyRemovedId, errs[0].GetId())
	require.Equal(t, checker.ERR, errs[0].GetLevel())
	require.Equal(t, "removed the optional property 'breed' from the response with the '200' status", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))

	d, osm, err = diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	checks = allChecksConfig()
	errs = checker.CheckBackwardCompatibility(checks, d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.WARN, errs[0].GetLevel())
}

// BC: removing an optional field from the response body is breaking (optional)
func TestBreaking_ResponseWriteOnlyPropertyRemoved(t *testing.T) {
	s1, err := open("../data/checker/response_property_removed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_removed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	checks := allChecksConfig().WithOptionalCheck(checker.ResponseOptionalWriteOnlyPropertyRemovedId)
	errs := checker.CheckBackwardCompatibility(checks, d, osm)

	require.Len(t, errs, 2)

	require.Equal(t, checker.ResponseOptionalWriteOnlyPropertyRemovedId, errs[0].GetId())
	require.Equal(t, checker.ERR, errs[0].GetLevel())
	require.Equal(t, "removed the optional write-only property 'other' from the response with the '200' status", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))

	d, osm, err = diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	checks = allChecksConfig()
	errs = checker.CheckBackwardCompatibility(checks, d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.WARN, errs[0].GetLevel())
}
