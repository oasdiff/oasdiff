package checker

import "github.com/oasdiff/oasdiff/diff"

const (
	ResponseBodyNotAddedId       = "response-body-not-added"
	ResponseBodyNotRemovedId     = "response-body-not-removed"
	ResponsePropertyNotAddedId   = "response-property-not-added"
	ResponsePropertyNotRemovedId = "response-property-not-removed"
)

func ResponsePropertyNotUpdatedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		result = append(result, singleSchemaKeywordChanges(info, DirectionResponse, notKeyword)...)
	})

	return result
}
