package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: modified webhook operations are covered by existing operation-level checkers
func TestWebhookOperationChangesDetected(t *testing.T) {
	s1, err := open("../data/checker/webhook_operation_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/webhook_operation_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	// Verify that the changes have webhook: prefix path
	foundWebhookPath := false
	for _, e := range errs {
		if apiChange, ok := e.(checker.ApiChange); ok {
			if apiChange.Path == "webhook:orderCreated" {
				foundWebhookPath = true
				break
			}
		}
	}
	require.True(t, foundWebhookPath, "expected changes with webhook:orderCreated path")

	// Verify specific changes detected (request body became required, new optional property added)
	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.RequestBodyBecameRequiredId] || ids[checker.NewOptionalRequestPropertyId],
		"expected at least one of: request-body-became-required or new-optional-request-property")
}

// CL: no changes in webhooks should produce no webhook-related changes
func TestWebhookOperationNoChanges(t *testing.T) {
	s1, err := open("../data/checker/webhook_operation_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s1)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)

	// No webhook-prefixed changes
	for _, e := range errs {
		if apiChange, ok := e.(checker.ApiChange); ok {
			require.NotContains(t, apiChange.Path, "webhook:", "unexpected webhook change detected")
		}
	}
}
