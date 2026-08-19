package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

func Test_ChecksNoTags(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru"), io.Discard, io.Discard))
}

func Test_ChecksTagsDirection(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags request"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags response"), io.Discard, io.Discard))
}

func Test_ChecksTagsAction(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags add"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags remove"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags change"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags generalize"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags specialize"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags increase"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags decrease"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags set"), io.Discard, io.Discard))
}

func Test_ChecksTagsArea(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags schema"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags parameters"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags requestBody"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags responses"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags paths"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags headers"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags security"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags tags"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags components"), io.Discard, io.Discard))
}

func Test_ChecksTagsKind(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags existence"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags requiredness"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags mutability"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags type"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags constraints"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags values"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags structure"), io.Discard, io.Discard))
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru --tags lifecycle"), io.Discard, io.Discard))
}

// Every tag must name exactly one dimension of its vocabulary: filtering
// groups requested values by dimension, so a value appearing in two
// dimensions would be ambiguous. A future dimension whose natural value
// collides must pick another name.
func TestTagVocabulariesUnique(t *testing.T) {
	for name, tags := range map[string][]string{
		"checks changelog":          internal.GetChangelogTagsForTest(),
		"checks changelog coverage": internal.GetCoverageTagsForTest(),
	} {
		seen := map[string]bool{}
		for _, tag := range tags {
			require.False(t, seen[tag], "%s: tag %q appears in two dimensions", name, tag)
			seen[tag] = true
		}
	}
}

// Values of the same dimension are ORed, dimensions are ANDed.
func TestTags_OrWithinDimension(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog --format json --tags request,response,decrease"), &stdout, io.Discard))

	var rows []map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &rows))
	require.NotEmpty(t, rows)
	sawRequest, sawResponse := false, false
	for _, row := range rows {
		require.Contains(t, []any{"request", "response"}, row["direction"])
		require.Contains(t, row["actions"], "decrease")
		sawRequest = sawRequest || row["direction"] == "request"
		sawResponse = sawResponse || row["direction"] == "response"
	}
	require.True(t, sawRequest, "OR within the direction dimension must include request rows")
	require.True(t, sawResponse, "OR within the direction dimension must include response rows")
}
