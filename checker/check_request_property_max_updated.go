package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMaxDecreasedId             = "request-body-max-decreased"
	RequestBodyMaxIncreasedId             = "request-body-max-increased"
	RequestPropertyMaxDecreasedId         = "request-property-max-decreased"
	RequestReadOnlyPropertyMaxDecreasedId = "request-read-only-property-max-decreased"
	RequestPropertyMaxIncreasedId         = "request-property-max-increased"
)

func RequestPropertyMaxDecreasedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

			baseSource, revisionSource := operationSources(operationsSources, operationItem.Base, operationItem.Revision)
			modifiedMediaTypes := operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified
			for mediaType, mediaTypeDiff := range modifiedMediaTypes {
				mediaTypeDetails := formatMediaTypeDetails(mediaType, len(modifiedMediaTypes))
				if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.MaxDiff != nil {
					maxDiff := mediaTypeDiff.SchemaDiff.MaxDiff
					if maxDiff.From != nil &&
						maxDiff.To != nil {
						if IsDecreasedValue(maxDiff) {
							result = append(result, NewApiChange(
								RequestBodyMaxDecreasedId,
								config,
								[]any{maxDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
						} else {
							result = append(result, NewApiChange(
								RequestBodyMaxIncreasedId,
								config,
								[]any{maxDiff.From, maxDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
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
						if maxDiff.From == nil ||
							maxDiff.To == nil {
							return
						}

						propName := propertyFullName(propertyPath, propertyName)

						id := RequestPropertyMaxDecreasedId
						if propertyDiff.Revision.ReadOnly {
							id = RequestReadOnlyPropertyMaxDecreasedId
						}

						if IsDecreasedValue(maxDiff) {
							result = append(result, NewApiChange(
								id,
								config,
								[]any{propName, maxDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
						} else {
							result = append(result, NewApiChange(
								RequestPropertyMaxIncreasedId,
								config,
								[]any{propName, maxDiff.From, maxDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
						}

					})
			}
		}
	}
	return result
}
