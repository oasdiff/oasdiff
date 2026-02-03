package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyBecameRequiredId          = "response-property-became-required"
	ResponseWriteOnlyPropertyBecameRequiredId = "response-write-only-property-became-required"
)

func ResponsePropertyBecameRequiredCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
						for _, changedRequiredPropertyName := range mediaTypeDiff.SchemaDiff.RequiredDiff.Added {
							if mediaTypeDiff.SchemaDiff.Base.Properties[changedRequiredPropertyName] == nil {
								// added properties are processed by ResponseRequiredPropertyUpdatedCheck check
								continue
							}
							if mediaTypeDiff.SchemaDiff.Revision.Properties[changedRequiredPropertyName] == nil {
								// removed properties processed by the ResponseRequiredPropertyUpdatedCheck check
								continue
							}
							id := ResponsePropertyBecameRequiredId
							if mediaTypeDiff.SchemaDiff.Revision.Properties[changedRequiredPropertyName].Value.WriteOnly {
								id = ResponseWriteOnlyPropertyBecameRequiredId
							}

							result = append(result, NewApiChange(
								id,
								config,
								[]any{propertyFullName("", changedRequiredPropertyName), responseStatus},
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
							for _, changedRequiredPropertyName := range requiredDiff.Added {
								if propertyDiff.Base.Properties[changedRequiredPropertyName] == nil {
									// added properties are processed by ResponseRequiredPropertyUpdatedCheck check
									continue
								}
								if propertyDiff.Revision.Properties[changedRequiredPropertyName] == nil {
									// removed properties processed by the ResponseRequiredPropertyUpdatedCheck check
									continue
								}
								id := ResponsePropertyBecameRequiredId
								if propertyDiff.Base.Properties[changedRequiredPropertyName].Value.WriteOnly {
									id = ResponseWriteOnlyPropertyBecameRequiredId
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
