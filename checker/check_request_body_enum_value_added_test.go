package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func enumSchema(values ...any) *openapi3.Schema {
	return &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: values}
}

func positionChanges(t *testing.T, position string, base, revision *openapi3.Schema) checker.Changes {
	t.Helper()
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), positionDoc(position, base), positionDoc(position, revision))
	require.NoError(t, err)
	return checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)
}

// adding a value to a request body enum accepts more, so it is safe
func TestRequestBodyEnumValueAdded(t *testing.T) {
	changes := positionChanges(t, "request-body", enumSchema("a", "b"), enumSchema("a", "b", "c"))
	require.Len(t, changes, 1)
	require.Equal(t, checker.RequestBodyEnumValueAddedId, changes[0].GetId())
	require.Equal(t, checker.INFO, changes[0].GetLevel())
	require.Equal(t, "request body enum value added `c`", changes[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// an enum added to an unconstrained request body is reported once, as
// request-body-became-enum, not as each value added
func TestRequestBodyEnumAddedIsNotEachValueAdded(t *testing.T) {
	changes := positionChanges(t, "request-body", &openapi3.Schema{Type: &openapi3.Types{"string"}}, enumSchema("a", "b"))
	require.Len(t, changes, 1)
	require.Equal(t, checker.RequestBodyBecameEnumId, changes[0].GetId())
}
