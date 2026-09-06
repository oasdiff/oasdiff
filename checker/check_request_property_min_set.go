package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMinSetId              = "request-body-min-set"
	RequestPropertyMinSetId          = "request-property-min-set"
	RequestBodyExclusiveMinSetId     = "request-body-exclusive-min-set"
	RequestPropertyExclusiveMinSetId = "request-property-exclusive-min-set"
)

func RequestPropertyMinSetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if value, ok := minimumBound.WasSet(info.schemaDiff); ok {
			_, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "minimum")
			result = append(result, info.newChange(
				RequestBodyMinSetId,
				[]any{value},
				boundSetComment,
			).WithSources(nil, revisionSource))
		}
		if value, ok := exclusiveMinimumBound.WasSet(info.schemaDiff); ok {
			_, exRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "exclusiveMinimum")
			result = append(result, info.newChange(
				RequestBodyExclusiveMinSetId,
				[]any{value},
				boundSetComment,
			).WithSources(nil, exRevisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			propName := propertyFullName(p.propertyPath, p.propertyName)

			if value, ok := minimumBound.WasSet(p.propertyDiff); ok {
				_, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "minimum")
				result = append(result, p.newChange(
					RequestPropertyMinSetId,
					[]any{propName, value},
					boundSetComment,
				).WithSources(nil, propRevisionSource))
			}

			if value, ok := exclusiveMinimumBound.WasSet(p.propertyDiff); ok {
				_, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "exclusiveMinimum")
				result = append(result, p.newChange(
					RequestPropertyExclusiveMinSetId,
					[]any{propName, value},
					boundSetComment,
				).WithSources(nil, propRevisionSource))
			}
		})
	})

	return result
}
