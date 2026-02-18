package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMinItemsSetId     = "request-body-min-items-set"
	RequestPropertyMinItemsSetId = "request-property-min-items-set"
)

func RequestPropertyMinItemsSetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
					if minItemsDiff.From == nil &&
						minItemsDiff.To != nil {
						result = append(result, NewApiChange(
							RequestBodyMinItemsSetId,
							config,
							[]any{minItemsDiff.To},
							commentId(RequestBodyMinItemsSetId),
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, revisionSource))
					}
				}

				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						minItemsDiff := propertyDiff.MinItemsDiff
						if minItemsDiff == nil {
							return
						}
						if minItemsDiff.From != nil ||
							minItemsDiff.To == nil {
							return
						}
						if propertyDiff.Revision.ReadOnly {
							return
						}

						result = append(result, NewApiChange(
							RequestPropertyMinItemsSetId,
							config,
							[]any{propertyFullName(propertyPath, propertyName), minItemsDiff.To},
							commentId(RequestPropertyMinItemsSetId),
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
