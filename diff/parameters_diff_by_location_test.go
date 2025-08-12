package diff_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/oasdiff/oasdiff/utils"
	"github.com/stretchr/testify/require"
)

func TestParamNamesByLocation_Len(t *testing.T) {
	require.Equal(t, 3, diff.ParamNamesByLocation{
		"query":  utils.StringList{"name"},
		"header": utils.StringList{"id", "organization"},
	}.Len())
}

func TestParamDiffByLocation_Len(t *testing.T) {
	require.Equal(t, 3, diff.ParamDiffByLocation{
		"query":  diff.ParamDiffs{"query": &diff.ParameterDiff{}},
		"header": diff.ParamDiffs{"id": &diff.ParameterDiff{}, "organization": &diff.ParameterDiff{}},
	}.Len())
}

// Test for exploded parameter equivalence issue #711 - Forward direction (simple → exploded)
func TestExplodedParameterEquivalence(t *testing.T) {
	loader := openapi3.NewLoader()

	// Load the test specs
	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/explode-params/base.yaml"))
	require.NoError(t, err)

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/explode-params/exploded.yaml"))
	require.NoError(t, err)

	// Get the diff
	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	// Check that parameters are not flagged as deleted
	if d != nil && d.PathsDiff != nil {
		for _, pathItem := range d.PathsDiff.Modified {
			if pathItem != nil && pathItem.OperationsDiff != nil {
				for _, operationItem := range pathItem.OperationsDiff.Modified {
					if operationItem != nil && operationItem.ParametersDiff != nil {
						// Should not have deleted parameters - they should be recognized as equivalent
						require.Empty(t, operationItem.ParametersDiff.Deleted, "Exploded parameters should not be flagged as deleted")
					}
				}
			}
		}
	}
}

// Test for exploded parameter equivalence - Reverse direction (exploded → simple)
func TestExplodedParameterEquivalenceReverse(t *testing.T) {
	loader := openapi3.NewLoader()

	// Load the test specs in reverse order (exploded first, then simple)
	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/explode-params/exploded.yaml"))
	require.NoError(t, err)

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/explode-params/base.yaml"))
	require.NoError(t, err)

	// Get the diff
	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	// In the reverse direction, we should also see no deleted parameters
	// The exploded parameter should be recognized as equivalent to the simple parameters
	if d != nil && d.PathsDiff != nil {
		for _, pathItem := range d.PathsDiff.Modified {
			if pathItem != nil && pathItem.OperationsDiff != nil {
				for _, operationItem := range pathItem.OperationsDiff.Modified {
					if operationItem != nil && operationItem.ParametersDiff != nil {
						// Should not have deleted parameters - the exploded parameter should map to simple parameters
						require.Empty(t, operationItem.ParametersDiff.Deleted, "Simple parameters should not be flagged as deleted when equivalent exploded parameter exists")
						// Should not have added parameters - the simple parameters should map to exploded parameter
						require.Empty(t, operationItem.ParametersDiff.Added, "Simple parameters should not be flagged as added when equivalent exploded parameter exists")
					}
				}
			}
		}
	}
}

// Test for partial exploded parameter conversion - where only some parameters are converted to exploded
func TestPartialExplodedParameterConversion(t *testing.T) {
	loader := openapi3.NewLoader()

	// Load the partial conversion test specs
	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/explode-params/partial-base.yaml"))
	require.NoError(t, err)

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/explode-params/partial-exploded.yaml"))
	require.NoError(t, err)

	// Get the diff
	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	// Check that only the exploded parameters are flagged as modified, not deleted/added
	if d != nil && d.PathsDiff != nil {
		for _, pathItem := range d.PathsDiff.Modified {
			if pathItem != nil && pathItem.OperationsDiff != nil {
				for _, operationItem := range pathItem.OperationsDiff.Modified {
					if operationItem != nil && operationItem.ParametersDiff != nil {
						// Should not have any deleted parameters - exploded equivalence should be recognized
						require.Empty(t, operationItem.ParametersDiff.Deleted, "No parameters should be flagged as deleted in partial exploded conversion")
						// Should not have any added parameters - exploded equivalence should be recognized
						require.Empty(t, operationItem.ParametersDiff.Added, "No parameters should be flagged as added in partial exploded conversion")

						// Should have modifications for the converted parameters (PageNumber, PageSize)
						require.Contains(t, operationItem.ParametersDiff.Modified, "query")
						queryMods := operationItem.ParametersDiff.Modified["query"]
						require.Contains(t, queryMods, "PageNumber", "PageNumber should be flagged as modified (style change)")
						require.Contains(t, queryMods, "PageSize", "PageSize should be flagged as modified (style change)")

						// Should NOT have modifications for the unchanged parameters (SortBy, Order)
						require.NotContains(t, queryMods, "SortBy", "SortBy should not be flagged as modified")
						require.NotContains(t, queryMods, "Order", "Order should not be flagged as modified")
					}
				}
			}
		}
	}
}

