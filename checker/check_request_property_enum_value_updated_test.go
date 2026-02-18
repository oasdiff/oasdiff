package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: removing request property enum values
func TestRequestPropertyEnumValueRemovedCheck(t *testing.T) {
	s1, err := open("../data/checker/request_property_enum_value_updated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_enum_value_updated_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["category"].Value.Enum = []any{"dog", "cat"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyEnumValueUpdatedCheck), d, osm, checker.ERR)
	require.Len(t, errs, 1)

	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestPropertyEnumValueRemovedId,
		Level:       checker.ERR,
		Args:        []any{"bird", "category"},
		Operation:   "POST",
		OperationId: "updatePet",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_enum_value_updated_base.yaml"),
	}, errs[0])
	require.Equal(t, "removed the enum value 'bird' of the request property 'category'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// CL: removing request read-only property enum values
func TestRequestReadOnlyPropertyEnumValueRemovedCheck(t *testing.T) {
	s1, err := open("../data/checker/request_property_enum_value_updated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_enum_value_updated_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["category"].Value.Enum = []any{"dog", "cat"}
	s2.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["category"].Value.ReadOnly = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyEnumValueUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)

	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestReadOnlyPropertyEnumValueRemovedId,
		Level:       checker.INFO,
		Args:        []any{"bird", "category"},
		Operation:   "POST",
		OperationId: "updatePet",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_enum_value_updated_base.yaml"),
	}, errs[0])
	require.Equal(t, "removed the enum value 'bird' of the request read-only property 'category'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// CL: adding request property enum values
func TestRequestPropertyEnumValueAddedCheck(t *testing.T) {
	s1, err := open("../data/checker/request_property_enum_value_updated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_enum_value_updated_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["category"].Value.Enum = []any{"dog", "cat"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyEnumValueUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)

	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestPropertyEnumValueAddedId,
		Level:       checker.INFO,
		Args:        []any{"bird", "category"},
		Operation:   "POST",
		OperationId: "updatePet",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_enum_value_updated_base.yaml"),
	}, errs[0])
	require.Equal(t, "added the new 'bird' enum value to the request property 'category'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}
