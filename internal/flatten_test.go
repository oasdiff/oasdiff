package internal_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

func Test_FlattenCmdOK(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff flatten ../data/allof/simple.yaml"), io.Discard, io.Discard))
}

func Test_FlattenCmdInvalid(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 122, internal.Run(cmdToArgs("oasdiff flatten ../data/allof/invalid.yaml"), io.Discard, &stderr))
	require.Equal(t, `Error: failed to flatten allOf in "../data/allof/invalid.yaml": unable to resolve Type conflict: all Type values must be identical
`, stderr.String())
}
