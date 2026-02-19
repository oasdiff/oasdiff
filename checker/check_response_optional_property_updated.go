package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"slices"
)

const (
	ResponseOptionalPropertyRemovedId          = "response-optional-property-removed"
	ResponseOptionalWriteOnlyPropertyRemovedId = "response-optional-write-only-property-removed"
	ResponseOptionalPropertyAddedId            = "response-optional-property-added"
	ResponseOptionalWriteOnlyPropertyAddedId   = "response-optional-write-only-property-added"
)

func ResponseOptionalPropertyUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {

			if operationItem.ResponsesDiff == nil {
				continue
			}

			baseSource, revisionSource := operationSources(operationsSources, operationItem.Base, operationItem.Revision)
			for responseStatus, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff.ContentDiff == nil ||
					responseDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}

				modifiedMediaTypes := responseDiff.ContentDiff.MediaTypeModified
				for mediaType, mediaTypeDiff := range modifiedMediaTypes {
					mediaTypeDetails := formatMediaTypeDetails(mediaType, len(modifiedMediaTypes))
					CheckDeletedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, parent *diff.SchemaDiff) {
							id := ResponseOptionalPropertyRemovedId
							if propertyItem.WriteOnly {
								id = ResponseOptionalWriteOnlyPropertyRemovedId
							}
							if slices.Contains(parent.Base.Required, propertyName) {
								// covered by response-required-property-removed
								return
							}

							result = append(result, NewApiChange(
								id,
								config,
								[]any{propertyFullName(propertyPath, propertyName), responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
						})
					CheckAddedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, parent *diff.SchemaDiff) {
							id := ResponseOptionalPropertyAddedId
							if propertyItem.WriteOnly {
								id = ResponseOptionalWriteOnlyPropertyAddedId
							}

							if slices.Contains(parent.Revision.Required, propertyName) {
								// covered by response-required-property-added
								return
							}

							result = append(result, NewApiChange(
								id,
								config,
								[]any{propertyFullName(propertyPath, propertyName), responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
						})
				}
			}
		}
	}
	return result
}
