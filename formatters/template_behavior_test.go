package formatters_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/formatters"
	"github.com/stretchr/testify/require"
)

func TestSupportsTemplateMethod(t *testing.T) {
	// Test cases: format name and whether it should support templates
	testCases := []struct {
		format           string
		supportsTemplate bool
	}{
		{"text", false},
		{"json", false},
		{"yaml", false},
		{"markdown", true},
		{"markup", true},
		{"html", true},
		{"singleline", false},
		{"githubactions", false},
		{"junit", false},
	}

	for _, tc := range testCases {
		t.Run(tc.format, func(t *testing.T) {
			f, err := formatters.Lookup(tc.format, formatters.DefaultFormatterOpts())
			require.NoError(t, err, "should be able to lookup formatter for %s", tc.format)

			actual := f.SupportsTemplate()
			require.Equal(t, tc.supportsTemplate, actual, "SupportsTemplate() for %s should return %v", tc.format, tc.supportsTemplate)
		})
	}
}

func TestGetSupportedTemplateFormats(t *testing.T) {
	supportedFormats := formatters.GetSupportedTemplateFormats()

	expectedFormats := []string{"html", "markdown", "markup"}
	require.Equal(t, expectedFormats, supportedFormats, "GetSupportedTemplateFormats should return sorted list of template-supporting formats")

	// Verify all returned formats actually support templates
	for _, format := range supportedFormats {
		f, err := formatters.Lookup(format, formatters.DefaultFormatterOpts())
		require.NoError(t, err, "should be able to lookup format %s", format)
		require.True(t, f.SupportsTemplate(), "format %s should support templates", format)
	}
}
