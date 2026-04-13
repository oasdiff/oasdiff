package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: adding dependent schema to request body
func TestRequestBodyDependentSchemaAdded(t *testing.T) {
	s1, err := open("../data/checker/dependent_schemas_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/dependent_schemas_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyDependentSchemasUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.GetId() == checker.RequestBodyDependentSchemaAddedId {
			found = true
			break
		}
	}
	require.True(t, found, "expected request-body-dependent-schema-added")
}

// CL: removing dependent schema from request body
func TestRequestBodyDependentSchemaRemoved(t *testing.T) {
	s1, err := open("../data/checker/dependent_schemas_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/dependent_schemas_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyDependentSchemasUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.GetId() == checker.RequestBodyDependentSchemaRemovedId {
			found = true
			break
		}
	}
	require.True(t, found, "expected request-body-dependent-schema-removed")
}

// CL: adding dependent schema to request property
func TestRequestPropertyDependentSchemaAdded(t *testing.T) {
	s1, err := open("../data/checker/dependent_schemas_prop_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/dependent_schemas_prop_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyDependentSchemasUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.RequestPropertyDependentSchemaAddedId), "expected request-property-dependent-schema-added")
}

// CL: removing dependent schema from request property
func TestRequestPropertyDependentSchemaRemoved(t *testing.T) {
	s1, err := open("../data/checker/dependent_schemas_prop_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/dependent_schemas_prop_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyDependentSchemasUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.RequestPropertyDependentSchemaRemovedId), "expected request-property-dependent-schema-removed")
}
