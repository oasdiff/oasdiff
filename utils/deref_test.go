package utils_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/utils"
	"github.com/stretchr/testify/require"
)

func TestDeref(t *testing.T) {
	require.Nil(t, utils.Deref[string](nil))
	s := "x"
	require.Equal(t, "x", utils.Deref(&s))
	zero := uint64(0)
	require.Equal(t, uint64(0), utils.Deref(&zero)) // zero value is not absence
}
