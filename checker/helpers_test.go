package checker_test

import (
	"fmt"
	"testing"

	"github.com/oasdiff/kin-openapi/openapi3"
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

// getDeprecationFile returns the path to a file under data/deprecation/.
func getDeprecationFile(file string) string {
	return fmt.Sprintf("../data/deprecation/%s", file)
}

// getReqPropFile returns the path to a file under data/required-properties/.
func getReqPropFile(file string) string {
	return fmt.Sprintf("../data/required-properties/%s", file)
}

// getParameterDeprecationFile returns the path to a file under data/param-deprecation/.
func getParameterDeprecationFile(file string) string {
	return fmt.Sprintf("../data/param-deprecation/%s", file)
}

func singleCheckConfig(c checker.BackwardCompatibilityCheck) *checker.Config {
	return checker.NewConfig(checker.BackwardCompatibilityChecks{c}).WithSingleCheck(c)
}

func allChecksConfig() *checker.Config {
	return checker.NewConfig(checker.GetAllChecks())
}
