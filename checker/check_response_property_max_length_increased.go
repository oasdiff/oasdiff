package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyMaxLengthIncreasedId     = "response-body-max-length-increased"
	ResponsePropertyMaxLengthIncreasedId = "response-property-max-length-increased"
)

func ResponsePropertyMaxLengthIncreasedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if maxLengthDiff := info.schemaDiff.MaxLengthDiff; maxLengthDiff != nil &&
			maxLengthDiff.From != nil &&
			maxLengthDiff.To != nil &&
			isIncreasedValue(maxLengthDiff) {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "maxLength")
			result = append(result, info.newChange(
				ResponseBodyMaxLengthIncreasedId,
				[]any{maxLengthDiff.From, maxLengthDiff.To},
				"",
			).WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			maxLengthDiff := p.propertyDiff.MaxLengthDiff
			if maxLengthDiff == nil {
				return
			}
			if maxLengthDiff.To == nil ||
				maxLengthDiff.From == nil {
				return
			}
			if !isIncreasedValue(maxLengthDiff) {
				return
			}

			if p.propertyDiff.Revision.WriteOnly {
				return
			}

			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "maxLength")
			result = append(result, p.newChange(
				ResponsePropertyMaxLengthIncreasedId,
				[]any{propertyFullName(p.propertyPath, p.propertyName), maxLengthDiff.From, maxLengthDiff.To, info.responseStatus},
				"",
			).WithSources(propBaseSource, propRevisionSource))
		})
	})

	return result
}
