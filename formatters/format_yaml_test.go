package formatters_test

import (
	"strings"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/load"
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

	out, err := yamlFormatter.RenderChangelog(testChanges, formatters.NewRenderOpts(), "", "")
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

	out, err := yamlFormatter.RenderChangelog(testChanges, formatters.RenderOpts{WrapInObject: true}, "", "")
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

func TestYamlFormatter_RenderChangelog_WithSources(t *testing.T) {
	testChanges := checker.Changes{
		checker.ApiChange{
			Id:        "change_id", // Use ID that MockLocalizer recognizes
			Level:     checker.ERR,
			Operation: "POST",
			Path:      "/api/test",
			Source:    &load.Source{}, // Need this to avoid nil pointer dereference
			CommonChange: checker.CommonChange{
				BaseSource:     checker.NewSource("base.yaml", 10, 5),
				RevisionSource: checker.NewSource("revision.yaml", 12, 7),
			},
		},
	}

	out, err := yamlFormatter.RenderChangelog(testChanges, formatters.NewRenderOpts(), "", "")
	require.NoError(t, err)

	expected := `- id: change_id
  text: This is a breaking change.
  level: 3
  operation: POST
  path: /api/test
  section: paths
  baseSource:
    file: base.yaml
    line: 10
    column: 5
  revisionSource:
    file: revision.yaml
    line: 12
    column: 7
`
	require.Equal(t, expected, string(out))
}
