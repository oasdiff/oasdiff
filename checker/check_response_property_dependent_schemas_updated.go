package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyDependentSchemaAddedId       = "response-body-dependent-schema-added"
	ResponseBodyDependentSchemaRemovedId     = "response-body-dependent-schema-removed"
	ResponsePropertyDependentSchemaAddedId   = "response-property-dependent-schema-added"
	ResponsePropertyDependentSchemaRemovedId = "response-property-dependent-schema-removed"
)

func ResponsePropertyDependentSchemasUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

					if mediaTypeDiff.SchemaDiff.DependentSchemasDiff != nil {
						for _, name := range mediaTypeDiff.SchemaDiff.DependentSchemasDiff.Added {
							appendResultItem(ResponseBodyDependentSchemaAddedId, name, responseStatus)
						}
						for _, name := range mediaTypeDiff.SchemaDiff.DependentSchemasDiff.Deleted {
							appendResultItem(ResponseBodyDependentSchemaRemovedId, name, responseStatus)
						}
					}

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							if propertyDiff.DependentSchemasDiff == nil {
								return
							}
							propName := propertyFullName(propertyPath, propertyName)
							for _, name := range propertyDiff.DependentSchemasDiff.Added {
								appendResultItem(ResponsePropertyDependentSchemaAddedId, name, propName, responseStatus)
							}
							for _, name := range propertyDiff.DependentSchemasDiff.Deleted {
								appendResultItem(ResponsePropertyDependentSchemaRemovedId, name, propName, responseStatus)
							}
						})
				}
			}
		}
	}
	return result
}
