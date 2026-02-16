package diff_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func loadSpec(t *testing.T, path string) *openapi3.T {
	t.Helper()
	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromFile(path)
	require.NoError(t, err)
	return spec
}

func TestWebhooks_Empty(t *testing.T) {
	require.True(t, (*diff.WebhooksDiff)(nil).Empty())
	require.True(t, (&diff.WebhooksDiff{}).Empty())
}

func TestWebhooks_AddedDeleted(t *testing.T) {
	s1 := loadSpec(t, "../data/openapi31/base.yaml")
	s2 := loadSpec(t, "../data/openapi31/revision.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d)
	require.NotNil(t, d.WebhooksDiff)

	// orderCreated was deleted
	require.Contains(t, d.WebhooksDiff.Deleted, "orderCreated")

	// paymentReceived was added
	require.Contains(t, d.WebhooksDiff.Added, "paymentReceived")

	// newUser was modified
	require.Contains(t, d.WebhooksDiff.Modified, "newUser")
}

func TestWebhooks_Modified(t *testing.T) {
	s1 := loadSpec(t, "../data/openapi31/base.yaml")
	s2 := loadSpec(t, "../data/openapi31/revision.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d.WebhooksDiff)
	require.NotNil(t, d.WebhooksDiff.Modified["newUser"])

	// Check that operations diff was detected
	pathDiff := d.WebhooksDiff.Modified["newUser"]
	require.NotNil(t, pathDiff.OperationsDiff)
	require.NotNil(t, pathDiff.OperationsDiff.Modified["POST"])
}

func TestWebhooks_Summary(t *testing.T) {
	s1 := loadSpec(t, "../data/openapi31/base.yaml")
	s2 := loadSpec(t, "../data/openapi31/revision.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	summary := d.GetSummary()
	webhooksSummary := summary.GetSummaryDetails(diff.WebhooksDetail)
	require.Equal(t, 1, webhooksSummary.Added)
	require.Equal(t, 1, webhooksSummary.Deleted)
	require.Equal(t, 1, webhooksSummary.Modified)
}

func TestWebhooks_NoDiff(t *testing.T) {
	s1 := loadSpec(t, "../data/openapi31/base.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s1)
	require.NoError(t, err)
	require.Nil(t, d)
}

func TestWebhooks_BothEmpty(t *testing.T) {
	// Load specs without webhooks
	s1 := loadSpec(t, "../data/simple1.yaml")
	s2 := loadSpec(t, "../data/simple2.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d)
	// WebhooksDiff should be nil when both specs have no webhooks
	require.Nil(t, d.WebhooksDiff)
}

func TestWebhooks_EmptyWithData(t *testing.T) {
	// Test Empty() with populated but empty diff
	webhooksDiff := &diff.WebhooksDiff{
		Added:    []string{},
		Deleted:  []string{},
		Modified: diff.ModifiedWebhooks{},
	}
	require.True(t, webhooksDiff.Empty())

	// Test Empty() with non-empty Added
	webhooksDiff.Added = []string{"webhook1"}
	require.False(t, webhooksDiff.Empty())

	// Reset and test with non-empty Deleted
	webhooksDiff.Added = []string{}
	webhooksDiff.Deleted = []string{"webhook2"}
	require.False(t, webhooksDiff.Empty())

	// Reset and test with non-empty Modified
	webhooksDiff.Deleted = []string{}
	webhooksDiff.Modified = diff.ModifiedWebhooks{"webhook3": nil}
	require.False(t, webhooksDiff.Empty())
}

func TestWebhooks_OnlyInBase(t *testing.T) {
	// Test case where webhooks exist in base but not in revision
	s1 := loadSpec(t, "../data/openapi31/base.yaml")
	s2 := loadSpec(t, "../data/simple1.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d)
	require.NotNil(t, d.WebhooksDiff)
	require.Len(t, d.WebhooksDiff.Deleted, 2)
	require.Empty(t, d.WebhooksDiff.Added)
}

func TestWebhooks_OnlyInRevision(t *testing.T) {
	// Test case where webhooks exist in revision but not in base
	s1 := loadSpec(t, "../data/simple1.yaml")
	s2 := loadSpec(t, "../data/openapi31/base.yaml")

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d)
	require.NotNil(t, d.WebhooksDiff)
	require.Len(t, d.WebhooksDiff.Added, 2)
	require.Empty(t, d.WebhooksDiff.Deleted)
}
