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
		var revisionSource *Source
		if pathItem := diffReport.WebhooksDiff.Revision[addedWebhook]; pathItem != nil {
			revisionSource = sourceFromOrigin(pathItem.Origin)
		}
		result = append(result, ComponentChange{
			Id:        WebhookAddedId,
			Level:     config.getLogLevel(WebhookAddedId),
			Args:      []any{addedWebhook},
			Component: ComponentWebhooks,
		}.WithSources(nil, revisionSource))
	}

	for _, deletedWebhook := range diffReport.WebhooksDiff.Deleted {
		var baseSource *Source
		if pathItem := diffReport.WebhooksDiff.Base[deletedWebhook]; pathItem != nil {
			baseSource = sourceFromOrigin(pathItem.Origin)
		}
		result = append(result, ComponentChange{
			Id:        WebhookRemovedId,
			Level:     config.getLogLevel(WebhookRemovedId),
			Args:      []any{deletedWebhook},
			Component: ComponentWebhooks,
		}.WithSources(baseSource, nil))
	}

	return result
}
