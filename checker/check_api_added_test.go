package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// new paths or path operations
func TestApiAdded_DetectsNewPathsAndNewOperations(t *testing.T) {
	loader := newLoaderWithOriginTracking()
	s1, err := open("../data/new_endpoints/base.yaml", loader)
	require.NoError(t, err)

	s2, err := open("../data/new_endpoints/revision.yaml", loader)
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIAddedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 2)

	require.IsType(t, checker.ApiChange{}, errs[0])
	e0 := errs[0].(checker.ApiChange)
	require.Equal(t, "endpoint-added", e0.Id)
	require.Equal(t, "POST", e0.Operation)
	require.Equal(t, "/api/test2", e0.Path)
	require.Empty(t, e0.GetBaseSource())
	require.Equal(t, e0.GetRevisionSource().File, "../data/new_endpoints/revision.yaml")
	require.Equal(t, e0.GetRevisionSource().Line, 21)
	require.Equal(t, e0.GetRevisionSource().Column, 5)

	require.IsType(t, checker.ApiChange{}, errs[1])
	e1 := errs[1].(checker.ApiChange)
	require.Equal(t, "endpoint-added", e1.Id)
	require.Equal(t, "GET", e1.Operation)
	require.Equal(t, "/api/test3", e1.Path)
	require.Empty(t, e1.GetBaseSource())
	require.Equal(t, &checker.Source{File: "../data/new_endpoints/revision.yaml", Line: 27, Column: 5, EndLine: 30, EndColumn: 26}, e1.GetRevisionSource())
}

// new paths or path operations
func TestApiAdded_DetectsModifiedPathsWithPathParam(t *testing.T) {
	loader := newLoaderWithOriginTracking()
	s1, err := open("../data/new_endpoints/base_with_path_param.yaml", loader)
	require.NoError(t, err)

	s2, err := open("../data/new_endpoints/revision_with_path_param.yaml", loader)
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIAddedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)

	require.IsType(t, checker.ApiChange{}, errs[0])
	e0 := errs[0].(checker.ApiChange)
	require.Equal(t, "endpoint-added", e0.Id)
	require.Equal(t, "POST", e0.Operation)
	require.Equal(t, "/api/test/{id}", e0.Path)
	require.Empty(t, e0.GetBaseSource())
	require.Equal(t, &checker.Source{File: "../data/new_endpoints/revision_with_path_param.yaml", Line: 15, Column: 5, EndLine: 18, EndColumn: 26}, e0.GetRevisionSource())
}
