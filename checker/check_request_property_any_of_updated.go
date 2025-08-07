package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

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
						RequestBodyAnyOfAddedId,
						config,
						[]any{mediaTypeDiff.SchemaDiff.AnyOfDiff.Added.String()},
						"",
						operationsSources,
						operationItem.Revision,
						operation,
						path,
					))
				}

				if mediaTypeDiff.SchemaDiff.AnyOfDiff != nil && len(mediaTypeDiff.SchemaDiff.AnyOfDiff.Deleted) > 0 {
					result = append(result, NewApiChange(
						RequestBodyAnyOfRemovedId,
						config,
						[]any{mediaTypeDiff.SchemaDiff.AnyOfDiff.Deleted.String()},
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

						propName := propertyFullName(propertyPath, propertyName)

						if len(propertyDiff.AnyOfDiff.Added) > 0 {
							result = append(result, NewApiChange(
								RequestPropertyAnyOfAddedId,
								config,
								[]any{propertyDiff.AnyOfDiff.Added.String(), propName},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							))
						}

						if len(propertyDiff.AnyOfDiff.Deleted) > 0 {
							result = append(result, NewApiChange(
								RequestPropertyAnyOfRemovedId,
								config,
								[]any{propertyDiff.AnyOfDiff.Deleted.String(), propName},
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
	return result
}
