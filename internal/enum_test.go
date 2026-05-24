package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Enum flags accept their values in any case and normalize to the canonical
// allowed value, so callers (and downstream lookups like formatters.Lookup
// and checker.NewLevel) get a known string regardless of how it was typed.
func Test_enumValue_Set_CaseInsensitive(t *testing.T) {
	for _, input := range []string{"WARN", "warn", "Warn", "wArN"} {
		v := newEnumValue([]string{"ERR", "WARN", "INFO"}, "ERR")
		require.NoError(t, v.Set(input))
		require.Equal(t, "WARN", v.String(), "input %q should normalize to the canonical allowed value", input)
	}
}

// Normalization works toward a lowercase canonical too (e.g. format values).
func Test_enumValue_Set_NormalizesToCanonical(t *testing.T) {
	v := newEnumValue([]string{"yaml", "json"}, "yaml")
	require.NoError(t, v.Set("JSON"))
	require.Equal(t, "json", v.String())
}

func Test_enumValue_Set_Invalid(t *testing.T) {
	v := newEnumValue([]string{"ERR", "WARN", "INFO"}, "ERR")
	err := v.Set("nope")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not one of the allowed values")
}
