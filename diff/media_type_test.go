package diff_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func TestParseMediaType(t *testing.T) {
	mediaType := "image/png;charset=utf-8"
	parts, err := diff.ParseMediaType(mediaType)
	require.NoError(t, err)
	require.Equal(t, "image", parts.Type)
	require.Equal(t, "png", parts.Subtype)
	require.Equal(t, "utf-8", parts.Parameters["charset"])
}

func TestParseMediaTypeWithMultipleParameters(t *testing.T) {
	mediaType := "image/png;charset=utf-8;boundary=123"
	parts, err := diff.ParseMediaType(mediaType)
	require.NoError(t, err)
	require.Equal(t, "image", parts.Type)
	require.Equal(t, "png", parts.Subtype)
	require.Equal(t, "utf-8", parts.Parameters["charset"])
	require.Equal(t, "123", parts.Parameters["boundary"])
}

func TestParseMediaTypeWithSuffix(t *testing.T) {
	mediaType := "image/png+json"
	parts, err := diff.ParseMediaType(mediaType)
	require.NoError(t, err)
	require.Equal(t, "image", parts.Type)
	require.Equal(t, "png", parts.Subtype)
	require.Equal(t, "json", parts.Suffix)
}

func TestParseMediaTypeWithInvalidMediaType(t *testing.T) {
	mediaType := "image/png+json+"
	_, err := diff.ParseMediaType(mediaType)
	require.Error(t, err)
}

func getMediaTypeContained(t *testing.T, mediaType1, mediaType2 string) bool {
	t.Helper()
	contained, err := diff.IsMediaTypeContained(mediaType1, mediaType2)
	require.NoError(t, err)
	return contained
}

func TestIsMediaTypeContained(t *testing.T) {
	require.True(t, getMediaTypeContained(t, "application/xml", "application/xml"))
	require.False(t, getMediaTypeContained(t, "application/xml", "application/json"))
}

func TestIsMediaTypeContainedWithSuffix(t *testing.T) {
	require.True(t, getMediaTypeContained(t, "application/xml", "application/atom+xml"))
	require.True(t, getMediaTypeContained(t, "application/json", "application/problem+json"))
	require.False(t, getMediaTypeContained(t, "application/problem+json", "application/json"))
	require.False(t, getMediaTypeContained(t, "application/xml", "application/problem+json"))
	require.True(t, getMediaTypeContained(t, "application/problem+json", "application/problem+json"))
}

// multiple suffixes are not supported
func TestIsMediaTypeContainedWithMultipleSuffixes(t *testing.T) {
	_, err := diff.IsMediaTypeContained("application/problem+json", "application/mitigation+problem+json")
	require.Error(t, err)
}
