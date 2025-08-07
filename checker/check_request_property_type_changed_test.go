package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/oasdiff/oasdiff/utils"
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
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestBodyTypeChangedId,
		Level:       checker.ERR,
		Args:        []any{utils.StringList{"object"}, "", utils.StringList{"array"}, ""},
		Operation:   "POST",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_type_changed_base.yaml"),
		OperationId: "addPet",
	}, errs[0])
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
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestBodyTypeChangedId,
		Level:       checker.ERR,
		Args:        []any{utils.StringList{"object"}, "", utils.StringList{"object"}, "uuid"},
		Operation:   "POST",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_type_changed_base.yaml"),
		OperationId: "addPet",
	}, errs[0])
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
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestPropertyTypeChangedId,
		Level:       checker.ERR,
		Args:        []any{"age", utils.StringList{"integer"}, "int32", utils.StringList{"string"}, "string"},
		Operation:   "POST",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_type_changed_revision.yaml"),
		OperationId: "addPet",
	}, errs[0])
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
	require.Len(t, errs, 2)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestPropertyTypeChangedId,
		Level:       checker.ERR,
		Args:        []any{"colors", utils.StringList{"array"}, "", utils.StringList{"object"}, ""},
		Operation:   "POST",
		Path:        "/dogs",
		Source:      load.NewSource("../data/checker/request_property_type_changed_revision_array_to_object.yaml"),
		OperationId: "addDog",
	}, errs[0])
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestBodyTypeChangedId,
		Level:       checker.ERR,
		Args:        []any{utils.StringList{"array"}, "", utils.StringList{"object"}, ""},
		Operation:   "POST",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_type_changed_revision_array_to_object.yaml"),
		OperationId: "addPet",
	}, errs[1])
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
	require.Len(t, errs, 2)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestPropertyTypeChangedId,
		Level:       checker.ERR,
		Args:        []any{"colors", utils.StringList{"object"}, "", utils.StringList{"array"}, ""},
		Operation:   "POST",
		Path:        "/dogs",
		Source:      load.NewSource("../data/checker/request_property_type_changed_base_array_to_object.yaml"),
		OperationId: "addDog",
	}, errs[0])
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestBodyTypeChangedId,
		Level:       checker.ERR,
		Args:        []any{utils.StringList{"object"}, "", utils.StringList{"array"}, ""},
		Operation:   "POST",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_type_changed_base_array_to_object.yaml"),
		OperationId: "addPet",
	}, errs[1])
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
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestPropertyTypeChangedId,
		Level:       checker.ERR,
		Args:        []any{"age", utils.StringList{"integer"}, "int32", utils.StringList{"integer"}, "uuid"},
		Operation:   "POST",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_type_changed_base.yaml"),
		OperationId: "addPet",
	}, errs[0])
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
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestPropertyTypeGeneralizedId,
		Level:       checker.INFO,
		Args:        []any{"age", utils.StringList{"integer"}, "int32", utils.StringList{"number"}, "int32"},
		Operation:   "POST",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_type_changed_base.yaml"),
		OperationId: "addPet",
	}, errs[0])
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
