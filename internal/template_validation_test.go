package internal_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

func cmdToArgsLocal(cmd string) []string {
	return strings.Fields(cmd)
}

func TestTemplateValidationInChangelog(t *testing.T) {
	// Create a temporary template file
	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "test-template.md")
	templateContent := "Custom template content: {{ .BaseVersion }}"
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	require.NoError(t, err)

	testCases := []struct {
		name          string
		format        string
		expectError   bool
		expectedCode  int
		errorContains string
	}{
		{
			name:        "markdown with template should work",
			format:      "markdown",
			expectError: false,
		},
		{
			name:        "markup with template should work",
			format:      "markup",
			expectError: false,
		},
		{
			name:        "html with template should work",
			format:      "html",
			expectError: false,
		},
		{
			name:          "json with template should fail",
			format:        "json",
			expectError:   true,
			expectedCode:  111,
			errorContains: "template flag is not supported for format \"json\"",
		},
		{
			name:          "yaml with template should fail",
			format:        "yaml",
			expectError:   true,
			expectedCode:  111,
			errorContains: "template flag is not supported for format \"yaml\"",
		},
		{
			name:          "text with template should fail",
			format:        "text",
			expectError:   true,
			expectedCode:  111,
			errorContains: "template flag is not supported for format \"text\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var stderr bytes.Buffer
			cmd := "oasdiff changelog ../data/run_test/changelog_base.yaml ../data/run_test/changelog_revision.yaml --format " + tc.format + " --template " + templatePath

			exitCode := internal.Run(cmdToArgsLocal(cmd), io.Discard, &stderr)

			if tc.expectError {
				require.Equal(t, tc.expectedCode, exitCode, "expected error code %d for format %s with template", tc.expectedCode, tc.format)
				require.Contains(t, stderr.String(), tc.errorContains, "error message should contain expected text")
			} else {
				require.Equal(t, 0, exitCode, "expected success for format %s with template", tc.format)
			}
		})
	}
}
