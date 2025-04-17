package formatters_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var jUnitFormatter = formatters.JUnitFormatter{
	Localizer: MockLocalizer,
}

func TestJUnitLineLookup(t *testing.T) {
	f, err := formatters.Lookup(string(formatters.FormatJUnit), formatters.DefaultFormatterOpts())
	require.NoError(t, err)
	require.IsType(t, formatters.JUnitFormatter{}, f)
}

func TestJUnitFormatter_RenderChangelog_Success(t *testing.T) {
	testChanges := checker.Changes{}

	// check output
	output, err := jUnitFormatter.RenderChangelog(testChanges, formatters.NewRenderOpts(), "", "")
	assert.NoError(t, err)
	expectedOutput := `<?xml version="1.0" encoding="UTF-8"?>
<testsuites>
  <testsuite package="com.oasdiff" time="0" tests="0" errors="0" failures="0" name="OASDiff">
    <testcase name="no breaking changes detected" classname="OASDiff" time="0"></testcase>
  </testsuite>
</testsuites>`
	assert.Equal(t, expectedOutput, string(output))
}

func TestJUnitFormatter_NotImplemented(t *testing.T) {

	var err error
	_, err = jUnitFormatter.RenderDiff(nil, formatters.NewRenderOpts())
	assert.Error(t, err)

	_, err = jUnitFormatter.RenderSummary(nil, formatters.NewRenderOpts())
	assert.Error(t, err)

	_, err = jUnitFormatter.RenderChecks(nil, formatters.NewRenderOpts())
	assert.Error(t, err)

	_, err = jUnitFormatter.RenderFlatten(nil, formatters.NewRenderOpts())
	assert.Error(t, err)
}
