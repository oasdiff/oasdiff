package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: adding and removing webhooks
func TestWebhookAddedAndRemoved(t *testing.T) {
	s1, err := open("../data/checker/webhook_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/webhook_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.WebhookUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 2)
	require.ElementsMatch(t, []checker.ComponentChange{{
		Id:        checker.WebhookRemovedId,
		Args:      []any{"orderDeleted"},
		Level:     checker.ERR,
		Component: checker.ComponentWebhooks,
	}, {
		Id:        checker.WebhookAddedId,
		Args:      []any{"orderUpdated"},
		Level:     checker.INFO,
		Component: checker.ComponentWebhooks,
	}}, errs)
}

// CL: no webhook changes
func TestWebhookNoChanges(t *testing.T) {
	s1, err := open("../data/checker/webhook_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/webhook_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.WebhookUpdatedCheck), d, osm, checker.INFO)
	require.Empty(t, errs)
}
