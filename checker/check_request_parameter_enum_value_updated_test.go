package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: removing an enum value from request parameter
func TestRequestParameterEnumValueRemovedCheck(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_enum_value_updated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_enum_value_updated_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterEnumValueUpdatedCheck), d, osm, checker.ERR)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestParameterEnumValueRemovedId,
		Args:        []any{"available", "query", "status"},
		Operation:   "GET",
		Path:        "/test",
		Source:      load.NewSource("../data/checker/request_parameter_enum_value_updated_revision.yaml"),
		OperationId: "getTest",
	}, errs)
}

// Regression: object-valued enum entries must not produce a false breaking change due to __origin__ metadata.
func TestRequestParameterObjectEnumNoFalseBreakingChange(t *testing.T) {
	s, err := open("../data/checker/object_enum_same.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s, s)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterEnumValueUpdatedCheck), d, osm, checker.INFO)
	require.Empty(t, errs)
}

// CL: adding an enum value to request parameter
func TestRequestParameterEnumValueAddedCheck(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_enum_value_updated_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_enum_value_updated_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterEnumValueUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestParameterEnumValueAddedId,
		Args:        []any{"available", "query", "status"},
		Operation:   "GET",
		Path:        "/test",
		Source:      load.NewSource("../data/checker/request_parameter_enum_value_updated_base.yaml"),
		OperationId: "getTest",
	}, errs)
}

// CL: removing an enum value from a deepObject request parameter property
func TestRequestParameterPropertyEnumValueRemovedCheck(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_property_enum_value_updated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_property_enum_value_updated_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterEnumValueUpdatedCheck), d, osm, checker.ERR)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestParameterPropertyEnumValueRemovedId,
		Args:        []any{"value-b", "origin", "query", "filter"},
		Operation:   "GET",
		Path:        "/test",
		Source:      load.NewSource("../data/checker/request_parameter_property_enum_value_updated_revision.yaml"),
		OperationId: "getTest",
	}, errs)
}

// CL: adding an enum value to a deepObject request parameter property
// Swap base/revision to exercise the added-value path
func TestRequestParameterPropertyEnumValueAddedCheck(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_property_enum_value_updated_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_property_enum_value_updated_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterEnumValueUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestParameterPropertyEnumValueAddedId,
		Args:        []any{"value-b", "origin", "query", "filter"},
		Operation:   "GET",
		Path:        "/test",
		Source:      load.NewSource("../data/checker/request_parameter_property_enum_value_updated_base.yaml"),
		OperationId: "getTest",
	}, errs)
}

// CL: verifying the localized message for parameter property enum value removed
func TestRequestParameterPropertyEnumValueRemovedMessage(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_property_enum_value_updated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_property_enum_value_updated_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterEnumValueUpdatedCheck), d, osm, checker.ERR)
	require.Len(t, errs, 1)
	require.Equal(t, "removed the enum value `value-b` from the property `origin` of the `query` request parameter `filter`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}
