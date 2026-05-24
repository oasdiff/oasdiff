package formatters_test

import (
	"encoding/json"
	"testing"

	"github.com/TwiN/go-color"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

func sampleFindings() formatters.Findings {
	return formatters.Findings{
		{Id: "missing-field", Text: "field x is missing", Level: checker.ERR, Section: "info", Source: formatters.Source{File: "spec.yaml", Line: 4, Column: 3}},
		{Id: "version-mismatch", Text: "field webhooks is for 3.1+", Level: checker.WARN, Section: "webhooks", Source: formatters.Source{File: "spec.yaml"}},
		{Id: "bad-example", Text: "invalid example:\nSchema:\n  stuff", Level: checker.INFO, Operation: "GET", Path: "/x", Section: "paths", Source: formatters.Source{File: "spec.yaml", Line: 9, Column: 7}},
	}
}

func TestFindings_GetLevelCount(t *testing.T) {
	counts := sampleFindings().GetLevelCount()
	require.Equal(t, 1, counts[checker.ERR])
	require.Equal(t, 1, counts[checker.WARN])
	require.Equal(t, 1, counts[checker.INFO])
}

func TestFindings_HasLevelOrHigher(t *testing.T) {
	require.True(t, sampleFindings().HasLevelOrHigher(checker.ERR))
	require.True(t, formatters.Findings{{Level: checker.WARN}}.HasLevelOrHigher(checker.WARN))
	require.False(t, formatters.Findings{{Level: checker.INFO}}.HasLevelOrHigher(checker.WARN))
	require.False(t, formatters.Findings{}.HasLevelOrHigher(checker.INFO))
}

func TestTEXTFormatter_RenderValidate(t *testing.T) {
	out, err := formatters.TEXTFormatter{Localizer: MockLocalizer}.RenderValidate(sampleFindings(), formatters.RenderOpts{ColorMode: checker.ColorNever})
	require.NoError(t, err)

	s := string(out)
	require.Contains(t, s, "3 findings: 1 error, 1 warning, 1 info")
	require.Contains(t, s, "error\t[missing-field] at spec.yaml:4:3")
	require.Contains(t, s, "warning\t[version-mismatch] at spec.yaml\n") // no line/column suffix
	require.Contains(t, s, "info\t[bad-example] at spec.yaml:9:7")
	require.Contains(t, s, "\tin API GET /x") // operation context line
	require.Contains(t, s, "\n\t\tSchema:")   // multi-line message indented under the operation
	require.NotContains(t, s, "\x1b[")        // no color when ColorNever
}

func TestTEXTFormatter_RenderValidate_Color(t *testing.T) {
	out, err := formatters.TEXTFormatter{Localizer: MockLocalizer}.RenderValidate(sampleFindings(), formatters.RenderOpts{ColorMode: checker.ColorAlways})
	require.NoError(t, err)

	s := string(out)
	require.Contains(t, s, color.InRed("error"))            // severity label colorized
	require.Contains(t, s, color.InYellow("missing-field")) // rule id yellow
	require.Contains(t, s, color.InGreen("GET"))            // endpoint green
}

func TestYAMLFormatter_RenderValidate(t *testing.T) {
	out, err := formatters.YAMLFormatter{Localizer: MockLocalizer}.RenderValidate(sampleFindings(), formatters.NewRenderOpts())
	require.NoError(t, err)

	var got []map[string]any
	require.NoError(t, yaml.Unmarshal(out, &got))
	require.Len(t, got, 3)
	require.Equal(t, "missing-field", got[0]["id"])
	src := got[0]["source"].(map[string]any)
	require.Equal(t, "spec.yaml", src["file"])
}

func TestJSONFormatter_RenderValidate(t *testing.T) {
	out, err := formatters.JSONFormatter{Localizer: MockLocalizer}.RenderValidate(sampleFindings(), formatters.NewRenderOpts())
	require.NoError(t, err)

	var got []map[string]any
	require.NoError(t, json.Unmarshal(out, &got))
	require.Len(t, got, 3)
	require.Equal(t, float64(3), got[0]["level"]) // checker.ERR
}

// Formatters that don't support validate output fall back to the
// not-implemented default.
func TestRenderValidate_NotImplemented(t *testing.T) {
	_, err := formatters.MarkupFormatter{}.RenderValidate(sampleFindings(), formatters.NewRenderOpts())
	require.Error(t, err)
}
