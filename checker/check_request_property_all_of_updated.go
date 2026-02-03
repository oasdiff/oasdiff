package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyAllOfAddedId       = "request-body-all-of-added"
	RequestBodyAllOfRemovedId     = "request-body-all-of-removed"
	RequestPropertyAllOfAddedId   = "request-property-all-of-added"
	RequestPropertyAllOfRemovedId = "request-property-all-of-removed"
)

func RequestPropertyAllOfUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

				if mediaTypeDiff.SchemaDiff.AllOfDiff != nil && len(mediaTypeDiff.SchemaDiff.AllOfDiff.Added) > 0 {
					result = append(result, NewApiChange(
						RequestBodyAllOfAddedId,
						config,
						[]any{mediaTypeDiff.SchemaDiff.AllOfDiff.Added.String()},
						"",
						operationsSources,
						operationItem.Revision,
						operation,
						path,
					))
				}

				if mediaTypeDiff.SchemaDiff.AllOfDiff != nil && len(mediaTypeDiff.SchemaDiff.AllOfDiff.Deleted) > 0 {
					result = append(result, NewApiChange(
						RequestBodyAllOfRemovedId,
						config,
						[]any{mediaTypeDiff.SchemaDiff.AllOfDiff.Deleted.String()},
						"",
						operationsSources,
						operationItem.Revision,
						operation,
						path,
					))
				}

				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						if propertyDiff.AllOfDiff == nil {
							return
						}

						propName := propertyFullName(propertyPath, propertyName)

						if len(propertyDiff.AllOfDiff.Added) > 0 {
							result = append(result, NewApiChange(
								RequestPropertyAllOfAddedId,
								config,
								[]any{propertyDiff.AllOfDiff.Added.String(), propName},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithDetails(mediaTypeDetails))
						}

						if len(propertyDiff.AllOfDiff.Deleted) > 0 {
							result = append(result, NewApiChange(
								RequestPropertyAllOfRemovedId,
								config,
								[]any{propertyDiff.AllOfDiff.Deleted.String(), propName},
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
