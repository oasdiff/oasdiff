package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyMaxLengthUnsetId     = "response-body-max-length-unset"
	ResponsePropertyMaxLengthUnsetId = "response-property-max-length-unset"
)

func ResponsePropertyMaxLengthUnsetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if maxLengthDiff := info.schemaDiff.MaxLengthDiff; maxLengthDiff != nil &&
			maxLengthDiff.From != nil && maxLengthDiff.To == nil {
			baseSource, _ := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "maxLength")
			result = append(result, info.newChange(
				ResponseBodyMaxLengthUnsetId,
				[]any{maxLengthDiff.From},
				"",
			).WithSources(baseSource, nil))
		}

		info.walkProperties(func(p propertyInfo) {
			maxLengthDiff := p.propertyDiff.MaxLengthDiff
			if maxLengthDiff == nil || maxLengthDiff.To != nil || maxLengthDiff.From == nil {
				return
			}
			if p.propertyDiff.Revision.WriteOnly {
				return
			}

			propBaseSource, _ := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "maxLength")
			result = append(result, p.newChange(
				ResponsePropertyMaxLengthUnsetId,
				[]any{propertyFullName(p.propertyPath, p.propertyName), maxLengthDiff.From, info.responseStatus},
				"",
			).WithSources(propBaseSource, nil))
		})
	})

	return result
}
