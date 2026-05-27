package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyPatternPropertyAddedId       = "request-body-pattern-property-added"
	RequestBodyPatternPropertyRemovedId     = "request-body-pattern-property-removed"
	RequestPropertyPatternPropertyAddedId   = "request-property-pattern-property-added"
	RequestPropertyPatternPropertyRemovedId = "request-property-pattern-property-removed"
)

func RequestPropertyPatternPropertiesUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.PatternPropertiesDiff != nil {
			patPropsDiff := info.schemaDiff.PatternPropertiesDiff
			for _, pattern := range patPropsDiff.Added {
				revisionSource := SchemaMapItemSource(operationsSources, info.operationItem.Revision, patPropsDiff.Revision, pattern)
				result = append(result, info.newChange(RequestBodyPatternPropertyAddedId, []any{pattern}, "").
					WithSources(nil, revisionSource))
			}
			for _, pattern := range patPropsDiff.Deleted {
				baseSource := SchemaMapItemSource(operationsSources, info.operationItem.Base, patPropsDiff.Base, pattern)
				result = append(result, info.newChange(RequestBodyPatternPropertyRemovedId, []any{pattern}, "").
					WithSources(baseSource, nil))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.PatternPropertiesDiff == nil {
				return
			}
			propName := propertyFullName(p.propertyPath, p.propertyName)
			patPropsDiff := p.propertyDiff.PatternPropertiesDiff
			for _, pattern := range patPropsDiff.Added {
				revisionSource := SchemaMapItemSource(operationsSources, info.operationItem.Revision, patPropsDiff.Revision, pattern)
				result = append(result, p.newChange(RequestPropertyPatternPropertyAddedId, []any{pattern, propName}, "").
					WithSources(nil, revisionSource))
			}
			for _, pattern := range patPropsDiff.Deleted {
				baseSource := SchemaMapItemSource(operationsSources, info.operationItem.Base, patPropsDiff.Base, pattern)
				result = append(result, p.newChange(RequestPropertyPatternPropertyRemovedId, []any{pattern, propName}, "").
					WithSources(baseSource, nil))
			}
		})
	})

	return result
}
