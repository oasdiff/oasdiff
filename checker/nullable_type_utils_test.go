package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

// The response side reports became-not-nullable for the type-array form
// (type: [string, "null"] to type: string), the same as the request side;
// before the nullabilityChange classifier unified the four emission blocks,
// this case was silent (the transition claimed the raw type change and no
// response reporter fired).
func TestResponseTypeArrayBecameNotNullable(t *testing.T) {
	ids := nullableWrapChanges(t, "../data/checker/nullable_type_array_response_base.yaml", "../data/checker/nullable_type_array_response_revision.yaml")
	require.Equal(t, []string{checker.ResponsePropertyBecameNotNullableId}, ids)
}

// And the reverse of the same pair: became-nullable.
func TestResponseTypeArrayBecameNullable(t *testing.T) {
	ids := nullableWrapChanges(t, "../data/checker/nullable_type_array_response_revision.yaml", "../data/checker/nullable_type_array_response_base.yaml")
	require.Equal(t, []string{checker.ResponsePropertyBecameNullableId}, ids)
}
