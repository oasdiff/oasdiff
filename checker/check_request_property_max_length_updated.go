package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMaxLengthDecreasedId             = "request-body-max-length-decreased"
	RequestBodyMaxLengthIncreasedId             = "request-body-max-length-increased"
	RequestPropertyMaxLengthDecreasedId         = "request-property-max-length-decreased"
	RequestReadOnlyPropertyMaxLengthDecreasedId = "request-read-only-property-max-length-decreased"
	RequestPropertyMaxLengthIncreasedId         = "request-property-max-length-increased"
)

func RequestPropertyMaxLengthUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "maxLength")
		if maxLengthDiff := info.schemaDiff.MaxLengthDiff; maxLengthDiff != nil &&
			maxLengthDiff.From != nil &&
			maxLengthDiff.To != nil {
			if IsDecreasedValue(maxLengthDiff) {
				result = append(result, info.newChange(
					RequestBodyMaxLengthDecreasedId,
					[]any{maxLengthDiff.To},
					"",
				).WithSources(baseSource, revisionSource))
			} else {
				result = append(result, info.newChange(
					RequestBodyMaxLengthIncreasedId,
					[]any{maxLengthDiff.From, maxLengthDiff.To},
					"",
				).WithSources(baseSource, revisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			maxLengthDiff := p.propertyDiff.MaxLengthDiff
			if maxLengthDiff == nil {
				return
			}
			if maxLengthDiff.From == nil ||
				maxLengthDiff.To == nil {
				return
			}

			propName := propertyFullName(p.propertyPath, p.propertyName)
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "maxLength")

			if IsDecreasedValue(maxLengthDiff) {

				id := RequestPropertyMaxLengthDecreasedId

				if p.propertyDiff.Revision.ReadOnly {
					id = RequestReadOnlyPropertyMaxLengthDecreasedId
				}

				result = append(result, p.newChange(
					id,
					[]any{propName, maxLengthDiff.To},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			} else {
				result = append(result, p.newChange(
					RequestPropertyMaxLengthIncreasedId,
					[]any{propName, maxLengthDiff.From, maxLengthDiff.To},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
