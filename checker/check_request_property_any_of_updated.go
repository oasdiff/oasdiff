package checker

import "github.com/oasdiff/oasdiff/diff"

const (
	RequestBodyAnyOfAddedId       = "request-body-any-of-added"
	RequestBodyAnyOfRemovedId     = "request-body-any-of-removed"
	RequestPropertyAnyOfAddedId   = "request-property-any-of-added"
	RequestPropertyAnyOfRemovedId = "request-property-any-of-removed"
)

func RequestPropertyAnyOfUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

				// Check for suppression by ListOfTypes checker
				if shouldSuppressOneOfSchemaChangedForListOfTypes(mediaTypeDiff.SchemaDiff) {
					continue
				}

				if mediaTypeDiff.SchemaDiff.AnyOfDiff != nil {
					added := filterValidationEquivalentAddedSubschemas(
						mediaTypeDiff.SchemaDiff.AnyOfDiff.Added,
						mediaTypeDiff.SchemaDiff.Base.AnyOf,
						mediaTypeDiff.SchemaDiff.Revision.AnyOf,
					)
					if len(added) > 0 {
						baseSource, revisionSource := SubschemaSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "anyOf", -1, added[0].Index)
						result = append(result, NewApiChange(
							RequestBodyAnyOfAddedId,
							config,
							[]any{added.String()},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, revisionSource))
					}

					deleted := filterValidationEquivalentDeletedSubschemas(
						mediaTypeDiff.SchemaDiff.AnyOfDiff.Deleted,
						mediaTypeDiff.SchemaDiff.Base.AnyOf,
						mediaTypeDiff.SchemaDiff.Revision.AnyOf,
					)
					if len(deleted) > 0 {
						baseSource, revisionSource := SubschemaSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "anyOf", deleted[0].Index, -1)
						result = append(result, NewApiChange(
							RequestBodyAnyOfRemovedId,
							config,
							[]any{deleted.String()},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, revisionSource))
					}
				}

				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						if propertyDiff.AnyOfDiff == nil {
							return
						}

						// Check for suppression by ListOfTypes checker
						if shouldSuppressPropertyOneOfSchemaChangedForListOfTypes(propertyDiff) {
							return
						}

						propName := propertyFullName(propertyPath, propertyName)

						added := filterValidationEquivalentAddedSubschemas(
							propertyDiff.AnyOfDiff.Added,
							propertyDiff.Base.AnyOf,
							propertyDiff.Revision.AnyOf,
						)
						if len(added) > 0 {
							propBaseSource, propRevisionSource := SubschemaSources(operationsSources, operationItem, propertyDiff, "anyOf", -1, added[0].Index)
							result = append(result, NewApiChange(
								RequestPropertyAnyOfAddedId,
								config,
								[]any{added.String(), propName},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(propBaseSource, propRevisionSource).WithDetails(mediaTypeDetails))
						}

						deleted := filterValidationEquivalentDeletedSubschemas(
							propertyDiff.AnyOfDiff.Deleted,
							propertyDiff.Base.AnyOf,
							propertyDiff.Revision.AnyOf,
						)
						if len(deleted) > 0 {
							propBaseSource, propRevisionSource := SubschemaSources(operationsSources, operationItem, propertyDiff, "anyOf", deleted[0].Index, -1)
							result = append(result, NewApiChange(
								RequestPropertyAnyOfRemovedId,
								config,
								[]any{deleted.String(), propName},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(propBaseSource, propRevisionSource).WithDetails(mediaTypeDetails))
						}
					})
			}
		}
	}
	return result
}
