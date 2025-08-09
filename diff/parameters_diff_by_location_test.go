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
