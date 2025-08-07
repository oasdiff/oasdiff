package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyOneOfAddedId       = "response-body-one-of-added"
	ResponseBodyOneOfRemovedId     = "response-body-one-of-removed"
	ResponsePropertyOneOfAddedId   = "response-property-one-of-added"
	ResponsePropertyOneOfRemovedId = "response-property-one-of-removed"
)

func ResponsePropertyOneOfUpdated(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
				for _, mediaTypeDiff := range modifiedMediaTypes {
					if mediaTypeDiff.SchemaDiff == nil {
						continue
					}

					// Check for suppression by ListOfTypes checker
					if shouldSuppressOneOfSchemaChangedForListOfTypes(mediaTypeDiff.SchemaDiff) {
						continue
					}

					if mediaTypeDiff.SchemaDiff.OneOfDiff != nil && len(mediaTypeDiff.SchemaDiff.OneOfDiff.Added) > 0 {
						result = append(result, NewApiChange(
							ResponseBodyOneOfAddedId,
							config,
							[]any{mediaTypeDiff.SchemaDiff.OneOfDiff.Added.String(), responseStatus},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						))
					}

					if mediaTypeDiff.SchemaDiff.OneOfDiff != nil && len(mediaTypeDiff.SchemaDiff.OneOfDiff.Deleted) > 0 {
						result = append(result, NewApiChange(
							ResponseBodyOneOfRemovedId,
							config,
							[]any{mediaTypeDiff.SchemaDiff.OneOfDiff.Deleted.String(), responseStatus},
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
							if propertyDiff.OneOfDiff == nil {
								return
							}

							// Check for suppression by ListOfTypes checker
							if shouldSuppressPropertyOneOfSchemaChangedForListOfTypes(propertyDiff) {
								return
							}

							propName := propertyFullName(propertyPath, propertyName)

							if len(propertyDiff.OneOfDiff.Added) > 0 {
								result = append(result, NewApiChange(
									ResponsePropertyOneOfAddedId,
									config,
									[]any{propertyDiff.OneOfDiff.Added.String(), propName, responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								))
							}

							if len(propertyDiff.OneOfDiff.Deleted) > 0 {
								result = append(result, NewApiChange(
									ResponsePropertyOneOfRemovedId,
									config,
									[]any{propertyDiff.OneOfDiff.Deleted.String(), propName, responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								))
							}
						})
				}
			}
		}
	}
	return result
}
