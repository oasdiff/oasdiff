package nullability_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/internal/nullability"
	"github.com/stretchr/testify/require"
)

func TestTypeArrayChanges(t *testing.T) {
	require.True(t, nullability.AddedToTypes([]string{"null"}, []string{"string"}))
	require.False(t, nullability.AddedToTypes([]string{"string", "null"}, nil), "the base had no type, so it already accepted null")
	require.False(t, nullability.AddedToTypes([]string{"integer"}, []string{"string"}))

	require.True(t, nullability.RemovedFromTypes([]string{"null"}, []string{"string"}))
	require.False(t, nullability.RemovedFromTypes([]string{"string", "null"}, nil), "the revision has no type, so it still accepts null")

	require.True(t, nullability.OnlyNullChanged([]string{"null"}, nil))
	require.True(t, nullability.OnlyNullChanged(nil, []string{"null"}))
	require.False(t, nullability.OnlyNullChanged([]string{"null", "integer"}, nil))
}

func TestWithoutNull(t *testing.T) {
	require.Equal(t, []string{"string"}, nullability.WithoutNull([]string{"string", "null"}))
	require.Empty(t, nullability.WithoutNull([]string{"null"}))
}
