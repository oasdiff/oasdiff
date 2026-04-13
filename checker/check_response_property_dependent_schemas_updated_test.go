package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: adding dependent schema to response body
func TestResponseBodyDependentSchemaAdded(t *testing.T) {
	s1, err := open("../data/checker/dependent_schemas_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/dependent_schemas_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyDependentSchemasUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.GetId() == checker.ResponseBodyDependentSchemaAddedId {
			found = true
			break
		}
	}
	require.True(t, found, "expected response-body-dependent-schema-added")
}

// CL: removing dependent schema from response body
func TestResponseBodyDependentSchemaRemoved(t *testing.T) {
	s1, err := open("../data/checker/dependent_schemas_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/dependent_schemas_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyDependentSchemasUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponseBodyDependentSchemaRemovedId), "expected response-body-dependent-schema-removed")
}

// CL: adding dependent schema to response property
func TestResponsePropertyDependentSchemaAdded(t *testing.T) {
	s1, err := open("../data/checker/dependent_schemas_prop_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/dependent_schemas_prop_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyDependentSchemasUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponsePropertyDependentSchemaAddedId), "expected response-property-dependent-schema-added")
}

// CL: removing dependent schema from response property
func TestResponsePropertyDependentSchemaRemoved(t *testing.T) {
	s1, err := open("../data/checker/dependent_schemas_prop_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/dependent_schemas_prop_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyDependentSchemasUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponsePropertyDependentSchemaRemovedId), "expected response-property-dependent-schema-removed")
}
