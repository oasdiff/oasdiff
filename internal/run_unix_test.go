//go:build unix

package internal_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

func Test_InvalidFileUnix(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 102, internal.Run(cmdToArgs("oasdiff diff no-file ../data/openapi-test3.yaml"), io.Discard, &stderr))
	require.Equal(t, `Error: failed to load base spec from "no-file": open no-file: no such file or directory
`, stderr.String())
}

func Test_ComposedModeInvalidFileUnix(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 122, internal.Run(cmdToArgs("oasdiff diff ../data/allof/* ../data/allof/* --composed --flatten"), io.Discard, &stderr))
	require.Equal(t, `Error: failed to flatten allOf in "../data/allof/invalid.yaml": unable to resolve Type conflict: all Type values must be identical
`, stderr.String())
}
