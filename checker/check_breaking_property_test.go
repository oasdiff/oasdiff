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

func getReqPropFile(file string) string {
	return fmt.Sprintf("../data/required-properties/%s", file)
}

// BC: new required property in request header is breaking
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
	require.Len(t, errs, 1)
	require.Equal(t, checker.NewRequiredRequestHeaderPropertyId, errs[0].GetId())
}

// BC: new optional property in request header is not breaking
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

// BC: changing an existing property in request header to required is breaking
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
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestHeaderPropertyBecameRequiredId, errs[0].GetId())
}

// BC: changing an existing property in request header to optional is not breaking
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

// BC: changing an existing property in response body to optional is breaking
func TestBreaking_RespBodyRequiredPropertyDisabled(t *testing.T) {
	s1, err := open(getReqPropFile("response-base.json"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("response-revision.json"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponsePropertyBecameOptionalId, errs[0].GetId())
}

// BC: changing a request body to enum is breaking
func TestBreaking_ReqBodyBecameEnum(t *testing.T) {
	s1, err := open("../data/enums/request-body-no-enum.yaml")
	require.NoError(t, err)

	s2, err := open("../data/enums/request-body-enum.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestBodyBecameEnumId, errs[0].GetId())
}

// BC: adding an enum value to request body is not breaking
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

// BC: changing a request body type and changing it to enum simultaneously is breaking
func TestBreaking_ReqBodyBecameEnumAndTypeChanged(t *testing.T) {
	s1, err := open("../data/enums/request-body-no-enum.yaml")
	require.NoError(t, err)

	s2, err := open("../data/enums/request-body-enum-int.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 2)
	require.Equal(t, checker.RequestBodyBecameEnumId, errs[0].GetId())
	require.Equal(t, checker.RequestBodyTypeChangedId, errs[1].GetId())
}

// BC: changing an existing property in request body to enum is breaking
func TestBreaking_ReqPropertyBecameEnum(t *testing.T) {
	s1, err := open("../data/enums/request-property-no-enum.yaml")
	require.NoError(t, err)

	s2, err := open("../data/enums/request-property-enum.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertyBecameEnumId, errs[0].GetId())
}

// BC: changing an existing path param to enum is breaking
func TestBreaking_ReqParameterBecameEnum(t *testing.T) {
	s1, err := open("../data/enums/request-parameter-op-no-enum.yaml")
	require.NoError(t, err)

	s2, err := open("../data/enums/request-parameter-op-enum.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterBecameEnumId, errs[0].GetId())
}

// BC: changing an existing property in request header to enum is breaking
func TestBreaking_ReqParameterHeaderPropertyBecameEnum(t *testing.T) {
	s1, err := open("../data/enums/request-parameter-property-no-enum.yaml")
	require.NoError(t, err)

	s2, err := open("../data/enums/request-parameter-property-enum.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestHeaderPropertyBecameEnumId, errs[0].GetId())
}

// BC: changing a response body to nullable is breaking
func TestBreaking_RespBodyNullable(t *testing.T) {
	s1, err := open("../data/nullable/base-body.yaml")
	require.NoError(t, err)

	s2, err := open("../data/nullable/revision-body.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponseBodyBecameNullableId, errs[0].GetId())
}

// BC: changing a request property to not nullable is breaking
func TestBreaking_ReqBodyPropertyNotNullable(t *testing.T) {
	s1, err := open("../data/nullable/base-req.yaml")
	require.NoError(t, err)

	s2, err := open("../data/nullable/revision-req.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertyBecomeNotNullableId, errs[0].GetId())
}

// BC: changing a response property to nullable is breaking
func TestBreaking_RespBodyPropertyNullable(t *testing.T) {
	s1, err := open("../data/nullable/base-property.yaml")
	require.NoError(t, err)

	s2, err := open("../data/nullable/revision-property.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponsePropertyBecameNullableId, errs[0].GetId())
}

// BC: changing an embedded response property to nullable is breaking
func TestBreaking_RespBodyEmbeddedPropertyNullable(t *testing.T) {
	s1, err := open("../data/nullable/base-embedded-property.yaml")
	require.NoError(t, err)

	s2, err := open("../data/nullable/revision-embedded-property.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponsePropertyBecameNullableId, errs[0].GetId())
}

// BC: changing a required property in response body to optional and also deleting it is breaking
func TestBreaking_RespBodyDeleteAndDisableRequiredProperty(t *testing.T) {
	s1, err := open(getReqPropFile("response-del-required-prop-base.yaml"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("response-del-required-prop-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
}

// BC: adding a non-existent required property in request body is not breaking
func TestBreaking_ReqBodyNewRequiredPropertyNew(t *testing.T) {
	s1, err := open(getReqPropFile("request-new-required-prop-base.yaml"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("request-new-required-prop-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: changing an existing property in response body to required is not breaking
func TestBreaking_RespBodyRequiredPropertyEnabled(t *testing.T) {
	s1, err := open(getReqPropFile("response-revision.json"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("response-base.json"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: changing an existing property in request body to optional is not breaking
func TestBreaking_ReqBodyRequiredPropertyDisabled(t *testing.T) {
	s1, err := open(getReqPropFile("request-base.yaml"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("request-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: changing an existing property in request body to required is breaking
func TestBreaking_ReqBodyRequiredPropertyEnabled(t *testing.T) {
	s1, err := open(getReqPropFile("request-revision.yaml"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("request-base.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertyBecameRequiredId, errs[0].GetId())
}

// BC: adding a new required property in request body is breaking
func TestBreaking_ReqBodyNewRequiredProperty(t *testing.T) {
	s1, err := open(getReqPropFile("request-new-base.yaml"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("request-new-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.NewRequiredRequestPropertyId, errs[0].GetId())
}

// BC: deleting a required property in request is breaking with warn
func TestBreaking_ReqBodyDeleteRequiredProperty(t *testing.T) {
	s1, err := open(getReqPropFile("request-new-revision.yaml"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("request-new-base.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertyRemovedId, errs[0].GetId())
	require.Equal(t, checker.WARN, errs[0].GetLevel())
}

// BC: deleting an embedded optional property in request is breaking with warn
func TestBreaking_ReqBodyDeleteRequiredProperty2(t *testing.T) {
	s1, err := open(getReqPropFile("request-property-items.yaml"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("request-property-items-2.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Contains(t, errs, checker.ApiChange{
		Id:        checker.RequestPropertyRemovedId,
		Args:      []any{"roleAssignments/items/role"},
		Level:     checker.WARN,
		Operation: "POST",
		Path:      "/api/roleMappings",
		Source:    load.NewSource("../data/required-properties/request-property-items-2.yaml"),
	})
}

// BC: adding a new required property in response body is not breaking
func TestBreaking_RespBodyNewRequiredProperty(t *testing.T) {
	s1, err := open(getReqPropFile("response-new-base.json"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("response-new-revision.json"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: deleting a required property in response body is breaking
func TestBreaking_RespBodyDeleteRequiredProperty(t *testing.T) {
	s1, err := open(getReqPropFile("response-new-revision.json"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("response-new-base.json"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponseRequiredPropertyRemovedId, errs[0].GetId())
}

// BC: adding a new required property under AllOf in response body is not breaking
func TestBreaking_RespBodyNewAllOfRequiredProperty(t *testing.T) {
	s1, err := open(getReqPropFile("response-allof-base.json"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("response-allof-revision.json"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: deleting a required property under AllOf in response body is breaking
func TestBreaking_RespBodyDeleteAllOfRequiredProperty(t *testing.T) {
	s1, err := open(getReqPropFile("response-allof-revision.json"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("response-allof-base.json"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponseRequiredPropertyRemovedId, errs[0].GetId())
}

// BC: adding a new required read-only property in request body is not breaking
func TestBreaking_ReadOnlyNewRequiredProperty(t *testing.T) {
	s1, err := open(getReqPropFile("read-only-new-base.yaml"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("read-only-new-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: changing an existing read-only property in request body to required is not breaking
func TestBreaking_ReadOnlyPropertyRequiredEnabled(t *testing.T) {
	s1, err := open(getReqPropFile("read-only-base.yaml"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("read-only-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: deleting a required write-only property in response body is not breaking
func TestBreaking_WriteOnlyDeleteRequiredProperty(t *testing.T) {
	s1, err := open(getReqPropFile("write-only-delete-base.yaml"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("write-only-delete-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertyRemovedId, errs[0].GetId())
	require.Equal(t, checker.WARN, errs[0].GetLevel())
}

// BC: deleting a non-required non-write-only property in response body is breaking with warning
func TestBreaking_WriteOnlyDeleteNonRequiredProperty(t *testing.T) {
	s1, err := open(getReqPropFile("write-only-delete-partial-base.yaml"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("write-only-delete-partial-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 3)
	require.Equal(t, checker.RequestPropertyRemovedId, errs[0].GetId())
	require.Equal(t, checker.WARN, errs[0].GetLevel())
	require.Equal(t, checker.ResponseOptionalPropertyRemovedId, errs[1].GetId())
	require.Equal(t, checker.WARN, errs[1].GetLevel())
	require.Equal(t, checker.ResponseOptionalPropertyRemovedId, errs[2].GetId())
	require.Equal(t, checker.WARN, errs[2].GetLevel())
}

// BC: changing an existing write-only property in response body to optional is not breaking
func TestBreaking_WriteOnlyPropertyRequiredDisabled(t *testing.T) {
	s1, err := open(getReqPropFile("write-only-base.yaml"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("write-only-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: changing an existing required property in response body to write-only is not breaking
func TestBreaking_RequiredPropertyWriteOnlyEnabled(t *testing.T) {
	s1, err := open(getReqPropFile("write-only-changed-base.yaml"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("write-only-changed-revision.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: changing an existing required property in response body to not-write-only is breaking
func TestBreaking_RequiredPropertyWriteOnlyDisabled(t *testing.T) {
	s1, err := open(getReqPropFile("write-only-changed-revision.yaml"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("write-only-changed-base.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 2)
	require.Equal(t, checker.ResponseRequiredPropertyBecameNonWriteOnlyId, errs[0].GetId())
	require.Equal(t, checker.WARN, errs[0].GetLevel())
	require.Equal(t, checker.ResponseRequiredPropertyBecameNonWriteOnlyId, errs[1].GetId())
	require.Equal(t, checker.WARN, errs[1].GetLevel())
}

// BC: changing an existing property in request body to required is breaking
func TestBreaking_Body(t *testing.T) {
	s1, err := open(getReqPropFile("body1.yaml"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("body2.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertyBecameRequiredId, errs[0].GetId())
	require.Equal(t, []any{"id"}, errs[0].GetArgs())
}

// BC: changing an existing property in request body items to required is breaking
func TestBreaking_Items(t *testing.T) {
	s1, err := open(getReqPropFile("items1.yaml"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("items2.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertyBecameRequiredId, errs[0].GetId())
	require.Equal(t, []any{"/items/id"}, errs[0].GetArgs())
}

// BC: changing an existing property in request body items to required with a default value is not breaking
func TestBreaking_ItemsWithDefault(t *testing.T) {
	s1, err := open(getReqPropFile("items1.yaml"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("items3.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: changing an existing property in request body anyOf to required is breaking
func TestBreaking_AnyOf(t *testing.T) {
	s1, err := open(getReqPropFile("anyOf1.yaml"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("anyOf2.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertyBecameRequiredId, errs[0].GetId())
}

// BC: changing an existing property under another property in request body to required is breaking
func TestBreaking_NestedProp(t *testing.T) {
	s1, err := open(getReqPropFile("nested-property1.yaml"))
	require.NoError(t, err)

	s2, err := open(getReqPropFile("nested-property2.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertyBecameRequiredId, errs[0].GetId())
}

// BC: changing a response property to optional under AllOf, AnyOf or OneOf is breaking
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
