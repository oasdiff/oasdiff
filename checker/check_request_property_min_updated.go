package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMinIncreasedId                      = "request-body-min-increased"
	RequestBodyMinDecreasedId                      = "request-body-min-decreased"
	RequestPropertyMinIncreasedId                  = "request-property-min-increased"
	RequestReadOnlyPropertyMinIncreasedId          = "request-read-only-property-min-increased"
	RequestPropertyMinDecreasedId                  = "request-property-min-decreased"
	RequestBodyExclusiveMinIncreasedId             = "request-body-exclusive-min-increased"
	RequestBodyExclusiveMinDecreasedId             = "request-body-exclusive-min-decreased"
	RequestPropertyExclusiveMinIncreasedId         = "request-property-exclusive-min-increased"
	RequestReadOnlyPropertyExclusiveMinIncreasedId = "request-read-only-property-exclusive-min-increased"
	RequestPropertyExclusiveMinDecreasedId         = "request-property-exclusive-min-decreased"
)

func RequestPropertyMinIncreasedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if minDiff := info.schemaDiff.MinDiff; minDiff != nil &&
			minDiff.From != nil && minDiff.To != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "minimum")
			if IsIncreasedValue(minDiff) {
				result = append(result, info.newChange(
					RequestBodyMinIncreasedId,
					[]any{minDiff.To},
					"",
				).WithSources(baseSource, revisionSource))
			} else {
				result = append(result, info.newChange(
					RequestBodyMinDecreasedId,
					[]any{minDiff.From, minDiff.To},
					"",
				).WithSources(baseSource, revisionSource))
			}
		}
		if exMinDiff := info.schemaDiff.ExclusiveMinDiff; exMinDiff != nil &&
			exMinDiff.From != nil && exMinDiff.To != nil {
			exBaseSource, exRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "exclusiveMinimum")
			if IsIncreasedValue(exMinDiff) {
				result = append(result, info.newChange(
					RequestBodyExclusiveMinIncreasedId,
					[]any{exMinDiff.To},
					"",
				).WithSources(exBaseSource, exRevisionSource))
			} else {
				result = append(result, info.newChange(
					RequestBodyExclusiveMinDecreasedId,
					[]any{exMinDiff.From, exMinDiff.To},
					"",
				).WithSources(exBaseSource, exRevisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			propName := propertyFullName(p.propertyPath, p.propertyName)

			if minDiff := p.propertyDiff.MinDiff; minDiff != nil &&
				minDiff.From != nil && minDiff.To != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "minimum")
				if IsIncreasedValue(minDiff) {
					id := RequestPropertyMinIncreasedId
					if p.propertyDiff.Revision.ReadOnly {
						id = RequestReadOnlyPropertyMinIncreasedId
					}
					result = append(result, p.newChange(
						id,
						[]any{propName, minDiff.To},
						"",
					).WithSources(propBaseSource, propRevisionSource))
				} else {
					result = append(result, p.newChange(
						RequestPropertyMinDecreasedId,
						[]any{propName, minDiff.From, minDiff.To},
						"",
					).WithSources(propBaseSource, propRevisionSource))
				}
			}

			if exMinDiff := p.propertyDiff.ExclusiveMinDiff; exMinDiff != nil &&
				exMinDiff.From != nil && exMinDiff.To != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "exclusiveMinimum")
				if IsIncreasedValue(exMinDiff) {
					id := RequestPropertyExclusiveMinIncreasedId
					if p.propertyDiff.Revision.ReadOnly {
						id = RequestReadOnlyPropertyExclusiveMinIncreasedId
					}
					result = append(result, p.newChange(
						id,
						[]any{propName, exMinDiff.To},
						"",
					).WithSources(propBaseSource, propRevisionSource))
				} else {
					result = append(result, p.newChange(
						RequestPropertyExclusiveMinDecreasedId,
						[]any{propName, exMinDiff.From, exMinDiff.To},
						"",
					).WithSources(propBaseSource, propRevisionSource))
				}
			}
		})
	})

	return result
}
