package diff

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeysDiff(t *testing.T) {
	m1 := map[string]int{"c": 1, "a": 1, "b": 1, "e": 1}
	m2 := map[string]int{"d": 1, "b": 1, "f": 1, "a": 1}

	for range 10 {
		added, deleted, common := keysDiff(m1, m2)
		require.Equal(t, []string{"d", "f"}, added)
		require.Equal(t, []string{"c", "e"}, deleted)
		require.Equal(t, []string{"a", "b"}, common)
	}
}

func TestKeysDiff_Empty(t *testing.T) {
	added, deleted, common := keysDiff(map[string]int{}, map[string]int(nil))
	require.Empty(t, added)
	require.Empty(t, deleted)
	require.Empty(t, common)
}
