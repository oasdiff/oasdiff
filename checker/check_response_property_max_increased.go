package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyMaxIncreasedId     = "response-body-max-increased"
	ResponsePropertyMaxIncreasedId = "response-property-max-increased"
)

func ResponsePropertyMaxIncreasedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
					if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.MaxDiff != nil {
						maxDiff := mediaTypeDiff.SchemaDiff.MaxDiff
						if maxDiff.From != nil &&
							maxDiff.To != nil {
							if IsIncreasedValue(maxDiff) {
								result = append(result, NewApiChange(
									ResponseBodyMaxIncreasedId,
									config,
									[]any{maxDiff.From, maxDiff.To},
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
							maxDiff := propertyDiff.MaxDiff
							if maxDiff == nil {
								return
							}
							if maxDiff.To == nil ||
								maxDiff.From == nil {
								return
							}
							if !IsIncreasedValue(maxDiff) {
								return
							}

							if propertyDiff.Revision.WriteOnly {
								return
							}

							result = append(result, NewApiChange(
								ResponsePropertyMaxIncreasedId,
								config,
								[]any{propertyFullName(propertyPath, propertyName), maxDiff.From, maxDiff.To, responseStatus},
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
