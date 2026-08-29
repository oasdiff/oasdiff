package checker

import (
	"math"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMultipleOfSetId             = "request-body-multiple-of-set"
	RequestPropertyMultipleOfSetId         = "request-property-multiple-of-set"
	RequestBodyMultipleOfUnsetId           = "request-body-multiple-of-unset"
	RequestPropertyMultipleOfUnsetId       = "request-property-multiple-of-unset"
	RequestBodyMultipleOfChangedId         = "request-body-multiple-of-changed"
	RequestPropertyMultipleOfChangedId     = "request-property-multiple-of-changed"
	RequestBodyMultipleOfGeneralizedId     = "request-body-multiple-of-generalized"
	RequestPropertyMultipleOfGeneralizedId = "request-property-multiple-of-generalized"
)

// isIntegerMultiple reports whether a is an integer multiple of b, in which
// case every multiple of a is also a multiple of b.
func isIntegerMultiple(a, b float64) bool {
	if b == 0 {
		return false
	}
	ratio := a / b
	return math.Abs(ratio-math.Round(ratio)) <= 1e-9*math.Max(math.Abs(ratio), 1)
}

func RequestPropertyMultipleOfUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if multipleOfDiff := info.schemaDiff.MultipleOfDiff; multipleOfDiff != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "multipleOf")
			switch {
			case multipleOfDiff.From == nil:
				result = append(result, info.newChange(
					RequestBodyMultipleOfSetId,
					[]any{multipleOfDiff.To},
					commentId(RequestBodyMultipleOfSetId),
				).WithSources(nil, revisionSource))
			case multipleOfDiff.To == nil:
				result = append(result, info.newChange(
					RequestBodyMultipleOfUnsetId,
					[]any{multipleOfDiff.From},
					"",
				).WithSources(baseSource, revisionSource))
			case isIntegerMultiple(multipleOfDiff.From.(float64), multipleOfDiff.To.(float64)):
				result = append(result, info.newChange(
					RequestBodyMultipleOfGeneralizedId,
					[]any{multipleOfDiff.From, multipleOfDiff.To},
					commentId(RequestBodyMultipleOfGeneralizedId),
				).WithSources(baseSource, revisionSource))
			default:
				result = append(result, info.newChange(
					RequestBodyMultipleOfChangedId,
					[]any{multipleOfDiff.From, multipleOfDiff.To},
					"",
				).WithSources(baseSource, revisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			multipleOfDiff := p.propertyDiff.MultipleOfDiff
			if multipleOfDiff == nil {
				return
			}
			// narrowing a read-only property does not affect requests
			if p.propertyDiff.Revision.ReadOnly && multipleOfDiff.To != nil {
				return
			}

			propName := propertyFullName(p.propertyPath, p.propertyName)
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "multipleOf")
			switch {
			case multipleOfDiff.From == nil:
				result = append(result, p.newChange(
					RequestPropertyMultipleOfSetId,
					[]any{propName, multipleOfDiff.To},
					commentId(RequestPropertyMultipleOfSetId),
				).WithSources(nil, propRevisionSource))
			case multipleOfDiff.To == nil:
				result = append(result, p.newChange(
					RequestPropertyMultipleOfUnsetId,
					[]any{propName, multipleOfDiff.From},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			case isIntegerMultiple(multipleOfDiff.From.(float64), multipleOfDiff.To.(float64)):
				result = append(result, p.newChange(
					RequestPropertyMultipleOfGeneralizedId,
					[]any{propName, multipleOfDiff.From, multipleOfDiff.To},
					commentId(RequestPropertyMultipleOfGeneralizedId),
				).WithSources(propBaseSource, propRevisionSource))
			default:
				result = append(result, p.newChange(
					RequestPropertyMultipleOfChangedId,
					[]any{propName, multipleOfDiff.From, multipleOfDiff.To},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
