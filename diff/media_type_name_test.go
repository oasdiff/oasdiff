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

func TestParseMediaTypeNameInvalidMissingSlash(t *testing.T) {
	_, err := diff.ParseMediaTypeName("invalid")
	require.Error(t, err)
}

func TestParseMediaTypeNameInvalidNoType(t *testing.T) {
	_, err := diff.ParseMediaTypeName("/invalid")
	require.Error(t, err)
}

func TestParseMediaTypeNameInvalidNoSubtype(t *testing.T) {
	_, err := diff.ParseMediaTypeName("invalid/")
	require.Error(t, err)
}

func TestParseMediaTypeNameInvalidEmptySuffix(t *testing.T) {
	_, err := diff.ParseMediaTypeName("image/png+")
	require.Error(t, err)
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
	require.Equal(t, "json", parts.Suffix)
}

// multiple suffixes are not supported
func TestParseMediaTypeNameWithMultipleSuffixes(t *testing.T) {
	_, err := diff.ParseMediaTypeName("image/png+json+xml")
	require.Error(t, err)
}

func TestParseMediaTypeNameWithParams(t *testing.T) {
	name, err := diff.ParseMediaTypeName("application/xml;q=0.9")
	require.NoError(t, err)
	require.Equal(t, "application", name.Type)
	require.Equal(t, "xml", name.Subtype)
	require.Equal(t, "0.9", name.Parameters["q"])
}

func TestIsMediaTypeNameContained(t *testing.T) {
	require.True(t, diff.IsMediaTypeNameContained("application/xml", "application/xml"))
	require.False(t, diff.IsMediaTypeNameContained("application/xml", "application/json"))
}

func TestIsMediaTypeNameDifferentTypes(t *testing.T) {
	require.False(t, diff.IsMediaTypeNameContained("application/xml", "image/png"))
}

func TestIsMediaTypeNameDifferentSuffix(t *testing.T) {
	require.False(t, diff.IsMediaTypeNameContained("application/atom+json", "application/atom+xml"))
}

func TestIsMediaTypeNameContainedWithSuffix(t *testing.T) {
	require.True(t, diff.IsMediaTypeNameContained("application/xml", "application/atom+xml"))
	require.True(t, diff.IsMediaTypeNameContained("application/json", "application/problem+json"))
	require.False(t, diff.IsMediaTypeNameContained("application/problem+json", "application/json"))
	require.False(t, diff.IsMediaTypeNameContained("application/xml", "application/problem+json"))
	require.True(t, diff.IsMediaTypeNameContained("application/problem+json", "application/problem+json"))
}

// Parameter values are ignored when determining if one media type contains another.
func TestIsMediaTypeNameContainedWithParams(t *testing.T) {
	require.True(t, diff.IsMediaTypeNameContained("application/xml;q=0.9", "application/xml;q=0.9"))
	require.True(t, diff.IsMediaTypeNameContained("application/xml;q=0.9", "application/xml;q=0.8"))
}

func TestIsMediaTypeNameContainedWithInvalidMediaType1(t *testing.T) {
	require.False(t, diff.IsMediaTypeNameContained("invalid", "application/xml"))
}

func TestIsMediaTypeNameContainedWithInvalidMediaType2(t *testing.T) {
	require.False(t, diff.IsMediaTypeNameContained("application/xml", "invalid"))
}
