package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	WebhookAddedId    = "webhook-added"
	WebhookRemovedId  = "webhook-removed"
	ComponentWebhooks = "webhooks"
)

func WebhookUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	if diffReport.WebhooksDiff == nil {
		return result
	}

	for _, addedWebhook := range diffReport.WebhooksDiff.Added {
		result = append(result, ComponentChange{
			Id:        WebhookAddedId,
			Level:     config.getLogLevel(WebhookAddedId),
			Args:      []any{addedWebhook},
			Component: ComponentWebhooks,
		})
	}

	for _, deletedWebhook := range diffReport.WebhooksDiff.Deleted {
		result = append(result, ComponentChange{
			Id:        WebhookRemovedId,
			Level:     config.getLogLevel(WebhookRemovedId),
			Args:      []any{deletedWebhook},
			Component: ComponentWebhooks,
		})
	}

	return result
}
