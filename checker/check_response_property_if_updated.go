package checker

import "github.com/oasdiff/oasdiff/diff"

const (
	ResponseBodyIfAddedId       = "response-body-if-added"
	ResponseBodyIfRemovedId     = "response-body-if-removed"
	ResponseBodyThenAddedId     = "response-body-then-added"
	ResponseBodyThenRemovedId   = "response-body-then-removed"
	ResponseBodyElseAddedId     = "response-body-else-added"
	ResponseBodyElseRemovedId   = "response-body-else-removed"
	ResponsePropertyIfAddedId   = "response-property-if-added"
	ResponsePropertyIfRemovedId = "response-property-if-removed"

	ResponsePropertyThenAddedId   = "response-property-then-added"
	ResponsePropertyThenRemovedId = "response-property-then-removed"
	ResponsePropertyElseAddedId   = "response-property-else-added"
	ResponsePropertyElseRemovedId = "response-property-else-removed"
)

func ResponsePropertyIfUpdatedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		result = append(result, singleSchemaKeywordChanges(info, DirectionResponse, ifKeyword, thenKeyword, elseKeyword)...)
	})

	return result
}
