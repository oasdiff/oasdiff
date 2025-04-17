package formatters_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/stretchr/testify/require"
)

var jsonFormatter = formatters.JSONFormatter{
	Localizer: MockLocalizer,
}

func TestJsonLookup(t *testing.T) {
	f, err := formatters.Lookup(string(formatters.FormatJSON), formatters.DefaultFormatterOpts())
	require.NoError(t, err)
	require.IsType(t, formatters.JSONFormatter{}, f)
}

func TestJsonFormatter_RenderChangelog(t *testing.T) {
	testChanges := checker.Changes{
		checker.ComponentChange{
			Id:    "change_id",
			Level: checker.ERR,
		},
	}

	out, err := jsonFormatter.RenderChangelog(testChanges, formatters.NewRenderOpts(), "", "")
	require.NoError(t, err)
	require.Equal(t, "[{\"id\":\"change_id\",\"text\":\"This is a breaking change.\",\"level\":3,\"section\":\"components\"}]", string(out))
}

func TestJsonFormatter_RenderChecks(t *testing.T) {
	checks := formatters.Checks{
		{
			Id:          "change_id",
			Level:       "info",
			Description: "This is a breaking change.",
		},
	}

	out, err := jsonFormatter.RenderChecks(checks, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Equal(t, "[{\"id\":\"change_id\",\"level\":\"info\",\"description\":\"This is a breaking change.\"}]", string(out))
}

func TestJsonFormatter_RenderDiff(t *testing.T) {
	out, err := jsonFormatter.RenderDiff(nil, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Empty(t, string(out))
}

func TestJsonFormatter_RenderFlatten(t *testing.T) {
	out, err := jsonFormatter.RenderFlatten(nil, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Empty(t, string(out))
}

func TestJsonFormatter_RenderSummary(t *testing.T) {
	out, err := jsonFormatter.RenderSummary(nil, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Equal(t, `{"diff":false}`, string(out))
}
