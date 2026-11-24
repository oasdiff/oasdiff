package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// closedObjectChanges returns the level of each closed-object addition between
// two fixtures, keyed by property name.
func closedObjectChanges(t *testing.T, base, revision string, options ...load.Option) map[any]checker.Level {
	t.Helper()
	s1, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../data/checker/closed_object_"+base+".yaml"), options...)
	require.NoError(t, err)
	s2, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../data/checker/closed_object_"+revision+".yaml"), options...)
	require.NoError(t, err)
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)
	result := map[any]checker.Level{}
	for _, e := range errs {
		if e.GetId() == checker.ResponsePropertyAddedToClosedObjectId {
			result[e.(checker.ApiChange).Args[0]] = e.GetLevel()
		}
	}
	return result
}

// required and optional additions to closed objects are reported at any depth; the open object is not
func TestResponsePropertyAddedToClosedObject(t *testing.T) {
	require.Equal(t, map[any]checker.Level{
		"b":        checker.ERR,
		"closed/y": checker.ERR,
	}, closedObjectChanges(t, "base", "revision"))
}

// removing a property from a closed response object is handled by the removal checks
func TestResponsePropertyRemovedFromClosedObject(t *testing.T) {
	require.Empty(t, closedObjectChanges(t, "revision", "base"))
}

// unevaluatedProperties: false closes an object like additionalProperties: false,
// and also closes the schemas it applies in place through allOf and oneOf. A name
// the closing schema already evaluated, a property nested under a property, and an
// object whose unevaluatedProperties is a schema are not reported. An addition
// under allOf carries the allOf disclaimer, so it is reported as a warning.
func TestResponsePropertyAddedToClosedObjectUnevaluated(t *testing.T) {
	require.Equal(t, map[any]checker.Level{
		"b": checker.ERR,
		"oneOf[#/components/schemas/BranchA]/extra": checker.ERR,
		"allOf[#/components/schemas/Parent]/p2":     checker.WARN,
	}, closedObjectChanges(t, "unevaluated_base", "unevaluated_revision"))
}

// with allOf flattened, the addition is reported as an error on the merged schema
func TestResponsePropertyAddedToClosedObjectUnevaluatedFlattened(t *testing.T) {
	require.Equal(t, map[any]checker.Level{
		"b": checker.ERR,
		"oneOf[#/components/schemas/BranchA]/extra": checker.ERR,
		"p2": checker.ERR,
	}, closedObjectChanges(t, "unevaluated_base", "unevaluated_revision", load.WithFlattenAllOf()))
}

// removing a property from an object closed by unevaluatedProperties is not reported here
func TestResponsePropertyRemovedFromClosedObjectUnevaluated(t *testing.T) {
	require.Empty(t, closedObjectChanges(t, "unevaluated_revision", "unevaluated_base"))
}
