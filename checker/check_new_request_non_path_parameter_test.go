package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// new header, query and cookie request params
func TestNewRequestNonPathParameter_DetectsNewPathsAndNewOperations(t *testing.T) {
	s1, err := open("../data/request_params/base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/request_params/optional-request-params.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(&diff.Config{}, s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.NewRequestNonPathParameterCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 3)

	require.IsType(t, checker.ApiChange{}, errs[0])
	e0 := errs[0].(checker.ApiChange)
	require.Equal(t, checker.NewOptionalRequestParameterId, e0.Id)
	require.Equal(t, "GET", e0.Operation)
	require.Equal(t, "/api/test1", e0.Path)
	require.Equal(t, checker.INFO, e0.Level)
	require.Contains(t, e0.SingleLineError(checker.NewDefaultLocalizer(), checker.ColorNever), "X-NewRequestHeaderParam")

	require.IsType(t, checker.ApiChange{}, errs[1])
	e1 := errs[1].(checker.ApiChange)
	require.Equal(t, checker.NewOptionalRequestParameterId, e1.Id)
	require.Equal(t, "GET", e1.Operation)
	require.Equal(t, "/api/test2", e1.Path)
	require.Equal(t, checker.INFO, e1.Level)
	require.Contains(t, e1.SingleLineError(checker.NewDefaultLocalizer(), checker.ColorNever), "newQueryParam")

	require.IsType(t, checker.ApiChange{}, errs[2])
	e2 := errs[2].(checker.ApiChange)
	require.Equal(t, checker.NewOptionalRequestParameterId, e2.Id)
	require.Equal(t, "GET", e2.Operation)
	require.Equal(t, "/api/test3", e2.Path)
	require.Equal(t, checker.INFO, e2.Level)
	require.Contains(t, e2.SingleLineError(checker.NewDefaultLocalizer(), checker.ColorNever), "csrf-token")
}
