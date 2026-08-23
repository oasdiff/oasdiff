package validate

import (
	"fmt"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// originKey reads kin's `Origin` field by reflection, so it is coupled to that
// field's name and type rather than to a list of error types. This pins the
// coupling: if a kin release renames or retypes the field, this fails instead
// of every finding silently losing its location.
func TestOriginKey_ReadsTheOriginField(t *testing.T) {
	loc := &openapi3.Location{Line: 42, Column: 7}
	origin := &openapi3.Origin{Key: loc}

	// Two unrelated kin error types, neither refined by locationForKinError,
	// both resolving through the reflective path.
	require.Equal(t, loc, originKey(&openapi3.DuplicateTagError{Name: "pets", Origin: origin}))
	require.Equal(t, loc, originKey(&openapi3.ConflictingPathsError{Origin: origin}))
}

// kin wraps errors in section / path / operation context, so the Origin sits on
// a leaf rather than on the error the caller holds.
func TestOriginKey_WalksTheUnwrapChain(t *testing.T) {
	loc := &openapi3.Location{Line: 11, Column: 3}
	leaf := &openapi3.DuplicateTagError{Name: "pets", Origin: &openapi3.Origin{Key: loc}}

	require.Equal(t, loc, originKey(fmt.Errorf("path: %w", fmt.Errorf("section: %w", leaf))))
}

// An error with no Origin at all resolves to nothing rather than to a
// misleading location borrowed from elsewhere.
func TestOriginKey_NoOrigin(t *testing.T) {
	require.Nil(t, originKey(&openapi3.WebhookNilError{}))
	require.Nil(t, originKey(fmt.Errorf("plain error")))
	require.Nil(t, originKey(nil))
}

// A nil Origin field, which is what a load without origin tracking produces,
// is not a location.
func TestOriginKey_NilOrigin(t *testing.T) {
	require.Nil(t, originKey(&openapi3.DuplicateTagError{Name: "pets"}))
}
