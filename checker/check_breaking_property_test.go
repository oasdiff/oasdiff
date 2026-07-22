package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// new required property in request header is breaking
func TestBreaking_NewRequiredProperty(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Schema.Value.Properties["courseId"] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:        &openapi3.Types{"string"},
			Description: "Unique ID of the course",
		},
	}
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Schema.Value.Required = []string{"courseId"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.NewRequiredRequestHeaderPropertyId)
}

// new optional property in request header is not breaking
func TestBreaking_NewNonRequiredProperty(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Schema.Value.Properties["courseId"] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:        &openapi3.Types{"string"},
			Description: "Unique ID of the course",
		},
	}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// changing an existing property in request header to required is breaking
func TestBreaking_PropertyRequiredEnabled(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	sr := openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:        &openapi3.Types{"string"},
			Description: "Unique ID of the course",
		},
	}

	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Schema.Value.Properties["courseId"] = &sr
	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Schema.Value.Required = []string{}

	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Schema.Value.Properties["courseId"] = &sr
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Schema.Value.Required = []string{"courseId"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.RequestHeaderPropertyBecameRequiredId)
}

// changing an existing property in request header to optional is not breaking
func TestBreaking_PropertyRequiredDisabled(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	sr := openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:        &openapi3.Types{"string"},
			Description: "Unique ID of the course",
		},
	}

	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Schema.Value.Properties["courseId"] = &sr
	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Schema.Value.Required = []string{"courseId"}

	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Schema.Value.Properties["courseId"] = &sr
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Schema.Value.Required = []string{}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// changing an existing property in response body to optional is breaking
func TestBreaking_RespBodyRequiredPropertyDisabled(t *testing.T) {
	s1, err := open(requiredPropertyFile("response-base.json"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("response-revision.json"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.ResponsePropertyBecameOptionalId)
}

// changing a request body to enum is breaking
func TestBreaking_ReqBodyBecameEnum(t *testing.T) {
	s1, err := open("../data/enums/request-body-no-enum.yaml")
	require.NoError(t, err)

	s2, err := open("../data/enums/request-body-enum.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.RequestBodyBecameEnumId)
}

// adding an enum value to request body is not breaking
func TestBreaking_ReqBodyEnumValueAdded(t *testing.T) {
	s1, err := open("../data/enums/request-body-enum.yaml")
	require.NoError(t, err)

	s2, err := open("../data/enums/request-body-enum-revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// changing a request body type and changing it to enum simultaneously is breaking
func TestBreaking_ReqBodyBecameEnumAndTypeChanged(t *testing.T) {
	s1, err := open("../data/enums/request-body-no-enum.yaml")
	require.NoError(t, err)

	s2, err := open("../data/enums/request-body-enum-int.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 2)
	requireChange(t, errs, checker.RequestBodyBecameEnumId)
	requireChange(t, errs, checker.RequestBodyTypeChangedId)
}

// changing an existing property in request body to enum is breaking
func TestBreaking_ReqPropertyBecameEnum(t *testing.T) {
	s1, err := open("../data/enums/request-property-no-enum.yaml")
	require.NoError(t, err)

	s2, err := open("../data/enums/request-property-enum.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.RequestPropertyBecameEnumId)
}

// changing an existing path param to enum is breaking
func TestBreaking_ReqParameterBecameEnum(t *testing.T) {
	s1, err := open("../data/enums/request-parameter-op-no-enum.yaml")
	require.NoError(t, err)

	s2, err := open("../data/enums/request-parameter-op-enum.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.RequestParameterBecameEnumId)
}

// changing an existing property in request header to enum is breaking
func TestBreaking_ReqParameterHeaderPropertyBecameEnum(t *testing.T) {
	s1, err := open("../data/enums/request-parameter-property-no-enum.yaml")
	require.NoError(t, err)

	s2, err := open("../data/enums/request-parameter-property-enum.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.RequestHeaderPropertyBecameEnumId)
}

// changing a response body to nullable is breaking
func TestBreaking_RespBodyNullable(t *testing.T) {
	s1, err := open("../data/nullable/base-body.yaml")
	require.NoError(t, err)

	s2, err := open("../data/nullable/revision-body.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.ResponseBodyBecameNullableId)
}

// changing a request property to not nullable is breaking
func TestBreaking_ReqBodyPropertyNotNullable(t *testing.T) {
	s1, err := open("../data/nullable/base-req.yaml")
	require.NoError(t, err)

	s2, err := open("../data/nullable/revision-req.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.RequestPropertyBecomeNotNullableId)
}

// changing a response property to nullable is breaking
func TestBreaking_RespBodyPropertyNullable(t *testing.T) {
	s1, err := open("../data/nullable/base-property.yaml")
	require.NoError(t, err)

	s2, err := open("../data/nullable/revision-property.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.ResponsePropertyBecameNullableId)
}

// changing an embedded response property to nullable is breaking
func TestBreaking_RespBodyEmbeddedPropertyNullable(t *testing.T) {
	s1, err := open("../data/nullable/base-embedded-property.yaml")
	require.NoError(t, err)

	s2, err := open("../data/nullable/revision-embedded-property.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.ResponsePropertyBecameNullableId)
}

// changing a required property in response body to optional and also deleting it is breaking
func TestBreaking_RespBodyDeleteAndDisableRequiredProperty(t *testing.T) {
	s1, err := open(requiredPropertyFile("response-del-required-prop-base.yaml"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("response-del-required-prop-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
}

// adding a non-existent required property in request body is not breaking
func TestBreaking_ReqBodyNewRequiredPropertyNew(t *testing.T) {
	s1, err := open(requiredPropertyFile("request-new-required-prop-base.yaml"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("request-new-required-prop-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// changing an existing property in response body to required is not breaking
func TestBreaking_RespBodyRequiredPropertyEnabled(t *testing.T) {
	s1, err := open(requiredPropertyFile("response-revision.json"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("response-base.json"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// changing an existing property in request body to optional is not breaking
func TestBreaking_ReqBodyRequiredPropertyDisabled(t *testing.T) {
	s1, err := open(requiredPropertyFile("request-base.yaml"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("request-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// changing an existing property in request body to required is breaking
func TestBreaking_ReqBodyRequiredPropertyEnabled(t *testing.T) {
	s1, err := open(requiredPropertyFile("request-revision.yaml"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("request-base.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.RequestPropertyBecameRequiredId)
}

// adding a new required property in request body is breaking
func TestBreaking_ReqBodyNewRequiredProperty(t *testing.T) {
	s1, err := open(requiredPropertyFile("request-new-base.yaml"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("request-new-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.NewRequiredRequestPropertyId)
}

// deleting a required property in request is breaking with warn
func TestBreaking_ReqBodyDeleteRequiredProperty(t *testing.T) {
	s1, err := open(requiredPropertyFile("request-new-revision.yaml"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("request-new-base.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.RequestPropertyRemovedId)
	require.Equal(t, checker.WARN, errs[0].GetLevel())
}

// deleting an embedded optional property in request is breaking with warn
func TestBreaking_ReqBodyDeleteRequiredProperty2(t *testing.T) {
	s1, err := open(requiredPropertyFile("request-property-items.yaml"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("request-property-items-2.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	requireApiChange(t, checker.ApiChange{
		Id:        checker.RequestPropertyRemovedId,
		Args:      []any{"roleAssignments/items/role"},
		Operation: "POST",
		Path:      "/api/roleMappings",
		Source:    load.NewSource("../data/required-properties/request-property-items-2.yaml"),
	}, requireChange(t, errs, checker.RequestPropertyRemovedId))
}

// adding a new required property in response body is not breaking
func TestBreaking_RespBodyNewRequiredProperty(t *testing.T) {
	s1, err := open(requiredPropertyFile("response-new-base.json"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("response-new-revision.json"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// deleting a required property in response body is breaking
func TestBreaking_RespBodyDeleteRequiredProperty(t *testing.T) {
	s1, err := open(requiredPropertyFile("response-new-revision.json"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("response-new-base.json"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.ResponseRequiredPropertyRemovedId)
}

// adding a new required property under AllOf in response body is not breaking
func TestBreaking_RespBodyNewAllOfRequiredProperty(t *testing.T) {
	s1, err := open(requiredPropertyFile("response-allof-base.json"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("response-allof-revision.json"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// deleting a required property under AllOf in response body is breaking
func TestBreaking_RespBodyDeleteAllOfRequiredProperty(t *testing.T) {
	s1, err := open(requiredPropertyFile("response-allof-revision.json"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("response-allof-base.json"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.ResponseRequiredPropertyRemovedId)
}

// adding a new required read-only property in request body is not breaking
func TestBreaking_ReadOnlyNewRequiredProperty(t *testing.T) {
	s1, err := open(requiredPropertyFile("read-only-new-base.yaml"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("read-only-new-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// changing an existing read-only property in request body to required is not breaking
func TestBreaking_ReadOnlyPropertyRequiredEnabled(t *testing.T) {
	s1, err := open(requiredPropertyFile("read-only-base.yaml"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("read-only-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// deleting a required write-only property in response body is not breaking
func TestBreaking_WriteOnlyDeleteRequiredProperty(t *testing.T) {
	s1, err := open(requiredPropertyFile("write-only-delete-base.yaml"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("write-only-delete-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.RequestPropertyRemovedId)
	require.Equal(t, checker.WARN, errs[0].GetLevel())
}

// deleting a non-required non-write-only property in response body is breaking with warning
func TestBreaking_WriteOnlyDeleteNonRequiredProperty(t *testing.T) {
	s1, err := open(requiredPropertyFile("write-only-delete-partial-base.yaml"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("write-only-delete-partial-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 3)
	requireChange(t, errs, checker.RequestPropertyRemovedId)
	require.Equal(t, checker.WARN, errs[0].GetLevel())
	requireChange(t, errs, checker.ResponseOptionalPropertyRemovedId)
	require.Equal(t, checker.WARN, errs[1].GetLevel())
	requireChange(t, errs, checker.ResponseOptionalPropertyRemovedId)
	require.Equal(t, checker.WARN, errs[2].GetLevel())
}

// changing an existing write-only property in response body to optional is not breaking
func TestBreaking_WriteOnlyPropertyRequiredDisabled(t *testing.T) {
	s1, err := open(requiredPropertyFile("write-only-base.yaml"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("write-only-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// changing an existing required property in response body to write-only is not breaking
func TestBreaking_RequiredPropertyWriteOnlyEnabled(t *testing.T) {
	s1, err := open(requiredPropertyFile("write-only-changed-base.yaml"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("write-only-changed-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// changing an existing required property in response body to not-write-only is breaking
func TestBreaking_RequiredPropertyWriteOnlyDisabled(t *testing.T) {
	s1, err := open(requiredPropertyFile("write-only-changed-revision.yaml"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("write-only-changed-base.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 2)
	requireChange(t, errs, checker.ResponseRequiredPropertyBecameNonWriteOnlyId)
	require.Equal(t, checker.WARN, errs[0].GetLevel())
	requireChange(t, errs, checker.ResponseRequiredPropertyBecameNonWriteOnlyId)
	require.Equal(t, checker.WARN, errs[1].GetLevel())
}

// changing an existing property in request body to required is breaking
func TestBreaking_Body(t *testing.T) {
	s1, err := open(requiredPropertyFile("body1.yaml"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("body2.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.RequestPropertyBecameRequiredId)
	require.Equal(t, []any{"id"}, errs[0].GetArgs())
}

// changing an existing property in request body items to required is breaking
func TestBreaking_Items(t *testing.T) {
	s1, err := open(requiredPropertyFile("items1.yaml"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("items2.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.RequestPropertyBecameRequiredId)
	require.Equal(t, []any{"items/id"}, errs[0].GetArgs())
}

// changing an existing property in request body items to required is breaking even when the
// property has a default: a request that omits it is invalid under the new contract, and the
// default is a server-side fallback, not a validity rule.
func TestBreaking_ItemsWithDefault(t *testing.T) {
	s1, err := open(requiredPropertyFile("items1.yaml"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("items3.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	requireSingleChange(t, errs, checker.RequestPropertyBecameRequiredWithDefaultId)
	require.Equal(t, []any{"items/id"}, errs[0].GetArgs())
	require.Equal(t, checker.ERR, errs[0].GetLevel())
}

// changing an existing property in request body anyOf to required is breaking
func TestBreaking_AnyOf(t *testing.T) {
	s1, err := open(requiredPropertyFile("anyOf1.yaml"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("anyOf2.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.RequestPropertyBecameRequiredId)
}

// changing an existing property under another property in request body to required is breaking
func TestBreaking_NestedProp(t *testing.T) {
	s1, err := open(requiredPropertyFile("nested-property1.yaml"))
	require.NoError(t, err)

	s2, err := open(requiredPropertyFile("nested-property2.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	requireSingleChange(t, errs, checker.RequestPropertyBecameRequiredId)
}

// changing a response property to optional under AllOf, AnyOf or OneOf is breaking
func TestBreaking_OneOf(t *testing.T) {
	s1, err := open("../data/x-of/base.json")
	require.NoError(t, err)

	s2, err := open("../data/x-of/revision.json")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 3)
}
