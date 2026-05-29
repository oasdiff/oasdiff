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
		_, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "maximum")
		if maxDiff := info.schemaDiff.MaxDiff; maxDiff != nil &&
			maxDiff.From == nil &&
			maxDiff.To != nil {
			result = append(result, info.newChange(
				RequestBodyMaxSetId,
				[]any{maxDiff.To},
				commentId(RequestBodyMaxSetId),
			).WithSources(nil, revisionSource))
		}
		if exMaxDiff := info.schemaDiff.ExclusiveMaxDiff; exMaxDiff != nil &&
			exMaxDiff.From == nil &&
			exMaxDiff.To != nil {
			_, exRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "exclusiveMaximum")
			result = append(result, info.newChange(
				RequestBodyExclusiveMaxSetId,
				[]any{exMaxDiff.To},
				commentId(RequestBodyExclusiveMaxSetId),
			).WithSources(nil, exRevisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.Revision.ReadOnly {
				return
			}
			propName := propertyFullName(p.propertyPath, p.propertyName)

			if maxDiff := p.propertyDiff.MaxDiff; maxDiff != nil &&
				maxDiff.From == nil &&
				maxDiff.To != nil {
				_, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "maximum")
				result = append(result, p.newChange(
					RequestPropertyMaxSetId,
					[]any{propName, maxDiff.To},
					commentId(RequestPropertyMaxSetId),
				).WithSources(nil, propRevisionSource))
			}

			if exMaxDiff := p.propertyDiff.ExclusiveMaxDiff; exMaxDiff != nil &&
				exMaxDiff.From == nil &&
				exMaxDiff.To != nil {
				_, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "exclusiveMaximum")
				result = append(result, p.newChange(
					RequestPropertyExclusiveMaxSetId,
					[]any{propName, exMaxDiff.To},
					commentId(RequestPropertyExclusiveMaxSetId),
				).WithSources(nil, propRevisionSource))
			}
		})
	})

	return result
}
