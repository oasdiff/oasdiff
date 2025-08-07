package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyListOfTypesWidenedId      = "response-body-list-of-types-widened"
	ResponseBodyListOfTypesNarrowedId     = "response-body-list-of-types-narrowed"
	ResponsePropertyListOfTypesWidenedId  = "response-property-list-of-types-widened"
	ResponsePropertyListOfTypesNarrowedId = "response-property-list-of-types-narrowed"
)

func ResponsePropertyListOfTypesChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
				if responseDiff.ContentDiff == nil ||
					responseDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}

				modifiedMediaTypes := responseDiff.ContentDiff.MediaTypeModified
				for mediaType, mediaTypeDiff := range modifiedMediaTypes {
					if mediaTypeDiff.SchemaDiff == nil {
						continue
					}

					// Check response body schema
					changes := checkBodyListOfTypesChange(
						mediaTypeDiff.SchemaDiff,
						mediaType,
						responseStatus,
						config,
						operationsSources,
						operationItem,
						operation,
						path,
						false, // isRequest
					)
					result = append(result, changes...)

					// Check response body properties
					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							if propertyDiff == nil || propertyDiff.Revision == nil {
								return
							}

							changes := checkPropertyListOfTypesChange(
								propertyPath,
								propertyName,
								propertyDiff,
								mediaType,
								responseStatus,
								config,
								operationsSources,
								operationItem,
								operation,
								path,
								false, // isRequest
							)
							result = append(result, changes...)
						})
				}
			}
		}
	}
	return result
}
