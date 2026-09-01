package diff

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMediaTypeNameDiff_ParametersOnly(t *testing.T) {
	d, err := getMediaTypeNameDiff("text/html", "text/html; charset=utf-8")
	require.NoError(t, err)
	require.False(t, d.BareNameChanged())
	require.Equal(t, "text/html", d.BaseBareName())
	require.False(t, d.ParametersDiff.Mixed())
}

func TestMediaTypeNameDiff_BareNameChanged(t *testing.T) {
	d, err := getMediaTypeNameDiff("application/json; q=1", "application/problem+json; q=1")
	require.NoError(t, err)
	require.True(t, d.BareNameChanged())
	require.Equal(t, "application/json", d.BaseBareName())
}

func TestMediaTypeNameDiff_MixedParameterChanges(t *testing.T) {
	d, err := getMediaTypeNameDiff("text/html; charset=utf-8", "text/html; q=1")
	require.NoError(t, err)
	require.False(t, d.BareNameChanged())
	require.True(t, d.ParametersDiff.Mixed())
}

func TestStringMapDiff_MixedNil(t *testing.T) {
	var d *StringMapDiff
	require.False(t, d.Mixed())
}
