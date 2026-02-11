package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: adding if/then/else to request body
func TestRequestBodyIfAdded(t *testing.T) {
	s1, err := open("../data/checker/if_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/if_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyIfUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.RequestBodyIfAddedId], "expected request-body-if-added")
	require.True(t, ids[checker.RequestBodyThenAddedId], "expected request-body-then-added")
	require.True(t, ids[checker.RequestBodyElseAddedId], "expected request-body-else-added")
}

// CL: removing if/then/else from request body
func TestRequestBodyIfRemoved(t *testing.T) {
	s1, err := open("../data/checker/if_added_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/if_added_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyIfUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.RequestBodyIfRemovedId], "expected request-body-if-removed")
	require.True(t, ids[checker.RequestBodyThenRemovedId], "expected request-body-then-removed")
	require.True(t, ids[checker.RequestBodyElseRemovedId], "expected request-body-else-removed")
}
