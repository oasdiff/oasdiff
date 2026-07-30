package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

// The listing flags moved to the subcommand, so they have to work there.
func Test_ChecksChangelogAcceptsTheListingFlags(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(
		cmdToArgs("oasdiff checks changelog -l ru --tags decrease,parameters --severity info,warn,error --format json"),
		&stdout, io.Discard))

	var checks []map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &checks))
	require.NotEmpty(t, checks)
}
