package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyAnyOfAddedId       = "response-body-any-of-added"
	ResponseBodyAnyOfRemovedId     = "response-body-any-of-removed"
	ResponsePropertyAnyOfAddedId   = "response-property-any-of-added"
	ResponsePropertyAnyOfRemovedId = "response-property-any-of-removed"
)

func ResponsePropertyAnyOfUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.AnyOfDiff != nil {
			if added := info.schemaDiff.AnyOfDiff.Added; len(added) > 0 {
				baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "anyOf", -1, added[0].Index)
				result = append(result, info.newChange(ResponseBodyAnyOfAddedId, []any{added.String(), info.responseStatus}, "").
					WithSources(baseSource, revisionSource))
			}
			if deleted := info.schemaDiff.AnyOfDiff.Deleted; len(deleted) > 0 {
				baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "anyOf", deleted[0].Index, -1)
				result = append(result, info.newChange(ResponseBodyAnyOfRemovedId, []any{deleted.String(), info.responseStatus}, "").
					WithSources(baseSource, revisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.AnyOfDiff == nil {
				return
			}
			propName := propertyFullName(p.propertyPath, p.propertyName)

			if added := p.propertyDiff.AnyOfDiff.Added; len(added) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "anyOf", -1, added[0].Index)
				result = append(result, p.newChange(ResponsePropertyAnyOfAddedId, []any{added.String(), propName, info.responseStatus}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
			if deleted := p.propertyDiff.AnyOfDiff.Deleted; len(deleted) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "anyOf", deleted[0].Index, -1)
				result = append(result, p.newChange(ResponsePropertyAnyOfRemovedId, []any{deleted.String(), propName, info.responseStatus}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
