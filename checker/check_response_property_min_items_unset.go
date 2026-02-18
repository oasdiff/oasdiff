package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyMinItemsUnsetId     = "response-body-min-items-unset"
	ResponsePropertyMinItemsUnsetId = "response-property-min-items-unset"
)

func ResponsePropertyMinItemsUnsetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
			baseSource, revisionSource := operationSources(operationsSources, operationItem.Base, operationItem.Revision)
			for responseStatus, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff == nil ||
					responseDiff.ContentDiff == nil ||
					responseDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}
				modifiedMediaTypes := responseDiff.ContentDiff.MediaTypeModified
				for _, mediaTypeDiff := range modifiedMediaTypes {
					if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.MinItemsDiff != nil {
						minItemsDiff := mediaTypeDiff.SchemaDiff.MinItemsDiff
						if minItemsDiff.From != nil &&
							minItemsDiff.To == nil {
							result = append(result, NewApiChange(
								ResponseBodyMinItemsUnsetId,
								config,
								[]any{minItemsDiff.From},
								"",
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
							if minItemsDiff.To != nil ||
								minItemsDiff.From == nil {
								return
							}
							if propertyDiff.Revision.WriteOnly {
								return
							}

							result = append(result, NewApiChange(
								ResponsePropertyMinItemsUnsetId,
								config,
								[]any{propertyFullName(propertyPath, propertyName), minItemsDiff.From, responseStatus},
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
	}
	return result
}
