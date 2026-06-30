package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMaxDecreasedId                      = "request-body-max-decreased"
	RequestBodyMaxIncreasedId                      = "request-body-max-increased"
	RequestPropertyMaxDecreasedId                  = "request-property-max-decreased"
	RequestReadOnlyPropertyMaxDecreasedId          = "request-read-only-property-max-decreased"
	RequestPropertyMaxIncreasedId                  = "request-property-max-increased"
	RequestBodyExclusiveMaxDecreasedId             = "request-body-exclusive-max-decreased"
	RequestBodyExclusiveMaxIncreasedId             = "request-body-exclusive-max-increased"
	RequestPropertyExclusiveMaxDecreasedId         = "request-property-exclusive-max-decreased"
	RequestReadOnlyPropertyExclusiveMaxDecreasedId = "request-read-only-property-exclusive-max-decreased"
	RequestPropertyExclusiveMaxIncreasedId         = "request-property-exclusive-max-increased"
)

func RequestPropertyMaxDecreasedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if maxDiff := info.schemaDiff.MaxDiff; maxDiff != nil &&
			maxDiff.From != nil && maxDiff.To != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "maximum")
			if isDecreasedValue(maxDiff) {
				result = append(result, info.newChange(
					RequestBodyMaxDecreasedId,
					[]any{maxDiff.To},
					"",
				).WithSources(baseSource, revisionSource))
			} else {
				result = append(result, info.newChange(
					RequestBodyMaxIncreasedId,
					[]any{maxDiff.From, maxDiff.To},
					"",
				).WithSources(baseSource, revisionSource))
			}
		}
		if exMaxDiff := info.schemaDiff.ExclusiveMaxDiff; exMaxDiff != nil &&
			exMaxDiff.From != nil && exMaxDiff.To != nil {
			exBaseSource, exRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "exclusiveMaximum")
			if isDecreasedValue(exMaxDiff) {
				result = append(result, info.newChange(
					RequestBodyExclusiveMaxDecreasedId,
					[]any{exMaxDiff.To},
					"",
				).WithSources(exBaseSource, exRevisionSource))
			} else {
				result = append(result, info.newChange(
					RequestBodyExclusiveMaxIncreasedId,
					[]any{exMaxDiff.From, exMaxDiff.To},
					"",
				).WithSources(exBaseSource, exRevisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			propName := propertyFullName(p.propertyPath, p.propertyName)

			if maxDiff := p.propertyDiff.MaxDiff; maxDiff != nil &&
				maxDiff.From != nil && maxDiff.To != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "maximum")
				if isDecreasedValue(maxDiff) {
					id := RequestPropertyMaxDecreasedId
					if p.propertyDiff.Revision.ReadOnly {
						id = RequestReadOnlyPropertyMaxDecreasedId
					}
					result = append(result, p.newChange(
						id,
						[]any{propName, maxDiff.To},
						"",
					).WithSources(propBaseSource, propRevisionSource))
				} else {
					result = append(result, p.newChange(
						RequestPropertyMaxIncreasedId,
						[]any{propName, maxDiff.From, maxDiff.To},
						"",
					).WithSources(propBaseSource, propRevisionSource))
				}
			}

			if exMaxDiff := p.propertyDiff.ExclusiveMaxDiff; exMaxDiff != nil &&
				exMaxDiff.From != nil && exMaxDiff.To != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "exclusiveMaximum")
				if isDecreasedValue(exMaxDiff) {
					id := RequestPropertyExclusiveMaxDecreasedId
					if p.propertyDiff.Revision.ReadOnly {
						id = RequestReadOnlyPropertyExclusiveMaxDecreasedId
					}
					result = append(result, p.newChange(
						id,
						[]any{propName, exMaxDiff.To},
						"",
					).WithSources(propBaseSource, propRevisionSource))
				} else {
					result = append(result, p.newChange(
						RequestPropertyExclusiveMaxIncreasedId,
						[]any{propName, exMaxDiff.From, exMaxDiff.To},
						"",
					).WithSources(propBaseSource, propRevisionSource))
				}
			}
		})
	})

	return result
}
