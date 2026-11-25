package checker

import "github.com/oasdiff/oasdiff/diff"

const (
	RequestBodyPropertyNamesAddedId       = "request-body-property-names-added"
	RequestBodyPropertyNamesRemovedId     = "request-body-property-names-removed"
	RequestPropertyPropertyNamesAddedId   = "request-property-property-names-added"
	RequestPropertyPropertyNamesRemovedId = "request-property-property-names-removed"
)

func RequestPropertyPropertyNamesUpdatedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		result = append(result, singleSchemaKeywordChanges(info, DirectionRequest, propertyNamesKeyword)...)
	})

	return result
}
