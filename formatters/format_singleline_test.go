package formatters_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var singleLineFormatter = formatters.SingleLineFormatter{
	Localizer: MockLocalizer,
}

func TestSingleLineLookup(t *testing.T) {
	f, err := formatters.Lookup(string(formatters.FormatSingleLine), formatters.DefaultFormatterOpts())
	require.NoError(t, err)
	require.IsType(t, formatters.SingleLineFormatter{}, f)
}

func TestSingleLineFormatter_RenderChangelog(t *testing.T) {
	testChanges := checker.Changes{
		checker.ComponentChange{
			Id:        "change_id",
			Level:     checker.ERR,
			Component: "test",
		},
	}

	out, err := singleLineFormatter.RenderChangelog(testChanges, formatters.NewRenderOpts(), "", "")
	require.NoError(t, err)
	require.Equal(t, "1 changes: 1 error, 0 warning, 0 info\nerror, in components/test This is a breaking change. [change_id]. \n\n", string(out))
}

func TestSingleLineFormatter_NotImplemented(t *testing.T) {
	var err error

	_, err = singleLineFormatter.RenderChecks(nil, formatters.NewRenderOpts())
	assert.Error(t, err)

	_, err = singleLineFormatter.RenderDiff(nil, formatters.NewRenderOpts())
	assert.Error(t, err)

	_, err = singleLineFormatter.RenderFlatten(nil, formatters.NewRenderOpts())
	assert.Error(t, err)

	_, err = singleLineFormatter.RenderSummary(nil, formatters.NewRenderOpts())
	assert.Error(t, err)
}
