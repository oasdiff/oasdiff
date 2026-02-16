package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyContainsAddedId            = "response-body-contains-added"
	ResponseBodyContainsRemovedId          = "response-body-contains-removed"
	ResponseBodyMinContainsIncreasedId     = "response-body-min-contains-increased"
	ResponseBodyMinContainsDecreasedId     = "response-body-min-contains-decreased"
	ResponseBodyMaxContainsIncreasedId     = "response-body-max-contains-increased"
	ResponseBodyMaxContainsDecreasedId     = "response-body-max-contains-decreased"
	ResponsePropertyContainsAddedId        = "response-property-contains-added"
	ResponsePropertyContainsRemovedId      = "response-property-contains-removed"
	ResponsePropertyMinContainsIncreasedId = "response-property-min-contains-increased"
	ResponsePropertyMinContainsDecreasedId = "response-property-min-contains-decreased"
	ResponsePropertyMaxContainsIncreasedId = "response-property-max-contains-increased"
	ResponsePropertyMaxContainsDecreasedId = "response-property-max-contains-decreased"
)

func ResponsePropertyContainsUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}

	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}

		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.ResponsesDiff == nil || operationItem.ResponsesDiff.Modified == nil {
				continue
			}

			for responseStatus, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff.ContentDiff == nil || responseDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}

				modifiedMediaTypes := responseDiff.ContentDiff.MediaTypeModified
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

					if mediaTypeDiff.SchemaDiff.ContainsDiff != nil {
						if mediaTypeDiff.SchemaDiff.ContainsDiff.SchemaAdded {
							appendResultItem(ResponseBodyContainsAddedId, responseStatus)
						}
						if mediaTypeDiff.SchemaDiff.ContainsDiff.SchemaDeleted {
							appendResultItem(ResponseBodyContainsRemovedId, responseStatus)
						}
					}

					if mediaTypeDiff.SchemaDiff.MinContainsDiff != nil {
						if IsIncreasedValue(mediaTypeDiff.SchemaDiff.MinContainsDiff) {
							appendResultItem(ResponseBodyMinContainsIncreasedId, mediaTypeDiff.SchemaDiff.MinContainsDiff.From, mediaTypeDiff.SchemaDiff.MinContainsDiff.To, responseStatus)
						}
						if IsDecreasedValue(mediaTypeDiff.SchemaDiff.MinContainsDiff) {
							appendResultItem(ResponseBodyMinContainsDecreasedId, mediaTypeDiff.SchemaDiff.MinContainsDiff.From, mediaTypeDiff.SchemaDiff.MinContainsDiff.To, responseStatus)
						}
					}

					if mediaTypeDiff.SchemaDiff.MaxContainsDiff != nil {
						if IsIncreasedValue(mediaTypeDiff.SchemaDiff.MaxContainsDiff) {
							appendResultItem(ResponseBodyMaxContainsIncreasedId, mediaTypeDiff.SchemaDiff.MaxContainsDiff.From, mediaTypeDiff.SchemaDiff.MaxContainsDiff.To, responseStatus)
						}
						if IsDecreasedValue(mediaTypeDiff.SchemaDiff.MaxContainsDiff) {
							appendResultItem(ResponseBodyMaxContainsDecreasedId, mediaTypeDiff.SchemaDiff.MaxContainsDiff.From, mediaTypeDiff.SchemaDiff.MaxContainsDiff.To, responseStatus)
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
									appendPropResultItem(ResponsePropertyContainsAddedId, propName, responseStatus)
								}
								if propertyDiff.ContainsDiff.SchemaDeleted {
									appendPropResultItem(ResponsePropertyContainsRemovedId, propName, responseStatus)
								}
							}

							if propertyDiff.MinContainsDiff != nil {
								if IsIncreasedValue(propertyDiff.MinContainsDiff) {
									appendPropResultItem(ResponsePropertyMinContainsIncreasedId, propName, propertyDiff.MinContainsDiff.From, propertyDiff.MinContainsDiff.To, responseStatus)
								}
								if IsDecreasedValue(propertyDiff.MinContainsDiff) {
									appendPropResultItem(ResponsePropertyMinContainsDecreasedId, propName, propertyDiff.MinContainsDiff.From, propertyDiff.MinContainsDiff.To, responseStatus)
								}
							}

							if propertyDiff.MaxContainsDiff != nil {
								if IsIncreasedValue(propertyDiff.MaxContainsDiff) {
									appendPropResultItem(ResponsePropertyMaxContainsIncreasedId, propName, propertyDiff.MaxContainsDiff.From, propertyDiff.MaxContainsDiff.To, responseStatus)
								}
								if IsDecreasedValue(propertyDiff.MaxContainsDiff) {
									appendPropResultItem(ResponsePropertyMaxContainsDecreasedId, propName, propertyDiff.MaxContainsDiff.From, propertyDiff.MaxContainsDiff.To, responseStatus)
								}
							}
						})
				}
			}
		}
	}
	return result
}
