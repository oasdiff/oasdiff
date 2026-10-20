package diff_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/stretchr/testify/require"
)

// specPairs returns base and revision fixtures to compare: every ordered pair
// of the openapi-test specs, and each base/revision pair under data.
func specPairs(t *testing.T) [][2]string {
	t.Helper()

	var pairs [][2]string

	tests, err := filepath.Glob("../data/openapi-test*.yaml")
	require.NoError(t, err)
	for _, base := range tests {
		for _, revision := range tests {
			if base != revision {
				pairs = append(pairs, [2]string{base, revision})
			}
		}
	}

	bases, err := filepath.Glob("../data/*/base.yaml")
	require.NoError(t, err)
	checkerBases, err := filepath.Glob("../data/checker/*base*.yaml")
	require.NoError(t, err)
	for _, base := range append(bases, checkerBases...) {
		revision := filepath.Join(filepath.Dir(base), strings.Replace(filepath.Base(base), "base", "revision", 1))
		if _, err := os.Stat(revision); err == nil {
			pairs = append(pairs, [2]string{base, revision})
		}
	}

	return pairs
}

// diffBytes returns the diff of base and revision rendered as YAML and JSON,
// or nil when the pair can't be compared.
func diffBytes(t *testing.T, base, revision string) []byte {
	t.Helper()

	load := func(path string) *openapi3.T {
		loader := openapi3.NewLoader()
		loader.IsExternalRefsAllowed = true
		spec, err := loader.LoadFromFile(path)
		if err != nil {
			return nil
		}
		return spec
	}
	s1, s2 := load(base), load(revision)
	if s1 == nil || s2 == nil {
		return nil
	}

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	if err != nil {
		return nil
	}

	yamlBytes, err := formatters.YAMLFormatter{}.RenderDiff(d, formatters.RenderOpts{})
	require.NoError(t, err)
	// JSON can't hold the endpoints section, whose keys are structs
	d.EndpointsDiff = nil
	jsonBytes, err := formatters.JSONFormatter{}.RenderDiff(d, formatters.RenderOpts{})
	require.NoError(t, err)

	return append(yamlBytes, jsonBytes...)
}

// Go randomizes map iteration, so output built by ranging over a map differs
// between runs unless it is sorted.
func TestDiff_Deterministic(t *testing.T) {
	const runs = 10

	for _, pair := range specPairs(t) {
		want := diffBytes(t, pair[0], pair[1])
		if want == nil {
			continue
		}
		for range runs - 1 {
			require.Equal(t, string(want), string(diffBytes(t, pair[0], pair[1])), "%s -> %s", pair[0], pair[1])
		}
	}
}
