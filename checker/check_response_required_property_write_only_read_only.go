package checker

import (
	"github.com/oasdiff/oasdiff/diff"
	"slices"
)

const (
	ResponseRequiredPropertyBecameNonWriteOnlyId = "response-required-property-became-not-write-only"
	ResponseRequiredPropertyBecameWriteOnlyId    = "response-required-property-became-write-only"
	ResponseRequiredPropertyBecameReadOnlyId     = "response-required-property-became-read-only"
	ResponseRequiredPropertyBecameNonReadOnlyId  = "response-required-property-became-not-read-only"
)

func ResponseRequiredPropertyWriteOnlyReadOnlyCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							writeOnlyDiff := propertyDiff.WriteOnlyDiff
							if writeOnlyDiff == nil {
								return
							}
							if parent.Revision.Properties[propertyName] == nil {
								// removed properties processed by the ResponseRequiredPropertyUpdatedCheck check
								return
							}
							if !slices.Contains(parent.Base.Required, propertyName) {
								// skip non-required properties
								return
							}

							id := ResponseRequiredPropertyBecameNonWriteOnlyId
							comment := commentId(ResponseRequiredPropertyBecameNonWriteOnlyId)

							if writeOnlyDiff.To == true {
								id = ResponseRequiredPropertyBecameWriteOnlyId
								comment = ""
							}

							result = append(result, NewApiChange(
								id,
								config,
								[]any{propertyFullName(propertyPath, propertyName), responseStatus},
								comment,
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithDetails(mediaTypeDetails))
						})

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							readOnlyDiff := propertyDiff.ReadOnlyDiff
							if readOnlyDiff == nil {
								return
							}
							if parent.Revision.Properties[propertyName] == nil {
								// removed properties processed by the ResponseRequiredPropertyUpdatedCheck check
								return
							}
							if !slices.Contains(parent.Base.Required, propertyName) {
								// skip non-required properties
								return
							}

							id := ResponseRequiredPropertyBecameNonReadOnlyId

							if readOnlyDiff.To == true {
								id = ResponseRequiredPropertyBecameReadOnlyId
							}

							result = append(result, NewApiChange(
								id,
								config,
								[]any{propertyFullName(propertyPath, propertyName), responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithDetails(mediaTypeDetails))
						})
				}
			}
		}
	}

	return result
}
