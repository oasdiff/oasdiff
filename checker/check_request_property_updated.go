package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"slices"
)

const (
	RequestPropertyRemovedId                = "request-property-removed"
	NewRequiredRequestPropertyId            = "new-required-request-property"
	NewRequiredRequestPropertyWithDefaultId = "new-required-request-property-with-default"
	NewOptionalRequestPropertyId            = "new-optional-request-property"
)

func RequestPropertyUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
				CheckDeletedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, parent *diff.SchemaDiff) {
						if !propertyItem.ReadOnly {

							result = append(result, NewApiChange(
								RequestPropertyRemovedId,
								config,
								[]any{propertyFullName(propertyPath, propertyName)},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithDetails(mediaTypeDetails))
						}
					})
				CheckAddedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, parent *diff.SchemaDiff) {
						if propertyItem.ReadOnly {
							return
						}

						propName := propertyFullName(propertyPath, propertyName)

						if slices.Contains(parent.Revision.Required, propertyName) {
							if propertyItem.Default == nil {
								result = append(result, NewApiChange(
									NewRequiredRequestPropertyId,
									config,
									[]any{propName},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithDetails(mediaTypeDetails))
							} else {
								result = append(result, NewApiChange(
									NewRequiredRequestPropertyWithDefaultId,
									config,
									[]any{propName},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithDetails(mediaTypeDetails))
							}
						} else {
							result = append(result, NewApiChange(
								NewOptionalRequestPropertyId,
								config,
								[]any{propName},
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
