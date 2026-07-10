package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// Deleting a value from an x-extensible-enum property is breaking
func TestRequestPropertyXExtensibleEnumValueRemovedCheck(t *testing.T) {
	s1, err := open("../data/checker/request_property_extensible_enum_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/request_property_extensible_enum_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestPropertyXExtensibleEnumValueRemovedCheck), d, osm)
	requireSingleChange(t, errs, checker.RequestPropertyXExtensibleEnumValueRemovedId)
}

// the change points at the exact removed x-extensible-enum value in the base spec
func TestRequestPropertyXExtensibleEnumValueRemoved_WithSources(t *testing.T) {
	s1, err := open("../data/checker/request_property_extensible_enum_base.yaml", newLoaderWithOriginTracking())
	require.NoError(t, err)

	s2, err := open("../data/checker/request_property_extensible_enum_revision.yaml", newLoaderWithOriginTracking())
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestPropertyXExtensibleEnumValueRemovedCheck), d, osm)
	requireSingleChange(t, errs, checker.RequestPropertyXExtensibleEnumValueRemovedId)

	require.NotEmpty(t, errs[0].GetBaseSource())
	require.Equal(t, "../data/checker/request_property_extensible_enum_base.yaml", errs[0].GetBaseSource().File)
	require.Equal(t, 20, errs[0].GetBaseSource().Line) // the removed "sold" value
	require.Empty(t, errs[0].GetRevisionSource())
}
