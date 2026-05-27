package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyAllOfAddedId       = "request-body-all-of-added"
	RequestBodyAllOfRemovedId     = "request-body-all-of-removed"
	RequestPropertyAllOfAddedId   = "request-property-all-of-added"
	RequestPropertyAllOfRemovedId = "request-property-all-of-removed"
)

func RequestPropertyAllOfUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.AllOfDiff != nil && len(info.schemaDiff.AllOfDiff.Added) > 0 {
			baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "allOf", -1, info.schemaDiff.AllOfDiff.Added[0].Index)
			result = append(result, info.newChange(RequestBodyAllOfAddedId, []any{info.schemaDiff.AllOfDiff.Added.String()}, "").
				WithSources(baseSource, revisionSource))
		}
		if info.schemaDiff.AllOfDiff != nil && len(info.schemaDiff.AllOfDiff.Deleted) > 0 {
			baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "allOf", info.schemaDiff.AllOfDiff.Deleted[0].Index, -1)
			result = append(result, info.newChange(RequestBodyAllOfRemovedId, []any{info.schemaDiff.AllOfDiff.Deleted.String()}, "").
				WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.AllOfDiff == nil {
				return
			}
			propName := propertyFullName(p.propertyPath, p.propertyName)

			if len(p.propertyDiff.AllOfDiff.Added) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "allOf", -1, p.propertyDiff.AllOfDiff.Added[0].Index)
				result = append(result, p.newChange(RequestPropertyAllOfAddedId, []any{p.propertyDiff.AllOfDiff.Added.String(), propName}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
			if len(p.propertyDiff.AllOfDiff.Deleted) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "allOf", p.propertyDiff.AllOfDiff.Deleted[0].Index, -1)
				result = append(result, p.newChange(RequestPropertyAllOfRemovedId, []any{p.propertyDiff.AllOfDiff.Deleted.String(), propName}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
