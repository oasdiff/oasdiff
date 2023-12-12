package checker

import (
	"github.com/tufin/oasdiff/diff"
	"github.com/tufin/oasdiff/load"
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
			source := (*operationsSources)[operationItem.Revision]
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
							result = append(result, ApiChange{
								Id:          ResponseBodyMinItemsUnsetId,
								Level:       ERR,
								Args:        []any{minItemsDiff.From},
								Operation:   operation,
								OperationId: operationItem.Revision.OperationID,
								Path:        path,
								Source:      load.NewSource(source),
							})
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

							result = append(result, ApiChange{
								Id:          ResponsePropertyMinItemsUnsetId,
								Level:       ERR,
								Args:        []any{propertyFullName(propertyPath, propertyName), minItemsDiff.From, responseStatus},
								Operation:   operation,
								OperationId: operationItem.Revision.OperationID,
								Path:        path,
								Source:      load.NewSource(source),
							})
						})
				}

			}

		}
	}
	return result
}
