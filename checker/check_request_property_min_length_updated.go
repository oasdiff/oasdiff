package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMinLengthIncreasedId     = "request-body-min-length-increased"
	RequestBodyMinLengthDecreasedId     = "request-body-min-length-decreased"
	RequestPropertyMinLengthIncreasedId = "request-property-min-length-increased"
	RequestPropertyMinLengthDecreasedId = "request-property-min-length-decreased"
)

func RequestPropertyMinLengthUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if minLengthDiff := info.schemaDiff.MinLengthDiff; minLengthDiff != nil &&
			minLengthDiff.From != nil && minLengthDiff.To != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "minLength")
			id := RequestBodyMinLengthDecreasedId
			if isIncreasedValue(minLengthDiff) {
				id = RequestBodyMinLengthIncreasedId
			}
			result = append(result, info.newChange(
				id,
				[]any{minLengthDiff.From, minLengthDiff.To},
				"",
			).WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			minLengthDiff := p.propertyDiff.MinLengthDiff
			if minLengthDiff == nil || minLengthDiff.From == nil || minLengthDiff.To == nil {
				return
			}

			propName := propertyFullName(p.propertyPath, p.propertyName)
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "minLength")
			id := RequestPropertyMinLengthIncreasedId
			if isDecreasedValue(minLengthDiff) {
				id = RequestPropertyMinLengthDecreasedId
			}
			result = append(result, p.newChange(
				id,
				[]any{propName, minLengthDiff.From, minLengthDiff.To},
				"",
			).WithSources(propBaseSource, propRevisionSource))
		})
	})

	return result
}
