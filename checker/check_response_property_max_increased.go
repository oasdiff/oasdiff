package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyMaxIncreasedId              = "response-body-max-increased"
	ResponsePropertyMaxIncreasedId          = "response-property-max-increased"
	ResponseBodyExclusiveMaxIncreasedId     = "response-body-exclusive-max-increased"
	ResponsePropertyExclusiveMaxIncreasedId = "response-property-exclusive-max-increased"
)

func ResponsePropertyMaxIncreasedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if maxDiff := info.schemaDiff.MaxDiff; maxDiff != nil &&
			maxDiff.From != nil && maxDiff.To != nil && IsIncreasedValue(maxDiff) {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "maximum")
			result = append(result, info.newChange(
				ResponseBodyMaxIncreasedId,
				[]any{maxDiff.From, maxDiff.To},
				"",
			).WithSources(baseSource, revisionSource))
		}
		if exMaxDiff := info.schemaDiff.ExclusiveMaxDiff; exMaxDiff != nil &&
			exMaxDiff.From != nil && exMaxDiff.To != nil && IsIncreasedValue(exMaxDiff) {
			exBaseSource, exRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "exclusiveMaximum")
			result = append(result, info.newChange(
				ResponseBodyExclusiveMaxIncreasedId,
				[]any{exMaxDiff.From, exMaxDiff.To},
				"",
			).WithSources(exBaseSource, exRevisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.Revision.WriteOnly {
				return
			}
			propName := propertyFullName(p.propertyPath, p.propertyName)

			if maxDiff := p.propertyDiff.MaxDiff; maxDiff != nil &&
				maxDiff.To != nil && maxDiff.From != nil && IsIncreasedValue(maxDiff) {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "maximum")
				result = append(result, p.newChange(
					ResponsePropertyMaxIncreasedId,
					[]any{propName, maxDiff.From, maxDiff.To, info.responseStatus},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}

			if exMaxDiff := p.propertyDiff.ExclusiveMaxDiff; exMaxDiff != nil &&
				exMaxDiff.To != nil && exMaxDiff.From != nil && IsIncreasedValue(exMaxDiff) {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "exclusiveMaximum")
				result = append(result, p.newChange(
					ResponsePropertyExclusiveMaxIncreasedId,
					[]any{propName, exMaxDiff.From, exMaxDiff.To, info.responseStatus},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
