package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyAllOfAddedId       = "response-body-all-of-added"
	ResponseBodyAllOfRemovedId     = "response-body-all-of-removed"
	ResponsePropertyAllOfAddedId   = "response-property-all-of-added"
	ResponsePropertyAllOfRemovedId = "response-property-all-of-removed"
)

func ResponsePropertyAllOfUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.AllOfDiff != nil && len(info.schemaDiff.AllOfDiff.Added) > 0 {
			baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "allOf", -1, info.schemaDiff.AllOfDiff.Added[0].Index)
			result = append(result, info.newChange(ResponseBodyAllOfAddedId, []any{info.schemaDiff.AllOfDiff.Added.String(), info.responseStatus}, "").
				WithSources(baseSource, revisionSource))
		}
		if info.schemaDiff.AllOfDiff != nil && len(info.schemaDiff.AllOfDiff.Deleted) > 0 {
			baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "allOf", info.schemaDiff.AllOfDiff.Deleted[0].Index, -1)
			result = append(result, info.newChange(ResponseBodyAllOfRemovedId, []any{info.schemaDiff.AllOfDiff.Deleted.String(), info.responseStatus}, "").
				WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.AllOfDiff == nil {
				return
			}
			propName := propertyFullName(p.propertyPath, p.propertyName)

			if len(p.propertyDiff.AllOfDiff.Added) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "allOf", -1, p.propertyDiff.AllOfDiff.Added[0].Index)
				result = append(result, p.newChange(ResponsePropertyAllOfAddedId, []any{p.propertyDiff.AllOfDiff.Added.String(), propName, info.responseStatus}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
			if len(p.propertyDiff.AllOfDiff.Deleted) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "allOf", p.propertyDiff.AllOfDiff.Deleted[0].Index, -1)
				result = append(result, p.newChange(ResponsePropertyAllOfRemovedId, []any{p.propertyDiff.AllOfDiff.Deleted.String(), propName, info.responseStatus}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
