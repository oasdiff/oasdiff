package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

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

func RequestPropertyIfUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}

	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}

		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.RequestBodyDiff == nil ||
				operationItem.RequestBodyDiff.ContentDiff == nil ||
				operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified == nil {
				continue
			}

			modifiedMediaTypes := operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified
			for mediaType, mediaTypeDiff := range modifiedMediaTypes {
				mediaTypeDetails := formatMediaTypeDetails(mediaType, len(modifiedMediaTypes))
				if mediaTypeDiff.SchemaDiff == nil {
					continue
				}

				appendResultItem := func(messageId string, a ...any) {
					result = append(result, NewApiChange(
						messageId,
						config,
						a,
						"",
						operationsSources,
						operationItem.Revision,
						operation,
						path,
					).WithDetails(mediaTypeDetails))
				}

				checkConditionalSchemaDiff(mediaTypeDiff.SchemaDiff.IfDiff, RequestBodyIfAddedId, RequestBodyIfRemovedId, appendResultItem)
				checkConditionalSchemaDiff(mediaTypeDiff.SchemaDiff.ThenDiff, RequestBodyThenAddedId, RequestBodyThenRemovedId, appendResultItem)
				checkConditionalSchemaDiff(mediaTypeDiff.SchemaDiff.ElseDiff, RequestBodyElseAddedId, RequestBodyElseRemovedId, appendResultItem)

				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						propName := propertyFullName(propertyPath, propertyName)

						appendPropResultItem := func(messageId string, a ...any) {
							result = append(result, NewApiChange(
								messageId,
								config,
								a,
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithDetails(mediaTypeDetails))
						}

						checkConditionalSchemaDiffProperty(propertyDiff.IfDiff, RequestPropertyIfAddedId, RequestPropertyIfRemovedId, propName, appendPropResultItem)
						checkConditionalSchemaDiffProperty(propertyDiff.ThenDiff, RequestPropertyThenAddedId, RequestPropertyThenRemovedId, propName, appendPropResultItem)
						checkConditionalSchemaDiffProperty(propertyDiff.ElseDiff, RequestPropertyElseAddedId, RequestPropertyElseRemovedId, propName, appendPropResultItem)
					})
			}
		}
	}
	return result
}

func checkConditionalSchemaDiff(schemaDiff *diff.SchemaDiff, addedId, removedId string, appendResultItem func(string, ...any)) {
	if schemaDiff == nil {
		return
	}
	if schemaDiff.SchemaAdded {
		appendResultItem(addedId)
	}
	if schemaDiff.SchemaDeleted {
		appendResultItem(removedId)
	}
}

func checkConditionalSchemaDiffProperty(schemaDiff *diff.SchemaDiff, addedId, removedId, propName string, appendResultItem func(string, ...any)) {
	if schemaDiff == nil {
		return
	}
	if schemaDiff.SchemaAdded {
		appendResultItem(addedId, propName)
	}
	if schemaDiff.SchemaDeleted {
		appendResultItem(removedId, propName)
	}
}
