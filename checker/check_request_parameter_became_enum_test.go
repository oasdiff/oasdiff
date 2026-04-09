package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: changing request parameter type to enum
func TestRequestParameterBecameEnum(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_became_enum_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_became_enum_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterBecameEnumCheck), d, osm, checker.ERR)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestParameterBecameEnumId,
		Args:        []any{"path", "groupId"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_became_enum_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
	require.Equal(t, "the `path` request parameter `groupId` was restricted to a list of enum values", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}
