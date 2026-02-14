package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyDependentSchemaAddedId       = "request-body-dependent-schema-added"
	RequestBodyDependentSchemaRemovedId     = "request-body-dependent-schema-removed"
	RequestPropertyDependentSchemaAddedId   = "request-property-dependent-schema-added"
	RequestPropertyDependentSchemaRemovedId = "request-property-dependent-schema-removed"
)

func RequestPropertyDependentSchemasUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

				if mediaTypeDiff.SchemaDiff.DependentSchemasDiff != nil {
					for _, name := range mediaTypeDiff.SchemaDiff.DependentSchemasDiff.Added {
						appendResultItem(RequestBodyDependentSchemaAddedId, name)
					}
					for _, name := range mediaTypeDiff.SchemaDiff.DependentSchemasDiff.Deleted {
						appendResultItem(RequestBodyDependentSchemaRemovedId, name)
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
							appendResultItem(RequestPropertyDependentSchemaAddedId, name, propName)
						}
						for _, name := range propertyDiff.DependentSchemasDiff.Deleted {
							appendResultItem(RequestPropertyDependentSchemaRemovedId, name, propName)
						}
					})
			}
		}
	}
	return result
}
