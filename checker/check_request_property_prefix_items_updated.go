package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyPrefixItemsAddedId       = "request-body-prefix-items-added"
	RequestBodyPrefixItemsRemovedId     = "request-body-prefix-items-removed"
	RequestPropertyPrefixItemsAddedId   = "request-property-prefix-items-added"
	RequestPropertyPrefixItemsRemovedId = "request-property-prefix-items-removed"
)

func RequestPropertyPrefixItemsUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

				if mediaTypeDiff.SchemaDiff.PrefixItemsDiff != nil && len(mediaTypeDiff.SchemaDiff.PrefixItemsDiff.Added) > 0 {
					result = append(result, NewApiChange(
						RequestBodyPrefixItemsAddedId,
						config,
						[]any{mediaTypeDiff.SchemaDiff.PrefixItemsDiff.Added.String()},
						"",
						operationsSources,
						operationItem.Revision,
						operation,
						path,
					).WithDetails(mediaTypeDetails))
				}

				if mediaTypeDiff.SchemaDiff.PrefixItemsDiff != nil && len(mediaTypeDiff.SchemaDiff.PrefixItemsDiff.Deleted) > 0 {
					result = append(result, NewApiChange(
						RequestBodyPrefixItemsRemovedId,
						config,
						[]any{mediaTypeDiff.SchemaDiff.PrefixItemsDiff.Deleted.String()},
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
								RequestPropertyPrefixItemsAddedId,
								config,
								[]any{propertyDiff.PrefixItemsDiff.Added.String(), propName},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithDetails(mediaTypeDetails))
						}

						if len(propertyDiff.PrefixItemsDiff.Deleted) > 0 {
							result = append(result, NewApiChange(
								RequestPropertyPrefixItemsRemovedId,
								config,
								[]any{propertyDiff.PrefixItemsDiff.Deleted.String(), propName},
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
	return result
}
