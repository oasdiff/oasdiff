package load_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

func TestSource_NewStdin(t *testing.T) {
	require.True(t, load.NewSource("-").IsStdin())
}

func TestSource_NewFile(t *testing.T) {
	require.True(t, load.NewSource("../spec.yaml").IsFile())
}

func TestSource_String(t *testing.T) {
	require.Equal(t, "stdin", load.NewSource("-").String())
}

func TestSource_OutStdin(t *testing.T) {
	require.Equal(t, `stdin`, load.NewSource("-").Out())
}

func TestSource_Out(t *testing.T) {
	require.Equal(t, `"http://twitter.com"`, load.NewSource("http://twitter.com").Out())
}

func TestSource_IsGitRevision(t *testing.T) {
	require.True(t, load.NewSource("origin/main:openapi.yaml").IsGitRevision())
	require.True(t, load.NewSource("HEAD~1:spec.yaml").IsGitRevision())
	require.True(t, load.NewSource("HEAD:dir/spec.yaml").IsGitRevision())
}

func TestSource_IsNotGitRevision(t *testing.T) {
	require.False(t, load.NewSource("openapi.yaml").IsGitRevision())
	require.False(t, load.NewSource("-").IsGitRevision())
	require.False(t, load.NewSource("http://example.com/spec.yaml").IsGitRevision())
	require.False(t, load.NewSource("https://example.com/spec.yaml").IsGitRevision())
	require.False(t, load.NewSource(`C:\path\spec.yaml`).IsGitRevision())
}
