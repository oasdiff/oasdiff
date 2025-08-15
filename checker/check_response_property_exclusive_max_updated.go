package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyExclusiveMaxIncreasedId = "response-property-exclusive-max-increased"
	ResponsePropertyExclusiveMaxDecreasedId = "response-property-exclusive-max-decreased"
)

func ResponsePropertyExclusiveMaxUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
							exclusiveMaxDiff := propertyDiff.ExclusiveMaxDiff
							if exclusiveMaxDiff == nil {
								return
							}

							if propertyDiff.Revision.WriteOnly {
								return
							}

							from, fromOK := exclusiveMaxDiff.From.(bool)
							to, toOK := exclusiveMaxDiff.To.(bool)

							if !fromOK && !toOK {
								return
							}

							isIncreased := toOK && to
							isDecreased := fromOK && from && (!toOK || !to)

							if !isIncreased && !isDecreased {
								return
							}

							id := ResponsePropertyExclusiveMaxDecreasedId
							if isIncreased {
								id = ResponsePropertyExclusiveMaxIncreasedId
							}

							result = append(result, NewApiChange(
								id,
								config,
								[]any{propertyFullName(propertyPath, propertyName), exclusiveMaxDiff.From, exclusiveMaxDiff.To, responseStatus},
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