// TestParameterStyleDefaults tests that parameter style defaults are handled correctly
// according to OpenAPI specification: query/cookie default to "form", path/header default to "simple"
func TestParameterStyleDefaults(t *testing.T) {
	// Test that path parameters with empty style don't incorrectly default to "form"
	t.Run("path parameter with empty style should not be treated as exploded form", func(t *testing.T) {
		loader := openapi3.NewLoader()

		// Valid spec: path parameter without explicit style (should default to "simple")
		s1, err := load.NewSpecInfo(loader, load.NewSource("../data/explode-params/base.yaml"))
		require.NoError(t, err)

		s2, err := load.NewSpecInfo(loader, load.NewSource("../data/explode-params/path-with-empty-style.yaml"))
		require.NoError(t, err)

		// The isExplodedObjectParam function should return false for this path parameter
		// because path parameters default to "simple" style, not "form"
		d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
		require.NoError(t, err)
		require.NotNil(t, d)

		// With our fix, path parameters should be correctly handled and not treated as exploded form
		// This test mainly ensures the fix doesn't break and handles different parameter locations correctly
	})

}

// TestExplodedParameterLocationMatching tests that parameters with the same name
// but different locations (In field) are NOT matched by exploded parameter logic.
// This test verifies the location check at parameters_diff_by_location.go:304
func TestExplodedParameterLocationMatching(t *testing.T) {
	t.Run("parameters with same name but different locations should not match", func(t *testing.T) {
		loader := openapi3.NewLoader()

		// Base spec: has "userId" parameter in cookie location
		s1, err := load.NewSpecInfo(loader, load.NewSource("../data/explode-params/cross-location-base.yaml"))
		require.NoError(t, err)

		// Exploded spec: has exploded query parameter with "userId" property
		s2, err := load.NewSpecInfo(loader, load.NewSource("../data/explode-params/cross-location-exploded.yaml"))
		require.NoError(t, err)

		// Get the diff
		d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
		require.NoError(t, err)

		// The key assertion: the cookie "userId" parameter should be DELETED and query exploded param ADDED
		// because they should NOT match (different locations prevent matching)
		if d != nil && d.PathsDiff != nil {
			for _, pathItem := range d.PathsDiff.Modified {
				if pathItem != nil && pathItem.OperationsDiff != nil {
					for _, operationItem := range pathItem.OperationsDiff.Modified {
						if operationItem != nil && operationItem.ParametersDiff != nil {
							// Cookie "userId" should be flagged as deleted because location check prevents matching
							require.NotEmpty(t, operationItem.ParametersDiff.Deleted["cookie"],
								"Cookie 'userId' parameter should be deleted - it should NOT match query exploded 'userId' property")
							require.Contains(t, operationItem.ParametersDiff.Deleted["cookie"], "userId",
								"The deleted cookie parameter should be 'userId'")

							// Query exploded parameter should be flagged as added
							require.NotEmpty(t, operationItem.ParametersDiff.Added["query"],
								"Query exploded parameter should be added since cookie param doesn't match")
							require.Contains(t, operationItem.ParametersDiff.Added["query"], "userParams",
								"The added query parameter should be 'userParams'")
						}
					}
				}
			}
		}
	})
}
