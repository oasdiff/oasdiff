package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// BC: new header, query and cookie required request default param is breaking
func TestNewRequestNonPathParameter_DetectsNewRequiredPathsAndNewOperations(t *testing.T) {
	s1, err := open("../data/request_params/base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/request_params/required-request-params.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(&diff.Config{}, s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.NewRequestNonPathDefaultParameterCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 7)

	requireApiChanges(t, []checker.ApiChange{
		{
			Id:          checker.NewRequiredRequestDefaultParameterToExistingPathId,
			Args:        []any{"query", "version"},
			Level:       3,
			Operation:   "GET",
			OperationId: "getTest",
			Path:        "/api/test1",
			Source:      load.NewSource("../data/request_params/required-request-params.yaml"),
		},
		{
			Id:        checker.NewRequiredRequestDefaultParameterToExistingPathId,
			Args:      []any{"query", "version"},
			Level:     3,
			Operation: "POST",
			Path:      "/api/test1",
			Source:    load.NewSource("../data/request_params/required-request-params.yaml"),
		},
		{
			Id:          checker.NewRequiredRequestDefaultParameterToExistingPathId,
			Args:        []any{"query", "id"},
			Level:       3,
			Operation:   "GET",
			OperationId: "getTest",
			Path:        "/api/test2",
			Source:      load.NewSource("../data/request_params/required-request-params.yaml"),
		},
		{
			Id:          checker.NewRequiredRequestDefaultParameterToExistingPathId,
			Args:        []any{"header", "If-None-Match"},
			Level:       3,
			Operation:   "GET",
			OperationId: "getTest",
			Path:        "/api/test3",
			Source:      load.NewSource("../data/request_params/required-request-params.yaml"),
		},
		{
			Id:          checker.NewOptionalRequestDefaultParameterToExistingPathId,
			Args:        []any{"query", "optionalQueryParam"},
			Level:       1,
			Operation:   "GET",
			OperationId: "getTest",
			Path:        "/api/test1",
			Source:      load.NewSource("../data/request_params/required-request-params.yaml"),
		},
		{
			Id:        checker.NewOptionalRequestDefaultParameterToExistingPathId,
			Args:      []any{"query", "optionalQueryParam"},
			Level:     1,
			Operation: "POST",
			Path:      "/api/test1",
			Source:    load.NewSource("../data/request_params/required-request-params.yaml"),
		},
		{
			Id:          checker.NewOptionalRequestDefaultParameterToExistingPathId,
			Args:        []any{"header", "optionalHeaderParam"},
			Level:       1,
			Operation:   "GET",
			OperationId: "getTest",
			Path:        "/api/test2",
			Source:      load.NewSource("../data/request_params/required-request-params.yaml"),
		}}, errs)
}
