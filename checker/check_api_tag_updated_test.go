package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: adding a new tag
func TestTagAdded(t *testing.T) {
	s1, err := open("../data/checker/tag_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/tag_added_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Tags = []string{"newTag"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APITagUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.APITagAddedId,
		Args:        []any{"newTag"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/tag_added_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
	require.Equal(t, "api tag `newTag` added", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// CL: removing an existing tag
func TestTagRemoved(t *testing.T) {
	s1, err := open("../data/checker/tag_removed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/tag_removed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Tags = []string{}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APITagUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.APITagRemovedId,
		Args:        []any{"Test"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/tag_removed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
	require.Equal(t, "api tag `Test` removed", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))

}

// CL: updating an existing tag
func TestTagUpdated(t *testing.T) {
	s1, err := open("../data/checker/tag_removed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/tag_removed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Tags = []string{"newTag"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APITagUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 2)
	for cl := range errs {
		require.Equal(t, checker.INFO, errs[cl].GetLevel())
		if errs[cl].GetId() == checker.APITagRemovedId {
			requireApiChange(t, checker.ApiChange{
				Id:          checker.APITagRemovedId,
				Args:        []any{"Test"},
				Level:       checker.INFO,
				Operation:   "POST",
				Path:        "/api/v1.0/groups",
				Source:      load.NewSource("../data/checker/tag_removed_base.yaml"),
				OperationId: "createOneGroup",
			}, errs[cl])
		}

		if errs[cl].GetId() == checker.APITagAddedId {
			requireApiChange(t, checker.ApiChange{
				Id:          checker.APITagAddedId,
				Args:        []any{"newTag"},
				Level:       checker.INFO,
				Operation:   "POST",
				Path:        "/api/v1.0/groups",
				Source:      load.NewSource("../data/checker/tag_removed_base.yaml"),
				OperationId: "createOneGroup",
			}, errs[cl])
		}
	}
}
