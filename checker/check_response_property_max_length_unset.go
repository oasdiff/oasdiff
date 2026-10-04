package checker

import (
	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/checker/schemawalk"
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
			baseSource, _ := location.SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "maxLength")
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

			propBaseSource, _ := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "maxLength")
			result = append(result, p.newChange(
				ResponsePropertyMaxLengthUnsetId,
				[]any{schemawalk.PropertyFullName(p.propertyPath, p.propertyName), maxLengthDiff.From, info.responseStatus},
				"",
			).WithSources(propBaseSource, nil))
		})
	})

	return result
}
