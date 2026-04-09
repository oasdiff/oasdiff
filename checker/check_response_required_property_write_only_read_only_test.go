package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: changing required response property to write-only
func TestResponseRequiredPropertyBecameWriteOnly(t *testing.T) {
	s1, err := open("../data/checker/response_required_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/response_required_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["name"].Value.WriteOnly = true
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseRequiredPropertyWriteOnlyReadOnlyCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{

		Id:          "response-required-property-became-write-only",
		Args:        []any{"data/name", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_required_property_write_only_read_only_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: changing required response property to not write-only
func TestResponseRequiredPropertyBecameNotWriteOnly(t *testing.T) {
	s1, err := open("../data/checker/response_required_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/response_required_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["writeOnlyName"].Value.WriteOnly = false
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseRequiredPropertyWriteOnlyReadOnlyCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{

		Id:          checker.ResponseRequiredPropertyBecameNonWriteOnlyId,
		Args:        []any{"data/writeOnlyName", "200"},
		Comment:     checker.ResponseRequiredPropertyBecameNonWriteOnlyId + "-comment",
		Level:       checker.WARN,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_required_property_write_only_read_only_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
	require.Equal(t, "It is valid only if the property was always returned before the specification has been changed", errs[0].GetComment(checker.NewDefaultLocalizer()))
	require.Equal(t, "the response required property `data/writeOnlyName` became not write-only for the status `200`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// CL: changing required response property to read-only
func TestResponseRequiredPropertyBecameReadOnly(t *testing.T) {
	s1, err := open("../data/checker/response_required_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/response_required_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s1.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["id"].Value.ReadOnly = false
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseRequiredPropertyWriteOnlyReadOnlyCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{

		Id:          checker.ResponseRequiredPropertyBecameReadOnlyId,
		Args:        []any{"data/id", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_required_property_write_only_read_only_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: changing required response property to not read-only
func TestResponseRequiredPropertyBecameNonReadOnly(t *testing.T) {
	s1, err := open("../data/checker/response_required_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/response_required_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["id"].Value.ReadOnly = false

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseRequiredPropertyWriteOnlyReadOnlyCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{

		Id:          checker.ResponseRequiredPropertyBecameNonReadOnlyId,
		Args:        []any{"data/id", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_required_property_write_only_read_only_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}
