package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: adding prefixItems to response body
func TestResponseBodyPrefixItemsAdded(t *testing.T) {
	s1, err := open("../data/checker/prefix_items_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/prefix_items_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyPrefixItemsUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.GetId() == checker.ResponseBodyPrefixItemsAddedId {
			found = true
			break
		}
	}
	require.True(t, found, "expected response-body-prefix-items-added")
}

// CL: removing prefixItems from response body
func TestResponseBodyPrefixItemsRemoved(t *testing.T) {
	s1, err := open("../data/checker/prefix_items_added_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/prefix_items_added_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyPrefixItemsUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponseBodyPrefixItemsRemovedId), "expected response-body-prefix-items-removed")
}

// CL: adding prefixItems to response property
func TestResponsePropertyPrefixItemsAdded(t *testing.T) {
	s1, err := open("../data/checker/prefix_items_prop_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/prefix_items_prop_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyPrefixItemsUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponsePropertyPrefixItemsAddedId), "expected response-property-prefix-items-added")
}

// CL: removing prefixItems from response property
func TestResponsePropertyPrefixItemsRemoved(t *testing.T) {
	s1, err := open("../data/checker/prefix_items_prop_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/prefix_items_prop_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyPrefixItemsUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponsePropertyPrefixItemsRemovedId), "expected response-property-prefix-items-removed")
}
