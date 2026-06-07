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

// CL: exercise processDeletedPropertiesDiff traversal through additionalProperties.
// Regression: prior to the fix, processDeletedPropertiesDiff and
// processAddedPropertiesDiff did not recurse through AdditionalPropertiesDiff
// (only processModifiedPropertiesDiff did), so removing a required property from
// the value-type of a `dict[str, X]`-style schema went undetected.
func TestAdditionalPropertiesTraversalDeletedRequiredProperty(t *testing.T) {
	s1, err := open("../data/checker/additional_properties_traversal_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/additional_properties_traversal_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseRequiredPropertyUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponseRequiredPropertyRemovedId), "expected response-required-property-removed via additionalProperties traversal")
}

// CL: exercise processAddedPropertiesDiff traversal through additionalProperties.
// Symmetric to the deletion test above — swapping base/revision exercises the
// added-required-property path through AdditionalPropertiesDiff.
func TestAdditionalPropertiesTraversalAddedRequiredProperty(t *testing.T) {
	s1, err := open("../data/checker/additional_properties_traversal_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/additional_properties_traversal_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseRequiredPropertyUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponseRequiredPropertyAddedId), "expected response-required-property-added via additionalProperties traversal")
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

// CL: exercise processModifiedPropertiesDiff traversal through the `not` sub-schema
func TestNotSubSchemaTraversalModifiedProperty(t *testing.T) {
	s1, err := open("../data/checker/not_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/not_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMaxLengthUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	// maxLength change inside `not` should be detected
	require.True(t, containsId(errs, checker.RequestPropertyMaxLengthDecreasedId), "expected maxLength decreased via not sub-schema traversal")
}

// CL: exercise processAddedPropertiesDiff traversal through the `not` sub-schema
func TestNotSubSchemaTraversalAddedProperty(t *testing.T) {
	s1, err := open("../data/checker/not_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/not_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	// new required property added inside `not` should be detected
	require.True(t, containsId(errs, checker.NewRequiredRequestPropertyId), "expected new-required-request-property via not sub-schema traversal")
}

// CL: exercise processDeletedPropertiesDiff traversal through the `not` sub-schema
func TestNotSubSchemaTraversalDeletedProperty(t *testing.T) {
	s1, err := open("../data/checker/not_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/not_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	// property removed from inside `not` should be detected
	require.True(t, containsId(errs, checker.RequestPropertyRemovedId), "expected request-property-removed via not sub-schema traversal")
}
