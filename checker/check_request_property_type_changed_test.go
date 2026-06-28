package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: changing request body type
func TestRequestBodyTypeChangedCheck(t *testing.T) {
	s1, err := open("../data/checker/request_property_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_type_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Type = &openapi3.Types{"array"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyTypeChangedCheck), d, osm, checker.ERR)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyTypeChangedId,
		Args:        []any{"type", "object", "array"},
		Operation:   "POST",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_type_changed_base.yaml"),
		OperationId: "addPet",
	}, errs)
}

// CL: changing request body type
func TestRequestBodyFormatChangedCheck(t *testing.T) {
	s1, err := open("../data/checker/request_property_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_type_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Format = "uuid"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyTypeChangedCheck), d, osm, checker.ERR)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyTypeChangedId,
		Args:        []any{"format", "none", "uuid"},
		Operation:   "POST",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_type_changed_base.yaml"),
		OperationId: "addPet",
	}, errs)
}

// CL: changing request property type
func TestRequestPropertyTypeChangedCheck(t *testing.T) {
	s1, err := open("../data/checker/request_property_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_type_changed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyTypeChangedCheck), d, osm, checker.ERR)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyTypeChangedId,
		Args:        []any{"age", "type/format", "integer/int32", "string/string"},
		Operation:   "POST",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_type_changed_revision.yaml"),
		OperationId: "addPet",
	}, errs)
}

// CL: narrowing a request property union type is breaking
func TestRequestPropertyTypeUnionNarrowedCheck(t *testing.T) {
	s1, err := open("../data/checker/request_property_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_type_changed_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["age"].Value.Type = &openapi3.Types{"string", "integer"}
	s1.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["age"].Value.Format = ""
	s2.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["age"].Value.Type = &openapi3.Types{"string"}
	s2.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["age"].Value.Format = ""

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyTypeChangedCheck), d, osm, checker.ERR)
	require.Len(t, errs, 1)
	requireChange(t, errs, checker.RequestPropertyTypeChangedId)
}

// CL: removing a request property type constraint is not breaking
func TestRequestPropertyTypeDeletedCheck(t *testing.T) {
	s1, err := open("../data/checker/request_property_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_type_changed_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["age"].Value.Type = &openapi3.Types{"string"}
	s1.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["age"].Value.Format = ""
	s2.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["age"].Value.Type = nil
	s2.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["age"].Value.Format = ""

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	requireChange(t, errs, checker.RequestPropertyTypeGeneralizedId)
}

// CL: changing request body and property types from array to object
func TestRequestBodyAndPropertyTypesChangedCheckArrayToObject(t *testing.T) {
	s1, err := open("../data/checker/request_property_type_changed_base_array_to_object.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_type_changed_revision_array_to_object.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyTypeChangedCheck), d, osm, checker.ERR)
	requireApiChanges(t, []checker.ApiChange{
		{
			Id:          checker.RequestPropertyTypeChangedId,
			Args:        []any{"colors", "type", "array<integer>", "object"},
			Operation:   "POST",
			Path:        "/dogs",
			Source:      load.NewSource("../data/checker/request_property_type_changed_revision_array_to_object.yaml"),
			OperationId: "addDog",
		},
		{
			Id:          checker.RequestBodyTypeChangedId,
			Args:        []any{"type", "array<object>", "object"},
			Operation:   "POST",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/request_property_type_changed_revision_array_to_object.yaml"),
			OperationId: "addPet",
		},
	}, errs)
}

// CL: changing request body and property types from object to array
func TestRequestBodyAndPropertyTypesChangedCheckObjectToArray(t *testing.T) {
	s1, err := open("../data/checker/request_property_type_changed_revision_array_to_object.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_type_changed_base_array_to_object.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyTypeChangedCheck), d, osm, checker.ERR)
	requireApiChanges(t, []checker.ApiChange{
		{
			Id:          checker.RequestPropertyTypeChangedId,
			Args:        []any{"colors", "type", "object", "array<integer>"},
			Operation:   "POST",
			Path:        "/dogs",
			Source:      load.NewSource("../data/checker/request_property_type_changed_base_array_to_object.yaml"),
			OperationId: "addDog",
		},
		{
			Id:          checker.RequestBodyTypeChangedId,
			Args:        []any{"type", "object", "array<object>"},
			Operation:   "POST",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/request_property_type_changed_base_array_to_object.yaml"),
			OperationId: "addPet",
		},
	}, errs)
}

// CL: changing request property format
func TestRequestPropertyFormatChangedCheck(t *testing.T) {
	s1, err := open("../data/checker/request_property_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_type_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["age"].Value.Format = "uuid"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyTypeChangedCheck), d, osm, checker.ERR)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyTypeChangedId,
		Args:        []any{"age", "format", "int32", "uuid"},
		Operation:   "POST",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_type_changed_base.yaml"),
		OperationId: "addPet",
	}, errs)
}

// CL: generalizing request property format
func TestRequestPropertyFormatChangedCheckNonBreaking(t *testing.T) {
	s1, err := open("../data/checker/request_property_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_type_changed_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["age"].Value.Type = &openapi3.Types{"integer"}
	s2.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["age"].Value.Type = &openapi3.Types{"number"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyTypeChangedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyTypeGeneralizedId,
		Args:        []any{"age", "type", "integer", "number"},
		Operation:   "POST",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_type_changed_base.yaml"),
		OperationId: "addPet",
	}, errs)
}

// CL: no changes when paths diff is nil
func TestRequestPropertyTypeChangedNoPathsDiff(t *testing.T) {
	config := &checker.Config{}
	d := &diff.Diff{}
	osm := &diff.OperationsSourcesMap{}

	errs := checker.RequestPropertyTypeChangedCheck(d, osm, config)
	require.Len(t, errs, 0)
}

// CL: no changes when operations diff is nil
func TestRequestPropertyTypeChangedNoOperationsDiff(t *testing.T) {
	config := &checker.Config{}
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: diff.ModifiedPaths{
				"/test": &diff.PathDiff{},
			},
		},
	}
	osm := &diff.OperationsSourcesMap{}

	errs := checker.RequestPropertyTypeChangedCheck(d, osm, config)
	require.Len(t, errs, 0)
}

