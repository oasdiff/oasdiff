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
		if info.schemaDiff.AllOfDiff != nil {
			added := filterAnnotationOnlySubschemas(info.schemaDiff.AllOfDiff.Added, info.schemaDiff.Revision.AllOf)
			if len(added) > 0 {
				baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "allOf", -1, added[0].Index)
				result = append(result, info.newChange(RequestBodyAllOfAddedId, []any{added.String()}, "").
					WithSources(baseSource, revisionSource))
			}
			deleted := filterAnnotationOnlySubschemas(info.schemaDiff.AllOfDiff.Deleted, info.schemaDiff.Base.AllOf)
			if len(deleted) > 0 {
				baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "allOf", deleted[0].Index, -1)
				result = append(result, info.newChange(RequestBodyAllOfRemovedId, []any{deleted.String()}, "").
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
				result = append(result, p.newChange(RequestPropertyAllOfAddedId, []any{added.String(), propName}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
			deleted := filterAnnotationOnlySubschemas(p.propertyDiff.AllOfDiff.Deleted, p.propertyDiff.Base.AllOf)
			if len(deleted) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "allOf", deleted[0].Index, -1)
				result = append(result, p.newChange(RequestPropertyAllOfRemovedId, []any{deleted.String(), propName}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
