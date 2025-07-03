package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: changing request property pattern
func TestRequestPropertyPatternChanged(t *testing.T) {
	s1, err := open("../data/checker/request_property_pattern_added_or_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_pattern_added_or_changed_revision.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value.Pattern = "^[\\w\\s]+$"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyPatternUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:        checker.RequestPropertyPatternChangedId,
		Args:      []any{"name", "^\\w+$", "^[\\w\\s]+$"},
		Level:     checker.WARN,
		Operation: "POST",
		Path:      "/test",
		Source:    load.NewSource("../data/checker/request_property_pattern_added_or_changed_revision.yaml"),
		Comment:   checker.PatternChangedCommentId,
	}, errs[0])
	require.Equal(t, "This is a warning because adding or changing a pattern may restrict the accepted values and break existing clients. For pattern changes, it is difficult to automatically analyze if the new pattern is a superset of the previous pattern (e.g. changed from '[0-9]+' to '[0-9]*')", errs[0].GetComment(checker.NewDefaultLocalizer()))
}

// CL: generalizing request property pattern
func TestRequestPropertyPatternGeneralized(t *testing.T) {
	s1, err := open("../data/checker/request_property_pattern_added_or_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_pattern_added_or_changed_revision.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/test").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value.Pattern = ".*"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyPatternUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:        checker.RequestPropertyPatternGeneralizedId,
		Args:      []any{"name", "^\\w+$", ".*"},
		Level:     checker.INFO,
		Operation: "POST",
		Path:      "/test",
		Source:    load.NewSource("../data/checker/request_property_pattern_added_or_changed_revision.yaml"),
	}, errs[0])
}

// CL: adding request property pattern
func TestRequestPropertyPatternAdded(t *testing.T) {
	s1, err := open("../data/checker/request_property_pattern_added_or_changed_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_pattern_added_or_changed_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyPatternUpdatedCheck), d, osm, checker.ERR)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:        checker.RequestPropertyPatternAddedId,
		Args:      []any{"^\\w+$", "name"},
		Level:     checker.ERR,
		Operation: "POST",
		Path:      "/test",
		Source:    load.NewSource("../data/checker/request_property_pattern_added_or_changed_base.yaml"),
		Comment:   checker.PatternAddedCommentId,
	}, errs[0])
	require.Equal(t, "This is a breaking change because adding a pattern restriction to a previously unrestricted parameter will reject values that were previously accepted, breaking existing clients", errs[0].GetComment(checker.NewDefaultLocalizer()))
}

// CL: removing request property pattern
func TestRequestPropertyPatternRemoved(t *testing.T) {
	s1, err := open("../data/checker/request_property_pattern_added_or_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_pattern_added_or_changed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyPatternUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:        checker.RequestPropertyPatternRemovedId,
		Args:      []any{"^\\w+$", "name"},
		Level:     checker.INFO,
		Operation: "POST",
		Path:      "/test",
		Source:    load.NewSource("../data/checker/request_property_pattern_added_or_changed_revision.yaml"),
	}, errs[0])
}
