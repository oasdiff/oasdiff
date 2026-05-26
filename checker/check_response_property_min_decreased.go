package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyMinDecreasedId              = "response-body-min-decreased"
	ResponsePropertyMinDecreasedId          = "response-property-min-decreased"
	ResponseBodyExclusiveMinDecreasedId     = "response-body-exclusive-min-decreased"
	ResponsePropertyExclusiveMinDecreasedId = "response-property-exclusive-min-decreased"
)

func ResponsePropertyMinDecreasedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if minDiff := info.schemaDiff.MinDiff; minDiff != nil &&
			minDiff.From != nil && minDiff.To != nil && IsDecreasedValue(minDiff) {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "minimum")
			result = append(result, info.newChange(
				ResponseBodyMinDecreasedId,
				[]any{minDiff.From, minDiff.To},
				"",
			).WithSources(baseSource, revisionSource))
		}
		if exMinDiff := info.schemaDiff.ExclusiveMinDiff; exMinDiff != nil &&
			exMinDiff.From != nil && exMinDiff.To != nil && IsDecreasedValue(exMinDiff) {
			exBaseSource, exRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "exclusiveMinimum")
			result = append(result, info.newChange(
				ResponseBodyExclusiveMinDecreasedId,
				[]any{exMinDiff.From, exMinDiff.To},
				"",
			).WithSources(exBaseSource, exRevisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.Revision.WriteOnly {
				return
			}
			propName := propertyFullName(p.propertyPath, p.propertyName)

			if minDiff := p.propertyDiff.MinDiff; minDiff != nil &&
				minDiff.To != nil && minDiff.From != nil && IsDecreasedValue(minDiff) {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "minimum")
				result = append(result, p.newChange(
					ResponsePropertyMinDecreasedId,
					[]any{propName, minDiff.From, minDiff.To, info.responseStatus},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}

			if exMinDiff := p.propertyDiff.ExclusiveMinDiff; exMinDiff != nil &&
				exMinDiff.To != nil && exMinDiff.From != nil && IsDecreasedValue(exMinDiff) {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "exclusiveMinimum")
				result = append(result, p.newChange(
					ResponsePropertyExclusiveMinDecreasedId,
					[]any{propName, exMinDiff.From, exMinDiff.To, info.responseStatus},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
