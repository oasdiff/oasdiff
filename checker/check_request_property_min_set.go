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
		_, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "minimum")
		if minDiff := info.schemaDiff.MinDiff; minDiff != nil &&
			minDiff.From == nil &&
			minDiff.To != nil {
			result = append(result, info.newChange(
				RequestBodyMinSetId,
				[]any{minDiff.To},
				commentId(RequestBodyMinSetId),
			).WithSources(nil, revisionSource))
		}
		if exMinDiff := info.schemaDiff.ExclusiveMinDiff; exMinDiff != nil &&
			exMinDiff.From == nil &&
			exMinDiff.To != nil {
			_, exRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "exclusiveMinimum")
			result = append(result, info.newChange(
				RequestBodyExclusiveMinSetId,
				[]any{exMinDiff.To},
				commentId(RequestBodyExclusiveMinSetId),
			).WithSources(nil, exRevisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.Revision.ReadOnly {
				return
			}
			propName := propertyFullName(p.propertyPath, p.propertyName)

			if minDiff := p.propertyDiff.MinDiff; minDiff != nil &&
				minDiff.From == nil &&
				minDiff.To != nil {
				_, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "minimum")
				result = append(result, p.newChange(
					RequestPropertyMinSetId,
					[]any{propName, minDiff.To},
					commentId(RequestPropertyMinSetId),
				).WithSources(nil, propRevisionSource))
			}

			if exMinDiff := p.propertyDiff.ExclusiveMinDiff; exMinDiff != nil &&
				exMinDiff.From == nil &&
				exMinDiff.To != nil {
				_, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "exclusiveMinimum")
				result = append(result, p.newChange(
					RequestPropertyExclusiveMinSetId,
					[]any{propName, exMinDiff.To},
					commentId(RequestPropertyExclusiveMinSetId),
				).WithSources(nil, propRevisionSource))
			}
		})
	})

	return result
}
