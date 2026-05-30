package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyContainsAddedId            = "response-body-contains-added"
	ResponseBodyContainsRemovedId          = "response-body-contains-removed"
	ResponseBodyMinContainsIncreasedId     = "response-body-min-contains-increased"
	ResponseBodyMinContainsDecreasedId     = "response-body-min-contains-decreased"
	ResponseBodyMaxContainsIncreasedId     = "response-body-max-contains-increased"
	ResponseBodyMaxContainsDecreasedId     = "response-body-max-contains-decreased"
	ResponsePropertyContainsAddedId        = "response-property-contains-added"
	ResponsePropertyContainsRemovedId      = "response-property-contains-removed"
	ResponsePropertyMinContainsIncreasedId = "response-property-min-contains-increased"
	ResponsePropertyMinContainsDecreasedId = "response-property-min-contains-decreased"
	ResponsePropertyMaxContainsIncreasedId = "response-property-max-contains-increased"
	ResponsePropertyMaxContainsDecreasedId = "response-property-max-contains-decreased"
)

func ResponsePropertyContainsUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if containsDiff := info.schemaDiff.ContainsDiff; containsDiff != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "contains")
			if containsDiff.SchemaAdded {
				result = append(result, info.newChange(ResponseBodyContainsAddedId, []any{info.responseStatus}, "").
					WithSources(nil, revisionSource))
			}
			if containsDiff.SchemaDeleted {
				result = append(result, info.newChange(ResponseBodyContainsRemovedId, []any{info.responseStatus}, "").
					WithSources(baseSource, nil))
			}
		}

		if d := info.schemaDiff.MinContainsDiff; d != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "minContains")
			if IsIncreasedValue(d) {
				result = append(result, info.newChange(ResponseBodyMinContainsIncreasedId, []any{d.From, d.To, info.responseStatus}, "").
					WithSources(baseSource, revisionSource))
			}
			if IsDecreasedValue(d) {
				result = append(result, info.newChange(ResponseBodyMinContainsDecreasedId, []any{d.From, d.To, info.responseStatus}, "").
					WithSources(baseSource, revisionSource))
			}
		}

		if d := info.schemaDiff.MaxContainsDiff; d != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "maxContains")
			if IsIncreasedValue(d) {
				result = append(result, info.newChange(ResponseBodyMaxContainsIncreasedId, []any{d.From, d.To, info.responseStatus}, "").
					WithSources(baseSource, revisionSource))
			}
			if IsDecreasedValue(d) {
				result = append(result, info.newChange(ResponseBodyMaxContainsDecreasedId, []any{d.From, d.To, info.responseStatus}, "").
					WithSources(baseSource, revisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			propName := propertyFullName(p.propertyPath, p.propertyName)

			if containsDiff := p.propertyDiff.ContainsDiff; containsDiff != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "contains")
				if containsDiff.SchemaAdded {
					result = append(result, p.newChange(ResponsePropertyContainsAddedId, []any{propName, info.responseStatus}, "").
						WithSources(nil, propRevisionSource))
				}
				if containsDiff.SchemaDeleted {
					result = append(result, p.newChange(ResponsePropertyContainsRemovedId, []any{propName, info.responseStatus}, "").
						WithSources(propBaseSource, nil))
				}
			}

			if d := p.propertyDiff.MinContainsDiff; d != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "minContains")
				if IsIncreasedValue(d) {
					result = append(result, p.newChange(ResponsePropertyMinContainsIncreasedId, []any{propName, d.From, d.To, info.responseStatus}, "").
						WithSources(propBaseSource, propRevisionSource))
				}
				if IsDecreasedValue(d) {
					result = append(result, p.newChange(ResponsePropertyMinContainsDecreasedId, []any{propName, d.From, d.To, info.responseStatus}, "").
						WithSources(propBaseSource, propRevisionSource))
				}
			}

			if d := p.propertyDiff.MaxContainsDiff; d != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "maxContains")
				if IsIncreasedValue(d) {
					result = append(result, p.newChange(ResponsePropertyMaxContainsIncreasedId, []any{propName, d.From, d.To, info.responseStatus}, "").
						WithSources(propBaseSource, propRevisionSource))
				}
				if IsDecreasedValue(d) {
					result = append(result, p.newChange(ResponsePropertyMaxContainsDecreasedId, []any{propName, d.From, d.To, info.responseStatus}, "").
						WithSources(propBaseSource, propRevisionSource))
				}
			}
		})
	})

	return result
}
