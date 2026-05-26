package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMaxLengthSetId     = "request-body-max-length-set"
	RequestPropertyMaxLengthSetId = "request-property-max-length-set"
)

func RequestPropertyMaxLengthSetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if maxLengthDiff := info.schemaDiff.MaxLengthDiff; maxLengthDiff != nil &&
			maxLengthDiff.From == nil && maxLengthDiff.To != nil {
			_, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "maxLength")
			result = append(result, info.newChange(
				RequestBodyMaxLengthSetId,
				[]any{maxLengthDiff.To},
				commentId(RequestBodyMaxLengthSetId),
			).WithSources(nil, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			maxLengthDiff := p.propertyDiff.MaxLengthDiff
			if maxLengthDiff == nil || maxLengthDiff.From != nil || maxLengthDiff.To == nil {
				return
			}
			if p.propertyDiff.Revision.ReadOnly {
				return
			}

			_, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "maxLength")
			result = append(result, p.newChange(
				RequestPropertyMaxLengthSetId,
				[]any{propertyFullName(p.propertyPath, p.propertyName), maxLengthDiff.To},
				commentId(RequestPropertyMaxLengthSetId),
			).WithSources(nil, propRevisionSource))
		})
	})

	return result
}
