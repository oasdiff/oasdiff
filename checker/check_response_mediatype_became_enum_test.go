package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

// an enum added to an unconstrained response body narrows what the server
// returns: reported once, as safe, not as each value added
func TestResponseMediaTypeBecameEnum(t *testing.T) {
	changes := positionChanges(t, "response-body", &openapi3.Schema{Type: &openapi3.Types{"string"}}, enumSchema("a", "b"))
	require.Len(t, changes, 1)
	require.Equal(t, checker.ResponseMediaTypeBecameEnumId, changes[0].GetId())
	require.Equal(t, checker.INFO, changes[0].GetLevel())
	require.Equal(t, "the response schema `application/json` for the response status `200` was restricted to a list of enum values", changes[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// the reverse is response-mediatype-enum-removed, an error
func TestResponseMediaTypeBecameEnumReverse(t *testing.T) {
	changes := positionChanges(t, "response-body", enumSchema("a", "b"), &openapi3.Schema{Type: &openapi3.Types{"string"}})
	require.Len(t, changes, 1)
	require.Equal(t, checker.ResponseMediaTypeEnumRemovedId, changes[0].GetId())
	require.Equal(t, checker.ERR, changes[0].GetLevel())
}
