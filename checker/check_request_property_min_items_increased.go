package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMinItemsIncreasedId     = "request-body-min-items-increased"
	RequestPropertyMinItemsIncreasedId = "request-property-min-items-increased"
)

func RequestPropertyMinItemsIncreasedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
				if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.MinItemsDiff != nil {
					minItemsDiff := mediaTypeDiff.SchemaDiff.MinItemsDiff
					if minItemsDiff.From != nil &&
						minItemsDiff.To != nil {
						if IsIncreasedValue(minItemsDiff) {
							result = append(result, NewApiChange(
								RequestBodyMinItemsIncreasedId,
								config,
								[]any{minItemsDiff.To},
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
						minItemsDiff := propertyDiff.MinItemsDiff
						if minItemsDiff == nil {
							return
						}
						if minItemsDiff.From == nil ||
							minItemsDiff.To == nil {
							return
						}
						if propertyDiff.Revision.ReadOnly {
							return
						}
						if !IsIncreasedValue(minItemsDiff) {
							return
						}

						result = append(result, NewApiChange(
							RequestPropertyMinItemsIncreasedId,
							config,
							[]any{propertyFullName(propertyPath, propertyName), minItemsDiff.To},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, revisionSource))
					})
			}
		}
	}
	return result
}
