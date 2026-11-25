package checker

import "github.com/oasdiff/oasdiff/diff"

const (
	RequestBodyNotAddedId       = "request-body-not-added"
	RequestBodyNotRemovedId     = "request-body-not-removed"
	RequestPropertyNotAddedId   = "request-property-not-added"
	RequestPropertyNotRemovedId = "request-property-not-removed"
)

func RequestPropertyNotUpdatedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		result = append(result, singleSchemaKeywordChanges(info, DirectionRequest, notKeyword)...)
	})

	return result
}
