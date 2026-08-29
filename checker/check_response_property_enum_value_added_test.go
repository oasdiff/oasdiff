package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// adding an enum value to a response property
func TestResponsePropertyEnumValueAdded(t *testing.T) {
	s1, err := open("../data/checker/response_property_enum_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_enum_added_base.yaml")
	require.NoError(t, err)

	s2.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["typeEnum"].Value.Enum = []any{"Test"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyEnumValueAddedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponsePropertyEnumValueAddedId,
		Args:        []any{"Test", "data/typeEnum", "200"},
		Comment:     checker.ResponsePropertyEnumValueAddedId + "-comment",
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_enum_added_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
	require.Equal(t, checker.ERR, errs[0].GetLevel())
	require.Equal(t, "The server may now return a value the previous contract excluded, so a client written against it may not handle the response. If the value set is meant to grow, declare it with x-extensible-enum.", errs[0].GetComment(checker.NewDefaultLocalizer()))
}

// adding an enum value to a response write-only property
func TestResponseWriteOnlyPropertyEnumValueAdded(t *testing.T) {
	s1, err := open("../data/checker/response_property_enum_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_enum_added_base.yaml")
	require.NoError(t, err)

	s2.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["writeOnlyEnum"].Value.Enum = []any{"Test"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyEnumValueAddedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponseWriteOnlyPropertyEnumValueAddedId,
		Args:        []any{"Test", "data/writeOnlyEnum", "200"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_enum_added_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}
