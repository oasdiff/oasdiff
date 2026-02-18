package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMinIncreasedId             = "request-body-min-increased"
	RequestBodyMinDecreasedId             = "request-body-min-decreased"
	RequestPropertyMinIncreasedId         = "request-property-min-increased"
	RequestReadOnlyPropertyMinIncreasedId = "request-read-only-property-min-increased"
	RequestPropertyMinDecreasedId         = "request-property-min-decreased"
)

func RequestPropertyMinIncreasedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
				if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.MinDiff != nil {
					minDiff := mediaTypeDiff.SchemaDiff.MinDiff
					if minDiff.From != nil &&
						minDiff.To != nil {
						if IsIncreasedValue(minDiff) {
							result = append(result, NewApiChange(
								RequestBodyMinIncreasedId,
								config,
								[]any{minDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource))
						} else {
							result = append(result, NewApiChange(
								RequestBodyMinDecreasedId,
								config,
								[]any{minDiff.From, minDiff.To},
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
						minDiff := propertyDiff.MinDiff
						if minDiff == nil {
							return
						}
						if minDiff.From == nil ||
							minDiff.To == nil {
							return
						}

						propName := propertyFullName(propertyPath, propertyName)

						if IsIncreasedValue(minDiff) {

							id := RequestPropertyMinIncreasedId

							if propertyDiff.Revision.ReadOnly {
								id = RequestReadOnlyPropertyMinIncreasedId
							}

							result = append(result, NewApiChange(
								id,
								config,
								[]any{propName, minDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource))
						} else {
							result = append(result, NewApiChange(
								RequestPropertyMinDecreasedId,
								config,
								[]any{propName, minDiff.From, minDiff.To},
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
