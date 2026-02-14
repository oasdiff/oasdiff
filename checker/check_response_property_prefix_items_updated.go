package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyPrefixItemsAddedId       = "response-body-prefix-items-added"
	ResponseBodyPrefixItemsRemovedId     = "response-body-prefix-items-removed"
	ResponsePropertyPrefixItemsAddedId   = "response-property-prefix-items-added"
	ResponsePropertyPrefixItemsRemovedId = "response-property-prefix-items-removed"
)

func ResponsePropertyPrefixItemsUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

			for responseStatus, responsesDiff := range operationItem.ResponsesDiff.Modified {
				if responsesDiff.ContentDiff == nil || responsesDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}

				modifiedMediaTypes := responsesDiff.ContentDiff.MediaTypeModified
				for mediaType, mediaTypeDiff := range modifiedMediaTypes {
					mediaTypeDetails := formatMediaTypeDetails(mediaType, len(modifiedMediaTypes))
					if mediaTypeDiff.SchemaDiff == nil {
						continue
					}

					if mediaTypeDiff.SchemaDiff.PrefixItemsDiff != nil && len(mediaTypeDiff.SchemaDiff.PrefixItemsDiff.Added) > 0 {
						result = append(result, NewApiChange(
							ResponseBodyPrefixItemsAddedId,
							config,
							[]any{mediaTypeDiff.SchemaDiff.PrefixItemsDiff.Added.String(), responseStatus},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithDetails(mediaTypeDetails))
					}

					if mediaTypeDiff.SchemaDiff.PrefixItemsDiff != nil && len(mediaTypeDiff.SchemaDiff.PrefixItemsDiff.Deleted) > 0 {
						result = append(result, NewApiChange(
							ResponseBodyPrefixItemsRemovedId,
							config,
							[]any{mediaTypeDiff.SchemaDiff.PrefixItemsDiff.Deleted.String(), responseStatus},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithDetails(mediaTypeDetails))
					}

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							if propertyDiff.PrefixItemsDiff == nil {
								return
							}

							propName := propertyFullName(propertyPath, propertyName)

							if len(propertyDiff.PrefixItemsDiff.Added) > 0 {
								result = append(result, NewApiChange(
									ResponsePropertyPrefixItemsAddedId,
									config,
									[]any{propertyDiff.PrefixItemsDiff.Added.String(), propName, responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithDetails(mediaTypeDetails))
							}

							if len(propertyDiff.PrefixItemsDiff.Deleted) > 0 {
								result = append(result, NewApiChange(
									ResponsePropertyPrefixItemsRemovedId,
									config,
									[]any{propertyDiff.PrefixItemsDiff.Deleted.String(), propName, responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithDetails(mediaTypeDetails))
							}
						})
				}
			}
		}
	}
	return result
}
