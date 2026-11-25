package checker

import "github.com/oasdiff/oasdiff/diff"

const (
	RequestBodyIfAddedId       = "request-body-if-added"
	RequestBodyIfRemovedId     = "request-body-if-removed"
	RequestBodyThenAddedId     = "request-body-then-added"
	RequestBodyThenRemovedId   = "request-body-then-removed"
	RequestBodyElseAddedId     = "request-body-else-added"
	RequestBodyElseRemovedId   = "request-body-else-removed"
	RequestPropertyIfAddedId   = "request-property-if-added"
	RequestPropertyIfRemovedId = "request-property-if-removed"

	RequestPropertyThenAddedId   = "request-property-then-added"
	RequestPropertyThenRemovedId = "request-property-then-removed"
	RequestPropertyElseAddedId   = "request-property-else-added"
	RequestPropertyElseRemovedId = "request-property-else-removed"
)

func RequestPropertyIfUpdatedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		result = append(result, subschemaKeywordChanges(info, DirectionRequest, ifKeyword, thenKeyword, elseKeyword)...)
	})

	return result
}
