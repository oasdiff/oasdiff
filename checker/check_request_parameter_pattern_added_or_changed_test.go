package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: changing pattern of request parameters
func TestRequestParameterPatternChanged(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_pattern_added_or_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_pattern_added_or_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/test").Post.Parameters[0].Value.Schema.Value.Pattern = "^[\\w\\s]+$"
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterPatternAddedOrChangedCheck), d, osm, checker.WARN)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:        checker.RequestParameterPatternChangedId,
		Args:      []any{"query", "category", "^\\w+$", "^[\\w\\s]+$"},
		Comment:   checker.PatternChangedCommentId,
		Level:     checker.WARN,
		Operation: "POST",
		Path:      "/test",
		Source:    load.NewSource("../data/checker/request_parameter_pattern_added_or_changed_base.yaml"),
	}, errs[0])
	require.Equal(t, "changed the pattern of the 'query' request parameter 'category' from '^\\w+$' to '^[\\w\\s]+$'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "This is a warning because adding or changing a pattern may restrict the accepted values and break existing clients. For pattern changes, it is difficult to automatically analyze if the new pattern is a superset of the previous pattern (e.g. changed from '[0-9]+' to '[0-9]*')", errs[0].GetComment(checker.NewDefaultLocalizer()))
}

// CL: generalizing pattern of request parameters
func TestRequestParameterPatternGeneralized(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_pattern_added_or_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_pattern_added_or_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/test").Post.Parameters[0].Value.Schema.Value.Pattern = ".*"
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterPatternAddedOrChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:        checker.RequestParameterPatternGeneralizedId,
		Args:      []any{"query", "category", "^\\w+$", ".*"},
		Level:     checker.INFO,
		Operation: "POST",
		Path:      "/test",
		Source:    load.NewSource("../data/checker/request_parameter_pattern_added_or_changed_base.yaml"),
	}, errs[0])
	require.Equal(t, "changed the pattern of the 'query' request parameter 'category' from '^\\w+$' to a more general pattern '.*'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// CL: adding pattern to request parameters
func TestRequestParameterPatternAdded(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_pattern_added_or_changed_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_pattern_added_or_changed_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterPatternAddedOrChangedCheck), d, osm, checker.ERR)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:        checker.RequestParameterPatternAddedId,
		Args:      []any{"^\\w+$", "query", "category"},
		Comment:   checker.PatternAddedCommentId,
		Level:     checker.ERR,
		Operation: "POST",
		Path:      "/test",
		Source:    load.NewSource("../data/checker/request_parameter_pattern_added_or_changed_base.yaml"),
	}, errs[0])
	require.Equal(t, "This is a breaking change because adding a pattern restriction to a previously unrestricted parameter will reject values that were previously accepted, breaking existing clients", errs[0].GetComment(checker.NewDefaultLocalizer()))
}

// CL: removing pattern from request parameters
func TestRequestParameterPatternRemoved(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_pattern_added_or_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_pattern_added_or_changed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterPatternAddedOrChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:        checker.RequestParameterPatternRemovedId,
		Args:      []any{"^\\w+$", "query", "category"},
		Level:     checker.INFO,
		Operation: "POST",
		Path:      "/test",
		Source:    load.NewSource("../data/checker/request_parameter_pattern_added_or_changed_revision.yaml"),
	}, errs[0])
}
