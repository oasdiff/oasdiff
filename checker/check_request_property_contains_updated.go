package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyContainsAddedId            = "request-body-contains-added"
	RequestBodyContainsRemovedId          = "request-body-contains-removed"
	RequestBodyMinContainsIncreasedId     = "request-body-min-contains-increased"
	RequestBodyMinContainsDecreasedId     = "request-body-min-contains-decreased"
	RequestBodyMaxContainsIncreasedId     = "request-body-max-contains-increased"
	RequestBodyMaxContainsDecreasedId     = "request-body-max-contains-decreased"
	RequestPropertyContainsAddedId        = "request-property-contains-added"
	RequestPropertyContainsRemovedId      = "request-property-contains-removed"
	RequestPropertyMinContainsIncreasedId = "request-property-min-contains-increased"
	RequestPropertyMinContainsDecreasedId = "request-property-min-contains-decreased"
	RequestPropertyMaxContainsIncreasedId = "request-property-max-contains-increased"
	RequestPropertyMaxContainsDecreasedId = "request-property-max-contains-decreased"
)

func RequestPropertyContainsUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if containsDiff := info.schemaDiff.ContainsDiff; containsDiff != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "contains")
			if containsDiff.SchemaAdded {
				result = append(result, info.newChange(RequestBodyContainsAddedId, nil, "").
					WithSources(nil, revisionSource))
			}
			if containsDiff.SchemaDeleted {
				result = append(result, info.newChange(RequestBodyContainsRemovedId, nil, "").
					WithSources(baseSource, nil))
			}
		}

		if d := info.schemaDiff.MinContainsDiff; d != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "minContains")
			if isIncreasedValue(d) {
				result = append(result, info.newChange(RequestBodyMinContainsIncreasedId, []any{d.From, d.To}, "").
					WithSources(baseSource, revisionSource))
			}
			if isDecreasedValue(d) {
				result = append(result, info.newChange(RequestBodyMinContainsDecreasedId, []any{d.From, d.To}, "").
					WithSources(baseSource, revisionSource))
			}
		}

		if d := info.schemaDiff.MaxContainsDiff; d != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "maxContains")
			if isIncreasedValue(d) {
				result = append(result, info.newChange(RequestBodyMaxContainsIncreasedId, []any{d.From, d.To}, "").
					WithSources(baseSource, revisionSource))
			}
			if isDecreasedValue(d) {
				result = append(result, info.newChange(RequestBodyMaxContainsDecreasedId, []any{d.From, d.To}, "").
					WithSources(baseSource, revisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			propName := propertyFullName(p.propertyPath, p.propertyName)

			if containsDiff := p.propertyDiff.ContainsDiff; containsDiff != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "contains")
				if containsDiff.SchemaAdded {
					result = append(result, p.newChange(RequestPropertyContainsAddedId, []any{propName}, "").
						WithSources(nil, propRevisionSource))
				}
				if containsDiff.SchemaDeleted {
					result = append(result, p.newChange(RequestPropertyContainsRemovedId, []any{propName}, "").
						WithSources(propBaseSource, nil))
				}
			}

			if d := p.propertyDiff.MinContainsDiff; d != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "minContains")
				if isIncreasedValue(d) {
					result = append(result, p.newChange(RequestPropertyMinContainsIncreasedId, []any{propName, d.From, d.To}, "").
						WithSources(propBaseSource, propRevisionSource))
				}
				if isDecreasedValue(d) {
					result = append(result, p.newChange(RequestPropertyMinContainsDecreasedId, []any{propName, d.From, d.To}, "").
						WithSources(propBaseSource, propRevisionSource))
				}
			}

			if d := p.propertyDiff.MaxContainsDiff; d != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "maxContains")
				if isIncreasedValue(d) {
					result = append(result, p.newChange(RequestPropertyMaxContainsIncreasedId, []any{propName, d.From, d.To}, "").
						WithSources(propBaseSource, propRevisionSource))
				}
				if isDecreasedValue(d) {
					result = append(result, p.newChange(RequestPropertyMaxContainsDecreasedId, []any{propName, d.From, d.To}, "").
						WithSources(propBaseSource, propRevisionSource))
				}
			}
		})
	})

	return result
}
