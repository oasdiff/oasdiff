package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyContainsAddedId            = "request-body-contains-added"
	RequestBodyContainsRemovedId          = "request-body-contains-removed"
	RequestBodyMinContainsIncreasedId     = "request-body-min-contains-increased"
	RequestBodyMinContainsDecreasedId     = "request-body-min-contains-decreased"
	RequestBodyMaxContainsIncreasedId     = "request-body-max-contains-increased"
	RequestBodyMaxContainsDecreasedId     = "request-body-max-contains-decreased"
	RequestPropertyContainsAddedId        = "request-property-contains-added"
	RequestPropertyContainsRemovedId      = "request-property-contains-removed"
	RequestPropertyMinContainsIncreasedId = "request-property-min-contains-increased"
	RequestPropertyMinContainsDecreasedId = "request-property-min-contains-decreased"
	RequestPropertyMaxContainsIncreasedId = "request-property-max-contains-increased"
	RequestPropertyMaxContainsDecreasedId = "request-property-max-contains-decreased"
)

func RequestPropertyContainsUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

				// Contains add/remove
				if mediaTypeDiff.SchemaDiff.ContainsDiff != nil {
					if mediaTypeDiff.SchemaDiff.ContainsDiff.SchemaAdded {
						appendResultItem(RequestBodyContainsAddedId)
					}
					if mediaTypeDiff.SchemaDiff.ContainsDiff.SchemaDeleted {
						appendResultItem(RequestBodyContainsRemovedId)
					}
				}

				// MinContains
				if mediaTypeDiff.SchemaDiff.MinContainsDiff != nil {
					if IsIncreasedValue(mediaTypeDiff.SchemaDiff.MinContainsDiff) {
						appendResultItem(RequestBodyMinContainsIncreasedId, mediaTypeDiff.SchemaDiff.MinContainsDiff.From, mediaTypeDiff.SchemaDiff.MinContainsDiff.To)
					}
					if IsDecreasedValue(mediaTypeDiff.SchemaDiff.MinContainsDiff) {
						appendResultItem(RequestBodyMinContainsDecreasedId, mediaTypeDiff.SchemaDiff.MinContainsDiff.From, mediaTypeDiff.SchemaDiff.MinContainsDiff.To)
					}
				}

				// MaxContains
				if mediaTypeDiff.SchemaDiff.MaxContainsDiff != nil {
					if IsIncreasedValue(mediaTypeDiff.SchemaDiff.MaxContainsDiff) {
						appendResultItem(RequestBodyMaxContainsIncreasedId, mediaTypeDiff.SchemaDiff.MaxContainsDiff.From, mediaTypeDiff.SchemaDiff.MaxContainsDiff.To)
					}
					if IsDecreasedValue(mediaTypeDiff.SchemaDiff.MaxContainsDiff) {
						appendResultItem(RequestBodyMaxContainsDecreasedId, mediaTypeDiff.SchemaDiff.MaxContainsDiff.From, mediaTypeDiff.SchemaDiff.MaxContainsDiff.To)
					}
				}

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

						if propertyDiff.ContainsDiff != nil {
							if propertyDiff.ContainsDiff.SchemaAdded {
								appendPropResultItem(RequestPropertyContainsAddedId, propName)
							}
							if propertyDiff.ContainsDiff.SchemaDeleted {
								appendPropResultItem(RequestPropertyContainsRemovedId, propName)
							}
						}

						if propertyDiff.MinContainsDiff != nil {
							if IsIncreasedValue(propertyDiff.MinContainsDiff) {
								appendPropResultItem(RequestPropertyMinContainsIncreasedId, propName, propertyDiff.MinContainsDiff.From, propertyDiff.MinContainsDiff.To)
							}
							if IsDecreasedValue(propertyDiff.MinContainsDiff) {
								appendPropResultItem(RequestPropertyMinContainsDecreasedId, propName, propertyDiff.MinContainsDiff.From, propertyDiff.MinContainsDiff.To)
							}
						}

						if propertyDiff.MaxContainsDiff != nil {
							if IsIncreasedValue(propertyDiff.MaxContainsDiff) {
								appendPropResultItem(RequestPropertyMaxContainsIncreasedId, propName, propertyDiff.MaxContainsDiff.From, propertyDiff.MaxContainsDiff.To)
							}
							if IsDecreasedValue(propertyDiff.MaxContainsDiff) {
								appendPropResultItem(RequestPropertyMaxContainsDecreasedId, propName, propertyDiff.MaxContainsDiff.From, propertyDiff.MaxContainsDiff.To)
							}
						}
					})
			}
		}
	}
	return result
}
