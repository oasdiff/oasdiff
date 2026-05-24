package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

func TestColorMode_Always(t *testing.T) {
	colorMode, err := checker.NewColorMode("always")
	require.NoError(t, err)
	require.Equal(t, colorMode, checker.ColorAlways)
}

func TestColorMode_Never(t *testing.T) {
	colorMode, err := checker.NewColorMode("never")
	require.NoError(t, err)
	require.Equal(t, colorMode, checker.ColorNever)
}

func TestColorMode_Auto(t *testing.T) {
	colorMode, err := checker.NewColorMode("auto")
	require.NoError(t, err)
	require.Equal(t, colorMode, checker.ColorAuto)
}

func TestColorMode_Invalid(t *testing.T) {
	colorMode, err := checker.NewColorMode("invalid")
	require.Error(t, err)
	require.Equal(t, colorMode, checker.ColorInvalid)
}

func TestIsColorEnabled(t *testing.T) {
	require.True(t, checker.IsColorEnabled(checker.ColorAlways))
	require.False(t, checker.IsColorEnabled(checker.ColorNever))
}
