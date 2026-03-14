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
			for responseStatus, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff == nil ||
					responseDiff.ContentDiff == nil ||
					responseDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}
				modifiedMediaTypes := responseDiff.ContentDiff.MediaTypeModified
				for mediaType, mediaTypeDiff := range modifiedMediaTypes {
					mediaTypeDetails := formatMediaTypeDetails(mediaType, len(modifiedMediaTypes))
					if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.MinItemsDiff != nil {
						minItemsDiff := mediaTypeDiff.SchemaDiff.MinItemsDiff
						if minItemsDiff.From != nil &&
							minItemsDiff.To == nil {
							baseSource, _ := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "minItems")
							result = append(result, NewApiChange(
								ResponseBodyMinItemsUnsetId,
								config,
								[]any{minItemsDiff.From},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, nil).WithDetails(mediaTypeDetails))
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

							propBaseSource, _ := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "minItems")
							result = append(result, NewApiChange(
								ResponsePropertyMinItemsUnsetId,
								config,
								[]any{propertyFullName(propertyPath, propertyName), minItemsDiff.From, responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(propBaseSource, nil).WithDetails(mediaTypeDetails))
						})
				}

			}

		}
	}
	return result
}
