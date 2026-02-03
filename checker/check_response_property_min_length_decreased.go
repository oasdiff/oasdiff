package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyMinLengthDecreasedId     = "response-body-min-length-decreased"
	ResponsePropertyMinLengthDecreasedId = "response-property-min-length-decreased"
)

func ResponsePropertyMinLengthDecreasedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
				if responseDiff == nil ||
					responseDiff.ContentDiff == nil ||
					responseDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}
				modifiedMediaTypes := responseDiff.ContentDiff.MediaTypeModified
				for mediaType, mediaTypeDiff := range modifiedMediaTypes {
					mediaTypeDetails := formatMediaTypeDetails(mediaType, len(modifiedMediaTypes))
					if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.MinLengthDiff != nil {
						minLengthDiff := mediaTypeDiff.SchemaDiff.MinLengthDiff
						if minLengthDiff.From != nil &&
							minLengthDiff.To != nil {
							if IsDecreasedValue(minLengthDiff) {
								result = append(result, NewApiChange(
									ResponseBodyMinLengthDecreasedId,
									config,
									[]any{minLengthDiff.From, minLengthDiff.To},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithDetails(mediaTypeDetails))
							}
						}
					}

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							minLengthDiff := propertyDiff.MinLengthDiff
							if minLengthDiff == nil {
								return
							}
							if minLengthDiff.To == nil ||
								minLengthDiff.From == nil {
								return
							}
							if !IsDecreasedValue(minLengthDiff) {
								return
							}

							if propertyDiff.Revision.WriteOnly {
								return
							}

							result = append(result, NewApiChange(
								ResponsePropertyMinLengthDecreasedId,
								config,
								[]any{propertyFullName(propertyPath, propertyName), minLengthDiff.From, minLengthDiff.To, responseStatus},
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
