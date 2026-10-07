package diff_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// `security: []` overrides the root list although it has no alternatives, so
// declaring or dropping it is a change.
func TestSecurityRequirementsListAddedAndDeleted(t *testing.T) {
	withoutList := loadSecurityFixture(t, "global_api_key")
	withList := loadSecurityFixture(t, "global_api_key_op_anonymous")

	added, err := diff.Get(diff.NewConfig(), withoutList, withList)
	require.NoError(t, err)
	require.Equal(t, &diff.SecurityRequirementsDiff{
		Added:     diff.SecurityAlternatives{},
		Deleted:   diff.SecurityAlternatives{},
		Modified:  diff.ModifiedSecurityRequirements{},
		ListAdded: true,
	}, added.PathsDiff.Modified["/pets"].OperationsDiff.Modified["GET"].SecurityDiff)

	deleted, err := diff.Get(diff.NewConfig(), withList, withoutList)
	require.NoError(t, err)
	require.True(t, deleted.PathsDiff.Modified["/pets"].OperationsDiff.Modified["GET"].SecurityDiff.ListDeleted)
}
