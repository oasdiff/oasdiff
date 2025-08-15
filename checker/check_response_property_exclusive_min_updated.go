package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyExclusiveMinIncreasedId = "response-property-exclusive-min-increased"
	ResponsePropertyExclusiveMinDecreasedId = "response-property-exclusive-min-decreased"
)

func ResponsePropertyExclusiveMinUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
				for _, mediaTypeDiff := range modifiedMediaTypes {
					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							exclusiveMinDiff := propertyDiff.ExclusiveMinDiff
							if exclusiveMinDiff == nil {
								return
							}

							if propertyDiff.Revision.WriteOnly {
								return
							}

							from, fromOK := exclusiveMinDiff.From.(bool)
							to, toOK := exclusiveMinDiff.To.(bool)

							if !fromOK && !toOK {
								return
							}

							isIncreased := toOK && to
							isDecreased := fromOK && from && (!toOK || !to)

							if !isIncreased && !isDecreased {
								return
							}

							id := ResponsePropertyExclusiveMinDecreasedId
							if isIncreased {
								id = ResponsePropertyExclusiveMinIncreasedId
							}

							result = append(result, NewApiChange(
								id,
								config,
								[]any{propertyFullName(propertyPath, propertyName), exclusiveMinDiff.From, exclusiveMinDiff.To, responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							))
						})
				}
			}
		}
	}
	return result
}
