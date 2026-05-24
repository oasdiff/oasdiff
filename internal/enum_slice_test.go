package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Multi-value enum flags accept their tokens in any case and normalize each
// to the canonical allowed value, so downstream consumers receive the exact
// allowed strings regardless of how they were typed.
func Test_enumSliceValue_Set_CaseInsensitive(t *testing.T) {
	v := newEnumSliceValue([]string{"info", "warn", "error"}, nil)
	require.NoError(t, v.Set("ERROR,Warn"))
	require.Equal(t, "error,warn", v.String())
}

func Test_enumSliceValue_Set_Invalid(t *testing.T) {
	v := newEnumSliceValue([]string{"info", "warn", "error"}, nil)
	err := v.Set("error,nope")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not one of the allowed values")
}
