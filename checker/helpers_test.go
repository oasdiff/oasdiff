package checker_test

import (
	"fmt"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

const (
	securityScorePath      = "/api/{domain}/{project}/badges/security-score"
	securityScorePathSlash = securityScorePath + "/"
	installCommandPath     = "/api/{domain}/{project}/install-command"
)

// newLoaderWithOriginTracking returns a loader with origin tracking enabled.
// Use this instead of the deprecated openapi3.IncludeOrigin global.
func newLoaderWithOriginTracking() *openapi3.Loader {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	return loader
}

// open loads a spec from file, optionally using a pre-configured loader.
func open(file string, loaders ...*openapi3.Loader) (*load.SpecInfo, error) {
	var loader *openapi3.Loader
	if len(loaders) > 0 {
		loader = loaders[0]
	} else {
		loader = openapi3.NewLoader()
	}
	return load.NewSpecInfo(loader, load.NewSource(file))
}

// l loads a numbered openapi-test spec, optionally using a pre-configured loader.
func l(t *testing.T, v int, loaders ...*openapi3.Loader) *load.SpecInfo {
	t.Helper()
	var loader *openapi3.Loader
	if len(loaders) > 0 {
		loader = loaders[0]
	} else {
		loader = openapi3.NewLoader()
	}
	specInfo, err := load.NewSpecInfo(loader, load.NewSource(fmt.Sprintf("../data/openapi-test%d.yaml", v)))
	require.NoError(t, err)
	return specInfo
}

// d diffs two numbered openapi-test specs and runs all backward-compatibility checks.
func d(t *testing.T, config *diff.Config, v1, v2 int, loaders ...*openapi3.Loader) checker.Changes {
	t.Helper()
	l1 := l(t, v1, loaders...)
	l2 := l(t, v2, loaders...)
	d, osm, err := diff.GetWithOperationsSourcesMap(config, l1, l2)
	require.NoError(t, err)
	return checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
}

// dataFileFn returns a closure that resolves files under data/<subdir>/.
// Use it to declare a named handle per data subdirectory; call sites then
// pass only the file name.
func dataFileFn(subdir string) func(string) string {
	return func(file string) string {
		return fmt.Sprintf("../data/%s/%s", subdir, file)
	}
}

var (
	deprecationFile      = dataFileFn("deprecation")
	paramDeprecationFile = dataFileFn("param-deprecation")
	requiredPropertyFile = dataFileFn("required-properties")
	stabilityFile        = dataFileFn("stability")
)

// stabilityChanges diffs two stability fixtures and runs all backward-compatibility
// checks at the given threshold.
func stabilityChanges(t *testing.T, baseFile, revisionFile string, sl checker.StabilityLevel) checker.Changes {
	t.Helper()
	s1, err := open(stabilityFile(baseFile))
	require.NoError(t, err)
	s2, err := open(stabilityFile(revisionFile))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := allChecksConfig()
	config.StabilityLevel = sl
	return checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
}

func singleCheckConfig(c checker.BackwardCompatibilityCheck, opts ...checker.Option) *checker.Config {
	return checker.NewConfig(checker.BackwardCompatibilityChecks{c}, append([]checker.Option{checker.WithSingleCheck(c)}, opts...)...)
}

func allChecksConfig(opts ...checker.Option) *checker.Config {
	return checker.NewConfig(checker.GetAllChecks(), opts...)
}

// findChange returns the first Change with the given id, or nil if none match.
func findChange(changes checker.Changes, id string) checker.Change {
	for _, c := range changes {
		if c.GetId() == id {
			return c
		}
	}
	return nil
}

func containsId(changes checker.Changes, id string) bool {
	return findChange(changes, id) != nil
}
