package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// checkHeaderCase runs every check on a pair of specs whose header names
// differ only in case, after modify changes the revision.
func checkHeaderCase(t *testing.T, modify func(revision *openapi3.T)) checker.Changes {
	t.Helper()
	s1, err := open("../data/header-case/base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/header-case/revision.yaml")
	require.NoError(t, err)

	if modify != nil {
		modify(s2.Spec)
	}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	return checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)
}

// HTTP header names are case-insensitive: renaming a header parameter or a
// response header only in case is not a change.
func TestHeaderCase_RenameIsNotAChange(t *testing.T) {
	require.Empty(t, checkHeaderCase(t, nil))
}

// the diff reports the documents as written
func TestHeaderCase_DiffReportsRename(t *testing.T) {
	s1, err := open("../data/header-case/base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/header-case/revision.yaml")
	require.NoError(t, err)

	d, err := diff.Get(diff.NewConfig(), s1.Spec, s2.Spec)
	require.NoError(t, err)

	pathDiff := d.PathsDiff.Modified["/test"]
	require.Equal(t, diff.ParamNamesByLocation{"header": {"X-Case"}}, pathDiff.ParametersDiff.Deleted)
	require.Equal(t, diff.ParamNamesByLocation{"header": {"x-case"}}, pathDiff.ParametersDiff.Added)

	responseDiff := pathDiff.OperationsDiff.Modified["GET"].ResponsesDiff.Modified["200"]
	require.Equal(t, []string{"X-RateLimit-Limit"}, responseDiff.HeadersDiff.Deleted)
	require.Equal(t, []string{"x-ratelimit-limit"}, responseDiff.HeadersDiff.Added)
	require.Nil(t, responseDiff.HeadersDiffIgnoringCase)
}

// a header parameter renamed in case and changed is reported as one changed
// parameter, under its name in the base
func TestHeaderCase_RenamedParameterChanged(t *testing.T) {
	errs := checkHeaderCase(t, func(revision *openapi3.T) {
		revision.Paths.Value("/test").Get.Parameters[0].Value.Required = true
	})
	requireSingleApiChange(t, checker.ApiChange{
		Id:        checker.RequestParameterBecomeRequiredId,
		Args:      []any{"header", "X-Case"},
		Operation: "GET",
		Path:      "/test",
	}, errs)
}

// a response header renamed in case and changed is reported as one changed
// header, under its name in the base
func TestHeaderCase_RenamedResponseHeaderChanged(t *testing.T) {
	errs := checkHeaderCase(t, func(revision *openapi3.T) {
		revision.Paths.Value("/test").Get.Responses.Value("200").Value.Headers["x-ratelimit-limit"].Value.Schema.Value.Type = &openapi3.Types{"string"}
	})
	requireSingleApiChange(t, checker.ApiChange{
		Id:        checker.ResponseHeaderTypeChangedId,
		Args:      []any{"X-RateLimit-Limit", "type", "integer", "string", "200"},
		Operation: "GET",
		Path:      "/test",
	}, errs)
}

// the renamed header is found on both sides
func TestHeaderCase_RenamedResponseHeaderSources(t *testing.T) {
	s1, err := open("../data/header-case/base.yaml", newLoaderWithOriginTracking())
	require.NoError(t, err)
	s2, err := open("../data/header-case/revision.yaml", newLoaderWithOriginTracking())
	require.NoError(t, err)
	s2.Spec.Paths.Value("/test").Get.Responses.Value("200").Value.Headers["x-ratelimit-limit"].Value.Schema.Value.Type = &openapi3.Types{"string"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseHeaderTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, 23, errs[0].GetBaseSource().Line)
	require.Equal(t, 23, errs[0].GetRevisionSource().Line)
}

// a spec with lowercased header names gives the same result
func TestHeaderCase_LowercasedSpecs(t *testing.T) {
	s1, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../data/header-case/base.yaml"), load.WithLowercaseHeaders())
	require.NoError(t, err)
	s2, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../data/header-case/revision.yaml"), load.WithLowercaseHeaders())
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.Empty(t, checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO))
}
