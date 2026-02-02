package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: removing an enum value from a response property
func TestResponsePropertyEnumValueRemoved(t *testing.T) {
	s1, err := open("../data/checker/response_property_enum_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_enum_added_base.yaml")
	require.NoError(t, err)

	s2.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["typeEnum"].Value.Enum = []any{"TYPE1"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseParameterEnumValueRemovedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponsePropertyEnumValueRemovedId,
		Args:        []any{"TYPE2", "data/typeEnum", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_enum_added_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: removing an enum value from a response write-only property
func TestResponseWriteOnlyPropertyEnumValueRemoved(t *testing.T) {
	s1, err := open("../data/checker/response_property_enum_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_enum_added_base.yaml")
	require.NoError(t, err)

	s2.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["writeOnlyEnum"].Value.Enum = []any{"TYPE1"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseParameterEnumValueRemovedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponsePropertyEnumValueRemovedId,
		Args:        []any{"TYPE2", "data/writeOnlyEnum", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_enum_added_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}
