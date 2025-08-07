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
				for _, mediaTypeDiff := range modifiedMediaTypes {
					if mediaTypeDiff.SchemaDiff == nil {
						continue
					}

					// Check for suppression by ListOfTypes checker
					if shouldSuppressOneOfSchemaChangedForListOfTypes(mediaTypeDiff.SchemaDiff) {
						continue
					}

					if mediaTypeDiff.SchemaDiff.AnyOfDiff != nil && len(mediaTypeDiff.SchemaDiff.AnyOfDiff.Added) > 0 {

						result = append(result, NewApiChange(
							ResponseBodyAnyOfAddedId,
							config,
							[]any{mediaTypeDiff.SchemaDiff.AnyOfDiff.Added.String(), responseStatus},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						))
					}

					if mediaTypeDiff.SchemaDiff.AnyOfDiff != nil && len(mediaTypeDiff.SchemaDiff.AnyOfDiff.Deleted) > 0 {
						result = append(result, NewApiChange(
							ResponseBodyAnyOfRemovedId,
							config,
							[]any{mediaTypeDiff.SchemaDiff.AnyOfDiff.Deleted.String(), responseStatus},
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
							if propertyDiff.AnyOfDiff == nil {
								return
							}

							// Check for suppression by ListOfTypes checker
							if shouldSuppressPropertyOneOfSchemaChangedForListOfTypes(propertyDiff) {
								return
							}

							if len(propertyDiff.AnyOfDiff.Added) > 0 {

								result = append(result, NewApiChange(
									ResponsePropertyAnyOfAddedId,
									config,
									[]any{propertyDiff.AnyOfDiff.Added.String(), propertyFullName(propertyPath, propertyName), responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								))
							}

							if len(propertyDiff.AnyOfDiff.Deleted) > 0 {

								result = append(result, NewApiChange(
									ResponsePropertyAnyOfRemovedId,
									config,
									[]any{propertyDiff.AnyOfDiff.Deleted.String(), propertyFullName(propertyPath, propertyName), responseStatus},
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
