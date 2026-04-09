package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: changing response property pattern
func TestResponsePropertyPatternChanged(t *testing.T) {
	s1, err := open("../data/checker/response_pattern_added_or_changed_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/response_pattern_added_or_changed_revision.yaml")
	require.NoError(t, err)

	s2.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["name"].Value.WriteOnly = true
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePatternAddedOrChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponsePropertyPatternChangedId,
		Args:        []any{"data/created", "^[a-z]+$", "^(?:([a-z]+-)*([a-z]+)?)$", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_pattern_added_or_changed_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
	require.Equal(t, "the `data/created` response's property pattern was changed from `^[a-z]+$` to `^(?:([a-z]+-)*([a-z]+)?)$` for the status `200`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// CL: adding response property pattern
func TestResponsePropertyPatternAdded(t *testing.T) {
	s1, err := open("../data/checker/response_pattern_added_or_changed_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/response_pattern_added_or_changed_base.yaml")
	require.NoError(t, err)

	s1.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["created"].Value.Pattern = ""
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePatternAddedOrChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponsePropertyPatternAddedId,
		Args:        []any{"data/created", "^[a-z]+$", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_pattern_added_or_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
	require.Equal(t, "the `data/created` response's property pattern `^[a-z]+$` was added for the status `200`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// CL: removing response property pattern
func TestResponsePropertyPatternRemoved(t *testing.T) {
	s1, err := open("../data/checker/response_pattern_added_or_changed_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/response_pattern_added_or_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["created"].Value.Pattern = ""
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePatternAddedOrChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponsePropertyPatternRemovedId,
		Args:        []any{"data/created", "^[a-z]+$", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_pattern_added_or_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
	require.Equal(t, "the `data/created` response's property pattern `^[a-z]+$` was removed for the status `200`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}
