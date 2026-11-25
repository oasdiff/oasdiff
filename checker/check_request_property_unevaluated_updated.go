package checker

import "github.com/oasdiff/oasdiff/diff"

const (
	RequestBodyUnevaluatedItemsAddedId        = "request-body-unevaluated-items-added"
	RequestBodyUnevaluatedItemsRemovedId      = "request-body-unevaluated-items-removed"
	RequestBodyUnevaluatedPropertiesAddedId   = "request-body-unevaluated-properties-added"
	RequestBodyUnevaluatedPropertiesRemovedId = "request-body-unevaluated-properties-removed"

	RequestPropertyUnevaluatedItemsAddedId        = "request-property-unevaluated-items-added"
	RequestPropertyUnevaluatedItemsRemovedId      = "request-property-unevaluated-items-removed"
	RequestPropertyUnevaluatedPropertiesAddedId   = "request-property-unevaluated-properties-added"
	RequestPropertyUnevaluatedPropertiesRemovedId = "request-property-unevaluated-properties-removed"
)

func RequestPropertyUnevaluatedUpdatedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		result = append(result, singleSchemaKeywordChanges(info, DirectionRequest, unevaluatedItemsKeyword, unevaluatedPropertiesKeyword)...)
	})

	return result
}
