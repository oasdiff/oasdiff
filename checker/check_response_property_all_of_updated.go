package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyAllOfAddedId       = "response-body-all-of-added"
	ResponseBodyAllOfRemovedId     = "response-body-all-of-removed"
	ResponsePropertyAllOfAddedId   = "response-property-all-of-added"
	ResponsePropertyAllOfRemovedId = "response-property-all-of-removed"
)

func ResponsePropertyAllOfUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

					if mediaTypeDiff.SchemaDiff.AllOfDiff != nil && len(mediaTypeDiff.SchemaDiff.AllOfDiff.Added) > 0 {
						result = append(result, NewApiChange(
							ResponseBodyAllOfAddedId,
							config,
							[]any{mediaTypeDiff.SchemaDiff.AllOfDiff.Added.String(), responseStatus},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithDetails(mediaTypeDetails))
					}

					if mediaTypeDiff.SchemaDiff.AllOfDiff != nil && len(mediaTypeDiff.SchemaDiff.AllOfDiff.Deleted) > 0 {
						result = append(result, NewApiChange(
							ResponseBodyAllOfRemovedId,
							config,
							[]any{mediaTypeDiff.SchemaDiff.AllOfDiff.Deleted.String(), responseStatus},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithDetails(mediaTypeDetails))
					}

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							if propertyDiff.AllOfDiff == nil {
								return
							}

							if len(propertyDiff.AllOfDiff.Added) > 0 {

								result = append(result, NewApiChange(
									ResponsePropertyAllOfAddedId,
									config,
									[]any{propertyDiff.AllOfDiff.Added.String(), propertyFullName(propertyPath, propertyName), responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithDetails(mediaTypeDetails))
							}

							if len(propertyDiff.AllOfDiff.Deleted) > 0 {

								result = append(result, NewApiChange(
									ResponsePropertyAllOfRemovedId,
									config,
									[]any{propertyDiff.AllOfDiff.Deleted.String(), propertyFullName(propertyPath, propertyName), responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithDetails(mediaTypeDetails))
							}
						})
				}
			}
		}
	}
	return result
}
