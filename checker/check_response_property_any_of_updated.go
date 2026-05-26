package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyAnyOfAddedId       = "response-body-any-of-added"
	ResponseBodyAnyOfRemovedId     = "response-body-any-of-removed"
	ResponsePropertyAnyOfAddedId   = "response-property-any-of-added"
	ResponsePropertyAnyOfRemovedId = "response-property-any-of-removed"
)

func ResponsePropertyAnyOfUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

					// Check for suppression by ListOfTypes checker
					if shouldSuppressOneOfSchemaChangedForListOfTypes(mediaTypeDiff.SchemaDiff) {
						continue
					}

					if mediaTypeDiff.SchemaDiff.AnyOfDiff != nil {
						if added := mediaTypeDiff.SchemaDiff.AnyOfDiff.Added; len(added) > 0 {
							baseSource, revisionSource := SubschemaSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "anyOf", -1, added[0].Index)
							result = append(result, NewApiChange(
								ResponseBodyAnyOfAddedId,
								config,
								[]any{added.String(), responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
						}

						if deleted := mediaTypeDiff.SchemaDiff.AnyOfDiff.Deleted; len(deleted) > 0 {
							baseSource, revisionSource := SubschemaSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "anyOf", deleted[0].Index, -1)
							result = append(result, NewApiChange(
								ResponseBodyAnyOfRemovedId,
								config,
								[]any{deleted.String(), responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
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

							if added := propertyDiff.AnyOfDiff.Added; len(added) > 0 {
								propBaseSource, propRevisionSource := SubschemaSources(operationsSources, operationItem, propertyDiff, "anyOf", -1, added[0].Index)
								result = append(result, NewApiChange(
									ResponsePropertyAnyOfAddedId,
									config,
									[]any{added.String(), propertyFullName(propertyPath, propertyName), responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(propBaseSource, propRevisionSource).WithDetails(mediaTypeDetails))
							}

							if deleted := propertyDiff.AnyOfDiff.Deleted; len(deleted) > 0 {
								propBaseSource, propRevisionSource := SubschemaSources(operationsSources, operationItem, propertyDiff, "anyOf", deleted[0].Index, -1)
								result = append(result, NewApiChange(
									ResponsePropertyAnyOfRemovedId,
									config,
									[]any{deleted.String(), propertyFullName(propertyPath, propertyName), responseStatus},
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
	}
	return result
}
