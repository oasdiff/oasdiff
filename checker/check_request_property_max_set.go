package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMaxSetId              = "request-body-max-set"
	RequestPropertyMaxSetId          = "request-property-max-set"
	RequestBodyExclusiveMaxSetId     = "request-body-exclusive-max-set"
	RequestPropertyExclusiveMaxSetId = "request-property-exclusive-max-set"
)

func RequestPropertyMaxSetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if value, ok := maximumBound.WasSet(info.schemaDiff); ok {
			_, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "maximum")
			result = append(result, info.newChange(
				RequestBodyMaxSetId,
				[]any{value},
				boundSetComment,
			).WithSources(nil, revisionSource))
		}
		if value, ok := exclusiveMaximumBound.WasSet(info.schemaDiff); ok {
			_, exRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "exclusiveMaximum")
			result = append(result, info.newChange(
				RequestBodyExclusiveMaxSetId,
				[]any{value},
				boundSetComment,
			).WithSources(nil, exRevisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			propName := propertyFullName(p.propertyPath, p.propertyName)

			if value, ok := maximumBound.WasSet(p.propertyDiff); ok {
				_, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "maximum")
				result = append(result, p.newChange(
					RequestPropertyMaxSetId,
					[]any{propName, value},
					boundSetComment,
				).WithSources(nil, propRevisionSource))
			}

			if value, ok := exclusiveMaximumBound.WasSet(p.propertyDiff); ok {
				_, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "exclusiveMaximum")
				result = append(result, p.newChange(
					RequestPropertyExclusiveMaxSetId,
					[]any{propName, value},
					boundSetComment,
				).WithSources(nil, propRevisionSource))
			}
		})
	})

	return result
}
