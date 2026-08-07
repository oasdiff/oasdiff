package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// The checkers read operations from the diff, so an operation under the
// OpenAPI 3.2 QUERY field or under a custom method in additionalOperations is
// checked like any other. These pin that end to end: without the methods in
// the diff, a breaking change under them is silently reported as no change.

func oas32Changes(t *testing.T) checker.Changes {
	t.Helper()

	s1, err := open("../data/oas32/base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/oas32/revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	return checker.CheckBackwardCompatibilityUntilLevel(checker.NewConfig(checker.GetAllChecks()), d, osm, checker.ERR)
}

func TestBreakingChange_UnderQueryMethod(t *testing.T) {
	changes := oas32Changes(t)

	require.NotEmpty(t, changes, "a breaking change under QUERY must not be reported as no change")
	require.Contains(t, methodsOf(changes), openapi3.MethodQuery)
}

func TestBreakingChange_UnderCustomMethod(t *testing.T) {
	require.Contains(t, methodsOf(oas32Changes(t)), "PURGE")
}

func methodsOf(changes checker.Changes) []string {
	methods := make([]string, 0, len(changes))
	for _, change := range changes {
		methods = append(methods, change.GetOperation())
	}
	return methods
}
