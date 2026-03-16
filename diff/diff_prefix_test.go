package diff_test

import (
	"fmt"
	"testing"

	"github.com/oasdiff/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func getPrefixFile(file string) string {
	return fmt.Sprintf("../data/prefix/%s", file)
}

func TestPrefix_NoArgs(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getPrefixFile("simple1.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getPrefixFile("simple2.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(&diff.Config{}, s1, s2)
	require.NoError(t, err)
	require.NotEmpty(t, dd)
}

func TestPrefix_BasePrefix(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getPrefixFile("simple1.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getPrefixFile("simple2.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(&diff.Config{
		PathPrefixBase: "/tenant",
	}, s1, s2)
	require.NoError(t, err)
	require.Empty(t, dd)
}

func TestPrefix_RevisionPrefix(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getPrefixFile("simple2.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getPrefixFile("simple1.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(&diff.Config{
		PathPrefixRevision: "/tenant",
	}, s1, s2)
	require.NoError(t, err)
	require.Empty(t, dd)
}

func TestPrefix_BasePrefixModified(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getPrefixFile("simple1.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getPrefixFile("simple3.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(&diff.Config{
		PathPrefixBase: "/other",
	}, s1, s2)
	require.NoError(t, err)
	require.NotEmpty(t, dd)
}

func TestPrefix_RevisionStrip(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getPrefixFile("simple1.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getPrefixFile("simple2.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(&diff.Config{
		PathStripPrefixRevision: "/tenant",
	}, s1, s2)
	require.NoError(t, err)
	require.Empty(t, dd)
}

func TestPrefix_BaseStrip(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getPrefixFile("simple2.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getPrefixFile("simple1.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(&diff.Config{
		PathStripPrefixBase: "/tenant",
	}, s1, s2)
	require.NoError(t, err)
	require.Empty(t, dd)
}

func TestPrefix_StripAndPrefix(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile(getPrefixFile("simple2.yaml"))
	require.NoError(t, err)

	s2, err := loader.LoadFromFile(getPrefixFile("simple4.yaml"))
	require.NoError(t, err)

	dd, err := diff.Get(&diff.Config{
		PathStripPrefixBase: "/tenant",
		PathPrefixBase:      "/tenant1",
	}, s1, s2)
	require.NoError(t, err)
	require.Empty(t, dd)
}
