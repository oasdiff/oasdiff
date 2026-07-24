package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// A response header becoming nullable means a client that never expected null
// can now receive it: breaking, reported at ERR.
func TestResponseHeaderBecameNullable(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	responseHeaderSchema(s1).Nullable = false
	responseHeaderSchema(s2).Nullable = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponseHeaderBecameNullableId, errs[0].GetId())
	require.Equal(t, checker.ERR, errs[0].GetLevel())
	require.Equal(t,
		"the response header `X-RateLimit-Limit` became nullable for the status `default`",
		errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// A response header becoming not-nullable narrows the server's output: safe,
// reported at INFO.
func TestResponseHeaderBecameNotNullable(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	responseHeaderSchema(s1).Nullable = true
	responseHeaderSchema(s2).Nullable = false

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	verifyNonBreakingChangeIsChangelogEntry(t, d, osm, checker.ResponseHeaderBecameNotNullableId)
}
