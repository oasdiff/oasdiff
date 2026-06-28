package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// Adding a response header is non-breaking and reported at INFO (the mirror of
// the response-header-removed checks; see #1033).
func TestResponseHeaderAdded(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	// base lacks the header, revision has it -> added
	delete(s1.Spec.Paths.Value(installCommandPath).Get.Responses.Value("default").Value.Headers, "X-RateLimit-Limit")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	verifyNonBreakingChangeIsChangelogEntry(t, d, osm, checker.ResponseHeaderAddedId)
}
