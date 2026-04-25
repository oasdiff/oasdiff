package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: exercise processModifiedPropertiesDiff traversal through 3.1 sub-schema paths
// (if, then, else, contains, prefixItems, propertyNames, unevaluatedItems, unevaluatedProperties, contentSchema)
func TestSubSchemaTraversalMinLengthChanged(t *testing.T) {
	s1, err := open("../data/checker/sub_schema_traversal_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/sub_schema_traversal_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMinLengthUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	// minLength changes in sub-schemas should be detected
	require.True(t, containsId(errs, checker.RequestPropertyMinLengthIncreasedId), "expected minLength increased via sub-schema traversal")
}

// CL: exercise processAddedPropertiesDiff through 3.1 sub-schema paths
func TestSubSchemaTraversalAddedProperty(t *testing.T) {
	s1, err := open("../data/checker/sub_schema_added_property_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/sub_schema_added_property_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	// New required properties added in then/else sub-schemas should be detected
	require.True(t, containsId(errs, checker.NewRequiredRequestPropertyId), "expected new-required-request-property via sub-schema traversal")
}

// CL: exercise processDeletedPropertiesDiff through 3.1 sub-schema paths
func TestSubSchemaTraversalDeletedProperty(t *testing.T) {
	// Swapped base/revision: the "newField" properties are being removed
	s1, err := open("../data/checker/sub_schema_added_property_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/sub_schema_added_property_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyUpdatedCheck), d, osm, checker.INFO)
	// Properties deleted from request is INFO level (less restrictive)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.RequestPropertyRemovedId), "expected request-property-removed via sub-schema traversal")
}

// CL: exercise processModifiedPropertiesDiff traversal for minimum changes through if/then sub-schemas
func TestSubSchemaTraversalMinChanged(t *testing.T) {
	s1, err := open("../data/checker/sub_schema_traversal_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/sub_schema_traversal_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMinIncreasedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	// minimum change in then sub-schema should be detected
	require.True(t, containsId(errs, checker.RequestPropertyMinIncreasedId), "expected min increased in then sub-schema")
}