// CL: no changes when request body diff is nil
func TestRequestPropertyTypeChangedNoRequestBodyDiff(t *testing.T) {
	config := &checker.Config{}
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: diff.ModifiedPaths{
				"/test": &diff.PathDiff{
					OperationsDiff: &diff.OperationsDiff{
						Modified: diff.ModifiedOperations{
							"POST": &diff.MethodDiff{},
						},
					},
				},
			},
		},
	}
	osm := &diff.OperationsSourcesMap{}

	errs := checker.RequestPropertyTypeChangedCheck(d, osm, config)
	require.Len(t, errs, 0)
}

// CL: no changes when schema diff is nil
func TestRequestPropertyTypeChangedNoSchemaDiff(t *testing.T) {
	config := &checker.Config{}
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: diff.ModifiedPaths{
				"/test": &diff.PathDiff{
					OperationsDiff: &diff.OperationsDiff{
						Modified: diff.ModifiedOperations{
							"POST": &diff.MethodDiff{
								RequestBodyDiff: &diff.RequestBodyDiff{
									ContentDiff: &diff.ContentDiff{
										MediaTypeModified: diff.ModifiedMediaTypes{
											"application/json": &diff.MediaTypeDiff{},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	osm := &diff.OperationsSourcesMap{}

	errs := checker.RequestPropertyTypeChangedCheck(d, osm, config)
	require.Len(t, errs, 0)
}

// CL: no changes when property is read-only
func TestRequestPropertyTypeChangedReadOnlyProperty(t *testing.T) {
	s1, err := open("../data/checker/request_property_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_type_changed_base.yaml")
	require.NoError(t, err)

	// Make property read-only and change its type
	s2.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["age"].Value.ReadOnly = true
	s2.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["age"].Value.Type = &openapi3.Types{"string"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.RequestPropertyTypeChangedCheck(d, osm, &checker.Config{})
	require.Len(t, errs, 0)
}

// setRequestBodyType sets the request body schema type on the /test POST.
func setRequestBodyType(t *testing.T, s *load.SpecInfo, types *openapi3.Types) {
	t.Helper()
	s.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value.Type = types
}

// BC: widening a request type set ([string] -> [string, integer]) is not
// breaking; the server accepts every type it did before plus more. Mirror of the
// response narrowing case.
func TestRequestBodyTypeWideningMultiTypeNotBreaking(t *testing.T) {
	s1, err := open("../data/type-change/simple-request.yaml")
	require.NoError(t, err)
	s2, err := open("../data/type-change/simple-request.yaml")
	require.NoError(t, err)
	setRequestBodyType(t, s1, &openapi3.Types{"string"})
	setRequestBodyType(t, s2, &openapi3.Types{"string", "integer"})

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyTypeChangedCheck), d, osm, checker.INFO)
	require.False(t, containsId(errs, checker.RequestBodyTypeChangedId),
		"widening a request type set is non-breaking; must not report request-body-type-changed")
}

// BC: removing the type entirely from a request ([string] -> no type) is not
// breaking; the server now accepts any value (a generalization).
func TestRequestBodyTypeRemovedNotBreaking(t *testing.T) {
	s1, err := open("../data/type-change/simple-request.yaml")
	require.NoError(t, err)
	s2, err := open("../data/type-change/simple-request.yaml")
	require.NoError(t, err)
	setRequestBodyType(t, s1, &openapi3.Types{"string"})
	setRequestBodyType(t, s2, nil)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyTypeChangedCheck), d, osm, checker.INFO)
	require.False(t, containsId(errs, checker.RequestBodyTypeChangedId),
		"removing the type from a request accepts any value; non-breaking")
}

// BC: narrowing a request type set ([string, integer] -> [string]) is breaking
// under a strongly-typed media type; a client sending integer is now rejected.
func TestRequestBodyTypeNarrowingStillBreaking(t *testing.T) {
	s1, err := open("../data/type-change/simple-request.yaml")
	require.NoError(t, err)
	s2, err := open("../data/type-change/simple-request.yaml")
	require.NoError(t, err)
	setRequestBodyType(t, s1, &openapi3.Types{"string", "integer"})
	setRequestBodyType(t, s2, &openapi3.Types{"string"})

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyTypeChangedCheck), d, osm, checker.ERR)
	require.True(t, containsId(errs, checker.RequestBodyTypeChangedId),
		"narrowing a request type set is breaking")
}

// BC: adding a type constraint to a previously untyped request (no type ->
// [string]) is breaking; it restricts the accepted values.
func TestRequestBodyTypeAddedFromUntypedStillBreaking(t *testing.T) {
	s1, err := open("../data/type-change/simple-request.yaml")
	require.NoError(t, err)
	s2, err := open("../data/type-change/simple-request.yaml")
	require.NoError(t, err)
	setRequestBodyType(t, s1, nil)
	setRequestBodyType(t, s2, &openapi3.Types{"string"})

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyTypeChangedCheck), d, osm, checker.ERR)
	require.True(t, containsId(errs, checker.RequestBodyTypeChangedId),
		"constraining a previously untyped request is breaking")
}

// BC: a request type widening that co-occurs with a breaking format change is
// breaking; the safe type axis must not mask the format axis. [integer] ->
// [integer, string] widens the type (not breaking on its own), but int64 ->
// int32 narrows the format (a client sending an int64 value is now rejected),
// which is breaking.
func TestRequestBodyTypeWidenWithBreakingFormatStillBreaking(t *testing.T) {
	s1, err := open("../data/type-change/simple-request.yaml")
	require.NoError(t, err)
	s2, err := open("../data/type-change/simple-request.yaml")
	require.NoError(t, err)
	base := s1.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value
	base.Type = &openapi3.Types{"integer"}
	base.Format = "int64"
	rev := s2.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value
	rev.Type = &openapi3.Types{"integer", "string"}
	rev.Format = "int32"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyTypeChangedCheck), d, osm, checker.ERR)
	require.True(t, containsId(errs, checker.RequestBodyTypeChangedId),
		"a type widening must not mask a co-occurring breaking format change (int64 -> int32)")
}
