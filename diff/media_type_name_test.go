package diff_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func TestParseMediaTypeName(t *testing.T) {
	mediaType := "image/png;charset=utf-8"
	parts, err := diff.ParseMediaTypeName(mediaType)
	require.NoError(t, err)
	require.Equal(t, "image", parts.Type)
	require.Equal(t, "png", parts.Subtype)
	require.Equal(t, "utf-8", parts.Parameters["charset"])
}

func TestParseMediaTypeNameWithMultipleParameters(t *testing.T) {
	mediaType := "image/png;charset=utf-8;boundary=123"
	parts, err := diff.ParseMediaTypeName(mediaType)
	require.NoError(t, err)
	require.Equal(t, "image", parts.Type)
	require.Equal(t, "png", parts.Subtype)
	require.Equal(t, "utf-8", parts.Parameters["charset"])
	require.Equal(t, "123", parts.Parameters["boundary"])
}

func TestParseMediaTypeNameWithSuffix(t *testing.T) {
	mediaType := "image/png+json"
	parts, err := diff.ParseMediaTypeName(mediaType)
	require.NoError(t, err)
	require.Equal(t, "image", parts.Type)
	require.Equal(t, "png", parts.Subtype)
	require.Equal(t, "json", parts.Suffixes[0])
}

func TestParseMediaTypeNameWithInvalidMediaType(t *testing.T) {
	mediaType := "image/png+json+"
	_, err := diff.ParseMediaTypeName(mediaType)
	require.Error(t, err)
}

func TestIsMediaTypeNameContained(t *testing.T) {
	require.True(t, diff.IsMediaTypeNameContained("application/xml", "application/xml"))
	require.False(t, diff.IsMediaTypeNameContained("application/xml", "application/json"))
}

func TestIsMediaTypeNameContainedWithSuffix(t *testing.T) {
	require.True(t, diff.IsMediaTypeNameContained("application/xml", "application/atom+xml"))
	require.True(t, diff.IsMediaTypeNameContained("application/json", "application/problem+json"))
	require.False(t, diff.IsMediaTypeNameContained("application/problem+json", "application/json"))
	require.False(t, diff.IsMediaTypeNameContained("application/xml", "application/problem+json"))
	require.True(t, diff.IsMediaTypeNameContained("application/problem+json", "application/problem+json"))
}

// multiple suffixes
func TestIsMediaTypeNameContainedWithMultipleSuffixes(t *testing.T) {
	require.True(t, diff.IsMediaTypeNameContained("image/png+json", "image/png+json+xml"))
	require.False(t, diff.IsMediaTypeNameContained("image/png+json+xml", "image/png+json"))
}
