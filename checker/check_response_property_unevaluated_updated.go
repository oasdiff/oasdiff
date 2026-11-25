package checker

import "github.com/oasdiff/oasdiff/diff"

const (
	ResponseBodyUnevaluatedItemsAddedId        = "response-body-unevaluated-items-added"
	ResponseBodyUnevaluatedItemsRemovedId      = "response-body-unevaluated-items-removed"
	ResponseBodyUnevaluatedPropertiesAddedId   = "response-body-unevaluated-properties-added"
	ResponseBodyUnevaluatedPropertiesRemovedId = "response-body-unevaluated-properties-removed"

	ResponsePropertyUnevaluatedItemsAddedId        = "response-property-unevaluated-items-added"
	ResponsePropertyUnevaluatedItemsRemovedId      = "response-property-unevaluated-items-removed"
	ResponsePropertyUnevaluatedPropertiesAddedId   = "response-property-unevaluated-properties-added"
	ResponsePropertyUnevaluatedPropertiesRemovedId = "response-property-unevaluated-properties-removed"
)

func ResponsePropertyUnevaluatedUpdatedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		result = append(result, singleSchemaKeywordChanges(info, DirectionResponse, unevaluatedItemsKeyword, unevaluatedPropertiesKeyword)...)
	})

	return result
}
