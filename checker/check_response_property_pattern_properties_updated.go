package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyPatternPropertyAddedId       = "response-body-pattern-property-added"
	ResponseBodyPatternPropertyRemovedId     = "response-body-pattern-property-removed"
	ResponsePropertyPatternPropertyAddedId   = "response-property-pattern-property-added"
	ResponsePropertyPatternPropertyRemovedId = "response-property-pattern-property-removed"
)

func ResponsePropertyPatternPropertiesUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.PatternPropertiesDiff != nil {
			patPropsDiff := info.schemaDiff.PatternPropertiesDiff
			for _, pattern := range patPropsDiff.Added {
				revisionSource := schemaMapItemSource(operationsSources, info.operationItem.Revision, patPropsDiff.Revision, pattern)
				result = append(result, info.newChange(ResponseBodyPatternPropertyAddedId, []any{pattern, info.responseStatus}, "").
					WithSources(nil, revisionSource))
			}
			for _, pattern := range patPropsDiff.Deleted {
				baseSource := schemaMapItemSource(operationsSources, info.operationItem.Base, patPropsDiff.Base, pattern)
				result = append(result, info.newChange(ResponseBodyPatternPropertyRemovedId, []any{pattern, info.responseStatus}, "").
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
				revisionSource := schemaMapItemSource(operationsSources, info.operationItem.Revision, patPropsDiff.Revision, pattern)
				result = append(result, p.newChange(ResponsePropertyPatternPropertyAddedId, []any{pattern, propName, info.responseStatus}, "").
					WithSources(nil, revisionSource))
			}
			for _, pattern := range patPropsDiff.Deleted {
				baseSource := schemaMapItemSource(operationsSources, info.operationItem.Base, patPropsDiff.Base, pattern)
				result = append(result, p.newChange(ResponsePropertyPatternPropertyRemovedId, []any{pattern, propName, info.responseStatus}, "").
					WithSources(baseSource, nil))
			}
		})
	})

	return result
}
