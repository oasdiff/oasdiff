package formatters_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var textFormatter = formatters.TEXTFormatter{
	Localizer: MockLocalizer,
}

func TestTextLookup(t *testing.T) {
	f, err := formatters.Lookup(string(formatters.FormatText), formatters.DefaultFormatterOpts())
	require.NoError(t, err)
	require.IsType(t, formatters.TEXTFormatter{}, f)
}

func TestTextFormatter_RenderChangelog(t *testing.T) {
	testChanges := checker.Changes{
		checker.ComponentChange{
			Id:        "change_id",
			Level:     checker.ERR,
			Component: "test",
		},
	}

	out, err := textFormatter.RenderChangelog(testChanges, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Equal(t, "1 changes: 1 error, 0 warning, 0 info\nerror\t[change_id]\n\tin components/test\n\t\tThis is a breaking change.\n\n", string(out))
}

func TestTextFormatter_RenderChecks(t *testing.T) {
	checks := formatters.Checks{
		{
			Id:          "change_id",
			Level:       "info",
			Description: "This is a breaking change.",
		},
	}

	out, err := textFormatter.RenderChecks(checks, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Equal(t, string(out), "ID        DESCRIPTION                LEVEL\nchange_id This is a breaking change. info\n")
}

func TestTextFormatter_RenderChangelog_EmptyChangesDifferentSpecs(t *testing.T) {
	out, err := textFormatter.RenderChangelog(checker.Changes{}, formatters.RenderOpts{})
	require.NoError(t, err)
	require.Equal(t, "No changes to report, but the specs are different.\nRun 'oasdiff diff' to see structural differences.", string(out))
}

func TestTextFormatter_RenderChangelog_EmptyChangesDifferentSpecs_BreakingMode(t *testing.T) {
	out, err := textFormatter.RenderChangelog(checker.Changes{}, formatters.RenderOpts{IsBreaking: true})
	require.NoError(t, err)
	require.Equal(t, "No breaking changes to report, but the specs are different.\nRun 'oasdiff diff' to see structural differences.", string(out))
}

func TestTextFormatter_RenderChangelog_EmptyChangesIdenticalSpecs(t *testing.T) {
	// DiffEmpty=true takes precedence; suggestion is suppressed because
	// there's nothing for `oasdiff diff` to show.
	out, err := textFormatter.RenderChangelog(checker.Changes{}, formatters.RenderOpts{DiffEmpty: true, IsBreaking: true})
	require.NoError(t, err)
	require.Equal(t, "No changes detected", string(out))
}

func TestTextFormatter_RenderDiff(t *testing.T) {
	out, err := textFormatter.RenderDiff(nil, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Equal(t, string(out), "No changes\n")
}

func TestTextFormatter_NotImplemented(t *testing.T) {
	var err error

	_, err = textFormatter.RenderFlatten(nil, formatters.NewRenderOpts())
	assert.Error(t, err)

	_, err = textFormatter.RenderSummary(nil, formatters.NewRenderOpts())
	assert.Error(t, err)
}

// An IDs-only listing (oasdiff checks validate) prints just the ID column: a
// DESCRIPTION or LEVEL header over blank cells would promise data that a
// validate rule has no static value for.
func TestTextFormatter_RenderChecks_IdOnlyDropsEmptyColumns(t *testing.T) {
	out, err := formatters.TEXTFormatter{Localizer: MockLocalizer}.RenderChecks(
		formatters.Checks{{Id: "some-rule"}, {Id: "other-rule"}}, formatters.NewRenderOpts())
	require.NoError(t, err)

	require.NotContains(t, string(out), "DESCRIPTION")
	require.NotContains(t, string(out), "LEVEL")
	require.Contains(t, string(out), "ID")
	require.Contains(t, string(out), "some-rule")
}

// A populated listing keeps every column.
func TestTextFormatter_RenderChecks_KeepsPopulatedColumns(t *testing.T) {
	out, err := formatters.TEXTFormatter{Localizer: MockLocalizer}.RenderChecks(
		formatters.Checks{{Id: "some-rule", Description: "d", Level: "error"}}, formatters.NewRenderOpts())
	require.NoError(t, err)

	for _, want := range []string{"ID", "DESCRIPTION", "LEVEL", "some-rule", "error"} {
		require.Contains(t, string(out), want)
	}
}
