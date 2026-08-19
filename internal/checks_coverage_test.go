package internal_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

func TestChecksCoverage_TextFormats(t *testing.T) {
	for _, format := range []string{"text", "json", "yaml"} {
		var stdout bytes.Buffer
		require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog coverage --format "+format), &stdout, &stdout), format)
		require.NotEmpty(t, stdout.String(), format)
	}
}

// the audit passes, so the uncovered listing is empty and renders as a
// valid empty document
func TestChecksCoverage_UncoveredIsEmpty(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog coverage --format json --tags uncovered"), &stdout, &stdout))
	require.JSONEq(t, "[]", stdout.String())
}

func TestChecksCoverage_TagsFilter(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog coverage --format json --tags covered,request,decrease"), &stdout, &stdout))

	var rows []checker.CoverageEdit
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &rows))
	require.NotEmpty(t, rows)
	for _, row := range rows {
		require.Equal(t, checker.CoverageCovered, row.Status)
		require.Equal(t, "request", row.Polarity)
		require.Equal(t, "decrease", row.Action)
		require.NotEmpty(t, row.Checks)
	}
}

func TestChecksCoverage_Patterns(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog coverage --format json --patterns"), &stdout, &stdout))

	var rows []checker.CoveragePattern
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &rows))
	require.NotEmpty(t, rows)
	for _, row := range rows {
		require.Contains(t, []string{"waiver", "non-contract"}, row.Kind)
		require.Positive(t, row.Edits, "stale pattern %q accounts for no edits", row.Pattern)
	}
}
