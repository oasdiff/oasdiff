package formatters_test

import (
	"strings"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/stretchr/testify/require"
)

var yamlFormatter = formatters.YAMLFormatter{
	Localizer: MockLocalizer,
}

func TestYamlLookup(t *testing.T) {
	f, err := formatters.Lookup(string(formatters.FormatYAML), formatters.DefaultFormatterOpts())
	require.NoError(t, err)
	require.IsType(t, formatters.YAMLFormatter{}, f)
}

func TestYamlFormatter_RenderChangelog(t *testing.T) {
	testChanges := checker.Changes{
		checker.ComponentChange{
			Id:    "change_id",
			Level: checker.ERR,
		},
	}

	out, err := yamlFormatter.RenderChangelog(testChanges, formatters.NewRenderOpts(), nil)
	require.NoError(t, err)
	require.Equal(t, "- id: change_id\n  text: This is a breaking change.\n  level: 3\n  section: components\n", string(out))
}

func TestYamlFormatter_RenderChangelogWithWrapInObject(t *testing.T) {
	testChanges := checker.Changes{
		checker.ComponentChange{
			Id:    "change_id",
			Level: checker.ERR,
		},
	}

	out, err := yamlFormatter.RenderChangelog(testChanges, formatters.RenderOpts{WrapInObject: true}, nil)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(out), "changes:"))
}

func TestYamlFormatter_RenderChecks(t *testing.T) {
	checks := formatters.Checks{
		{
			Id:          "change_id",
			Level:       "info",
			Description: "This is a breaking change.",
		},
	}

	out, err := yamlFormatter.RenderChecks(checks, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Equal(t, "- id: change_id\n  level: info\n  description: This is a breaking change.\n", string(out))
}

func TestYamlFormatter_RenderDiff(t *testing.T) {
	out, err := yamlFormatter.RenderDiff(nil, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Empty(t, string(out))
}

func TestYamlFormatter_RenderFlatten(t *testing.T) {
	out, err := yamlFormatter.RenderFlatten(nil, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Empty(t, string(out))
}

func TestYamlFormatter_RenderSummary(t *testing.T) {
	out, err := yamlFormatter.RenderSummary(nil, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Equal(t, string(out), "diff: false\n")
}
