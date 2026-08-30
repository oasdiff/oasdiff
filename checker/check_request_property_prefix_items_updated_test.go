package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// adding prefixItems to request body
func TestRequestBodyPrefixItemsAdded(t *testing.T) {
	s1, err := open("../data/checker/prefix_items_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/prefix_items_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyPrefixItemsUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.GetId() == checker.RequestBodyPrefixItemsAddedId {
			found = true
			break
		}
	}
	require.True(t, found, "expected request-body-prefix-items-added")
}

// removing prefixItems from request body (reverse)
func TestRequestBodyPrefixItemsRemoved(t *testing.T) {
	s1, err := open("../data/checker/prefix_items_added_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/prefix_items_added_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyPrefixItemsUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.GetId() == checker.RequestBodyPrefixItemsRemovedId {
			found = true
			break
		}
	}
	require.True(t, found, "expected request-body-prefix-items-removed")
}

// adding prefixItems to request property
func TestRequestPropertyPrefixItemsAdded(t *testing.T) {
	s1, err := open("../data/checker/prefix_items_prop_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/prefix_items_prop_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyPrefixItemsUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.RequestPropertyPrefixItemsAddedId), "expected request-property-prefix-items-added")
}

// removing prefixItems from request property
func TestRequestPropertyPrefixItemsRemoved(t *testing.T) {
	s1, err := open("../data/checker/prefix_items_prop_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/prefix_items_prop_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyPrefixItemsUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.RequestPropertyPrefixItemsRemovedId), "expected request-property-prefix-items-removed")
}

// reordering prefixItems entries changes which schema validates each position,
// so both positions are reported as type changes; the previous content-based
// pairing matched each entry with its identical twin and reported nothing
func TestRequestBodyPrefixItemsReordered(t *testing.T) {
	s1, err := open("../data/checker/prefix_items_reordered_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/prefix_items_reordered_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)

	require.Len(t, errs, 2)
	for _, e := range errs {
		require.Equal(t, checker.RequestPropertyTypeChangedId, e.GetId())
		require.Equal(t, checker.ERR, e.GetLevel())
	}
}

// a prefixItems entry that repeats the items schema constrains the position it
// covers exactly as items already did, so the accepted arrays are unchanged and
// there is nothing to report
func TestRequestBodyPrefixItemsSameAsItems(t *testing.T) {
	s1, err := open("../data/checker/prefix_items_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/prefix_items_same_as_items_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyPrefixItemsUpdatedCheck), d, osm, checker.INFO)
	require.Empty(t, errs)
}

// the direction of a prefixItems change depends on the items schema governing
// the position, so the verdict is a warning carrying the reason
func TestRequestBodyPrefixItemsAddedIsWarning(t *testing.T) {
	s1, err := open("../data/checker/prefix_items_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/prefix_items_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyPrefixItemsUpdatedCheck), d, osm, checker.INFO)

	change := requireChange(t, errs, checker.RequestBodyPrefixItemsAddedId)
	require.Equal(t, checker.WARN, change.GetLevel())
	require.NotEmpty(t, change.GetComment(checker.NewDefaultLocalizer()))
}
