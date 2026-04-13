package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: adding if/then/else to response body
func TestResponseBodyIfAdded(t *testing.T) {
	s1, err := open("../data/checker/if_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/if_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyIfUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.ResponseBodyIfAddedId], "expected response-body-if-added")
	require.True(t, ids[checker.ResponseBodyThenAddedId], "expected response-body-then-added")
	require.True(t, ids[checker.ResponseBodyElseAddedId], "expected response-body-else-added")
}

// CL: removing if/then/else from response body
func TestResponseBodyIfRemoved(t *testing.T) {
	s1, err := open("../data/checker/if_added_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/if_added_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyIfUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.ResponseBodyIfRemovedId], "expected response-body-if-removed")
	require.True(t, ids[checker.ResponseBodyThenRemovedId], "expected response-body-then-removed")
	require.True(t, ids[checker.ResponseBodyElseRemovedId], "expected response-body-else-removed")
}

// CL: adding if/then/else to response property
func TestResponsePropertyIfAdded(t *testing.T) {
	s1, err := open("../data/checker/if_added_prop_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/if_added_prop_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyIfUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.ResponsePropertyIfAddedId], "expected response-property-if-added")
	require.True(t, ids[checker.ResponsePropertyThenAddedId], "expected response-property-then-added")
	require.True(t, ids[checker.ResponsePropertyElseAddedId], "expected response-property-else-added")
}

// CL: removing if/then/else from response property
func TestResponsePropertyIfRemoved(t *testing.T) {
	s1, err := open("../data/checker/if_added_prop_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/if_added_prop_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyIfUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.ResponsePropertyIfRemovedId], "expected response-property-if-removed")
	require.True(t, ids[checker.ResponsePropertyThenRemovedId], "expected response-property-then-removed")
	require.True(t, ids[checker.ResponsePropertyElseRemovedId], "expected response-property-else-removed")
}
