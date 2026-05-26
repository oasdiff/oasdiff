package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyMinLengthDecreasedId     = "response-body-min-length-decreased"
	ResponsePropertyMinLengthDecreasedId = "response-property-min-length-decreased"
)

func ResponsePropertyMinLengthDecreasedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if minLengthDiff := info.schemaDiff.MinLengthDiff; minLengthDiff != nil &&
			minLengthDiff.From != nil && minLengthDiff.To != nil && IsDecreasedValue(minLengthDiff) {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "minLength")
			result = append(result, info.newChange(
				ResponseBodyMinLengthDecreasedId,
				[]any{minLengthDiff.From, minLengthDiff.To},
				"",
			).WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			minLengthDiff := p.propertyDiff.MinLengthDiff
			if minLengthDiff == nil || minLengthDiff.To == nil || minLengthDiff.From == nil {
				return
			}
			if !IsDecreasedValue(minLengthDiff) {
				return
			}
			if p.propertyDiff.Revision.WriteOnly {
				return
			}

			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "minLength")
			result = append(result, p.newChange(
				ResponsePropertyMinLengthDecreasedId,
				[]any{propertyFullName(p.propertyPath, p.propertyName), minLengthDiff.From, minLengthDiff.To, info.responseStatus},
				"",
			).WithSources(propBaseSource, propRevisionSource))
		})
	})

	return result
}
