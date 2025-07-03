package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// BC: Deleting a value from an x-extensible-enum parameter is breaking
func TestRequestParameterXExtensibleEnumValueRemoved(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_extensible_enum_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/request_parameter_extensible_enum_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterXExtensibleEnumValueRemovedCheck), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterXExtensibleEnumValueRemovedId, errs[0].GetId())
	require.Equal(t, "removed the x-extensible-enum value '2' from the 'path' request parameter 'groupId'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}
