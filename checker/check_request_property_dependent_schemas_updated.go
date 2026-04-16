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

				if mediaTypeDiff.SchemaDiff.DependentSchemasDiff != nil {
					depSchemasDiff := mediaTypeDiff.SchemaDiff.DependentSchemasDiff
					for _, name := range depSchemasDiff.Added {
						revisionSource := SchemaMapItemSource(operationsSources, operationItem.Revision, depSchemasDiff.Revision, name)
						result = append(result, NewApiChange(
							RequestBodyDependentSchemaAddedId,
							config,
							[]any{name},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
					}
					for _, name := range depSchemasDiff.Deleted {
						baseSource := SchemaMapItemSource(operationsSources, operationItem.Base, depSchemasDiff.Base, name)
						result = append(result, NewApiChange(
							RequestBodyDependentSchemaRemovedId,
							config,
							[]any{name},
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
						depSchemasDiff := propertyDiff.DependentSchemasDiff
						for _, name := range depSchemasDiff.Added {
							revisionSource := SchemaMapItemSource(operationsSources, operationItem.Revision, depSchemasDiff.Revision, name)
							result = append(result, NewApiChange(
								RequestPropertyDependentSchemaAddedId,
								config,
								[]any{name, propName},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
						}
						for _, name := range depSchemasDiff.Deleted {
							baseSource := SchemaMapItemSource(operationsSources, operationItem.Base, depSchemasDiff.Base, name)
							result = append(result, NewApiChange(
								RequestPropertyDependentSchemaRemovedId,
								config,
								[]any{name, propName},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, nil).WithDetails(mediaTypeDetails))
						}
					})
			}
		}
	}
	return result
}
