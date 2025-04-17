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

	out, err := textFormatter.RenderChangelog(testChanges, formatters.NewRenderOpts(), "", "")
	require.NoError(t, err)
	require.Equal(t, "1 changes: 1 error, 0 warning, 0 info\nerror\t[change_id] \t\n\tin components/test\n\t\tThis is a breaking change.\n\n", string(out))
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
