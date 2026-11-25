package checker

import "github.com/oasdiff/oasdiff/diff"

const (
	ResponseBodyPropertyNamesAddedId       = "response-body-property-names-added"
	ResponseBodyPropertyNamesRemovedId     = "response-body-property-names-removed"
	ResponsePropertyPropertyNamesAddedId   = "response-property-property-names-added"
	ResponsePropertyPropertyNamesRemovedId = "response-property-property-names-removed"
)

func ResponsePropertyPropertyNamesUpdatedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		result = append(result, subschemaKeywordChanges(info, DirectionResponse, propertyNamesKeyword)...)
	})

	return result
}
