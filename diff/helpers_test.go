package diff_test

import (
	"fmt"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func loadSpec(t *testing.T, path string) *openapi3.T {
	t.Helper()
	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromFile(path)
	require.NoError(t, err)
	return spec
}

// getPrefixFile returns the path to a file under data/prefix/.
func getPrefixFile(file string) string {
	return fmt.Sprintf("../data/prefix/%s", file)
}

// getXOfFile returns the path to a file under data/x-of/.
func getXOfFile(file string) string {
	return fmt.Sprintf("../data/x-of/%s", file)
}

// getXOfTitlesFile returns the path to a file under data/x-of-titles/.
func getXOfTitlesFile(file string) string {
	return fmt.Sprintf("../data/x-of-titles/%s", file)
}

// getSecurityFile returns the path to an api_security_anonymous_ fixture under data/checker/.
func getSecurityFile(name string) string {
	return fmt.Sprintf("../data/checker/api_security_anonymous_%s.yaml", name)
}
