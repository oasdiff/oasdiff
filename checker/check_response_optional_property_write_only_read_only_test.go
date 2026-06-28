package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: changing optional response property to write-only
func TestResponseOptionalPropertyBecameWriteOnly(t *testing.T) {
	s1, err := open("../data/checker/response_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/response_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["name"].Value.WriteOnly = true
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseOptionalPropertyWriteOnlyReadOnlyCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{

		Id:          checker.ResponseOptionalPropertyBecameWriteOnlyId,
		Args:        []any{"data/name", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_optional_property_write_only_read_only_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// CL: changing optional response property to not write-only
func TestResponseOptionalPropertyBecameNotWriteOnly(t *testing.T) {
	s1, err := open("../data/checker/response_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/response_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["writeOnlyName"].Value.WriteOnly = false
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseOptionalPropertyWriteOnlyReadOnlyCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{

		Id:          checker.ResponseOptionalPropertyBecameNonWriteOnlyId,
		Args:        []any{"data/writeOnlyName", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_optional_property_write_only_read_only_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// CL: changing optional response property to read-only
func TestResponseOptionalPropertyBecameReadOnly(t *testing.T) {
	s1, err := open("../data/checker/response_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/response_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s1.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["id"].Value.ReadOnly = false
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseOptionalPropertyWriteOnlyReadOnlyCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{

		Id:          checker.ResponseOptionalPropertyBecameReadOnlyId,
		Args:        []any{"data/id", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_optional_property_write_only_read_only_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// CL: changing optional response property to not read-only
func TestResponseOptionalPropertyBecameNonReadOnly(t *testing.T) {
	s1, err := open("../data/checker/response_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/response_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["id"].Value.ReadOnly = false

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseOptionalPropertyWriteOnlyReadOnlyCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{

		Id:          checker.ResponseOptionalPropertyBecameNonReadOnlyId,
		Args:        []any{"data/id", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_optional_property_write_only_read_only_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}
