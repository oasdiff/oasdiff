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

					if mediaTypeDiff.SchemaDiff.DependentSchemasDiff != nil {
						baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "dependentSchemas")
						for _, name := range mediaTypeDiff.SchemaDiff.DependentSchemasDiff.Added {
							result = append(result, NewApiChange(
								ResponseBodyDependentSchemaAddedId,
								config,
								[]any{name, responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
						}
						for _, name := range mediaTypeDiff.SchemaDiff.DependentSchemasDiff.Deleted {
							result = append(result, NewApiChange(
								ResponseBodyDependentSchemaRemovedId,
								config,
								[]any{name, responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, nil).WithDetails(mediaTypeDetails))
						}
					}

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							if propertyDiff.DependentSchemasDiff == nil {
								return
							}
							propName := propertyFullName(propertyPath, propertyName)
							propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "dependentSchemas")
							for _, name := range propertyDiff.DependentSchemasDiff.Added {
								result = append(result, NewApiChange(
									ResponsePropertyDependentSchemaAddedId,
									config,
									[]any{name, propName, responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(nil, propRevisionSource).WithDetails(mediaTypeDetails))
							}
							for _, name := range propertyDiff.DependentSchemasDiff.Deleted {
								result = append(result, NewApiChange(
									ResponsePropertyDependentSchemaRemovedId,
									config,
									[]any{name, propName, responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(propBaseSource, nil).WithDetails(mediaTypeDetails))
							}
						})
				}
			}
		}
	}
	return result
}
