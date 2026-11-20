package internal_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

// A refused external $ref gets its own exit code so a caller can offer the
// "set allow-external-refs" remedy without matching error text. The refusal
// comes from kin-openapi on an ordinary load and from oasdiff's own reader
// under --open, and the code is the same either way.
func TestRefusedExternalRef_ExitCode(t *testing.T) {
	const base = "../data/ref-chain-example/base/openapi.yaml"
	const revision = "../data/ref-chain-example/revision/openapi.yaml"

	for _, cmd := range []string{"diff", "breaking", "changelog", "summary"} {
		var stderr bytes.Buffer
		code := internal.Run(cmdToArgs("oasdiff "+cmd+" --allow-external-refs=false "+base+" "+revision), io.Discard, &stderr)
		require.Equal(t, 123, code, "%s: %s", cmd, stderr.String())
		require.Contains(t, stderr.String(), "external")
	}
}

// The same specs load when the refs are allowed, so the exit code above is the
// policy refusing them rather than a broken fixture.
func TestRefusedExternalRef_AllowedLoads(t *testing.T) {
	code := internal.Run(cmdToArgs("oasdiff diff --allow-external-refs=true ../data/ref-chain-example/base/openapi.yaml ../data/ref-chain-example/revision/openapi.yaml"), io.Discard, io.Discard)
	require.Zero(t, code)
}
