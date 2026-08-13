package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyAllOfAddedId                     = "response-body-all-of-added"
	ResponseBodyAllOfRemovedId                   = "response-body-all-of-removed"
	ResponsePropertyAllOfAddedId                 = "response-property-all-of-added"
	ResponsePropertyAllOfRemovedId               = "response-property-all-of-removed"
	ResponseBodyAllOfAddedAnnotationOnlyId       = "response-body-all-of-added-annotation-only"
	ResponseBodyAllOfRemovedAnnotationOnlyId     = "response-body-all-of-removed-annotation-only"
	ResponsePropertyAllOfAddedAnnotationOnlyId   = "response-property-all-of-added-annotation-only"
	ResponsePropertyAllOfRemovedAnnotationOnlyId = "response-property-all-of-removed-annotation-only"
)

func ResponsePropertyAllOfUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.AllOfDiff != nil {
			added, annotationOnlyAdded := splitSubschemasByAnnotationOnly(info.schemaDiff.AllOfDiff.Added, info.schemaDiff.Revision.AllOf)
			if len(added) > 0 {
				baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "allOf", -1, added[0].Index)
				result = append(result, info.newChange(ResponseBodyAllOfAddedId, []any{added.String(), info.responseStatus}, "").
					WithSources(baseSource, revisionSource))
			}
			if len(annotationOnlyAdded) > 0 {
				baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "allOf", -1, annotationOnlyAdded[0].Index)
				result = append(result, info.newChange(ResponseBodyAllOfAddedAnnotationOnlyId, []any{annotationOnlyAdded.String(), info.responseStatus}, "").
					WithSources(baseSource, revisionSource))
			}

			deleted, annotationOnlyDeleted := splitSubschemasByAnnotationOnly(info.schemaDiff.AllOfDiff.Deleted, info.schemaDiff.Base.AllOf)
			if len(deleted) > 0 {
				baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "allOf", deleted[0].Index, -1)
				result = append(result, info.newChange(ResponseBodyAllOfRemovedId, []any{deleted.String(), info.responseStatus}, "").
					WithSources(baseSource, revisionSource))
			}
			if len(annotationOnlyDeleted) > 0 {
				baseSource, revisionSource := SubschemaSources(operationsSources, info.operationItem, info.schemaDiff, "allOf", annotationOnlyDeleted[0].Index, -1)
				result = append(result, info.newChange(ResponseBodyAllOfRemovedAnnotationOnlyId, []any{annotationOnlyDeleted.String(), info.responseStatus}, "").
					WithSources(baseSource, revisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.AllOfDiff == nil {
				return
			}
			propName := propertyFullName(p.propertyPath, p.propertyName)

			added, annotationOnlyAdded := splitSubschemasByAnnotationOnly(p.propertyDiff.AllOfDiff.Added, p.propertyDiff.Revision.AllOf)
			if len(added) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "allOf", -1, added[0].Index)
				result = append(result, p.newChange(ResponsePropertyAllOfAddedId, []any{added.String(), propName, info.responseStatus}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
			if len(annotationOnlyAdded) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "allOf", -1, annotationOnlyAdded[0].Index)
				result = append(result, p.newChange(ResponsePropertyAllOfAddedAnnotationOnlyId, []any{annotationOnlyAdded.String(), propName, info.responseStatus}, "").
					WithSources(propBaseSource, propRevisionSource))
			}

			deleted, annotationOnlyDeleted := splitSubschemasByAnnotationOnly(p.propertyDiff.AllOfDiff.Deleted, p.propertyDiff.Base.AllOf)
			if len(deleted) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "allOf", deleted[0].Index, -1)
				result = append(result, p.newChange(ResponsePropertyAllOfRemovedId, []any{deleted.String(), propName, info.responseStatus}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
			if len(annotationOnlyDeleted) > 0 {
				propBaseSource, propRevisionSource := SubschemaSources(operationsSources, info.operationItem, p.propertyDiff, "allOf", annotationOnlyDeleted[0].Index, -1)
				result = append(result, p.newChange(ResponsePropertyAllOfRemovedAnnotationOnlyId, []any{annotationOnlyDeleted.String(), propName, info.responseStatus}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
