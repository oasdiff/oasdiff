package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMaxLengthSetId     = "request-body-max-length-set"
	RequestPropertyMaxLengthSetId = "request-property-max-length-set"
)

func RequestPropertyMaxLengthSetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
				if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.MaxLengthDiff != nil {
					maxLengthDiff := mediaTypeDiff.SchemaDiff.MaxLengthDiff
					if maxLengthDiff.From == nil &&
						maxLengthDiff.To != nil {
						result = append(result, NewApiChange(
							RequestBodyMaxLengthSetId,
							config,
							[]any{maxLengthDiff.To},
							commentId(RequestBodyMaxLengthSetId),
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
						maxLengthDiff := propertyDiff.MaxLengthDiff
						if maxLengthDiff == nil {
							return
						}
						if maxLengthDiff.From != nil ||
							maxLengthDiff.To == nil {
							return
						}
						if propertyDiff.Revision.ReadOnly {
							return
						}

						result = append(result, NewApiChange(
							RequestPropertyMaxLengthSetId,
							config,
							[]any{propertyFullName(propertyPath, propertyName), maxLengthDiff.To},
							commentId(RequestPropertyMaxLengthSetId),
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithDetails(mediaTypeDetails))
					})
			}
		}
	}
	return result
}
