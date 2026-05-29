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
		if info.schemaDiff.AllOfDiff != nil {
			added := filterAnnotationOnlySubschemas(info.schemaDiff.AllOfDiff.Added, info.schemaDiff.Revision.AllOf)
			if len(added) > 0 {
				baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "allOf", -1, added[0].Index)
				result = append(result, info.newChange(ResponseBodyAllOfAddedId, []any{added.String(), info.responseStatus}, "").
					WithSources(baseSource, revisionSource))
			}
			deleted := filterAnnotationOnlySubschemas(info.schemaDiff.AllOfDiff.Deleted, info.schemaDiff.Base.AllOf)
			if len(deleted) > 0 {
				baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "allOf", deleted[0].Index, -1)
				result = append(result, info.newChange(ResponseBodyAllOfRemovedId, []any{deleted.String(), info.responseStatus}, "").
					WithSources(baseSource, revisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.AllOfDiff == nil {
				return
			}
			propName := propertyFullName(p.propertyPath, p.propertyName)

			added := filterAnnotationOnlySubschemas(p.propertyDiff.AllOfDiff.Added, p.propertyDiff.Revision.AllOf)
			if len(added) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "allOf", -1, added[0].Index)
				result = append(result, p.newChange(ResponsePropertyAllOfAddedId, []any{added.String(), propName, info.responseStatus}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
			deleted := filterAnnotationOnlySubschemas(p.propertyDiff.AllOfDiff.Deleted, p.propertyDiff.Base.AllOf)
			if len(deleted) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "allOf", deleted[0].Index, -1)
				result = append(result, p.newChange(ResponsePropertyAllOfRemovedId, []any{deleted.String(), propName, info.responseStatus}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
