package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// adding prefixItems to response body
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

// removing prefixItems from response body
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

// adding prefixItems to response property
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

// removing prefixItems from response property
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

// a prefixItems entry that repeats the items schema constrains the position it
// covers exactly as items already did, so the described responses are
// unchanged and there is nothing to report
func TestResponseBodyPrefixItemsSameAsItems(t *testing.T) {
	s1, err := open("../data/checker/prefix_items_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/prefix_items_same_as_items_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyPrefixItemsUpdatedCheck), d, osm, checker.INFO)
	require.Empty(t, errs)
}

// the direction of a prefixItems change depends on the items schema governing
// the position, so the verdict is a warning carrying the reason
func TestResponseBodyPrefixItemsAddedIsWarning(t *testing.T) {
	s1, err := open("../data/checker/prefix_items_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/prefix_items_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyPrefixItemsUpdatedCheck), d, osm, checker.INFO)

	change := requireChange(t, errs, checker.ResponseBodyPrefixItemsAddedId)
	require.Equal(t, checker.WARN, change.GetLevel())
	require.NotEmpty(t, change.GetComment(checker.NewDefaultLocalizer()))
}
