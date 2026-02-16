package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyMinDecreasedId              = "response-body-min-decreased"
	ResponsePropertyMinDecreasedId          = "response-property-min-decreased"
	ResponseBodyExclusiveMinDecreasedId     = "response-body-exclusive-min-decreased"
	ResponsePropertyExclusiveMinDecreasedId = "response-property-exclusive-min-decreased"
)

func ResponsePropertyMinDecreasedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
					if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.MinDiff != nil {
						minDiff := mediaTypeDiff.SchemaDiff.MinDiff
						if minDiff.From != nil &&
							minDiff.To != nil {
							if IsDecreasedValue(minDiff) {
								result = append(result, NewApiChange(
									ResponseBodyMinDecreasedId,
									config,
									[]any{minDiff.From, minDiff.To},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithDetails(mediaTypeDetails))
							}
						}
					}
					if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.ExclusiveMinDiff != nil {
						exMinDiff := mediaTypeDiff.SchemaDiff.ExclusiveMinDiff
						if exMinDiff.From != nil &&
							exMinDiff.To != nil {
							if IsDecreasedValue(exMinDiff) {
								result = append(result, NewApiChange(
									ResponseBodyExclusiveMinDecreasedId,
									config,
									[]any{exMinDiff.From, exMinDiff.To},
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
							minDiff := propertyDiff.MinDiff
							if minDiff == nil {
								return
							}
							if minDiff.To == nil ||
								minDiff.From == nil {
								return
							}
							if !IsDecreasedValue(minDiff) {
								return
							}

							if propertyDiff.Revision.WriteOnly {
								return
							}

							result = append(result, NewApiChange(
								ResponsePropertyMinDecreasedId,
								config,
								[]any{propertyFullName(propertyPath, propertyName), minDiff.From, minDiff.To, responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithDetails(mediaTypeDetails))
						})

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							exMinDiff := propertyDiff.ExclusiveMinDiff
							if exMinDiff == nil {
								return
							}
							if exMinDiff.To == nil ||
								exMinDiff.From == nil {
								return
							}
							if !IsDecreasedValue(exMinDiff) {
								return
							}

							if propertyDiff.Revision.WriteOnly {
								return
							}

							result = append(result, NewApiChange(
								ResponsePropertyExclusiveMinDecreasedId,
								config,
								[]any{propertyFullName(propertyPath, propertyName), exMinDiff.From, exMinDiff.To, responseStatus},
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
