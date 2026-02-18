package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMaxLengthDecreasedId             = "request-body-max-length-decreased"
	RequestBodyMaxLengthIncreasedId             = "request-body-max-length-increased"
	RequestPropertyMaxLengthDecreasedId         = "request-property-max-length-decreased"
	RequestReadOnlyPropertyMaxLengthDecreasedId = "request-read-only-property-max-length-decreased"
	RequestPropertyMaxLengthIncreasedId         = "request-property-max-length-increased"
)

func RequestPropertyMaxLengthUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
			for _, mediaTypeDiff := range modifiedMediaTypes {
				if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.MaxLengthDiff != nil {
					maxLengthDiff := mediaTypeDiff.SchemaDiff.MaxLengthDiff
					if maxLengthDiff.From != nil &&
						maxLengthDiff.To != nil {
						if IsDecreasedValue(maxLengthDiff) {
							result = append(result, NewApiChange(
								RequestBodyMaxLengthDecreasedId,
								config,
								[]any{maxLengthDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource))
						} else {
							result = append(result, NewApiChange(
								RequestBodyMaxLengthIncreasedId,
								config,
								[]any{maxLengthDiff.From, maxLengthDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource))
						}
					}
				}

				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						maxLengthDiff := propertyDiff.MaxLengthDiff
						if maxLengthDiff == nil {
							return
						}
						if maxLengthDiff.From == nil ||
							maxLengthDiff.To == nil {
							return
						}

						propName := propertyFullName(propertyPath, propertyName)

						if IsDecreasedValue(maxLengthDiff) {

							id := RequestPropertyMaxLengthDecreasedId

							if propertyDiff.Revision.ReadOnly {
								id = RequestReadOnlyPropertyMaxLengthDecreasedId
							}

							result = append(result, NewApiChange(
								id,
								config,
								[]any{propName, maxLengthDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource))
						} else {
							result = append(result, NewApiChange(
								RequestPropertyMaxLengthIncreasedId,
								config,
								[]any{propName, maxLengthDiff.From, maxLengthDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource))
						}

					})
			}
		}
	}
	return result
}
