package internal_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

func Test_Summary(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff summary ../data/openapi-test1.yaml ../data/openapi-test3.yaml"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), `diff: true`)
}
