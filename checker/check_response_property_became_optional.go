package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyBecameOptionalId          = "response-property-became-optional"
	ResponseWriteOnlyPropertyBecameOptionalId = "response-write-only-property-became-optional"
)

func ResponsePropertyBecameOptionalCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {

			if operationItem.ResponsesDiff == nil {
				continue
			}

			for responseStatus, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff.ContentDiff == nil ||
					responseDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}

				modifiedMediaTypes := responseDiff.ContentDiff.MediaTypeModified
				for mediaType, mediaTypeDiff := range modifiedMediaTypes {
					mediaTypeDetails := formatMediaTypeDetails(mediaType, len(modifiedMediaTypes))
					if mediaTypeDiff.SchemaDiff == nil {
						continue
					}

					if mediaTypeDiff.SchemaDiff.RequiredDiff != nil {
						for _, changedRequiredPropertyName := range mediaTypeDiff.SchemaDiff.RequiredDiff.Deleted {
							id := ResponsePropertyBecameOptionalId
							if mediaTypeDiff.SchemaDiff.Revision.Properties[changedRequiredPropertyName] == nil {
								// removed properties processed by the ResponseRequiredPropertyUpdatedCheck check
								continue
							}
							if mediaTypeDiff.SchemaDiff.Revision.Properties[changedRequiredPropertyName].Value.WriteOnly {
								id = ResponseWriteOnlyPropertyBecameOptionalId
							}

							result = append(result, NewApiChange(
								id,
								config,
								[]any{changedRequiredPropertyName, responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithDetails(mediaTypeDetails))
						}
					}

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							requiredDiff := propertyDiff.RequiredDiff
							if requiredDiff == nil {
								return
							}
							for _, changedRequiredPropertyName := range requiredDiff.Deleted {

								if propertyDiff.Base.Properties[changedRequiredPropertyName] == nil {
									continue
								}

								if propertyDiff.Revision.Properties[changedRequiredPropertyName] == nil {
									// removed properties processed by the ResponseRequiredPropertyUpdatedCheck check
									continue
								}

								id := ResponsePropertyBecameOptionalId

								if propertyDiff.Base.Properties[changedRequiredPropertyName].Value.WriteOnly {
									id = ResponseWriteOnlyPropertyBecameOptionalId
								}

								result = append(result, NewApiChange(
									id,
									config,
									[]any{propertyFullName(propertyPath, propertyFullName(propertyName, changedRequiredPropertyName)), responseStatus},
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
